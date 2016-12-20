// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/go-account-lib/rights"
	"github.com/TheThingsNetwork/ttn/api"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb "github.com/TheThingsNetwork/ttn/api/handler"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/api/ratelimit"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type handlerManager struct {
	handler         *handler
	deviceManager   pb_lorawan.DeviceManagerClient
	devAddrManager  pb_lorawan.DevAddrManagerClient
	applicationRate *ratelimit.Registry
	clientRate      *ratelimit.Registry
}

func (h *handlerManager) validateTTNAuthAppContext(ctx context.Context, appID string) (context.Context, *claims.Claims, error) {
	md, err := api.MetadataFromContext(ctx)
	if err != nil {
		return ctx, nil, err
	}
	// If token is empty, try to get the access key and convert it into a token
	token, err := api.TokenFromMetadata(md)
	if err != nil || token == "" {
		key, err := api.KeyFromMetadata(md)
		if err != nil {
			return ctx, nil, errors.NewErrInvalidArgument("Metadata", "neither token nor key present")
		}
		token, err := h.handler.Component.ExchangeAppKeyForToken(appID, key)
		if err != nil {
			return ctx, nil, err
		}
		md = metadata.Join(md, metadata.Pairs("token", token))
		ctx = metadata.NewContext(ctx, md)
	}
	claims, err := h.handler.Component.ValidateTTNAuthContext(ctx)
	if err != nil {
		return ctx, nil, err
	}
	if h.clientRate.Limit(claims.Subject) {
		return ctx, claims, grpc.Errorf(codes.ResourceExhausted, "Rate limit for client reached")
	}
	if h.applicationRate.Limit(appID) {
		return ctx, claims, grpc.Errorf(codes.ResourceExhausted, "Rate limit for application reached")
	}
	return ctx, claims, nil
}

func (h *handlerManager) GetDevice(ctx context.Context, in *pb.DeviceIdentifier) (*pb.Device, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Device Identifier")
	}

	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppId)
	if err != nil {
		return nil, err
	}
	if !claims.AppRight(in.AppId, rights.Devices) {
		return nil, errors.NewErrPermissionDenied(fmt.Sprintf(`No "devices" rights to application "%s"`, in.AppId))
	}

	if _, err := h.handler.applications.Get(in.AppId); err != nil {
		return nil, errors.Wrap(err, "Application not registered to this Handler")
	}

	dev, err := h.handler.devices.Get(in.AppId, in.DevId)
	if err != nil {
		return nil, err
	}

	pbDev := &pb.Device{
		AppId: dev.AppID,
		DevId: dev.DevID,
		Device: &pb.Device_LorawanDevice{LorawanDevice: &pb_lorawan.Device{
			AppId:                 dev.AppID,
			AppEui:                &dev.AppEUI,
			DevId:                 dev.DevID,
			DevEui:                &dev.DevEUI,
			DevAddr:               &dev.DevAddr,
			NwkSKey:               &dev.NwkSKey,
			AppSKey:               &dev.AppSKey,
			AppKey:                &dev.AppKey,
			DisableFCntCheck:      dev.Options.DisableFCntCheck,
			Uses32BitFCnt:         dev.Options.Uses32BitFCnt,
			ActivationConstraints: dev.Options.ActivationConstraints,
		}},
	}

	nsDev, err := h.deviceManager.GetDevice(ctx, &pb_lorawan.DeviceIdentifier{
		AppEui: &dev.AppEUI,
		DevEui: &dev.DevEUI,
	})
	if errors.GetErrType(errors.FromGRPCError(err)) == errors.NotFound {
		// Re-register the device in the Broker (NetworkServer)
		h.handler.Ctx.WithFields(log.Fields{
			"AppID":  dev.AppID,
			"DevID":  dev.DevID,
			"AppEUI": dev.AppEUI,
			"DevEUI": dev.DevEUI,
		}).Warn("Re-registering missing device to Broker")
		nsDev = dev.GetLoRaWAN()
		_, err = h.deviceManager.SetDevice(ctx, nsDev)
		if err != nil {
			return nil, errors.Wrap(errors.FromGRPCError(err), "Could not re-register missing device to Broker")
		}
	} else if err != nil {
		return pbDev, errors.Wrap(errors.FromGRPCError(err), "Broker did not return device")
	}

	pbDev.GetLorawanDevice().FCntUp = nsDev.FCntUp
	pbDev.GetLorawanDevice().FCntDown = nsDev.FCntDown
	pbDev.GetLorawanDevice().LastSeen = nsDev.LastSeen

	return pbDev, nil
}

func (h *handlerManager) SetDevice(ctx context.Context, in *pb.Device) (*empty.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Device")
	}

	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppId)
	if err != nil {
		return nil, err
	}
	if !claims.AppRight(in.AppId, rights.Devices) {
		return nil, errors.NewErrPermissionDenied(fmt.Sprintf(`No "devices" rights to application "%s"`, in.AppId))
	}

	if _, err := h.handler.applications.Get(in.AppId); err != nil {
		return nil, errors.Wrap(err, "Application not registered to this Handler")
	}

	dev, err := h.handler.devices.Get(in.AppId, in.DevId)
	if err != nil && errors.GetErrType(err) != errors.NotFound {
		return nil, err
	}

	lorawan := in.GetLorawanDevice()
	if lorawan == nil {
		return nil, errors.NewErrInvalidArgument("Device", "No LoRaWAN Device")
	}

	if dev != nil { // When this is an update
		if dev.AppEUI != *lorawan.AppEui || dev.DevEUI != *lorawan.DevEui {
			// If the AppEUI or DevEUI is changed, we should remove the device from the NetworkServer and re-add it later
			_, err = h.deviceManager.DeleteDevice(ctx, &pb_lorawan.DeviceIdentifier{
				AppEui: &dev.AppEUI,
				DevEui: &dev.DevEUI,
			})
			if err != nil {
				return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not delete device")
			}
		}
		dev.StartUpdate()
	} else { // When this is a create
		existingDevices, err := h.handler.devices.ListForApp(in.AppId)
		if err != nil {
			return nil, err
		}
		for _, existingDevice := range existingDevices {
			if existingDevice.AppEUI == *lorawan.AppEui && existingDevice.DevEUI == *lorawan.DevEui {
				return nil, errors.NewErrAlreadyExists("Device with AppEUI and DevEUI")
			}
		}
		dev = new(device.Device)
	}

	dev.AppID = in.AppId
	dev.AppEUI = *lorawan.AppEui
	dev.DevID = in.DevId
	dev.DevEUI = *lorawan.DevEui

	dev.Options = device.Options{
		DisableFCntCheck:      lorawan.DisableFCntCheck,
		Uses32BitFCnt:         lorawan.Uses32BitFCnt,
		ActivationConstraints: lorawan.ActivationConstraints,
	}
	if dev.Options.ActivationConstraints == "" {
		dev.Options.ActivationConstraints = "local"
	}

	if lorawan.DevAddr != nil {
		dev.DevAddr = *lorawan.DevAddr
	}
	if lorawan.NwkSKey != nil {
		dev.NwkSKey = *lorawan.NwkSKey
	}
	if lorawan.AppSKey != nil {
		dev.AppSKey = *lorawan.AppSKey
	}

	if lorawan.AppKey != nil {
		if dev.AppKey != *lorawan.AppKey { // When the AppKey of an existing device is changed
			dev.UsedAppNonces = []device.AppNonce{}
			dev.UsedDevNonces = []device.DevNonce{}
		}
		dev.AppKey = *lorawan.AppKey
	}

	// Update the device in the Broker (NetworkServer)
	nsUpdated := dev.GetLoRaWAN()
	nsUpdated.FCntUp = lorawan.FCntUp
	nsUpdated.FCntDown = lorawan.FCntDown

	_, err = h.deviceManager.SetDevice(ctx, nsUpdated)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not set device")
	}

	err = h.handler.devices.Set(dev)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (h *handlerManager) DeleteDevice(ctx context.Context, in *pb.DeviceIdentifier) (*empty.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Device Identifier")
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppId)
	if err != nil {
		return nil, err
	}
	if !claims.AppRight(in.AppId, rights.Devices) {
		return nil, errors.NewErrPermissionDenied(fmt.Sprintf(`No "devices" rights to application "%s"`, in.AppId))
	}

	if _, err := h.handler.applications.Get(in.AppId); err != nil {
		return nil, errors.Wrap(err, "Application not registered to this Handler")
	}

	dev, err := h.handler.devices.Get(in.AppId, in.DevId)
	if err != nil {
		return nil, err
	}
	_, err = h.deviceManager.DeleteDevice(ctx, &pb_lorawan.DeviceIdentifier{AppEui: &dev.AppEUI, DevEui: &dev.DevEUI})
	if err != nil && errors.GetErrType(errors.FromGRPCError(err)) != errors.NotFound {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not delete device")
	}
	err = h.handler.devices.Delete(in.AppId, in.DevId)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (h *handlerManager) GetDevicesForApplication(ctx context.Context, in *pb.ApplicationIdentifier) (*pb.DeviceList, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Application Identifier")
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppId)
	if err != nil {
		return nil, err
	}
	if !claims.AppRight(in.AppId, rights.Devices) {
		return nil, errors.NewErrPermissionDenied(fmt.Sprintf(`No "devices" rights to application "%s"`, in.AppId))
	}

	if _, err := h.handler.applications.Get(in.AppId); err != nil {
		return nil, errors.Wrap(err, "Application not registered to this Handler")
	}

	devices, err := h.handler.devices.ListForApp(in.AppId)
	if err != nil {
		return nil, err
	}
	res := &pb.DeviceList{Devices: []*pb.Device{}}
	for _, dev := range devices {
		res.Devices = append(res.Devices, &pb.Device{
			AppId: dev.AppID,
			DevId: dev.DevID,
			Device: &pb.Device_LorawanDevice{LorawanDevice: &pb_lorawan.Device{
				AppId:   dev.AppID,
				AppEui:  &dev.AppEUI,
				DevId:   dev.DevID,
				DevEui:  &dev.DevEUI,
				DevAddr: &dev.DevAddr,
				NwkSKey: &dev.NwkSKey,
				AppSKey: &dev.AppSKey,
				AppKey:  &dev.AppKey,
			}},
		})
	}
	return res, nil
}

func (h *handlerManager) GetApplication(ctx context.Context, in *pb.ApplicationIdentifier) (*pb.Application, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.NewErrInvalidArgument("Application Identifier", err.Error())
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppId)
	if err != nil {
		return nil, err
	}
	if !claims.AppRight(in.AppId, rights.AppSettings) {
		return nil, errors.NewErrPermissionDenied(`No "settings" rights to application`)
	}
	app, err := h.handler.applications.Get(in.AppId)
	if err != nil {
		return nil, err
	}

	return &pb.Application{
		AppId:     app.AppID,
		Decoder:   app.Decoder,
		Converter: app.Converter,
		Validator: app.Validator,
		Encoder:   app.Encoder,
	}, nil
}

func (h *handlerManager) RegisterApplication(ctx context.Context, in *pb.ApplicationIdentifier) (*empty.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Application Identifier")
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppId)
	if err != nil {
		return nil, err
	}
	if !claims.AppRight(in.AppId, rights.AppSettings) {
		return nil, errors.NewErrPermissionDenied(`No "settings" rights to application`)
	}
	app, err := h.handler.applications.Get(in.AppId)
	if err != nil && errors.GetErrType(err) != errors.NotFound {
		return nil, err
	}
	if app != nil {
		return nil, errors.NewErrAlreadyExists("Application")
	}

	err = h.handler.applications.Set(&application.Application{
		AppID: in.AppId,
	})
	if err != nil {
		return nil, err
	}

	token, _ := api.TokenFromContext(ctx)
	err = h.handler.Discovery.AddAppID(in.AppId, token)
	if err != nil {
		h.handler.Ctx.WithField("AppID", in.AppId).WithError(err).Warn("Could not register Application with Discovery")
	}

	_, err = h.handler.ttnBrokerManager.RegisterApplicationHandler(ctx, &pb_broker.ApplicationHandlerRegistration{
		AppId:     in.AppId,
		HandlerId: h.handler.Identity.Id,
	})
	if err != nil {
		h.handler.Ctx.WithField("AppID", in.AppId).WithError(err).Warn("Could not register Application with Broker")
	}

	return &empty.Empty{}, nil

}

func (h *handlerManager) SetApplication(ctx context.Context, in *pb.Application) (*empty.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Application")
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppId)
	if err != nil {
		return nil, err
	}
	if !claims.AppRight(in.AppId, rights.AppSettings) {
		return nil, errors.NewErrPermissionDenied(`No "settings" rights to application`)
	}
	app, err := h.handler.applications.Get(in.AppId)
	if err != nil {
		return nil, err
	}

	app.StartUpdate()

	app.Decoder = in.Decoder
	app.Converter = in.Converter
	app.Validator = in.Validator
	app.Encoder = in.Encoder

	err = h.handler.applications.Set(app)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (h *handlerManager) DeleteApplication(ctx context.Context, in *pb.ApplicationIdentifier) (*empty.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Application Identifier")
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppId)
	if err != nil {
		return nil, err
	}
	if !claims.AppRight(in.AppId, rights.AppSettings) {
		return nil, errors.NewErrPermissionDenied(`No "settings" rights to application`)
	}
	_, err = h.handler.applications.Get(in.AppId)
	if err != nil {
		return nil, err
	}

	// Get and delete all devices for this application
	devices, err := h.handler.devices.ListForApp(in.AppId)
	if err != nil {
		return nil, err
	}
	for _, dev := range devices {
		_, err = h.deviceManager.DeleteDevice(ctx, &pb_lorawan.DeviceIdentifier{AppEui: &dev.AppEUI, DevEui: &dev.DevEUI})
		if err != nil {
			return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not delete device")
		}
		err = h.handler.devices.Delete(dev.AppID, dev.DevID)
		if err != nil {
			return nil, err
		}
	}

	// Delete the Application
	err = h.handler.applications.Delete(in.AppId)
	if err != nil {
		return nil, err
	}

	token, _ := api.TokenFromContext(ctx)
	err = h.handler.Discovery.RemoveAppID(in.AppId, token)
	if err != nil {
		h.handler.Ctx.WithField("AppID", in.AppId).WithError(errors.FromGRPCError(err)).Warn("Could not unregister Application from Discovery")
	}

	return &empty.Empty{}, nil
}

func (h *handlerManager) GetPrefixes(ctx context.Context, in *pb_lorawan.PrefixesRequest) (*pb_lorawan.PrefixesResponse, error) {
	res, err := h.devAddrManager.GetPrefixes(ctx, in)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not return prefixes")
	}
	return res, nil
}

func (h *handlerManager) GetDevAddr(ctx context.Context, in *pb_lorawan.DevAddrRequest) (*pb_lorawan.DevAddrResponse, error) {
	res, err := h.devAddrManager.GetDevAddr(ctx, in)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not return DevAddr")
	}
	return res, nil
}

func (h *handlerManager) GetStatus(ctx context.Context, in *pb.StatusRequest) (*pb.Status, error) {
	if h.handler.Identity.Id != "dev" {
		claims, err := h.handler.ValidateTTNAuthContext(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "No access")
		}
		if !claims.ComponentAccess(h.handler.Identity.Id) {
			return nil, errors.NewErrPermissionDenied(fmt.Sprintf("Claims do not grant access to %s", h.handler.Identity.Id))
		}
	}
	status := h.handler.GetStatus()
	if status == nil {
		return new(pb.Status), nil
	}
	return status, nil
}

func (h *handler) RegisterManager(s *grpc.Server) {
	server := &handlerManager{
		handler:        h,
		deviceManager:  pb_lorawan.NewDeviceManagerClient(h.ttnBrokerConn),
		devAddrManager: pb_lorawan.NewDevAddrManagerClient(h.ttnBrokerConn),
	}

	server.applicationRate = ratelimit.NewRegistry(5000, time.Hour)
	server.clientRate = ratelimit.NewRegistry(5000, time.Hour)

	pb.RegisterHandlerManagerServer(s, server)
	pb.RegisterApplicationManagerServer(s, server)
	pb_lorawan.RegisterDevAddrManagerServer(s, server)
}

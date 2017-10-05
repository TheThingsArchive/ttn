// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_handler "github.com/TheThingsNetwork/api/handler"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/go-account-lib/rights"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/ratelimit"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	gogo "github.com/gogo/protobuf/types"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type handlerManager struct {
	handler         *handler
	devAddrManager  pb_lorawan.DevAddrManagerClient
	applicationRate *ratelimit.Registry
	clientRate      *ratelimit.Registry
}

func checkAppRights(claims *claims.Claims, appID string, right types.Right) error {
	if !claims.AppRight(appID, right) {
		return errors.NewErrPermissionDenied(fmt.Sprintf(`No "%s" rights to Application "%s"`, right, appID))
	}
	return nil
}

func (h *handlerManager) validateTTNAuthAppContext(ctx context.Context, appID string) (context.Context, *claims.Claims, error) {
	md := ttnctx.MetadataFromIncomingContext(ctx)

	// If token is empty, try to get the access key and convert it into a token
	token, err := ttnctx.TokenFromMetadata(md)
	if err != nil || token == "" {
		key, err := ttnctx.KeyFromMetadata(md)
		if err != nil {
			return ctx, nil, errors.NewErrInvalidArgument("Metadata", "neither token nor key present")
		}
		token, err := h.handler.Component.ExchangeAppKeyForToken(appID, key)
		if err != nil {
			return ctx, nil, err
		}
		md = metadata.Join(md, metadata.Pairs("token", token))
		ctx = metadata.NewIncomingContext(ctx, md)
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

func (h *handlerManager) GetDevice(ctx context.Context, in *pb_handler.DeviceIdentifier) (*pb_handler.Device, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Device Identifier")
	}

	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	token, _ := ttnctx.TokenFromIncomingContext(ctx)
	err = checkAppRights(claims, in.AppID, rights.Devices)
	if err != nil {
		return nil, err
	}

	if _, err := h.handler.applications.Get(in.AppID); err != nil {
		return nil, errors.Wrap(err, "Application not registered to this Handler")
	}

	dev, err := h.handler.devices.Get(in.AppID, in.DevID)
	if err != nil {
		return nil, err
	}

	pbDev := dev.ToPb()

	nsDev, err := h.handler.ttnDeviceManager.GetDevice(ttnctx.OutgoingContextWithToken(ctx, token), &pb_lorawan.DeviceIdentifier{
		AppEUI: dev.AppEUI,
		DevEUI: dev.DevEUI,
	})
	if errors.GetErrType(errors.FromGRPCError(err)) == errors.NotFound {
		// Re-register the device in the Broker (NetworkServer)
		h.handler.Ctx.WithFields(ttnlog.Fields{
			"AppID":  dev.AppID,
			"DevID":  dev.DevID,
			"AppEUI": dev.AppEUI,
			"DevEUI": dev.DevEUI,
		}).Warn("Re-registering missing device to Broker")
		nsDev = dev.ToLoRaWANPb()
		nsDev.AppKey = nil
		nsDev.AppSKey = nil
		_, err = h.handler.ttnDeviceManager.SetDevice(ttnctx.OutgoingContextWithToken(ctx, token), nsDev)
		if err != nil {
			return nil, errors.Wrap(errors.FromGRPCError(err), "Could not re-register missing device to Broker")
		}
	} else if err != nil {
		return pbDev, errors.Wrap(errors.FromGRPCError(err), "Broker did not return device")
	}

	pbDev.GetLoRaWANDevice().FCntUp = nsDev.FCntUp
	pbDev.GetLoRaWANDevice().FCntDown = nsDev.FCntDown
	pbDev.GetLoRaWANDevice().LastSeen = nsDev.LastSeen

	return pbDev, nil
}

func (h *handlerManager) SetDevice(ctx context.Context, in *pb_handler.Device) (*gogo.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Device")
	}

	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	token, _ := ttnctx.TokenFromIncomingContext(ctx)
	err = checkAppRights(claims, in.AppID, rights.Devices)
	if err != nil {
		return nil, err
	}

	if _, err := h.handler.applications.Get(in.AppID); err != nil {
		return nil, errors.Wrap(err, "Application not registered to this Handler")
	}

	dev, err := h.handler.devices.Get(in.AppID, in.DevID)
	if err != nil && errors.GetErrType(err) != errors.NotFound {
		return nil, err
	}

	lorawan := in.GetLoRaWANDevice()
	if lorawan == nil {
		return nil, errors.NewErrInvalidArgument("Device", "No LoRaWAN Device")
	}

	var eventType types.EventType
	if dev != nil {
		eventType = types.UpdateEvent

		// Not allowed to update join nonces after device is created
		lorawan.UsedDevNonces, lorawan.UsedAppNonces = nil, nil

		// If the AppEUI or DevEUI is changed, we should remove the device from the NetworkServer and re-add it later
		if dev.AppEUI != lorawan.AppEUI || dev.DevEUI != lorawan.DevEUI {
			_, err = h.handler.ttnDeviceManager.DeleteDevice(ttnctx.OutgoingContextWithToken(ctx, token), &pb_lorawan.DeviceIdentifier{
				AppEUI: dev.AppEUI,
				DevEUI: dev.DevEUI,
			})
			if err != nil {
				return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not delete device")
			}
		}

		dev.StartUpdate()
	} else {
		eventType = types.CreateEvent
		existingDevices, err := h.handler.devices.ListForApp(in.AppID, nil)
		if err != nil {
			return nil, err
		}
		for _, existingDevice := range existingDevices {
			if existingDevice.AppEUI == lorawan.AppEUI && existingDevice.DevEUI == lorawan.DevEUI {
				return nil, errors.NewErrAlreadyExists("Device with AppEUI and DevEUI")
			}
		}
		dev = new(device.Device)
	}

	if lorawan.AppKey != nil && dev.AppKey != *lorawan.AppKey {
		// Reset join nonces when AppKey changes
		dev.UsedDevNonces, dev.UsedAppNonces = []device.DevNonce{}, []device.AppNonce{}
	}

	dev.FromPb(in)

	if dev.Options.ActivationConstraints == "" {
		dev.Options.ActivationConstraints = "local"
	}

	// Update the device in the Broker (NetworkServer)
	lorawanPb := dev.ToLoRaWANPb()
	lorawanPb.AppKey = nil
	lorawanPb.AppSKey = nil
	lorawanPb.UsedDevNonces = nil
	lorawanPb.UsedAppNonces = nil
	lorawanPb.FCntUp = lorawan.FCntUp
	lorawanPb.FCntDown = lorawan.FCntDown

	_, err = h.handler.ttnDeviceManager.SetDevice(ttnctx.OutgoingContextWithToken(ctx, token), lorawanPb)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not set device")
	}

	err = h.handler.devices.Set(dev)
	if err != nil {
		return nil, err
	}

	h.handler.qEvent <- &types.DeviceEvent{
		AppID: dev.AppID,
		DevID: dev.DevID,
		Event: eventType,
		Data:  nil, // Don't send potentially sensitive details over MQTT
	}

	return &gogo.Empty{}, nil
}

func (h *handlerManager) DeleteDevice(ctx context.Context, in *pb_handler.DeviceIdentifier) (*gogo.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Device Identifier")
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	token, _ := ttnctx.TokenFromIncomingContext(ctx)
	err = checkAppRights(claims, in.AppID, rights.Devices)
	if err != nil {
		return nil, err
	}

	if _, err := h.handler.applications.Get(in.AppID); err != nil {
		return nil, errors.Wrap(err, "Application not registered to this Handler")
	}

	dev, err := h.handler.devices.Get(in.AppID, in.DevID)
	if err != nil {
		return nil, err
	}
	_, err = h.handler.ttnDeviceManager.DeleteDevice(ttnctx.OutgoingContextWithToken(ctx, token), &pb_lorawan.DeviceIdentifier{AppEUI: dev.AppEUI, DevEUI: dev.DevEUI})
	if err != nil && errors.GetErrType(errors.FromGRPCError(err)) != errors.NotFound {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not delete device")
	}
	err = h.handler.devices.Delete(in.AppID, in.DevID)
	if err != nil {
		return nil, err
	}
	h.handler.qEvent <- &types.DeviceEvent{
		AppID: in.AppID,
		DevID: in.DevID,
		Event: types.DeleteEvent,
	}
	return &gogo.Empty{}, nil
}

func (h *handlerManager) GetDevicesForApplication(ctx context.Context, in *pb_handler.ApplicationIdentifier) (*pb_handler.DeviceList, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Application Identifier")
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	err = checkAppRights(claims, in.AppID, rights.Devices)
	if err != nil {
		return nil, err
	}

	if _, err := h.handler.applications.Get(in.AppID); err != nil {
		return nil, errors.Wrap(err, "Application not registered to this Handler")
	}

	limit, offset, err := ttnctx.LimitAndOffsetFromIncomingContext(ctx)
	if err != nil {
		return nil, err
	}

	opts := &storage.ListOptions{Limit: limit, Offset: offset}
	devices, err := h.handler.devices.ListForApp(in.AppID, opts)
	if err != nil {
		return nil, err
	}
	res := &pb_handler.DeviceList{Devices: []*pb_handler.Device{}}
	for _, dev := range devices {
		if dev == nil {
			continue
		}
		res.Devices = append(res.Devices, dev.ToPb())
	}

	total, selected := opts.GetTotalAndSelected()
	header := metadata.Pairs(
		"total", strconv.FormatUint(total, 10),
		"selected", strconv.FormatUint(selected, 10),
	)
	grpc.SendHeader(ctx, header)

	return res, nil
}

func (h *handlerManager) GetApplication(ctx context.Context, in *pb_handler.ApplicationIdentifier) (*pb_handler.Application, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.NewErrInvalidArgument("Application Identifier", err.Error())
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	err = checkAppRights(claims, in.AppID, rights.AppSettings)
	if err != nil {
		return nil, err
	}
	app, err := h.handler.applications.Get(in.AppID)
	if err != nil {
		return nil, err
	}

	res := &pb_handler.Application{
		AppID:         app.AppID,
		PayloadFormat: string(app.PayloadFormat),
		Decoder:       app.CustomDecoder,
		Converter:     app.CustomConverter,
		Validator:     app.CustomValidator,
		Encoder:       app.CustomEncoder,
	}
	if err := checkAppRights(claims, in.AppID, rights.Devices); err == nil {
		res.RegisterOnJoinAccessKey = app.RegisterOnJoinAccessKey
	} else if app.RegisterOnJoinAccessKey != "" {
		parts := strings.Split(app.RegisterOnJoinAccessKey, ".")
		if len(parts) == 2 {
			res.RegisterOnJoinAccessKey = parts[1] + "." + "<...>"
		} else {
			res.RegisterOnJoinAccessKey = "..."
		}
	}
	return res, nil
}

func (h *handlerManager) RegisterApplication(ctx context.Context, in *pb_handler.ApplicationIdentifier) (*gogo.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Application Identifier")
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	token, _ := ttnctx.TokenFromIncomingContext(ctx)
	err = checkAppRights(claims, in.AppID, rights.AppSettings)
	if err != nil {
		return nil, err
	}
	app, err := h.handler.applications.Get(in.AppID)
	if err != nil && errors.GetErrType(err) != errors.NotFound {
		return nil, err
	}
	if app != nil {
		return nil, errors.NewErrAlreadyExists("Application")
	}

	err = h.handler.applications.Set(&application.Application{
		AppID: in.AppID,
	})
	if err != nil {
		return nil, err
	}

	err = h.handler.Discovery.AddAppID(in.AppID, token)
	if err != nil {
		h.handler.Ctx.WithField("AppID", in.AppID).WithError(err).Warn("Could not register Application with Discovery")
	}

	_, err = h.handler.ttnBrokerManager.RegisterApplicationHandler(ttnctx.OutgoingContextWithToken(ctx, token), &pb_broker.ApplicationHandlerRegistration{
		AppID:     in.AppID,
		HandlerID: h.handler.Identity.ID,
	})
	if err != nil {
		h.handler.Ctx.WithField("AppID", in.AppID).WithError(err).Warn("Could not register Application with Broker")
	}

	return &gogo.Empty{}, nil

}

func (h *handlerManager) SetApplication(ctx context.Context, in *pb_handler.Application) (*gogo.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Application")
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	err = checkAppRights(claims, in.AppID, rights.AppSettings)
	if err != nil {
		return nil, err
	}
	app, err := h.handler.applications.Get(in.AppID)
	if err != nil {
		return nil, err
	}

	app.StartUpdate()

	app.PayloadFormat = application.PayloadFormat(in.PayloadFormat)
	app.CustomDecoder = in.Decoder
	app.CustomConverter = in.Converter
	app.CustomValidator = in.Validator
	app.CustomEncoder = in.Encoder

	if err := checkAppRights(claims, in.AppID, rights.Devices); err == nil {
		app.RegisterOnJoinAccessKey = in.RegisterOnJoinAccessKey
	}

	if app.PayloadFormat == "" && (app.CustomDecoder != "" || app.CustomConverter != "" || app.CustomValidator != "" || app.CustomEncoder != "") {
		app.PayloadFormat = application.PayloadFormatCustom
	}

	err = h.handler.applications.Set(app)
	if err != nil {
		return nil, err
	}

	return &gogo.Empty{}, nil
}

func (h *handlerManager) DeleteApplication(ctx context.Context, in *pb_handler.ApplicationIdentifier) (*gogo.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Application Identifier")
	}
	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	token, _ := ttnctx.TokenFromIncomingContext(ctx)
	err = checkAppRights(claims, in.AppID, rights.AppDelete)
	if err != nil {
		return nil, err
	}

	_, err = h.handler.applications.Get(in.AppID)
	if err != nil {
		return nil, err
	}

	// Get and delete all devices for this application
	devices, err := h.handler.devices.ListForApp(in.AppID, nil)
	if err != nil {
		return nil, err
	}
	for _, dev := range devices {
		_, err = h.handler.ttnDeviceManager.DeleteDevice(ttnctx.OutgoingContextWithToken(ctx, token), &pb_lorawan.DeviceIdentifier{AppEUI: dev.AppEUI, DevEUI: dev.DevEUI})
		if err != nil {
			return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not delete device")
		}
		err = h.handler.devices.Delete(dev.AppID, dev.DevID)
		if err != nil {
			return nil, err
		}
	}

	// Delete the Application
	err = h.handler.applications.Delete(in.AppID)
	if err != nil {
		return nil, err
	}

	err = h.handler.Discovery.RemoveAppID(in.AppID, token)
	if err != nil {
		h.handler.Ctx.WithField("AppID", in.AppID).WithError(errors.FromGRPCError(err)).Warn("Could not unregister Application from Discovery")
	}

	return &gogo.Empty{}, nil
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

func (h *handlerManager) GetStatus(ctx context.Context, in *pb_handler.StatusRequest) (*pb_handler.Status, error) {
	if h.handler.Identity.ID != "dev" {
		claims, err := h.handler.ValidateTTNAuthContext(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "No access")
		}
		if !claims.ComponentAccess(h.handler.Identity.ID) {
			return nil, errors.NewErrPermissionDenied(fmt.Sprintf("Claims do not grant access to %s", h.handler.Identity.ID))
		}
	}
	status := h.handler.GetStatus()
	if status == nil {
		return new(pb_handler.Status), nil
	}
	return status, nil
}

func (h *handler) RegisterManager(s *grpc.Server) {
	server := &handlerManager{
		handler:        h,
		devAddrManager: pb_lorawan.NewDevAddrManagerClient(h.ttnBrokerConn),
	}

	server.applicationRate = ratelimit.NewRegistry(5000, time.Hour)
	server.clientRate = ratelimit.NewRegistry(5000, time.Hour)

	pb_handler.RegisterHandlerManagerServer(s, server)
	pb_handler.RegisterApplicationManagerServer(s, server)
	pb_lorawan.RegisterDevAddrManagerServer(s, server)
}

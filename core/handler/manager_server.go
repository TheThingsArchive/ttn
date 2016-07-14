// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"errors"

	"github.com/TheThingsNetwork/ttn/api"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb "github.com/TheThingsNetwork/ttn/api/handler"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type handlerManager struct {
	*handler
}

var errf = grpc.Errorf

func (h *handlerManager) getDevice(ctx context.Context, in *pb.DeviceIdentifier) (*device.Device, error) {
	if !in.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Device Identifier")
	}
	claims, err := h.Component.ValidateContext(ctx)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(in.AppId) {
		return nil, errf(codes.Unauthenticated, "No access to this device")
	}
	dev, err := h.devices.Get(in.AppId, in.DevId)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(dev.AppID) {
		return nil, errf(codes.Unauthenticated, "No access to this device")
	}
	return dev, nil
}

func (h *handlerManager) GetDevice(ctx context.Context, in *pb.DeviceIdentifier) (*pb.Device, error) {
	dev, err := h.getDevice(ctx, in)
	if err != nil {
		return nil, err
	}

	nsDev, err := h.ttnDeviceManager.GetDevice(ctx, &pb_lorawan.DeviceIdentifier{
		AppEui: &dev.AppEUI,
		DevEui: &dev.DevEUI,
	})
	if err != nil {
		return nil, err
	}

	return &pb.Device{
		AppId: dev.AppID,
		DevId: dev.DevID,
		Device: &pb.Device_LorawanDevice{LorawanDevice: &pb_lorawan.Device{
			AppId:            dev.AppID,
			AppEui:           nsDev.AppEui,
			DevId:            dev.DevID,
			DevEui:           nsDev.DevEui,
			DevAddr:          nsDev.DevAddr,
			NwkSKey:          nsDev.NwkSKey,
			AppSKey:          &dev.AppSKey,
			AppKey:           &dev.AppKey,
			FCntUp:           nsDev.FCntUp,
			FCntDown:         nsDev.FCntDown,
			DisableFCntCheck: nsDev.DisableFCntCheck,
			Uses32BitFCnt:    nsDev.Uses32BitFCnt,
			LastSeen:         nsDev.LastSeen,
		}},
	}, nil
}

func (h *handlerManager) SetDevice(ctx context.Context, in *pb.Device) (*api.Ack, error) {
	_, err := h.getDevice(ctx, &pb.DeviceIdentifier{AppId: in.AppId, DevId: in.DevId})
	if err != nil && err != device.ErrNotFound {
		return nil, err
	}

	if !in.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Device")
	}

	updated := &device.Device{
		AppID: in.AppId,
		DevID: in.DevId,
	}

	lorawan := in.GetLorawanDevice()
	if lorawan == nil {
		err = grpcErrf(codes.InvalidArgument, "No LoRaWAN Device")
	}

	_, err = h.ttnDeviceManager.SetDevice(ctx, &pb_lorawan.Device{
		AppId:            in.AppId,
		AppEui:           lorawan.AppEui,
		DevEui:           lorawan.DevEui,
		DevAddr:          lorawan.DevAddr,
		NwkSKey:          lorawan.NwkSKey,
		FCntUp:           lorawan.FCntUp,
		FCntDown:         lorawan.FCntDown,
		DisableFCntCheck: lorawan.DisableFCntCheck,
		Uses32BitFCnt:    lorawan.Uses32BitFCnt,
	})
	if err != nil {
		return nil, err
	}

	updated.AppEUI = *lorawan.AppEui
	updated.DevEUI = *lorawan.DevEui

	if lorawan.DevAddr != nil && lorawan.NwkSKey != nil && lorawan.AppSKey != nil {
		updated.DevAddr = *lorawan.DevAddr
		updated.NwkSKey = *lorawan.NwkSKey
		updated.AppSKey = *lorawan.AppSKey
	}

	if lorawan.AppKey != nil {
		updated.AppKey = *lorawan.AppKey
	}

	err = h.devices.Set(updated)
	if err != nil {
		return nil, err
	}

	return &api.Ack{}, nil
}

func (h *handlerManager) DeleteDevice(ctx context.Context, in *pb.DeviceIdentifier) (*api.Ack, error) {
	dev, err := h.getDevice(ctx, in)
	if err != nil {
		return nil, err
	}
	_, err = h.ttnDeviceManager.DeleteDevice(ctx, &pb_lorawan.DeviceIdentifier{AppEui: &dev.AppEUI, DevEui: &dev.DevEUI})
	if err != nil {
		return nil, err
	}
	err = h.devices.Delete(in.AppId, in.DevId)
	if err != nil {
		return nil, err
	}
	return &api.Ack{}, nil
}

func (h *handlerManager) GetDevicesForApplication(ctx context.Context, in *pb.ApplicationIdentifier) (*pb.DeviceList, error) {
	if !in.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Application Identifier")
	}
	claims, err := h.Component.ValidateContext(ctx)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(in.AppId) {
		return nil, errf(codes.Unauthenticated, "No access to this application")
	}
	devices, err := h.devices.ListForApp(in.AppId)
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

func (h *handlerManager) getApplication(ctx context.Context, in *pb.ApplicationIdentifier) (*application.Application, error) {
	if !in.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Application Identifier")
	}
	claims, err := h.Component.ValidateContext(ctx)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(in.AppId) {
		return nil, errf(codes.Unauthenticated, "No access to this application")
	}
	app, err := h.applications.Get(in.AppId)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(app.AppID) {
		return nil, errf(codes.Unauthenticated, "No access to this application")
	}
	return app, nil
}

func (h *handlerManager) GetApplication(ctx context.Context, in *pb.ApplicationIdentifier) (*pb.Application, error) {
	app, err := h.getApplication(ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb.Application{
		AppId:     app.AppID,
		Decoder:   app.Decoder,
		Converter: app.Converter,
		Validator: app.Validator,
	}, nil
}

func (h *handlerManager) SetApplication(ctx context.Context, in *pb.Application) (*api.Ack, error) {
	app, err := h.getApplication(ctx, &pb.ApplicationIdentifier{AppId: in.AppId})
	if err != nil && err != application.ErrNotFound {
		return nil, err
	}

	if !in.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Application")
	}

	err = h.applications.Set(&application.Application{
		AppID:     in.AppId,
		Decoder:   in.Decoder,
		Converter: in.Converter,
		Validator: in.Validator,
	})
	if err != nil {
		return nil, err
	}

	if app == nil {
		// Add this application ID to the cache if needed
		h.applicationIdsLock.Lock()
		var alreadyInCache bool
		for _, id := range h.applicationIds {
			if id == in.AppId {
				alreadyInCache = true
				break
			}
		}
		if !alreadyInCache {
			h.applicationIds = append(h.applicationIds, in.AppId)
		}
		h.applicationIdsLock.Unlock()

		// If we had to add it, we also have to announce it to the Discovery and Broker
		if !alreadyInCache {
			h.announce()
			_, err := h.ttnBrokerManager.RegisterApplicationHandler(ctx, &pb_broker.ApplicationHandlerRegistration{
				AppId:     in.AppId,
				HandlerId: h.Identity.Id,
			})
			if err != nil {
				h.Ctx.WithField("AppID", in.AppId).WithError(err).Warn("Could not register Application with Broker")
			}
		}
	}

	return &api.Ack{}, nil
}

func (h *handlerManager) DeleteApplication(ctx context.Context, in *pb.ApplicationIdentifier) (*api.Ack, error) {
	_, err := h.getApplication(ctx, in)
	if err != nil {
		return nil, err
	}

	err = h.applications.Delete(in.AppId)
	if err != nil {
		return nil, err
	}

	// Remove this application ID from the cache
	h.applicationIdsLock.Lock()
	newApplicationIDList := make([]string, 0, len(h.applicationIds))
	for _, id := range h.applicationIds {
		if id != in.AppId {
			newApplicationIDList = append(newApplicationIDList, id)
		}
	}
	h.applicationIds = newApplicationIDList
	h.applicationIdsLock.Unlock()

	return &api.Ack{}, nil
}

func (h *handlerManager) GetStatus(ctx context.Context, in *pb.StatusRequest) (*pb.Status, error) {
	return nil, errors.New("Not Implemented")
}

func (b *handler) RegisterManager(s *grpc.Server) {
	server := &handlerManager{b}
	pb.RegisterHandlerManagerServer(s, server)
	pb.RegisterApplicationManagerServer(s, server)
}

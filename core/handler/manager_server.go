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

func (h *handlerManager) getDevice(ctx context.Context, in *pb_lorawan.DeviceIdentifier) (*device.Device, error) {
	if in.AppId == "" || in.AppEui == nil || in.DevEui == nil {
		return nil, errf(codes.InvalidArgument, "AppID, AppEUI and DevEUI are required")
	}
	claims, err := h.Component.ValidateContext(ctx)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(in.AppId) {
		return nil, errf(codes.Unauthenticated, "No access to this device")
	}
	dev, err := h.devices.Get(*in.AppEui, *in.DevEui)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(dev.AppID) {
		return nil, errf(codes.Unauthenticated, "No access to this device")
	}
	return dev, nil
}

func (h *handlerManager) GetDevice(ctx context.Context, in *pb_lorawan.DeviceIdentifier) (*pb_lorawan.Device, error) {
	dev, err := h.getDevice(ctx, in)
	if err != nil {
		return nil, err
	}

	nsDev, err := h.ttnDeviceManager.GetDevice(ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb_lorawan.Device{
		AppId:            dev.AppID,
		AppEui:           nsDev.AppEui,
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
	}, nil
}

func (h *handlerManager) SetDevice(ctx context.Context, in *pb_lorawan.Device) (*api.Ack, error) {
	_, err := h.getDevice(ctx, &pb_lorawan.DeviceIdentifier{AppId: in.AppId, AppEui: in.AppEui, DevEui: in.DevEui})
	if err != nil && err != device.ErrNotFound {
		return nil, err
	}

	_, err = h.ttnDeviceManager.SetDevice(ctx, &pb_lorawan.Device{
		AppId:            in.AppId,
		AppEui:           in.AppEui,
		DevEui:           in.DevEui,
		DevAddr:          in.DevAddr,
		NwkSKey:          in.NwkSKey,
		FCntUp:           in.FCntUp,
		FCntDown:         in.FCntDown,
		DisableFCntCheck: in.DisableFCntCheck,
		Uses32BitFCnt:    in.Uses32BitFCnt,
	})
	if err != nil {
		return nil, err
	}

	updated := &device.Device{
		AppID:  in.AppId,
		AppEUI: *in.AppEui,
		DevEUI: *in.DevEui,
	}

	if in.DevAddr != nil && in.NwkSKey != nil && in.AppSKey != nil {
		updated.DevAddr = *in.DevAddr
		updated.NwkSKey = *in.NwkSKey
		updated.AppSKey = *in.AppSKey
	}

	if in.AppKey != nil {
		updated.AppKey = *in.AppKey
	}

	err = h.devices.Set(updated)
	if err != nil {
		return nil, err
	}

	return &api.Ack{}, nil
}

func (h *handlerManager) DeleteDevice(ctx context.Context, in *pb_lorawan.DeviceIdentifier) (*api.Ack, error) {
	_, err := h.getDevice(ctx, in)
	if err != nil {
		return nil, err
	}
	_, err = h.ttnDeviceManager.DeleteDevice(ctx, in)
	if err != nil {
		return nil, err
	}
	err = h.devices.Delete(*in.AppEui, *in.DevEui)
	if err != nil {
		return nil, err
	}
	return &api.Ack{}, nil
}

func (h *handlerManager) getApplication(ctx context.Context, in *pb.ApplicationIdentifier) (*application.Application, error) {
	if in.AppId == "" {
		return nil, errf(codes.InvalidArgument, "AppID is required")
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
	pb_lorawan.RegisterDeviceManagerServer(s, server)
}

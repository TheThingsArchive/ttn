// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context"
)

// ListDevices implements the core.HandlerManagerServer interface
func (h component) ListDevices(bctx context.Context, req *core.ListDevicesHandlerReq) (*core.ListDevicesHandlerRes, error) {
	h.Ctx.Debug("Handle list devices request")

	// 1. Validate the request
	if len(req.AppEUI) != 8 {
		err := errors.New(errors.Structural, "Invalid request parameters")
		h.Ctx.WithError(err).Debug("Unable to handle list devices request")
		return new(core.ListDevicesHandlerRes), err
	}

	// 2. Validate token and retrieve devices from the storage
	if _, err := h.Broker.ValidateToken(context.Background(), &core.ValidateTokenBrokerReq{AppEUI: req.AppEUI, Token: req.Token}); err != nil {
		h.Ctx.WithError(err).Debug("Unable to handle list devices request")
		return new(core.ListDevicesHandlerRes), errors.New(errors.Operational, err)
	}
	entries, err := h.DevStorage.readAll(req.AppEUI)
	if err != nil {
		h.Ctx.WithError(err).Debug("Unable to handle list devices request")
		return new(core.ListDevicesHandlerRes), errors.New(errors.Operational, err)
	}

	// 3. Build the reply, separate OTAA from ABP
	var abp []*core.HandlerABPDevice
	var otaa []*core.HandlerOTAADevice
	for _, dev := range entries {
		d := new(devEntry)
		*d = dev
		if dev.AppKey == nil {
			abp = append(abp, &core.HandlerABPDevice{
				DevAddr:  d.DevAddr,
				NwkSKey:  d.NwkSKey[:],
				AppSKey:  d.AppSKey[:],
				FCntUp:   d.FCntUp,
				FCntDown: d.FCntDown,
				DevMode:  d.DevMode,
			})
		} else {
			otaa = append(otaa, &core.HandlerOTAADevice{
				DevEUI:   d.DevEUI,
				DevAddr:  d.DevAddr,
				NwkSKey:  d.NwkSKey[:],
				AppSKey:  d.AppSKey[:],
				AppKey:   d.AppKey[:],
				FCntUp:   d.FCntUp,
				FCntDown: d.FCntDown,
			})
		}
	}

	return &core.ListDevicesHandlerRes{ABP: abp, OTAA: otaa}, nil
}

// UpsertABP implements the core.HandlerManager interface
func (h component) UpsertABP(bctx context.Context, req *core.UpsertABPHandlerReq) (*core.UpsertABPHandlerRes, error) {
	h.Ctx.Debug("Handle upsert ABP request")

	// 1. Validate the request
	if len(req.AppEUI) != 8 || len(req.DevAddr) != 4 || len(req.NwkSKey) != 16 || len(req.AppSKey) != 16 {
		err := errors.New(errors.Structural, "Invalid request parameters")
		h.Ctx.WithError(err).Debug("Unable to handle ABP request")
		return new(core.UpsertABPHandlerRes), err
	}

	// 2. Forward to the broker first -> The Broker also does the token verification
	_, err := h.Broker.UpsertABP(context.Background(), &core.UpsertABPBrokerReq{
		Token:      req.Token,
		AppEUI:     req.AppEUI,
		DevAddr:    req.DevAddr,
		NwkSKey:    req.NwkSKey,
		NetAddress: h.PrivateNetAddrAnnounce,
		DevMode:    req.DevMode,
	})
	if err != nil {
		h.Ctx.WithError(err).Debug("Broker rejected ABP")
		return new(core.UpsertABPHandlerRes), errors.New(errors.Operational, err)
	}

	// 3. Save the device in the storage
	h.Ctx.WithField("AppEUI", req.AppEUI).WithField("DevAddr", req.DevAddr).Debug("Request accepted by broker. Registering device")
	entry := devEntry{
		AppEUI:   req.AppEUI,
		DevEUI:   append([]byte{0, 0, 0, 0}, req.DevAddr...),
		DevAddr:  req.DevAddr,
		FCntDown: 0,
		FCntUp:   0,
		DevMode:  req.DevMode,
	}
	copy(entry.NwkSKey[:], req.NwkSKey)
	copy(entry.AppSKey[:], req.AppSKey)
	if err = h.DevStorage.upsert(entry); err != nil {
		h.Ctx.WithError(err).Debug("Error while trying to handle valid request")
		return new(core.UpsertABPHandlerRes), errors.New(errors.Operational, err)
	}
	h.Processed.Remove(append([]byte{1}, append(entry.AppEUI, entry.DevEUI...)...))

	return new(core.UpsertABPHandlerRes), nil
}

// UpsertOTAA implements the core.HandlerManager interface
func (h component) UpsertOTAA(bctx context.Context, req *core.UpsertOTAAHandlerReq) (*core.UpsertOTAAHandlerRes, error) {
	h.Ctx.Debug("Handle upsert OTAA request")

	// 1. Validate the request
	if len(req.AppEUI) != 8 || len(req.DevEUI) != 8 || len(req.AppKey) != 16 {
		err := errors.New(errors.Structural, "Invalid request parameters")
		h.Ctx.WithError(err).Debug("Unable to handle OTAA request")
		return new(core.UpsertOTAAHandlerRes), err
	}

	// 2. Notify the broker -> The Broker also does the token verification
	_, err := h.Broker.ValidateOTAA(context.Background(), &core.ValidateOTAABrokerReq{
		Token:      req.Token,
		NetAddress: h.PrivateNetAddrAnnounce,
		AppEUI:     req.AppEUI,
	})
	if err != nil {
		h.Ctx.WithError(err).Debug("Broker rejected OTAA")
		return new(core.UpsertOTAAHandlerRes), errors.New(errors.Operational, err)
	}

	// 3. Save the device in the storage
	h.Ctx.WithField("AppEUI", req.AppEUI).WithField("DevEUI", req.DevEUI).Debug("Request accepted by broker. Registering device")
	var appKey [16]byte
	copy(appKey[:], req.AppKey)
	err = h.DevStorage.upsert(devEntry{
		AppEUI: req.AppEUI,
		DevEUI: req.DevEUI,
		AppKey: &appKey,
	})
	if err != nil {
		h.Ctx.WithError(err).Debug("Error while trying to handle valid request")
		return new(core.UpsertOTAAHandlerRes), err
	}

	return new(core.UpsertOTAAHandlerRes), nil
}

// GetDefaultDevice implements the core.HandlerManager Interface
func (h component) GetDefaultDevice(bctx context.Context, req *core.GetDefaultDeviceReq) (*core.GetDefaultDeviceRes, error) {
	h.Ctx.Debug("Handle get default device request")

	// 1. Validate the request
	if len(req.AppEUI) != 8 {
		err := errors.New(errors.Structural, "Invalid request parameters")
		h.Ctx.WithError(err).Debug("Unable to handle set default device request")
		return new(core.GetDefaultDeviceRes), err
	}

	// 2. Validate the token
	_, err := h.Broker.ValidateToken(context.Background(), &core.ValidateTokenBrokerReq{
		Token:  req.Token,
		AppEUI: req.AppEUI,
	})
	if err != nil {
		h.Ctx.WithError(err).Debug("Broker rejected token")
		return new(core.GetDefaultDeviceRes), errors.New(errors.Operational, err)
	}

	// 3. Get default device entry from storage
	entry, err := h.DevStorage.getDefault(req.AppEUI)
	if err != nil {
		h.Ctx.WithError(err).Debug("Error while trying to retrieve default device")
		return new(core.GetDefaultDeviceRes), err
	}
	if entry == nil {
		return new(core.GetDefaultDeviceRes), errors.New(errors.NotFound, "No default device found")
	}

	return &core.GetDefaultDeviceRes{AppKey: entry.AppKey[:]}, nil
}

// SetDefaultDevice implements the core.HandlerManager interface
func (h component) SetDefaultDevice(bctx context.Context, req *core.SetDefaultDeviceReq) (*core.SetDefaultDeviceRes, error) {
	h.Ctx.Debug("Handle set default device request")

	// 1. Validate the request
	if len(req.AppEUI) != 8 || len(req.AppKey) != 16 {
		err := errors.New(errors.Structural, "Invalid request parameters")
		h.Ctx.WithError(err).Debug("Unable to handle set default device request")
		return new(core.SetDefaultDeviceRes), err
	}

	// 2. Validate the token
	_, err := h.Broker.ValidateOTAA(context.Background(), &core.ValidateOTAABrokerReq{
		Token:      req.Token,
		AppEUI:     req.AppEUI,
		NetAddress: h.PrivateNetAddrAnnounce,
	})
	if err != nil {
		h.Ctx.WithError(err).Debug("Broker rejected token")
		return new(core.SetDefaultDeviceRes), errors.New(errors.Operational, err)
	}

	// 3. Set the default device in the storage
	h.Ctx.WithField("AppEUI", req.AppEUI).Debug("Valid token. Registering default device")
	var appKey [16]byte
	copy(appKey[:], req.AppKey)
	err = h.DevStorage.setDefault(req.AppEUI, &devDefaultEntry{
		AppKey: appKey,
	})
	if err != nil {
		h.Ctx.WithError(err).Debug("Storage error")
		return new(core.SetDefaultDeviceRes), errors.New(errors.Operational, err)
	}

	return new(core.SetDefaultDeviceRes), nil
}

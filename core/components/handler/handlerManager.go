// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// ListDevices implements the core.HandlerManagerServer interface
func (h component) ListDevices(bctx context.Context, req *core.ListDevicesHandlerReq) (*core.ListDevicesHandlerRes, error) {
	return new(core.ListDevicesHandlerRes), errors.New(errors.Implementation, "Not implemented")
}

// UpsertABP implements the core.HandlerManager interface
func (h component) UpsertABP(bctx context.Context, req *core.UpsertABPHandlerReq) (*core.UpsertABPHandlerRes, error) {
	h.Ctx.Debug("Handle Upsert ABP Request")

	// 1. Validate the request
	if len(req.AppEUI) != 8 || len(req.DevAddr) != 4 || len(req.NwkSKey) != 16 || len(req.AppSKey) != 16 {
		err := errors.New(errors.Structural, "Invalid request parameters")
		h.Ctx.WithError(err).Debug("Unable to handle ABP request")
		return new(core.UpsertABPHandlerRes), err
	}

	// 2. Forward to the broker firt -> The Broker also does the token verification
	var token string
	if meta, ok := metadata.FromContext(bctx); ok && len(meta["token"]) > 0 {
		token = meta["token"][0]
	}
	_, err := h.Broker.BeginToken(token).UpsertABP(context.Background(), &core.UpsertABPBrokerReq{
		AppEUI:     req.AppEUI,
		DevAddr:    req.DevAddr,
		NwkSKey:    req.NwkSKey,
		NetAddress: h.NetAddr,
	})
	h.Broker.EndToken()
	if err != nil {
		h.Ctx.WithError(err).Debug("Broker rejected ABP")
		return new(core.UpsertABPHandlerRes), errors.New(errors.Operational, err)
	}

	// 3. Insert the request in our own storage
	h.Ctx.WithField("AppEUI", req.AppEUI).WithField("DevAddr", req.DevAddr).Debug("Request accepted by broker. Registering Device.")
	entry := devEntry{
		AppEUI:   req.AppEUI,
		DevEUI:   append([]byte{0, 0, 0, 0}, req.DevAddr...),
		DevAddr:  req.DevAddr,
		FCntDown: 0,
	}
	copy(entry.NwkSKey[:], req.NwkSKey)
	copy(entry.AppSKey[:], req.AppSKey)
	if err = h.DevStorage.upsert(entry); err != nil {
		h.Ctx.WithError(err).Debug("Error while trying to save valid request")
		return new(core.UpsertABPHandlerRes), errors.New(errors.Operational, err)
	}

	// Done.
	return new(core.UpsertABPHandlerRes), nil
}

// UpsertOTAA implements the core.HandlerManager interface
func (h component) UpsertOTAA(bctx context.Context, req *core.UpsertOTAAHandlerReq) (*core.UpsertOTAAHandlerRes, error) {
	h.Ctx.Debug("Handle Upsert OTAA Request")

	// 1. Validate the request
	if len(req.AppEUI) != 8 || len(req.DevEUI) != 8 || len(req.AppKey) != 16 {
		err := errors.New(errors.Structural, "Invalid request parameters")
		h.Ctx.WithError(err).Debug("Unable to handle OTAA request")
		return new(core.UpsertOTAAHandlerRes), err
	}

	// 2. Notify the broker -> The Broker also does the token verification
	var token string
	h.Ctx.WithField("meta", bctx).Debug("Trying to get Meta")
	if meta, ok := metadata.FromContext(bctx); ok && len(meta["token"]) > 0 {
		token = meta["token"][0]
	}
	_, err := h.Broker.BeginToken(token).ValidateOTAA(context.Background(), &core.ValidateOTAABrokerReq{
		NetAddress: h.NetAddr,
		AppEUI:     req.AppEUI,
	})
	h.Broker.EndToken()
	if err != nil {
		h.Ctx.WithError(err).Debug("Broker rejected OTAA")
		return new(core.UpsertOTAAHandlerRes), errors.New(errors.Operational, err)
	}

	// 3. Insert the request in our own storage
	h.Ctx.WithField("AppEUI", req.AppEUI).WithField("DevEUI", req.DevEUI).Debug("Request accepted by broker. Registering Device.")
	var appKey [16]byte
	copy(appKey[:], req.AppKey)
	err = h.DevStorage.upsert(devEntry{
		AppEUI: req.AppEUI,
		DevEUI: req.DevEUI,
		AppKey: &appKey,
	})
	if err != nil {
		h.Ctx.WithError(err).Debug("Error while trying to save valid request")
		return new(core.UpsertOTAAHandlerRes), err
	}

	// 4. Done.
	return new(core.UpsertOTAAHandlerRes), nil
}

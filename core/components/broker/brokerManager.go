// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"regexp"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context"
)

// ListDevices implements the core.BrokerManagerServer interface
func (b component) ListDevices(bctx context.Context, req *core.ListDevicesBrokerReq) (*core.ListDevicesBrokerRes, error) {
	return new(core.ListDevicesBrokerRes), errors.New(errors.Implementation, "Not implemented")
}

// ValidateOTAA implements the core.BrokerManager interface
func (b component) ValidateOTAA(bctx context.Context, req *core.ValidateOTAABrokerReq) (*core.ValidateOTAABrokerRes, error) {
	b.Ctx.Debug("Handle ValidateOTAA request")

	// 1. Validate the request
	re := regexp.MustCompile("^([-\\w]+\\.?)+:\\d+$")
	if len(req.AppEUI) != 8 || !re.Match([]byte(req.NetAddress)) {
		err := errors.New(errors.Structural, "Invalid request parameters")
		b.Ctx.WithError(err).Debug("Unable to validate OTAA request")
		return new(core.ValidateOTAABrokerRes), err
	}

	// 2. Verify and validate the token
	// TODO

	// 3. Update the internal storage
	b.Ctx.WithField("AppEUI", req.AppEUI).Debug("Request accepted by broker. Registering / Updating App.")
	err := b.AppStorage.upsert(appEntry{
		Dialer: NewDialer([]byte(req.NetAddress)),
		AppEUI: req.AppEUI,
	})
	if err != nil {
		b.Ctx.WithError(err).Debug("Error while trying to save valid request")
		return new(core.ValidateOTAABrokerRes), errors.New(errors.Operational, err)
	}

	// 4. Done.
	return new(core.ValidateOTAABrokerRes), nil
}

// UpsertABP implements the core.BrokerManager interface
func (b component) UpsertABP(bctx context.Context, req *core.UpsertABPBrokerReq) (*core.UpsertABPBrokerRes, error) {
	b.Ctx.Debug("Handle ValidateOTAA request")

	// 1. Validate the request
	re := regexp.MustCompile("^([-\\w]+\\.?)+:\\d+$")
	if len(req.AppEUI) != 8 || !re.Match([]byte(req.NetAddress)) || len(req.DevAddr) != 4 || len(req.NwkSKey) != 16 {
		err := errors.New(errors.Structural, "Invalid request parameters")
		b.Ctx.WithError(err).Debug("Unable to proceed Upsert ABP request")
		return new(core.UpsertABPBrokerRes), err
	}

	// 2. Verify and validate the token
	// TODO

	// 3. Update the internal storage
	b.Ctx.WithField("AppEUI", req.AppEUI).WithField("DevAddr", req.DevAddr).Debug("Request accepted by broker. Registering device.")
	var nwkSKey [16]byte
	copy(nwkSKey[:], req.NwkSKey)
	err := b.NetworkController.upsert(devEntry{
		Dialer:  NewDialer([]byte(req.NetAddress)),
		AppEUI:  req.AppEUI,
		DevEUI:  append([]byte{0, 0, 0, 0}, req.DevAddr...),
		DevAddr: req.DevAddr,
		NwkSKey: nwkSKey,
		FCntUp:  0,
	})
	if err != nil {
		b.Ctx.WithError(err).Debug("Error while trying to save valid request")
		return new(core.UpsertABPBrokerRes), errors.New(errors.Operational, err)
	}

	// 4. Done.
	return new(core.UpsertABPBrokerRes), nil
}

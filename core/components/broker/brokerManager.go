// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"fmt"
	"regexp"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

// ValidateToken implements the core.BrokerManagerServer interface
func (b component) ValidateToken(bctx context.Context, req *core.ValidateTokenBrokerReq) (*core.ValidateTokenBrokerRes, error) {
	b.Ctx.Debug("Handle ValidateToken request")
	if len(req.AppEUI) != 8 {
		err := errors.New(errors.Structural, "Invalid request parameters")
		b.Ctx.WithError(err).Debug("Unable to handle ValidateToken request")
		return new(core.ValidateTokenBrokerRes), err
	}
	if err := b.validateToken(bctx, req.Token, req.AppEUI); err != nil {
		b.Ctx.WithError(err).Debug("Unable to handle ValidateToken request")
		return new(core.ValidateTokenBrokerRes), err
	}
	return new(core.ValidateTokenBrokerRes), nil
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
	if err := b.validateToken(bctx, req.Token, req.AppEUI); err != nil {
		return new(core.ValidateOTAABrokerRes), err
	}

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
	if err := b.validateToken(bctx, req.Token, req.AppEUI); err != nil {
		return new(core.UpsertABPBrokerRes), err
	}

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

// validateToken verify an OAuth Bearer token pass through metadata during RPC
func (b component) validateToken(ctx context.Context, token string, appEUI []byte) error {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return b.SecretKey[:], nil
	})
	if err != nil {
		return errors.New(errors.Structural, "Unable to parse token")
	}
	if !parsed.Valid || parsed.Claims["sub"] != fmt.Sprintf("%X", appEUI) {
		return errors.New(errors.Structural, "Invalid token.")
	}
	return nil
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"time"

	pb "github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/go-account-lib/rights"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/types"
	gogo "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func (h *handlerManager) SimulateUplink(ctx context.Context, in *pb.SimulatedUplinkMessage) (*gogo.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Uplink")
	}

	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	err = checkAppRights(claims, in.AppID, rights.Devices)
	if err != nil {
		return nil, err
	}

	log := h.handler.Ctx.WithFields(log.Fields{
		"AppID": in.AppID,
		"DevID": in.DevID,
	})

	dev, err := h.handler.devices.Get(in.AppID, in.DevID)
	if err != nil {
		return nil, err
	}

	uplink := &types.UplinkMessage{
		AppID:          in.AppID,
		DevID:          in.DevID,
		HardwareSerial: dev.DevEUI.String(),
		FPort:          uint8(in.Port),
		PayloadRaw:     in.Payload,
		Metadata: types.Metadata{
			Time: types.JSONTime(time.Now().UTC()),
			LocationMetadata: types.LocationMetadata{
				Latitude:  dev.Latitude,
				Longitude: dev.Longitude,
				Altitude:  dev.Altitude,
			},
		},
	}

	err = h.handler.ConvertFieldsUp(log, nil, uplink, dev)
	if err != nil {
		return nil, err
	}

	h.handler.qUp <- uplink

	return new(gogo.Empty), nil
}

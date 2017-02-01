// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"time"

	"github.com/TheThingsNetwork/go-account-lib/rights"
	"github.com/TheThingsNetwork/go-utils/log"
	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func (h *handlerManager) SimulateUplink(ctx context.Context, in *pb.SimulatedUplinkMessage) (*empty.Empty, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Uplink")
	}

	ctx, claims, err := h.validateTTNAuthAppContext(ctx, in.AppId)
	if err != nil {
		return nil, err
	}
	err = checkAppRights(claims, in.AppId, rights.Devices)
	if err != nil {
		return nil, err
	}

	log := h.handler.Ctx.WithFields(log.Fields{
		"AppID": in.AppId,
		"DevID": in.DevId,
	})

	dev, err := h.handler.devices.Get(in.AppId, in.DevId)
	if err != nil {
		return nil, err
	}

	uplink := &types.UplinkMessage{
		AppID:          in.AppId,
		DevID:          in.DevId,
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

	h.handler.mqttUp <- uplink
	if h.handler.amqpEnabled {
		h.handler.amqpUp <- uplink
	}

	return new(empty.Empty), nil
}

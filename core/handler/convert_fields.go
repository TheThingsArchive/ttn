// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/handler/functions"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// ConvertFieldsUp converts the payload to fields using payload functions
func (h *handler) ConvertFieldsUp(ctx ttnlog.Interface, _ *pb_broker.DeduplicatedUplinkMessage, appUp *types.UplinkMessage, _ *device.Device) error {
	// Find Application
	app, err := h.applications.Get(appUp.AppID)
	if err != nil {
		return nil // Do not process if application not found
	}

	functions := &CustomUplinkFunctions{
		Decoder:   app.CustomDecoder,
		Converter: app.CustomConverter,
		Validator: app.CustomValidator,
		Logger:    functions.Ignore,
	}

	fields, valid, err := functions.Process(appUp.PayloadRaw, appUp.FPort)
	if err != nil {

		// Emit the error
		h.mqttEvent <- &types.DeviceEvent{
			AppID: appUp.AppID,
			DevID: appUp.DevID,
			Event: types.UplinkErrorEvent,
			Data:  types.ErrorEventData{Error: err.Error()},
		}

		// Do not set fields if processing failed, but allow the handler to continue processing
		// without payload functions
		return nil
	}

	if !valid {
		return errors.NewErrInvalidArgument("Payload", "payload validator function returned false")
	}

	appUp.PayloadFields = fields

	return nil
}

// ConvertFieldsDown converts the fields into a payload
func (h *handler) ConvertFieldsDown(ctx ttnlog.Interface, appDown *types.DownlinkMessage, ttnDown *pb_broker.DownlinkMessage, _ *device.Device) error {
	if appDown.PayloadFields == nil || len(appDown.PayloadFields) == 0 {
		return nil
	}

	if appDown.PayloadRaw != nil {
		return errors.NewErrInvalidArgument("Downlink", "Both Fields and Payload provided")
	}

	app, err := h.applications.Get(appDown.AppID)
	if err != nil {
		return nil
	}

	functions := &CustomDownlinkFunctions{
		Encoder: app.CustomEncoder,
		Logger:  functions.Ignore,
	}

	message, _, err := functions.Process(appDown.PayloadFields, appDown.FPort)
	if err != nil {
		return err
	}

	appDown.PayloadRaw = message

	return nil
}

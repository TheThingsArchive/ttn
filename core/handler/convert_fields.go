// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"encoding/json"
	"fmt"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/cayennelpp"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/handler/functions"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// PayloadDecoder decodes raw payload to fields
type PayloadDecoder interface {
	Decode(payload []byte, fPort uint8) (map[string]interface{}, bool, error)
	Log() []*pb_handler.LogEntry
}

// PayloadEncoder encodes fields to raw payload
type PayloadEncoder interface {
	Encode(fields map[string]interface{}, fPort uint8) ([]byte, bool, error)
	Log() []*pb_handler.LogEntry
}

// ConvertFieldsUp converts the payload to fields using the application's payload formatter
func (h *handler) ConvertFieldsUp(ctx ttnlog.Interface, _ *pb_broker.DeduplicatedUplinkMessage, appUp *types.UplinkMessage, dev *device.Device) error {
	// Find Application
	app, err := h.applications.Get(appUp.AppID)
	if err != nil {
		return nil // Do not process if application not found
	}

	var decoder PayloadDecoder
	switch app.PayloadFormat {
	case application.PayloadFormatCustom:
		decoder = &CustomUplinkFunctions{
			Decoder:   app.CustomDecoder,
			Converter: app.CustomConverter,
			Validator: app.CustomValidator,
			Logger:    functions.Ignore,
		}
	case application.PayloadFormatCayenneLPP:
		decoder = &cayennelpp.Decoder{}
	default:
		return nil
	}

	fields, valid, err := decoder.Decode(appUp.PayloadRaw, appUp.FPort)
	if err != nil {
		// Emit the error
		h.qEvent <- &types.DeviceEvent{
			AppID: appUp.AppID,
			DevID: appUp.DevID,
			Event: types.UplinkErrorEvent,
			Data:  types.ErrorEventData{Error: fmt.Sprintf("Unable to decode payload fields: %s", err)},
		}

		// Do not set fields if processing failed, but allow the handler to continue processing
		// without payload formatting
		return nil
	}

	if !valid {
		return errors.NewErrInvalidArgument("Payload", "payload validator function returned false")
	}

	// Check if the functions return valid JSON
	_, err = json.Marshal(fields)
	if err != nil {
		// Emit the error
		h.qEvent <- &types.DeviceEvent{
			AppID: appUp.AppID,
			DevID: appUp.DevID,
			Event: types.UplinkErrorEvent,
			Data:  types.ErrorEventData{Error: fmt.Sprintf("Payload Function output cannot be marshaled to JSON: %s", err.Error())},
		}

		// Do not set fields if processing failed, but allow the handler to continue processing
		// without payload formatting
		return nil
	}

	appUp.PayloadFields = fields
	appUp.Attributes = dev.Attributes

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

	var encoder PayloadEncoder
	switch app.PayloadFormat {
	case application.PayloadFormatCustom:
		encoder = &CustomDownlinkFunctions{
			Encoder: app.CustomEncoder,
			Logger:  functions.Ignore,
		}
	case application.PayloadFormatCayenneLPP:
		encoder = &cayennelpp.Encoder{}
	default:
		return nil
	}

	raw, _, err := encoder.Encode(appDown.PayloadFields, appDown.FPort)
	if err != nil {
		return err
	}

	appDown.PayloadRaw = raw

	return nil
}

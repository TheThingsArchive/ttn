// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"errors"
	"fmt"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
	"github.com/robertkrimen/otto"
)

// ConvertFields converts the payload to fields using payload functions
func (h *handler) ConvertFields(ctx log.Interface, ttnUp *pb_broker.DeduplicatedUplinkMessage, appUp *mqtt.UplinkMessage) error {
	// Find Application
	app, err := h.applications.Get(appUp.AppEUI)
	if err != nil {
		return nil // Do not process if application not found
	}

	functions := &Functions{
		Decoder:   app.Decoder,
		Converter: app.Converter,
		Validator: app.Validator,
	}

	fields, valid, err := functions.Process(appUp.Payload)
	if err != nil {
		return nil // Do not set fields if processing failed
	}

	if !valid {
		return errors.New("ttn/handler: The processed payload is not valid")
	}

	appUp.Fields = fields

	return nil
}

// Functions decodes, converts and validates payload using JavaScript functions
type Functions struct {
	// Decoder is a JavaScript function that accepts the payload as byte array and
	// returns an object containing the decoded values
	Decoder string
	// Converter is a JavaScript function that accepts the data as decoded by
	// Decoder and returns an object containing the converted values
	Converter string
	// Validator is a JavaScript function that validates the data is converted by
	// Converter and returns a boolean value indicating the validity of the data
	Validator string
}

// Decode decodes the payload using the Decoder function into a map
func (f *Functions) Decode(payload []byte) (map[string]interface{}, error) {
	if f.Decoder == "" {
		return nil, errors.New("Decoder function not set")
	}

	vm := otto.New()
	vm.Set("payload", payload)
	value, err := vm.Run(fmt.Sprintf("(%s)(payload)", f.Decoder))
	if err != nil {
		return nil, err
	}

	if !value.IsObject() {
		return nil, errors.New("Decoder does not return an object")
	}

	v, _ := value.Export()
	return v.(map[string]interface{}), nil
}

// Convert converts the values in the specified map to a another map using the
// Converter function. If the Converter function is not set, this function
// returns the data as-is
func (f *Functions) Convert(data map[string]interface{}) (map[string]interface{}, error) {
	if f.Converter == "" {
		return data, nil
	}

	vm := otto.New()
	vm.Set("data", data)
	value, err := vm.Run(fmt.Sprintf("(%s)(data)", f.Converter))
	if err != nil {
		return nil, err
	}

	if !value.IsObject() {
		return nil, errors.New("Converter does not return an object")
	}

	v, _ := value.Export()
	return v.(map[string]interface{}), nil
}

// Validate validates the values in the specified map using the Validator
// function. If the Validator function is not set, this function returns true
func (f *Functions) Validate(data map[string]interface{}) (bool, error) {
	if f.Validator == "" {
		return true, nil
	}

	vm := otto.New()
	vm.Set("data", data)
	value, err := vm.Run(fmt.Sprintf("(%s)(data)", f.Validator))
	if err != nil {
		return false, err
	}

	if !value.IsBoolean() {
		return false, errors.New("Validator does not return a boolean")
	}

	return value.ToBoolean()
}

// Process decodes the specified payload, converts it and test the validity
func (f *Functions) Process(payload []byte) (map[string]interface{}, bool, error) {
	decoded, err := f.Decode(payload)
	if err != nil {
		return nil, false, err
	}

	converted, err := f.Convert(decoded)
	if err != nil {
		return nil, false, err
	}

	valid, err := f.Validate(converted)
	return converted, valid, err
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"reflect"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
	"github.com/robertkrimen/otto"
)

// ConvertFieldsUp converts the payload to fields using payload functions
func (h *handler) ConvertFieldsUp(ctx log.Interface, ttnUp *pb_broker.DeduplicatedUplinkMessage, appUp *mqtt.UplinkMessage) error {
	// Find Application
	app, err := h.applications.Get(ttnUp.AppId)
	if err != nil {
		return nil // Do not process if application not found
	}

	functions := &UplinkFunctions{
		Decoder:   app.Decoder,
		Converter: app.Converter,
		Validator: app.Validator,
	}

	fields, valid, err := functions.Process(appUp.Payload)
	if err != nil {
		return nil // Do not set fields if processing failed
	}

	if !valid {
		return core.NewErrInvalidArgument("Payload", "payload validator function returned false")
	}

	appUp.Fields = fields

	return nil
}

// UplinkFunctions decodes, converts and validates payload using JavaScript functions
type UplinkFunctions struct {
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

// timeOut is the maximum allowed time a payload function is allowed to run
var timeOut = 100 * time.Millisecond

// Decode decodes the payload using the Decoder function into a map
func (f *UplinkFunctions) Decode(payload []byte) (map[string]interface{}, error) {
	if f.Decoder == "" {
		return nil, core.NewErrInternal("Decoder function not set")
	}

	vm := otto.New()
	vm.Set("payload", payload)
	value, err := runUnsafeCode(vm, fmt.Sprintf("(%s)(payload)", f.Decoder), timeOut)
	if err != nil {
		return nil, err
	}

	if !value.IsObject() {
		return nil, core.NewErrInvalidArgument("Decoder", "does not return an object")
	}

	v, _ := value.Export()
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil, core.NewErrInvalidArgument("Decoder", "does not return an object")
	}
	return m, nil
}

// Convert converts the values in the specified map to a another map using the
// Converter function. If the Converter function is not set, this function
// returns the data as-is
func (f *UplinkFunctions) Convert(data map[string]interface{}) (map[string]interface{}, error) {
	if f.Converter == "" {
		return data, nil
	}

	vm := otto.New()
	vm.Set("data", data)
	value, err := runUnsafeCode(vm, fmt.Sprintf("(%s)(data)", f.Converter), timeOut)
	if err != nil {
		return nil, err
	}

	if !value.IsObject() {
		return nil, core.NewErrInvalidArgument("Converter", "does not return an object")
	}

	v, _ := value.Export()
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil, core.NewErrInvalidArgument("Converter", "does not return an object")
	}

	return m, nil
}

// Validate validates the values in the specified map using the Validator
// function. If the Validator function is not set, this function returns true
func (f *UplinkFunctions) Validate(data map[string]interface{}) (bool, error) {
	if f.Validator == "" {
		return true, nil
	}

	vm := otto.New()
	vm.Set("data", data)
	value, err := runUnsafeCode(vm, fmt.Sprintf("(%s)(data)", f.Validator), timeOut)
	if err != nil {
		return false, err
	}

	if !value.IsBoolean() {
		return false, core.NewErrInvalidArgument("Validator", "does not return a boolean")
	}

	return value.ToBoolean()
}

// Process decodes the specified payload, converts it and test the validity
func (f *UplinkFunctions) Process(payload []byte) (map[string]interface{}, bool, error) {
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

var errTimeOutExceeded = core.NewErrInternal("Code has been running to long")

func runUnsafeCode(vm *otto.Otto, code string, timeOut time.Duration) (value otto.Value, err error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if caught := recover(); caught != nil {
			if caught == errTimeOutExceeded {
				value = otto.Value{}
				err = core.NewErrInternal(fmt.Sprintf("Interrupted javascript execution after %v", duration))
				return
			}
			// if this is not the our timeout interrupt, raise the panic again
			// so someone else can handle it
			panic(caught)
		}
	}()

	vm.Interrupt = make(chan func(), 1)

	go func() {
		time.Sleep(timeOut)
		vm.Interrupt <- func() {
			panic(errTimeOutExceeded)
		}
	}()
	return vm.Run(code)
}

// DownlinkFunctions encodes payload using JavaScript functions
type DownlinkFunctions struct {
	// Encoder is a JavaScript function that accepts the payload as JSON and
	// returns an array of bytes
	Encoder string
}

// Encode encodes the map into a byte slice using the encoder payload function
// If no encoder function is set, this function returns an array.
func (f *DownlinkFunctions) Encode(payload map[string]interface{}) ([]byte, error) {
	if f.Encoder == "" {
		return nil, core.NewErrInternal("Encoder function not set")
	}

	vm := otto.New()
	vm.Set("payload", payload)
	value, err := runUnsafeCode(vm, fmt.Sprintf("(%s)(payload)", f.Encoder), timeOut)
	if err != nil {
		return nil, err
	}

	if !value.IsObject() {
		return nil, core.NewErrInvalidArgument("Encoder", "does not return an object")
	}

	v, err := value.Export()
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(v).Kind() != reflect.Slice {
		return nil, core.NewErrInvalidArgument("Encoder", "does not return an Array")
	}

	s := reflect.ValueOf(v)
	l := s.Len()

	res := make([]byte, l)

	var n int64
	for i := 0; i < l; i++ {
		el := s.Index(i).Interface()

		// type switch does not have fallthrough so we need
		// to check every element individually
		switch t := el.(type) {
		case byte:
			n = int64(t)
		case int:
			n = int64(t)
		case int8:
			n = int64(t)
		case int16:
			n = int64(t)
		case uint16:
			n = int64(t)
		case int32:
			n = int64(t)
		case uint32:
			n = int64(t)
		case int64:
			n = int64(t)
		case uint64:
			n = int64(t)
		case float32:
			n = int64(t)
			if float32(n) != t {
				return nil, core.NewErrInvalidArgument("Encoder", "should return an Array of integer numbers")
			}
		case float64:
			n = int64(t)
			if float64(n) != t {
				return nil, core.NewErrInvalidArgument("Encoder", "should return an Array of integer numbers")
			}
		default:
			return nil, core.NewErrInvalidArgument("Encoder", "should return an Array of integer numbers")
		}

		if n < 0 || n > 255 {
			return nil, core.NewErrInvalidArgument("Encoder Output", "Numbers in Array should be between 0 and 255")
		}

		res[i] = byte(n)
	}

	return res, nil
}

// Process encode the specified field, converts it into a valid payload
func (f *DownlinkFunctions) Process(payload map[string]interface{}) ([]byte, bool, error) {
	encoded, err := f.Encode(payload)
	if err != nil {
		return nil, false, err
	}

	return encoded, true, nil
}

// ConvertFieldsDown converts the fields into a payload
func (h *handler) ConvertFieldsDown(ctx log.Interface, appDown *mqtt.DownlinkMessage, ttnDown *pb_broker.DownlinkMessage) error {
	if appDown.Fields == nil {
		return nil
	}

	if appDown.Payload != nil {
		return core.NewErrInvalidArgument("Downlink", "Both Fields and Payload provided")
	}

	app, err := h.applications.Get(appDown.AppID)
	if err != nil {
		return nil
	}

	functions := &DownlinkFunctions{
		Encoder: app.Encoder,
	}

	message, _, err := functions.Process(appDown.Fields)
	if err != nil {
		return err
	}

	appDown.Payload = message

	return nil
}

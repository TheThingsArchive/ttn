// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package fields

import (
	"errors"
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
)

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

// timeOut is the maximum allowed time a payload function is allowed to run
var timeOut = 2 * time.Second

// Decode decodes the payload using the Decoder function into a map
func (f *Functions) Decode(payload []byte) (map[string]interface{}, error) {
	if f.Decoder == "" {
		return nil, errors.New("Decoder function not set")
	}

	vm := otto.New()
	vm.Set("payload", payload)
	value, err := RunUnsafeCode(vm, fmt.Sprintf("(%s)(payload)", f.Decoder), timeOut)
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
	value, err := RunUnsafeCode(vm, fmt.Sprintf("(%s)(data)", f.Converter), timeOut)
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
	value, err := RunUnsafeCode(vm, fmt.Sprintf("(%s)(data)", f.Validator), timeOut)
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

var timeOutExceeded = errors.New("Code has been running to long")

func RunUnsafeCode(vm *otto.Otto, code string, timeOut time.Duration) (value otto.Value, err error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if caught := recover(); caught != nil {
			if caught == timeOutExceeded {
				value = otto.Value{}
				err = fmt.Errorf("Interrupted javascript execution after %v", duration)
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
			panic(timeOutExceeded)
		}
	}()
	return vm.Run(code)
}

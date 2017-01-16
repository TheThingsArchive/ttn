// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"reflect"
	"regexp"

	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Validator interface is used to validate protos
type Validator interface {
	// Returns the validation error or nil if valid
	Validate() error
}

// Validate the given object if it implements the Validator interface
// Must not be called with nil values!
func Validate(in interface{}) error {
	if v, ok := in.(Validator); ok {
		return v.Validate()
	}
	return nil
}

var idRegexp = regexp.MustCompile("^[0-9a-z](?:[_-]?[0-9a-z]){1,35}$")

// ValidID returns true if the given ID is a valid application or device ID
func ValidID(id string) bool {
	return idRegexp.MatchString(id)
}

// NotEmptyAndValidID checks if the ID is not empty AND has a valid format
func NotEmptyAndValidID(id string, argument string) error {
	if id == "" {
		return errors.NewErrInvalidArgument(argument, "can not be empty")
	}
	if !ValidID(id) {
		return errors.NewErrInvalidArgument(argument, "has wrong format. IDs can contain lowercase letters, numbers, dashes and underscores and should have a maximum length of 36")
	}
	return nil
}

// NotNilAndValid checks if the given interface is not nil AND validates it
func NotNilAndValid(in interface{}, argument string) error {
	// Structs can not be nil and reflect.ValueOf(in).IsNil() would panic
	if reflect.ValueOf(in).Kind() == reflect.Struct {
		return errors.Wrap(Validate(in), "Invalid "+argument)
	}

	// We need to check for the interface to be nil and the value of the interface
	// See: https://stackoverflow.com/questions/13476349/check-for-nil-and-nil-interface-in-go
	if in == nil || reflect.ValueOf(in).IsNil() {
		return errors.NewErrInvalidArgument(argument, "can not be empty")
	}

	if err := Validate(in); err != nil {
		return errors.Wrap(Validate(in), "Invalid "+argument)
	}

	return nil
}

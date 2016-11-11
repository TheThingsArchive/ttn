package api

import (
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"reflect"
	"regexp"
)

var idRegexp = regexp.MustCompile("^[0-9a-z](?:[_-]?[0-9a-z]){1,35}$")

// ValidID returns true if the given ID is a valid application or device ID
func ValidID(id string) bool {
	return idRegexp.MatchString(id)
}

func NotEmptyAndValidId(id string, argument string) error {
	if id == "" {
		return errors.NewErrInvalidArgument(argument, "can not be empty")
	}
	if !ValidID(id) {
		return errors.NewErrInvalidArgument(argument, "has wrong format "+id)
	}
	return nil
}

func NotNilAndValid(in interface{}, argument string) error {
	// Structs can not be nil and reflect.ValueOf(in).IsNil() would panic
	if reflect.ValueOf(in).Kind() == reflect.Struct {
		return Validate(in)
	}

	// We need to check for the interface to be nil and the value of the interface
	// See: https://stackoverflow.com/questions/13476349/check-for-nil-and-nil-interface-in-go
	if in == nil || reflect.ValueOf(in).IsNil() {
		return errors.NewErrInvalidArgument(argument, "can not be empty")
	}
	return Validate(in)
}

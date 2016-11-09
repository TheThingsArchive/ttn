package api

import (
	"regexp"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

var idRegexp = regexp.MustCompile("^[0-9a-z](?:[_-]?[0-9a-z]){1,35}$")

// ValidID returns true if the given ID is a valid application or device ID
func ValidID(id string) bool {
	return idRegexp.Match([]byte(id))
}

func NotEmptyAndValidId(id string, argument string) error {
	if id == "" {
		return errors.NewErrInvalidArgument(argument, "can not be empty")
	}
	if !ValidID(id) {
		errors.NewErrInvalidArgument(argument, "has wrong format " + id)
	}
	return nil
}


func NotNilAndValid(in interface{}, argument string) error {
	if in == nil {
		return errors.NewErrInvalidArgument(argument, "can not be empty")
	}
	return Validate(in)
}

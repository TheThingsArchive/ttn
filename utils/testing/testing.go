// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package testing offers some handy methods to display check and cross symbols with colors in test
// logs.
package testing

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/apex/log"
)

var ErrStructural = pointer.String(string(errors.Structural))
var ErrOperational = pointer.String(string(errors.Operational))
var ErrNotFound = pointer.String(string(errors.NotFound))
var ErrBehavioural = pointer.String(string(errors.Behavioural))

func GetLogger(t *testing.T, tag string) log.Interface {
	logger := &log.Logger{
		Handler: NewLogHandler(t),
		Level:   log.DebugLevel,
	}
	return logger.WithField("tag", tag)
}

// Ok displays a green check symbol
func Ok(t *testing.T, tag string) {
	t.Log(fmt.Sprintf("\033[32;1m\u2714 ok | %s\033[0m", tag))
}

// Ko fails the test and display a red cross symbol
func Ko(t *testing.T, format string, a ...interface{}) {
	t.Fatalf("\033[31;1m\u2718 ko | \033[0m\033[31m%s\033[0m", fmt.Sprintf(format, a...))
}

// Desc displays the provided description in cyan
func Desc(t *testing.T, format string, a ...interface{}) {
	t.Logf("\033[36m%s\033[0m", fmt.Sprintf(format, a...))
}

// Check serves a for general comparison between two objects
func Check(t *testing.T, want, got interface{}, name string) {
	if !reflect.DeepEqual(want, got) {
		Ko(t, "%s don't match expectations.\nWant: %+v\nGot:  %+v", name, want, got)
	}
	Ok(t, fmt.Sprintf("Check %s", name))
}

// Check errors verify if a given string corresponds to a known error
func CheckErrors(t *testing.T, want *string, got error) {
	if got == nil {
		if want == nil {
			Ok(t, "Check errors")
			return
		}
		Ko(t, "Expected error to be {%s} but got nothing", *want)
		return
	}

	if want == nil {
		Ko(t, "Expected no error but got {%v}", got)
		return
	}

	if got.(errors.Failure).Nature == errors.Nature(*want) {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be {%s} but got {%v}", *want, got)
}

func FatalUnless(t *testing.T, err error) {
	if err != nil {
		Ko(t, "Unexpected error arised: %s", err)
	}
}

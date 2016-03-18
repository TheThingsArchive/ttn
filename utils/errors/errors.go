// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package errors

import (
	"fmt"
	"reflect"
	"time"
)

type Nature string

const (
	Structural     Nature = "Invalid structure"        // Requests, parameters or inputs are wrong. Retry won't work.
	NotFound       Nature = "Unable to find entity"    // Failed to lookup something, somewhere
	Behavioural    Nature = "Wrong behavior or result" // No error but the result isn't the one expected
	Operational    Nature = "Invalid operation"        // An operation failed due to external causes, a retry could work
	Implementation Nature = "Illegal call"             // Method not implemented or unsupported for the given structure
)

// Failure states for fault that occurs during a process.
type Failure struct {
	Nature    Nature    // Kind of error, used a comparator
	Timestamp time.Time // The moment the error was created
	Fault     error     // The source of the failure
}

// NewFailure creates a new Failure from a source error
func New(nat Nature, src interface{}) Failure {
	var fault error
	switch src.(type) {
	case string:
		fault = fmt.Errorf("%s", src.(string))
	case error:
		fault = src.(error)
	default:
		panic("Unexpected error source")
	}

	failure := Failure{
		Nature:    nat,
		Timestamp: time.Now(),
		Fault:     fault,
	}

	// Pop one level if we made a failure from a failure
	t := reflect.TypeOf(src)
	tf := reflect.TypeOf(failure)
	if t == tf {
		failure.Fault = src.(Failure).Fault
	}

	return failure
}

// Error implements the error built-in interface
func (err Failure) Error() string {
	if err.Fault == nil {
		return fmt.Sprintf("%s", err.Nature)
	}
	return fmt.Sprintf("%s: %s", err.Nature, err.Fault.Error())
}

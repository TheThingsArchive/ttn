// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package errors

import (
	"fmt"
	"reflect"
	"time"
)

// Failure states for fault that occurs during a process.
type Failure struct {
	Nature    string    // Kind of error, used a comparator
	Timestamp time.Time // The moment the error was created
	Fault     error     // The source of the failure
}

// NewFailure creates a new Failure from a source error
func New(k string, src interface{}) Failure {
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
		Nature:    k,
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

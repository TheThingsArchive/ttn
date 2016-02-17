// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package errors

import (
	"fmt"
	"reflect"
	"time"
)

type base struct {
	Nature    Nature    // Kind of error, used a comparator
	Timestamp time.Time // The moment the error was created
}

// Failure states for fault that aren't recoverable by the system itself.
// They consist in unexpected behavior
type Failure struct {
	base
	Fault error // The source of the failure
}

// Error states for fault that are explicitely created by the application in order to be handled
// elsewhere. They are recoverable and convey valuable pieces of information.
type Error struct {
	base
}

// Nature identify an error type with a simple tag
type Nature string

// NewFailure creates a new Failure from a source error
func NewFailure(k Nature, src interface{}) Failure {
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
		base: base{
			Nature:    k,
			Timestamp: time.Now(),
		},
		Fault: fault,
	}

	// Pop one level if we made a failure from a failure
	t := reflect.TypeOf(src)
	tf := reflect.TypeOf(failure)
	if t == tf {
		failure.Fault = src.(Failure).Fault
	}

	return failure
}

// NewError creates a new Error
func NewError(k Nature) Error {
	return Error{
		base: base{
			Nature:    k,
			Timestamp: time.Now(),
		},
	}
}

// Error implements the error built-in interface
func (err Failure) Error() string {
	if err.Fault == nil {
		return err.base.Error()
	}
	return fmt.Sprintf("%s: %s", err.Nature, err.Fault.Error())
}

// Error implements the error built-in interface
func (err base) Error() string {
	return fmt.Sprintf("%s", err.Nature)
}

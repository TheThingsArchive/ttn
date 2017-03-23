// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package errors

import (
	"fmt"
	"io"
	"strings"

	errs "github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type ErrType string

// These constants represent error types
const (
	AlreadyExists    ErrType = "already exists"
	Internal         ErrType = "internal"
	InvalidArgument  ErrType = "invalid argument"
	NotFound         ErrType = "not found"
	OutOfRange       ErrType = "out of range"
	PermissionDenied ErrType = "permission denied"
	Unknown          ErrType = "unknown"
)

// GetErrType returns the type of err
func GetErrType(err error) ErrType {
	switch errs.Cause(err).(type) {
	case *ErrAlreadyExists:
		return AlreadyExists
	case *ErrInternal:
		return Internal
	case *ErrInvalidArgument:
		return InvalidArgument
	case *ErrNotFound:
		return NotFound
	case *ErrPermissionDenied:
		return PermissionDenied
	}
	return Unknown
}

// IsPermissionDenied returns whether error type is PermissionDenied
func IsPermissionDenied(err error) bool {
	return GetErrType(err) == PermissionDenied
}

// IsNotFound returns whether error type is NotFound
func IsNotFound(err error) bool {
	return GetErrType(err) == NotFound
}

// IsInvalidArgument returns whether error type is InvalidArgument
func IsInvalidArgument(err error) bool {
	return GetErrType(err) == InvalidArgument
}

// IsInternal returns whether error type is Internal
func IsInternal(err error) bool {
	return GetErrType(err) == Internal
}

// IsAlreadyExists returns whether error type is AlreadyExists
func IsAlreadyExists(err error) bool {
	return GetErrType(err) == AlreadyExists
}

// BuildGRPCError returns the error with a GRPC code
func BuildGRPCError(err error) error {
	if err == nil {
		return nil
	}
	code := grpc.Code(err)
	if code != codes.Unknown {
		return err // it already is a gRPC error
	}
	switch errs.Cause(err).(type) {
	case *ErrAlreadyExists:
		code = codes.AlreadyExists
	case *ErrInternal:
		code = codes.Internal
	case *ErrInvalidArgument:
		code = codes.InvalidArgument
	case *ErrNotFound:
		code = codes.NotFound
	case *ErrPermissionDenied:
		code = codes.PermissionDenied
	}
	switch err {
	case context.Canceled:
		code = codes.Canceled
	case io.EOF:
		code = codes.OutOfRange
	}
	return grpc.Errorf(code, err.Error())
}

// FromGRPCError creates a regular error with the same type as the gRPC error
func FromGRPCError(err error) error {
	if err == nil {
		return nil
	}
	code := grpc.Code(err)
	desc := grpc.ErrorDesc(err)
	switch code {
	case codes.AlreadyExists:
		return NewErrAlreadyExists(strings.TrimSuffix(desc, " already exists"))
	case codes.Internal:
		return NewErrInternal(strings.TrimPrefix(desc, "Internal error: "))
	case codes.InvalidArgument:
		if split := strings.Split(desc, " not valid: "); len(split) == 2 {
			return NewErrInvalidArgument(split[0], split[1])
		}
		return NewErrInvalidArgument("Argument", desc)
	case codes.NotFound:
		return NewErrNotFound(strings.TrimSuffix(desc, " not found"))
	case codes.PermissionDenied:
		return NewErrPermissionDenied(strings.TrimPrefix(desc, "permission denied: "))
	case codes.Unknown: // This also includes all non-gRPC errors
		if desc == "EOF" {
			return io.EOF
		}
		return errs.New(desc)
	}
	return NewErrInternal(fmt.Sprintf("[%s] %s", code, desc))
}

// NewErrAlreadyExists returns a new ErrAlreadyExists for the given entitiy
func NewErrAlreadyExists(entity string) error {
	return &ErrAlreadyExists{entity: entity}
}

// ErrAlreadyExists indicates that an entity already exists
type ErrAlreadyExists struct {
	entity string
}

// Error implements the error interface
func (err ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s already exists", err.entity)
}

// NewErrInternal returns a new ErrInternal with the given message
func NewErrInternal(message string) error {
	return &ErrInternal{message: message}
}

// ErrInternal indicates that an internal error occured
type ErrInternal struct {
	message string
}

// Error implements the error interface
func (err ErrInternal) Error() string {
	return fmt.Sprintf("Internal error: %s", err.message)
}

// NewErrInvalidArgument returns a new ErrInvalidArgument for the given entitiy
func NewErrInvalidArgument(argument string, reason string) error {
	return &ErrInvalidArgument{argument: argument, reason: reason}
}

// ErrInvalidArgument indicates that an argument was invalid
type ErrInvalidArgument struct {
	argument string
	reason   string
}

// Error implements the error interface
func (err ErrInvalidArgument) Error() string {
	return fmt.Sprintf("%s not valid: %s", err.argument, err.reason)
}

// NewErrNotFound returns a new ErrNotFound for the given entitiy
func NewErrNotFound(entity string) error {
	return &ErrNotFound{entity: entity}
}

// ErrNotFound indicates that an entity was not found
type ErrNotFound struct {
	entity string
}

// Error implements the error interface
func (err ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", err.entity)
}

// NewErrPermissionDenied returns a new ErrPermissionDenied with the given reason
func NewErrPermissionDenied(reason string) error {
	return &ErrPermissionDenied{reason: reason}
}

// ErrPermissionDenied indicates that permissions were not sufficient
type ErrPermissionDenied struct {
	reason string
}

// Error implements the error interface
func (err ErrPermissionDenied) Error() string {
	return fmt.Sprintf("permission denied: %s", err.reason)
}

// Wrapf returns an error annotating err with the format specifier.
// If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	return errs.Wrapf(err, format, args...)
}

// Wrap returns an error annotating err with message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	return errs.Wrap(err, message)
}

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(message string) error {
	return errs.New(message)
}

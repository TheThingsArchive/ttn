// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package log provides some handy types and method to activate and deactivate specific log
// behavior within files in a transparent way.
package log

import (
	"fmt"
)

// Logger is a minimalist interface to represent logger
type Logger interface {
	Log(format string, a ...interface{})
}

// DebugLogger can be used in development to display loglines in the console
type DebugLogger struct {
	Tag string
}

// Log implements the Logger interface
func (l DebugLogger) Log(format string, a ...interface{}) {
	fmt.Printf("[ %v ]  ", l.Tag)
	fmt.Printf(format, a...)
}

// VoidLogger can be used to deactivate logs by displaying nothing
type VoidLogger struct{}

// Log implements the Logger interface
func (l VoidLogger) Log(format string, a ...interface{}) {}

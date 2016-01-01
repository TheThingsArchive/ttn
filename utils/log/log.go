// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package log provides some handy types and method to activate and deactivate specific log
// behavior within files in a transparent way.
package log

import (
	"fmt"
	"testing"
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
	fmt.Printf("\033[33m[ %s ]\033[0m ", l.Tag) // Tag printed in yellow
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

// TestLogger can be used in a test environnement to display log only on failure
type TestLogger struct {
	Tag string
	T   *testing.T
}

// Log implements the Logger interface
func (l TestLogger) Log(format string, a ...interface{}) {
	l.T.Logf("\033[33m[ %s ]\033[0m %s", l.Tag, fmt.Sprintf(format, a...)) // Tag printed in yellow
}

// VoidLogger can be used to deactivate logs by displaying nothing
type VoidLogger struct{}

// Log implements the Logger interface
func (l VoidLogger) Log(format string, a ...interface{}) {}

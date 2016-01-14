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
	Log(str string)
	Logf(format string, a ...interface{})
	LogEntry(level Level, msg string, meta Meta)
}

// DebugLogger can be used in development to display loglines in the console
type DebugLogger struct {
	Tag string
}

// Log implements the Logger interface
func (l DebugLogger) Log(str string) {
	fmt.Printf("\033[33m[ %s ]\033[0m ", l.Tag) // Tag printed in yellow
	fmt.Print(str)
	fmt.Print("\n")
}

// Logf implements the Logger interface
func (l DebugLogger) Logf(format string, a ...interface{}) {
	fmt.Printf("\033[33m[ %s ]\033[0m ", l.Tag) // Tag printed in yellow
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

// LogEntry implements the Logger interface
func (l DebugLogger) LogEntry(level Level, msg string, meta Meta) {
	l.Log(Entry(level, msg, meta).String())
}

// TestLogger can be used in a test environnement to display log only on failure
type TestLogger struct {
	Tag string
	T   *testing.T
}

// Log implements the Logger interface
func (l TestLogger) Log(str string) {
	l.T.Logf("\033[33m[ %s ]\033[0m %s", l.Tag, str) // Tag printed in yellow
}

// Logf implements the Logger interface
func (l TestLogger) Logf(format string, a ...interface{}) {
	l.T.Logf("\033[33m[ %s ]\033[0m %s", l.Tag, fmt.Sprintf(format, a...)) // Tag printed in yellow
}

// LogEntry implements the Logger interface
func (l TestLogger) LogEntry(level Level, msg string, meta Meta) {
	l.Log(Entry(level, msg, meta).String())
}

// MultiLogger aggregates several loggers log to each of them
type MultiLogger struct {
	Loggers []Logger
}

// Log implements the Logger interface
func (l MultiLogger) Log(str string) {
	for _, logger := range l.Loggers {
		logger.Log(str)
	}
}

// Logf implements the Logger interface
func (l MultiLogger) Logf(format string, a ...interface{}) {
	for _, logger := range l.Loggers {
		logger.Logf(format, a...)
	}
}

// LogEntry implements the Logger interface
func (l MultiLogger) LogEntry(level Level, msg string, meta Meta) {
	l.Log(Entry(level, msg, meta).String())
}

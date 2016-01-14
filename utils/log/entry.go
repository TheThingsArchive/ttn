// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"strings"
)

// Meta contains metadata for a log entry
type Meta map[string]interface{}

func (m Meta) String() string {
	if len(m) == 0 {
		return ""
	}

	metas := []string{}
	for k, v := range m {
		metas = append(metas, fmt.Sprintf("%s=%v", k, v))
	}

	return fmt.Sprint("\033[34m[ ", strings.Join(metas, ", "), " ]\033[0m")
}

// The Level for this log entry
type Level uint8

// Log levels
const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

func (l Level) String() string {
	switch l {
	case PanicLevel:
		return "\033[31m[ panic ]\033[0m" // Level printed in red
	case FatalLevel:
		return "\033[31m[ fatal ]\033[0m" // Level printed in red
	case ErrorLevel:
		return "\033[31m[ error ]\033[0m" // Level printed in red
	case WarnLevel:
		return "\033[33m[ warn ]\033[0m" // Level printed in yellow
	case DebugLevel:
		return "\033[34m[ debug ]\033[0m" // Level printed in blue
	default: // Default is InfoLevel
		return "\033[32m[ info ]\033[0m" // Level printed in green
	}
}

// entry type
type entry struct {
	Level   Level
	Message string
	Meta    Meta
}

// Log implements the Logger interface
func (e entry) String() string {
	return fmt.Sprintf("%s %s %s", e.Level, e.Message, e.Meta)
}

// Entry gives you a new log entry
func Entry(level Level, msg string, meta Meta) entry {
	return entry{
		Level:   level,
		Message: msg,
		Meta:    meta,
	}
}

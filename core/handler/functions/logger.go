// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package functions

import (
	pb_handler "github.com/TheThingsNetwork/api/handler"
	"github.com/robertkrimen/otto"
)

// Logger is something that can be logged to, saving the logs for later use
type Logger interface {
	// Log passes the console.log function call to the logger
	Log(call otto.FunctionCall)

	// Enter tells the Logger what function it is currently in
	Enter(function string)

	// Entries returns the log
	Entries() []*pb_handler.LogEntry
}

// console is a Logger that saves the logs as LogEntries
type EntryLogger struct {
	logs     []*pb_handler.LogEntry
	function string
}

// Console returns a new Logger that save the logs to
func NewEntryLogger() *EntryLogger {
	return &EntryLogger{
		logs: make([]*pb_handler.LogEntry, 0),
	}
}

// JSON stringifies a value inside of the otto vm, yielding better
// results than Export for Object-like class such as Date, but being much
// slower.
func JSON(val otto.Value) string {
	vm := otto.New()
	vm.Set("value", val)
	res, _ := vm.Run(`JSON.stringify(value)`)
	return res.String()
}

func (c *EntryLogger) Log(call otto.FunctionCall) {
	fields := []string{}
	for _, field := range call.ArgumentList {
		fields = append(fields, JSON(field))
	}

	c.logs = append(c.logs, &pb_handler.LogEntry{
		Function: c.function,
		Fields:   fields,
	})
}

func (c *EntryLogger) Enter(function string) {
	c.function = function
}

func (c *EntryLogger) Entries() []*pb_handler.LogEntry {
	return c.logs
}

type IgnoreLogger struct{}

var Ignore = &IgnoreLogger{}

func (c *IgnoreLogger) Log(call otto.FunctionCall)      {}
func (c *IgnoreLogger) Enter(function string)           {}
func (c *IgnoreLogger) Entries() []*pb_handler.LogEntry { return nil }

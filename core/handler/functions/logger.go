// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package functions

import (
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/robertkrimen/otto"
)

// Logger is something that can be logged to, saving the logs for later use
type Logger interface {
	// Log passes the console.log function call to the logger
	Log(call otto.FunctionCall)

	// Enter tells the Logger what function it is currently in
	Enter(function string)
}

// console is a Logger that saves the logs as LogEntries
type EntryLogger struct {
	Logs     []*pb_handler.LogEntry
	function string
}

// Console returns a new Logger that save the logs to
func NewEntryLogger() *EntryLogger {
	return &EntryLogger{
		Logs: make([]*pb_handler.LogEntry, 0),
	}
}

// vm is used for stringifying values
var vm = otto.New()

// JSON stringifies a value inside of the otto vm, yielding better
// results than Export for Object-like class such as Date, but being much
// slower.
func JSON(val otto.Value) string {
	vm.Set("value", val)
	res, _ := vm.Run(`JSON.stringify(value)`)
	return res.String()
}

func (c *EntryLogger) Log(call otto.FunctionCall) {
	fields := []string{}
	for _, field := range call.ArgumentList {
		fields = append(fields, JSON(field))
	}

	c.Logs = append(c.Logs, &pb_handler.LogEntry{
		Function: c.function,
		Fields:   fields,
	})
}

func (c *EntryLogger) Enter(function string) {
	c.function = function
}

type IgnoreLogger struct{}

var Ignore = &IgnoreLogger{}

func (c *IgnoreLogger) Log(call otto.FunctionCall) {}
func (c *IgnoreLogger) Enter(function string)      {}

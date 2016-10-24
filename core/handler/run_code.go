// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/robertkrimen/otto"
)

type Logger interface {
	Log(call otto.FunctionCall) otto.Value
}

type console struct {
	Logs [][]string
}

func JSON(val otto.Value) string {
	vm := otto.New()
	vm.Set("value", val)
	res, _ := vm.Run(`JSON.stringify(value)`)
	return res.String()
}

func (c *console) Log(call otto.FunctionCall) otto.Value {
	line := []string{}
	for _, arg := range call.ArgumentList {
		line = append(line, JSON(arg))
	}

	c.Logs = append(c.Logs, line)

	return otto.UndefinedValue()
}

type ignore struct{}

func (c *ignore) Log(call otto.FunctionCall) otto.Value {
	return otto.UndefinedValue()
}

func runUnsafeCodeWithLogger(vm *otto.Otto, code string, timeOut time.Duration, logger Logger) (otto.Value, error) {
	if logger == nil {
		logger = &ignore{}
	}

	vm.Set("__log", logger.Log)
	vm.Run("console.log = __log")
	start := time.Now()

	var value otto.Value
	var err error

	defer func() {
		duration := time.Since(start)
		if caught := recover(); caught != nil {
			if caught == errTimeOutExceeded {
				value = otto.Value{}
				err = errors.NewErrInternal(fmt.Sprintf("Interrupted javascript execution after %v", duration))
				return
			}
			// if this is not the our timeout interrupt, raise the panic again
			// so someone else can handle it
			panic(caught)
		}
	}()

	vm.Interrupt = make(chan func(), 1)

	go func() {
		time.Sleep(timeOut)
		vm.Interrupt <- func() {
			panic(errTimeOutExceeded)
		}
	}()
	val, err := vm.Run(code)

	return val, err
}

func runUnsafeCode(vm *otto.Otto, code string, timeOut time.Duration) (otto.Value, error) {
	return runUnsafeCodeWithLogger(vm, code, timeOut, &ignore{})
}

func runUnsafeCodeWithLogs(vm *otto.Otto, code string, timeOut time.Duration) (otto.Value, [][]string, error) {
	console := &console{}
	val, err := runUnsafeCodeWithLogger(vm, code, timeOut, console)
	return val, console.Logs, err
}

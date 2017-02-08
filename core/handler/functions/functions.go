// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package functions

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/robertkrimen/otto"
)

var errTimeOutExceeded = errors.NewErrInternal("Code has been running to long")

func RunCode(name, code string, env map[string]interface{}, timeout time.Duration, logger Logger) (val otto.Value, err error) {
	vm := otto.New()

	// load the environment
	for key, val := range env {
		vm.Set(key, val)
	}

	if logger == nil {
		logger = Ignore
	}
	logger.Enter(name)

	vm.Set("__log", func(call otto.FunctionCall) otto.Value {
		logger.Log(call)
		return otto.UndefinedValue()
	})
	vm.Run("console.log = __log")

	start := time.Now()

	defer func() {
		duration := time.Since(start)
		if caught := recover(); caught != nil {
			val = otto.Value{}
			if caught == errTimeOutExceeded {
				err = errors.NewErrInternal(fmt.Sprintf("Interrupted javascript execution for %s after %v", name, duration))
				return
			} else {
				err = errors.NewErrInternal(fmt.Sprintf("Fatal error in %s: %s", name, caught))
			}
		}
	}()

	vm.Interrupt = make(chan func(), 1)

	go func() {
		time.Sleep(timeout)
		vm.Interrupt <- func() {
			panic(errTimeOutExceeded)
		}
	}()

	val, err = vm.Run(code)
	if err != nil {
		return val, errors.NewErrInternal(fmt.Sprintf("%s threw error: %s", name, err))
	}

	return val, nil
}

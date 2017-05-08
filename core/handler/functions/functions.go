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

func RunCode(name, code string, env map[string]interface{}, timeout time.Duration, logger Logger) (val interface{}, err error) {
	vm := otto.New()
	vm.SetStackDepthLimit(32)

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
			val = nil
			switch {
			case caught == errTimeOutExceeded:
				err = errors.NewErrInternal(fmt.Sprintf("Interrupted javascript execution for %s after %v", name, duration))
			default:
				err = errors.NewErrInternal(fmt.Sprintf("Fatal error in %s: %s", name, caught))
			}
			return
		}
	}()

	vm.Interrupt = make(chan func(), 1)

	go func() {
		time.Sleep(timeout)
		vm.Interrupt <- func() {
			panic(errTimeOutExceeded)
		}
	}()

	oVal, err := vm.Run(code)
	if err != nil {
		return nil, errors.NewErrInternal(fmt.Sprintf("%s threw error: %s", name, err))
	}

	switch {
	case oVal.IsBoolean():
		return oVal.ToBoolean()
	case oVal.IsNull(), oVal.IsUndefined():
		return nil, nil
	case oVal.IsNumber():
		f, _ := oVal.ToFloat()
		if float64(int64(f)) == f {
			return oVal.ToInteger()
		}
		return f, nil
	case oVal.IsObject():
		return oVal.Export()
	case oVal.IsString():
		return oVal.ToString()
	}

	return nil, errors.NewErrInternal(fmt.Sprintf("%s return value invalid", name))
}

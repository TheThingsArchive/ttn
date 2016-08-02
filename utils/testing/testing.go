// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package testing offers some handy methods to display check and cross symbols with colors in test
// logs.
package testing

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/apex/log"
)

func GetLogger(t *testing.T, tag string) log.Interface {
	logger := &log.Logger{
		Handler: NewLogHandler(t),
		Level:   log.DebugLevel,
	}
	return logger.WithField("tag", tag)
}

// WaitGroup is an extension of sync.WaitGroup with a WaitFor function for testing
type WaitGroup struct {
	sync.WaitGroup
}

// WaitFor waits for the specified duration
func (wg *WaitGroup) WaitFor(d time.Duration) error {
	waitChan := make(chan bool)
	go func() {
		wg.Wait()
		fmt.Println("WG DONE")
		waitChan <- true
		close(waitChan)
	}()
	select {
	case <-waitChan:
		return nil
	case <-time.After(d):
		return errors.New("Wait timeout expired")
	}
}

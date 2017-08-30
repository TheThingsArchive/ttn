// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package testing offers some handy methods to display check and cross symbols with colors in test
// logs.
package testing

import (
	"errors"
	"sync"
	"testing"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	ttnapex "github.com/TheThingsNetwork/go-utils/log/apex"
	apexlog "github.com/apex/log"
)

func GetLogger(t *testing.T, tag string) ttnlog.Interface {
	logger := &apexlog.Logger{
		Handler: NewLogHandler(t),
		Level:   apexlog.DebugLevel,
	}
	return ttnapex.Wrap(logger).WithField("tag", tag)
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

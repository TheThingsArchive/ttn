// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package utils

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// for keeping track of how many goroutines are waiting
var waiting int32

// WaitGroup is an extension of sync.WaitGroup that allows waiting with a maximum duration
type WaitGroup struct {
	sync.WaitGroup
}

// WaitForMax waits until the WaitGroup is Done or the specified duration has elapsed
func (wg *WaitGroup) WaitForMax(d time.Duration) error {
	waitChan := make(chan struct{})
	go func() {
		atomic.AddInt32(&waiting, 1)
		wg.Wait()
		atomic.AddInt32(&waiting, -1)
		close(waitChan)
	}()
	select {
	case <-waitChan:
		return nil
	case <-time.After(d):
		return errors.New("Wait timeout expired")
	}
}

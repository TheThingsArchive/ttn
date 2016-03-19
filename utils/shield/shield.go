// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package shield

import (
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Interfaces materializes the public API of a shield.
type Interface interface {
	ThroughIn() error // Try to pass through, will fail if the shield has been passed too many times
	ThroughOut()      // Terminate a previous session
}

type shield struct {
	Queue chan struct{}
}

// New constructs a new shield with the given size
func New(size uint) Interface {
	return shield{Queue: make(chan struct{}, size)}
}

// Through implements the shield Interface
func (s shield) ThroughIn() error {
	select {
	case s.Queue <- struct{}{}:
		return nil
	default:
		return errors.New(errors.Operational, "Impossible to pass. Too many requests.")
	}
}

// Release implements the shield Interface
func (s shield) ThroughOut() {
	select {
	case <-s.Queue:
	default:
	}
}

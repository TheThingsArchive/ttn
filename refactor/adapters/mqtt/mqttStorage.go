// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import ()

// Storage defines an interface for the mqtt adapter
type Storage interface {
	Push(topic string, data []byte) error
	Pull(topic string) ([]byte, error)
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broadcast

import (
	"github.com/TheThingsNetwork/ttn/core"
)

type voidAckNacker struct{}

// Ack implements the core.AckNacker interface
func (v voidAckNacker) Ack(p core.Packet) error {
	return nil
}

// Nack implements the core.AckNacker interface
func (v voidAckNacker) Nack(p core.Packet) error {
	return nil
}

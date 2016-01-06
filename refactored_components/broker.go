// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/thethingsnetwork/core"
)

type Broker struct{}

func (b *Broker) NextUp() (*core.Packet, error) {
	return nil, nil
}

func (b *Broker) NextDown() (*core.Packet, error) {
	return nil, nil
}

func (b *Broker) Handle(p core.Packet, an core.AckNacker) error {
	return nil
}

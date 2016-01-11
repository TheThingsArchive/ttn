// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"github.com/thethingsnetwork/core"
)

type Broker struct{}

func NewBroker() (*Broker, error) {
	return nil, nil
}
func (b *Broker) HandleUp(p core.Packet, an core.AckNacker, a core.Adapter) error {
	return nil
}

func (b *Broker) HandleDown(p core.Packet, an core.AckNacker, a core.Adapter) error {
	return fmt.Errorf("Not Implemented")
}

func (b *Broker) Register(r core.Registration) error {
	return nil
}

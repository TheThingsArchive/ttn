// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/utils/log"
)

type Broker struct {
	loggers []log.Logger
	db      addressKeeper
}

func NewBroker(loggers ...log.Logger) (*Broker, error) {
	localDB, err := NewLocalDB(EXPIRY_DELAY)

	if err != nil {
		return nil, err
	}

	return &Broker{
		loggers: loggers,
		db:      localDB,
	}, nil
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

func (b *Broker) log(format string, i ...interface{}) {
	for _, logger := range b.loggers {
		logger.Log(format, i...)
	}
}

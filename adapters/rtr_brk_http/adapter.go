// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package rtr_brk_http
//
// Assume one endpoint url accessible through a POST http request
package rtr_brk_http

import (
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/utils/log"
)

type Adapter struct {
	router core.Router
	logger log.Logger
}

// New constructs a new Router-Broker-HTTP adapter
func New(router core.Router, broAddrs ...core.BrokerAddress) (*Adapter, error) {
	return nil, nil
}

// Connect implements the core.BrokerRouter interface
func (a *Adapter) Connect(router core.Router) {
	a.log("Connects to router %+v", router)
	a.router = router
}

// Broadcast implements the core.BrokerRouter interface
func (a *Adapter) Broadcast(packet semtech.Packet) {

}

// Forward implements the core.BrokerRouter interface
func (a *Adapter) Forward(packet semtech.Packet, broAddrs ...core.BrokerAddress) {

}

// log is nothing more than a shortcut / helper to access the logger
func (a Adapter) log(format string, i ...interface{}) {
	a.logger.Log(format, i...)
}

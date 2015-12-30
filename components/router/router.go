// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/utils/log"
	"time"
)

const (
	EXPIRY_DELAY = time.Hour * 8
)

type Router struct {
	PortUDP       uint
	PortHTTP      uint
	Logger        log.Logger
	addressKeeper addressKeeper
}

func New(portUDP, portHTTP uint) (*Router, error) {
	localDB, err := NewLocalDB(EXPIRY_DELAY)

	if err != nil {
		return nil, error
	}

	return &Router{
		PortUDP:       portUDP,
		PortHTTP:      portHTTP,
		addressKeeper: localDB,
		Logger:        log.VoidLogger{},
	}, nil
}

// HandleUplink implements the core.Router interface
func (r *Router) HandleUplink(upAdapter core.GatewayRouterAdapter, downAdapter core.RouterBrokerAdapter, packet semtech.Packet, gateway core.GatewayAddress) {
	switch packet.Identifier {
	case semtech.PULL_DATA:
		r.log("receives PULL_DATA, sending ack")
		upAdapter.Ack(r, semtech.Packet{
			Version:    semtech.VERSION,
			Identifier: semtech.PULL_ACK,
			Token:      packet.Token,
		}, gateway)
	case semtech.PUSH_DATA:
		// 1. Send an ack
		r.log("receives PUSH_DATA, sending ack")
		upAdapter.Ack(r, semtech.Packet{
			Version:    semtech.VERSION,
			Identifier: semtech.PUSH_ACK,
			Token:      packet.Token,
		}, gateway)

		// 2. Determine payloads related to different end-devices present in the packet
		// NOTE So far, Stats are ignored.
		if packet.Payload == nil || len(packet.Payload.RXPK) == 0 {
			r.log("Ignores inconsistent PUSH_DATA packet")
			return
		}

		payloads = make(map[semtech.DeviceAddress]semtech.Payload)
		for _, rxpk := range packet.Payload.RXPK {
			devAddr := rxpk.DevAddr()
			if devAddr == nil {
				r.log("Unable to determine end-device address for rxpk: %+v", rxpk)
				continue
			}

			if _, ok := payloads[*devAddr]; !ok {
				payloads[*devAddr] = semtech.Payload{
					RXPK: make([]semtech.RXPK, 0),
				}
			}

			payloads[*devAddr].RXPK = append(payloads[*devAddr].RXPK, rxpk)
		}

		// 3. Broadcast or Forward payloads depending wether or not the brokers are known
		for payload, devAddr := range payloads {
			brokers, err := r.addressKeeper.lookup(devAddr)
			if err != nil {
				r.log("Forward payload to known brokers %+v", payload)
				downAdapter.Forward(router, payload, brokers...)
				continue
			}

			r.log("Broadcast payload to all brokers %+v", payload)
			downAdapter.Broadcast(router, payload)
		}
	default:
		r.log("Unexpected packet receive from uplink %+v", packet)

	}
}

// HandleDownlink implements the core.Router interface
func (r *Router) HandleDownlink(downAdapter core.RouterBrokerAdapter, packet semtech.Packet, broker core.BrokerAddress) {
	// TODO MileStone 4
}

// RegisterDevice implements the core.Router interface
func (r *Router) RegisterDevice(devAddr core.DeviceAddress, broAddrs ...core.BrokerAddress) {
	// TODO MileStone 4
}

// RegisterDevice implements the core.Router interface
func (r *Router) HandleError(err interface{}) {
	switch err.(type) {
	case core.ErrAck:
	case core.ErrDownlink:
	case core.ErrForward:
	case core.ErrBroadcast:
	case core.ErrUplink:
	default:
		fmt.Println(err) // Wow, much handling, very reliable
	}
}

func (r *Router) log(format string, a ...interface{}) {
	if r.Logger == nil {
		return
	}
	r.Logger.Log(format, a...)
}

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
	EXPIRY_DELAY   = time.Hour * 8
	UP_POOL_SIZE   = 1
	DOWN_POOL_SIZE = 1
)

// Router represents a concrete router of TTN architecture. Use the New() method to create a new
// one and then connect it to its adapters.
type Router struct {
	brokers       []core.BrokerAddress // Brokers known by the router
	Logger        log.Logger           // Specify a logger for the router. NOTE Having this exported isn't thread-safe.
	addressKeeper addressKeeper        // Local storage that maps end-device addresses to broker addresses
	up            chan upMsg           // Internal communication channel which sends data to the up adapter
	down          chan downMsg         // Internal communication channel which sends data to the down adapter
}

// upMsg materializes messages that flow along the up channel
type upMsg struct {
	packet  semtech.Packet      // The packet to transfer
	gateway core.GatewayAddress // The recipient gateway to reach
}

// downMsg materializes messages that flow along the down channel
type downMsg struct {
	payload semtech.Payload      // The payload to transfer
	brokers []core.BrokerAddress // The recipient broker to reach. If nil or empty, assume that all broker should be reached
}

// New constructs a Router and setup its internal structure
func New(brokers ...core.BrokerAddress) (*Router, error) {
	localDB, err := NewLocalDB(EXPIRY_DELAY)

	if err != nil {
		return nil, error
	}

	if len(brokers) == 0 {
		return nil, fmt.Errorf("The router should be connected to at least one broker")
	}

	return &Router{
		brokers:       brokers,
		addressKeeper: localDB,
		up:            make(chan upMsg),
		down:          make(chan downMsg),
		Logger:        log.VoidLogger{},
	}, nil
}

// HandleUplink implements the core.Router interface
func (r *Router) HandleUplink(packet semtech.Packet, gateway core.GatewayAddress) {
	r.ensure()

	switch packet.Identifier {
	case semtech.PULL_DATA:
		r.log("receives PULL_DATA, sending ack")
		r.up <- upMsg{
			packet: semtech.Packet{
				Version:    semtech.VERSION,
				Identifier: semtech.PULL_ACK,
				Token:      packet.Token,
			},
			gateway: gateway,
		}
	case semtech.PUSH_DATA:
		// 1. Send an ack
		r.log("receives PUSH_DATA, sending ack")
		r.up <- upMsg{
			packet: semtech.Packet{
				Version:    semtech.VERSION,
				Identifier: semtech.PUSH_ACK,
				Token:      packet.Token,
			},
			gateway: gateway,
		}

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
				r.down <- downMsg{
					payload: payload,
					brokers: brokers,
				}
				continue
			}

			r.log("Broadcast payload to all brokers %+v", payload)
			r.down <- downMsg{payload: payload}
		}
	default:
		r.log("Unexpected packet receive from uplink %+v", packet)

	}
}

// HandleDownlink implements the core.Router interface
func (r *Router) HandleDownlink(payload semtech.Payload, broker core.BrokerAddress) {
	// TODO MileStone 4
}

// RegisterDevice implements the core.Router interface
func (r *Router) RegisterDevice(devAddr semtech.DeviceAddress, broAddrs ...core.BrokerAddress) {
	r.ensure()
	r.addressKeeper.store(devAddr, broAddrs) // TODO handle the error
}

// RegisterDevice implements the core.Router interface
func (r *Router) HandleError(err interface{}) {
	r.ensure()

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

// Connect implements the core.Router interface
func (r *Router) Connect(upAdapter core.GatewayRouterAdapter, downAdapter core.RouterBrokerAdapter) error {
	r.ensure()

	for i := 0; i < UP_POOL_SIZE; i += 1 {
		go r.connectUpAdapter(upAdapter)
	}

	for i := 0; i < DOWN_POOL_SIZE; i += 1 {
		go r.connectDownAdapter(downAdapter)
	}

	return nil
}

// Consume messages sent to r.up channel
func (r *Router) connectUpAdapter(upAdapter core.GatewayRouterAdapter) {
	for msg := range r.up {
		upAdapter.Ack(r, msg.packet, msg.gateway)
	}
}

// Consume messages sent to r.down channel
func (r *Router) connectDownAdapter(downAdapter core.RouterBrokerAdapter) {
	for msg := range r.down {
		if len(msg.brokers) == 0 {
			downAdapter.Broadcast(r, msg.payload, r.Brokers...)
			continue
		}
		downAdapter.Forward(r, msg.payload, msg.Brokers...)
	}
}

// ensure checks whether or not the current Router has been created via New(). It panics if not.
func (r *Router) ensure() bool {
	if r == nil || r.addressKeeper == nil {
		panic("Call method on non-initialized Router")
	}
}

// log is a shortcut to access the router logger
func (r *Router) log(format string, a ...interface{}) {
	if r.Logger == nil {
		return
	}
	r.Logger.Log(format, a...)
}

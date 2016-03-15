// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"golang.org/x/net/context"
)

type adapter struct {
	ctx log.Interface
}

type replier func(data []byte) error

// Start constructs and launches a new udp adapter
func Start(bindNet string, router core.RouterServer, ctx log.Interface) error {
	// Create the udp connection and start listening with a goroutine
	var udpConn *net.UDPConn
	addr, err := net.ResolveUDPAddr("udp", bindNet)
	if udpConn, err = net.ListenUDP("udp", addr); err != nil {
		ctx.WithError(err).Error("Unable to start server")
		return errors.New(errors.Operational, fmt.Sprintf("Invalid bind address %v", bindNet))
	}

	ctx.WithField("bind", bindNet).Info("Starting Server")

	a := adapter{ctx: ctx}
	go a.listen(udpConn, router)
	return nil
}

// makeReply curryfies a writing to udp connection by binding the address and connection
func makeReply(addr *net.UDPAddr, conn *net.UDPConn) replier {
	return func(data []byte) error {
		_, err := conn.WriteToUDP(data, addr)
		return err
	}
}

// listen Handle incoming packets and forward them.
//
// Runs in its own goroutine.
func (a adapter) listen(conn *net.UDPConn, router core.RouterServer) {
	defer conn.Close()
	a.ctx.WithField("address", conn.LocalAddr()).Debug("Starting accept loop")

	for {
		buf := make([]byte, 5000)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil { // Problem with the connection
			a.ctx.WithError(err).Error("Connection error")
			continue
		}

		a.ctx.Debug("Incoming datagram")
		go func(data []byte, reply replier, router core.RouterServer) {
			pkt := new(semtech.Packet)
			if err := pkt.UnmarshalBinary(data); err != nil {
				a.ctx.WithError(err).Debug("Unable to handle datagram")
			}

			switch pkt.Identifier {
			case semtech.PULL_DATA:
				err = a.handlePullData(*pkt, reply)
			case semtech.PUSH_DATA:
				err = a.handlePushData(*pkt, reply, router)
			default:
				err = errors.New(errors.Implementation, "Unhandled packet type")
			}

			if err != nil {
				a.ctx.WithError(err).Debug("Unable to handle datagram")
			}
		}(buf[:n], makeReply(addr, conn), router)
	}
}

// Handle a PULL_DATA packet coming from a gateway
func (a adapter) handlePullData(pkt semtech.Packet, reply replier) error {
	stats.MarkMeter("semtech_adapter.pull_data")
	stats.MarkMeter(fmt.Sprintf("semtech_adapter.gateways.%X.pull_data", pkt.GatewayId))
	stats.SetString(fmt.Sprintf("semtech_adapter.gateways.%X.last_pull_data", pkt.GatewayId), "date", time.Now().UTC().Format(time.RFC3339))

	data, err := semtech.Packet{
		Version:    semtech.VERSION,
		Token:      pkt.Token,
		Identifier: semtech.PULL_ACK,
	}.MarshalBinary()

	if err != nil {
		return errors.New(errors.Structural, err)
	}

	return reply(data)
}

// Handle a PUSH_DATA packet coming from a gateway
func (a adapter) handlePushData(pkt semtech.Packet, reply replier, router core.RouterServer) error {
	stats.MarkMeter("semtech_adapter.push_data")
	stats.MarkMeter(fmt.Sprintf("semtech_adapter.gateways.%X.push_data", pkt.GatewayId))
	stats.SetString(fmt.Sprintf("semtech_adapter.gateways.%X.last_push_data", pkt.GatewayId), "date", time.Now().UTC().Format(time.RFC3339))

	// AckNowledge with a PUSH_ACK
	data, err := semtech.Packet{
		Version:    semtech.VERSION,
		Token:      pkt.Token,
		Identifier: semtech.PUSH_ACK,
	}.MarshalBinary()
	if err != nil || reply(data) != nil || pkt.Payload == nil {
		return errors.New(errors.Operational, "Unable to process PUSH_DATA packet")
	}

	// Process Stat payload
	if pkt.Payload.Stat != nil {
		go router.HandleStats(context.Background(), &core.StatsReq{
			GatewayID: pkt.GatewayId,
			Metadata:  extractMetadata(*pkt.Payload.Stat, new(core.StatsMetadata)).(*core.StatsMetadata),
		})
	}

	// Process rxpks payloads
	cherr := make(chan error, len(pkt.Payload.RXPK))
	wait := sync.WaitGroup{}
	wait.Add(len(pkt.Payload.RXPK))
	for _, rxpk := range pkt.Payload.RXPK {
		go func(rxpk semtech.RXPK) {
			defer wait.Done()
			if err := a.handleDataUp(rxpk, pkt.GatewayId, reply, router); err != nil {
				cherr <- err
			}
		}(rxpk)
	}

	// Retrieve any errors
	wait.Wait()
	close(cherr)
	return <-cherr
}

func (a adapter) handleDataUp(rxpk semtech.RXPK, gid []byte, reply replier, router core.RouterServer) error {
	dataRouterReq, err := a.newDataRouterReq(rxpk, gid)
	if err != nil {
		return errors.New(errors.Structural, err)
	}
	resp, err := router.HandleData(context.Background(), dataRouterReq)
	if err != nil {
		errors.New(errors.Operational, err)
	}
	return a.handleDataDown(resp, reply)
}

func (a adapter) handleDataDown(resp *core.DataRouterRes, reply replier) error {
	if resp == nil { // No response
		return nil
	}

	txpk, err := a.newTXPK(*resp)
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	data, err := semtech.Packet{
		Version:    semtech.VERSION,
		Identifier: semtech.PULL_RESP,
		Payload:    &semtech.Payload{TXPK: &txpk},
	}.MarshalBinary()
	if err != nil {
		return errors.New(errors.Structural, err)
	}
	return reply(data)
}

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
	Components
	pool *gatewayPool
}

// Components defines a structure to make the instantiation easier to read
type Components struct {
	Ctx    log.Interface
	Router core.RouterServer
}

// Options defines a structure to make the instantiation easier to read
type Options struct {
	// NetAddr refers to the udp address + port the adapter will have to listen
	NetAddr string
	// MaxReconnectionDelay defines the delay of the longest attempt to reconnect a lost connection
	// before giving up.
	MaxReconnectionDelay time.Duration
}

// straightMarshaler can be used to easily obtain a binary marshaler from a sequence of byte
type straightMarshaler []byte

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (s straightMarshaler) MarshalBinary() ([]byte, error) {
	return s, nil
}

// Start constructs and launches a new udp adapter
func Start(c Components, o Options) (err error) {
	// Create the udp connection and start listening with a goroutine
	var udpConn *net.UDPConn
	if udpConn, err = tryConnect(o.NetAddr); err != nil {
		c.Ctx.WithError(err).Error("Unable to start server")
		return errors.New(errors.Operational, fmt.Sprintf("Invalid bind address %v", o.NetAddr))
	}

	c.Ctx.WithField("bind", o.NetAddr).Info("Starting Server")
	a := adapter{Components: c, pool: newPool()}
	go a.listen(o.NetAddr, udpConn, o.MaxReconnectionDelay)
	return nil
}

// tryConnect attempt to connect to a udp connection
func tryConnect(netAddr string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", netAddr)
	if err != nil {
		return nil, err
	}
	return net.ListenUDP("udp", addr)
}

// listen Handle incoming packets and forward them.
//
// Runs in its own goroutine.
func (a adapter) listen(netAddr string, conn *net.UDPConn, maxReconnectionDelay time.Duration) {
	var err error
	a.Ctx.WithField("address", conn.LocalAddr()).Debug("Starting accept loop")
	for {
		// Read on the UDP connection
		buf := make([]byte, 5000)
		n, addr, fault := conn.ReadFromUDP(buf)
		if fault != nil { // Problem with the connection
			for delay := time.Millisecond * 25; delay < maxReconnectionDelay; delay *= 10 {
				a.Ctx.Infof("UDP connection lost. Trying to reconnect in %s", delay)
				<-time.After(delay)
				conn, err = tryConnect(netAddr)
				if err == nil {
					a.Ctx.Info("UDP connection recovered")
					break
				}
			}
			a.Ctx.WithError(fault).Error("Unable to restore UDP connection")
			break
		}

		ctx := a.Ctx.WithField("Source", addr.IP.String())

		// Handle the incoming datagram
		go func(data []byte, conn *net.UDPConn) {
			pkt := new(semtech.Packet)
			if err := pkt.UnmarshalBinary(data); err != nil {
				ctx.WithError(err).Debug("Invalid datagram")
			}

			gtwConn := a.pool.GetOrCreate(pkt.GatewayId)
			gtwConn.SetConn(conn)

			switch pkt.Identifier {
			case semtech.PULL_DATA:
				gtwConn.SetDownlinkAddr(addr)
				err = a.handlePullData(*pkt, gtwConn)
			case semtech.PUSH_DATA:
				gtwConn.SetUplinkAddr(addr)
				err = a.handlePushData(*pkt, gtwConn)
			default:
				err = errors.New(errors.Implementation, "Unhandled packet type")
			}
			if err != nil {
				ctx.WithError(err).Debug("Unable to handle datagram")
			}
		}(buf[:n], conn)
	}

	if conn != nil {
		_ = conn.Close()
	}
}

// Handle a PULL_DATA packet coming from a gateway
func (a adapter) handlePullData(pkt semtech.Packet, reply replier) error {
	stats.MarkMeter("semtech_adapter.pull_data")
	stats.MarkMeter(fmt.Sprintf("semtech_adapter.gateways.%X.pull_data", pkt.GatewayId))
	stats.SetString(fmt.Sprintf("semtech_adapter.gateways.%X.last_pull_data", pkt.GatewayId), "date", time.Now().UTC().Format(time.RFC3339))
	ctx := a.Ctx.WithField("GatewayID", pkt.GatewayId)
	ctx.Debug("Handle PULL_DATA")

	data, err := semtech.Packet{
		Version:    semtech.VERSION,
		Token:      pkt.Token,
		Identifier: semtech.PULL_ACK,
	}.MarshalBinary()

	if err != nil || reply.WriteToDownlink(data) != nil {
		ctx.Debug("Unable to send PULL_ACK")
		return errors.New(errors.Operational, "Unable to send PULL_ACK")
	}

	return nil
}

// Handle a PUSH_DATA packet coming from a gateway
func (a adapter) handlePushData(pkt semtech.Packet, reply replier) error {
	stats.MarkMeter("semtech_adapter.push_data")
	stats.MarkMeter(fmt.Sprintf("semtech_adapter.gateways.%X.push_data", pkt.GatewayId))
	stats.SetString(fmt.Sprintf("semtech_adapter.gateways.%X.last_push_data", pkt.GatewayId), "date", time.Now().UTC().Format(time.RFC3339))
	ctx := a.Ctx.WithField("GatewayID", pkt.GatewayId)
	ctx.Debug("Handle PUSH_DATA")

	// AckNowledge with a PUSH_ACK
	data, err := semtech.Packet{
		Version:    semtech.VERSION,
		Token:      pkt.Token,
		Identifier: semtech.PUSH_ACK,
	}.MarshalBinary()

	if err != nil || reply.WriteToUplink(data) != nil || pkt.Payload == nil {
		ctx.Debug("Unable to send PUSH_ACK")
		return errors.New(errors.Operational, "Unable to send PUSH_ACK")
	}

	// Process Stat payload
	if pkt.Payload.Stat != nil {
		ctx.Debug("Handle stat")
		go a.Router.HandleStats(context.Background(), &core.StatsReq{
			GatewayID: pkt.GatewayId,
			Metadata:  extractMetadata(*pkt.Payload.Stat, new(core.StatsMetadata)).(*core.StatsMetadata),
		})
	}

	// Process rxpks payloads
	wait := sync.WaitGroup{}

	if len(pkt.Payload.RXPK) > 0 {
		wait.Add(len(pkt.Payload.RXPK))
		ctx.Debug("Handle rxpk")
		for _, rxpk := range pkt.Payload.RXPK {
			go func(rxpk semtech.RXPK) {
				defer wait.Done()
				if err := a.handleDataUp(rxpk, pkt.GatewayId, reply); err != nil {
					ctx.WithError(err).Debug("rxpk not processed")
				}
			}(rxpk)
		}
	}

	// Retrieve any errors
	wait.Wait()
	return nil
}

func (a adapter) handleDataUp(rxpk semtech.RXPK, gid []byte, reply replier) error {
	ctx := a.Ctx.WithField("GatewayID", gid)

	itf, err := toLoRaWANPayload(rxpk, gid, ctx)
	if err != nil {
		ctx.WithError(err).Debug("Invalid up RXPK packet")
		return errors.New(errors.Structural, err)
	}

	switch itf.(type) {
	case *core.DataRouterReq:
		ctx.Debug("Handle uplink data")
		resp, err := a.Router.HandleData(context.Background(), itf.(*core.DataRouterReq))
		if err != nil {
			return errors.New(errors.Operational, err)
		}
		return a.handleDataDown(resp, reply)
	case *core.JoinRouterReq:
		ctx.Debug("Handle join request")
		resp, err := a.Router.HandleJoin(context.Background(), itf.(*core.JoinRouterReq))
		if err != nil {
			return errors.New(errors.Operational, err)
		}
		return a.handleJoinAccept(resp, reply)
	default:
		ctx.Warn("Unhandled LoRaWAN Payload type")
		return errors.New(errors.Implementation, "Unhandled LoRaWAN Payload type")
	}
}

func (a adapter) handleDataDown(resp *core.DataRouterRes, reply replier) error {
	ctx := a.Ctx.WithField("GatewayID", reply.DestinationID())

	if resp == nil || resp.Payload == nil { // No response
		ctx.Debug("No response to send")
		return nil
	}

	payload, err := core.NewLoRaWANData(resp.Payload, false)
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	txpk, err := newTXPK(payload, resp.Metadata, ctx)
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	ctx.WithFields(log.Fields{
		"Metadata": resp.Metadata,
		"DevAddr":  resp.Payload.MACPayload.FHDR.DevAddr,
	}).Debug("Send txpk")

	data, err := semtech.Packet{
		Version:    semtech.VERSION,
		Identifier: semtech.PULL_RESP,
		Payload:    &semtech.Payload{TXPK: &txpk},
	}.MarshalBinary()
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	return reply.WriteToDownlink(data)
}

func (a adapter) handleJoinAccept(resp *core.JoinRouterRes, reply replier) error {
	ctx := a.Ctx.WithField("GatewayID", reply.DestinationID())

	if resp == nil || resp.Payload == nil {
		return errors.New(errors.Structural, "Invalid Join-Accept response. Expected a payload")
	}

	txpk, err := newTXPK(straightMarshaler(resp.Payload.Payload), resp.Metadata, a.Ctx)
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	ctx.WithField("Metadata", resp.Metadata).Debug("Send join-accept")

	data, err := semtech.Packet{
		Version:    semtech.VERSION,
		Identifier: semtech.PULL_RESP,
		Payload:    &semtech.Payload{TXPK: &txpk},
	}.MarshalBinary()
	if err != nil {
		return errors.New(errors.Structural, err)
	}
	return reply.WriteToDownlink(data)
}

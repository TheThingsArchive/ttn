// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/udp"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/brocaar/lorawan"
)

// Semtech implements the Semtech protocol and make a bridge between gateways and routers
type Semtech struct{}

// Handle implements the udp.Handler interface
func (s Semtech) Handle(conn chan<- udp.MsgUDP, packets chan<- udp.MsgReq, msg udp.MsgUDP) error {
	pkt := new(semtech.Packet)
	err := pkt.UnmarshalBinary(msg.Data)
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	switch pkt.Identifier {
	case semtech.PULL_DATA: // PULL_DATA -> Respond to the recipient with an ACK
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
		conn <- udp.MsgUDP{
			Addr: msg.Addr,
			Data: data,
		}
	case semtech.PUSH_DATA: // PUSH_DATA -> Transfer all RXPK to the component
		stats.MarkMeter("semtech_adapter.push_data")
		stats.MarkMeter(fmt.Sprintf("semtech_adapter.gateways.%X.push_data", pkt.GatewayId))
		stats.SetString(fmt.Sprintf("semtech_adapter.gateways.%X.last_push_data", pkt.GatewayId), "date", time.Now().UTC().Format(time.RFC3339))

		data, err := semtech.Packet{
			Version:    semtech.VERSION,
			Token:      pkt.Token,
			Identifier: semtech.PUSH_ACK,
		}.MarshalBinary()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		conn <- udp.MsgUDP{
			Addr: msg.Addr,
			Data: data,
		}

		if pkt.Payload == nil {
			return errors.New(errors.Structural, "Unable to process empty PUSH_DATA payload")
		}

		// Handle stat payload
		if pkt.Payload.Stat != nil {
			spacket, err := core.NewSPacket(pkt.GatewayId, extractMetadata(*pkt.Payload.Stat))
			if err == nil {
				data, err := spacket.MarshalBinary()
				if err == nil {
					go func() {
						packets <- udp.MsgReq{Data: data, Chresp: nil}
					}()
				}
			}
		}

		// Handle rxpks payloads
		cherr := make(chan error, len(pkt.Payload.RXPK))
		wait := sync.WaitGroup{}
		wait.Add(len(pkt.Payload.RXPK))
		for _, rxpk := range pkt.Payload.RXPK {
			go func(rxpk semtech.RXPK) {
				defer wait.Done()
				pktOut, err := rxpk2packet(rxpk, pkt.GatewayId)
				if err != nil {
					cherr <- errors.New(errors.Structural, err)
					return
				}
				data, err := pktOut.MarshalBinary()
				if err != nil {
					cherr <- errors.New(errors.Structural, err)
					return
				}
				chresp := make(chan udp.MsgRes)
				packets <- udp.MsgReq{Data: data, Chresp: chresp}
				select {
				case resp, ok := <-chresp:
					if !ok {
						// No response
						return
					}

					itf, err := core.UnmarshalPacket(resp)
					if err != nil {
						cherr <- errors.New(errors.Structural, err)
						return
					}
					pkt, ok := itf.(core.RPacket) // NOTE Here we'll handle join-accept
					if !ok {
						cherr <- errors.New(errors.Structural, "Unhandled packet type")
						return
					}
					txpk, err := packet2txpk(pkt)
					if err != nil {
						cherr <- errors.New(errors.Structural, err)
						return
					}

					data, err := semtech.Packet{
						Version:    semtech.VERSION,
						Identifier: semtech.PULL_RESP,
						Payload:    &semtech.Payload{TXPK: &txpk},
					}.MarshalBinary()
					if err != nil {
						cherr <- errors.New(errors.Structural, err)
						return
					}
					conn <- udp.MsgUDP{Addr: msg.Addr, Data: data}
				case <-time.After(time.Second * 2):
				}
			}(rxpk)
		}
		wait.Wait()
		close(cherr)
		if err := <-cherr; err != nil {
			return err
		}
	default:
		return errors.New(errors.Implementation, "Unhandled packet type")
	}
	return nil
}

func rxpk2packet(p semtech.RXPK, gid []byte) (core.Packet, error) {
	// First, we have to get the physical payload which is encoded in the Data field
	if p.Data == nil {
		return nil, errors.New(errors.Structural, "There's no data in the packet")
	}

	// RXPK Data are base64 encoded, yet without the trailing "==" if any.....
	encoded := *p.Data
	switch len(encoded) % 4 {
	case 2:
		encoded += "=="
	case 3:
		encoded += "="
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	payload := lorawan.NewPHYPayload(true)
	if err = payload.UnmarshalBinary(raw); err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	// At the end, our converted packet hold the same metadata than the RXPK packet but the Data
	// which as been completely transformed into a lorawan Physical Payload.
	return core.NewRPacket(payload, gid, extractMetadata(p))
}

func packet2txpk(p core.RPacket) (semtech.TXPK, error) {
	// Step 1, convert the physical payload to a base64 string (without the padding)
	raw, err := p.Payload().MarshalBinary()
	if err != nil {
		return semtech.TXPK{}, errors.New(errors.Structural, err)
	}

	data := strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")
	txpk := semtech.TXPK{Data: pointer.String(data)}

	// Step 2, copy every compatible metadata from the packet to the TXPK packet.
	// We are possibly loosing information here.
	injectMetadata(&txpk, p.Metadata())

	return txpk, nil
}

func injectMetadata(ptr interface{}, metadata core.Metadata) {
	v := reflect.ValueOf(metadata)
	t := v.Type()
	d := reflect.ValueOf(ptr).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i).Name
		if d.FieldByName(field).CanSet() {
			d.FieldByName(field).Set(v.Field(i))
		}
	}
}

func extractMetadata(xpk interface{}) core.Metadata {
	metadata := core.Metadata{}
	v := reflect.ValueOf(xpk)
	t := v.Type()
	m := reflect.ValueOf(&metadata).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i).Name
		if m.FieldByName(field).CanSet() {
			m.FieldByName(field).Set(v.Field(i))
		}
	}

	return metadata
}

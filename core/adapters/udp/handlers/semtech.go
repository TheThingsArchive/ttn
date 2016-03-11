// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
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

		for _, rxpk := range pkt.Payload.RXPK {
			go func(rxpk semtech.RXPK) {
				pktOut, err := rxpk2packet(rxpk, pkt.GatewayId)
				if err != nil {
					// TODO Log error
					return
				}
				data, err := pktOut.MarshalBinary()
				if err != nil {
					// TODO Log error
					return
				}
				chresp := make(chan udp.MsgRes)
				packets <- udp.MsgReq{Data: data, Chresp: chresp}
				select {
				case resp := <-chresp:
					itf, err := core.UnmarshalPacket(resp)
					if err != nil {
						return
					}
					pkt, ok := itf.(core.RPacket) // NOTE Here we'll handle join-accept
					if !ok {
						return
					}
					txpk, err := packet2txpk(pkt)
					if err != nil {
						// TODO Log error
						return
					}

					data, err := semtech.Packet{
						Version:    semtech.VERSION,
						Identifier: semtech.PULL_RESP,
						Payload:    &semtech.Payload{TXPK: &txpk},
					}.MarshalBinary()
					if err != nil {
						// TODO Log error
						return
					}
					conn <- udp.MsgUDP{Addr: msg.Addr, Data: data}
				case <-time.After(time.Second * 2):
				}
			}(rxpk)
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

	// Then, we interpret every other known field as a metadata and store them into an appropriate
	// metadata object.
	metadata := core.Metadata{}
	rxpkValue := reflect.ValueOf(p)
	rxpkStruct := rxpkValue.Type()
	metas := reflect.ValueOf(&metadata).Elem()
	for i := 0; i < rxpkStruct.NumField(); i++ {
		field := rxpkStruct.Field(i).Name
		if metas.FieldByName(field).CanSet() {
			metas.FieldByName(field).Set(rxpkValue.Field(i))
		}
	}

	// At the end, our converted packet hold the same metadata than the RXPK packet but the Data
	// which as been completely transformed into a lorawan Physical Payload.
	return core.NewRPacket(payload, gid, metadata)
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
	metadataValue := reflect.ValueOf(p.Metadata())
	metadataStruct := metadataValue.Type()
	txpkStruct := reflect.ValueOf(&txpk).Elem()
	for i := 0; i < metadataStruct.NumField(); i++ {
		field := metadataStruct.Field(i).Name
		if txpkStruct.FieldByName(field).CanSet() {
			txpkStruct.FieldByName(field).Set(metadataValue.Field(i))
		}
	}

	return txpk, nil
}

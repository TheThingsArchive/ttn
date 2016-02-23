// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/base64"
	"reflect"
	"strings"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	core "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/refactor/adapters/udp"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/brocaar/lorawan"
)

type Semtech struct{}

// HandleNack implements the udp.Handler interface
func (s Semtech) HandleNack(chresp chan<- udp.HandlerMsg) {
	close(chresp) // There is no notion of nack in the semtech protocol
}

// HandleAck implements the udp.Handler interface
func (s Semtech) HandleAck(packet *core.Packet, chresp chan<- udp.HandlerMsg) {
	defer close(chresp)

	if packet == nil {
		return
	}

	// For the downlink, we have to send a PULL_RESP packet which hold a TXPK.
	txpk, err := packet2txpk(*packet)
	if err != nil {
		chresp <- udp.HandlerMsg{Type: udp.HANDLER_ERROR, Data: []byte(err.Error())}
		return
	}

	// Step 3, marshal the packet and send it to the gateway
	raw, err := semtech.Packet{
		Version:    semtech.VERSION,
		Identifier: semtech.PULL_RESP,
		Payload:    &semtech.Payload{TXPK: &txpk},
	}.MarshalBinary()

	if err != nil {
		chresp <- udp.HandlerMsg{Type: udp.HANDLER_ERROR, Data: []byte(err.Error())}
		return
	}

	chresp <- udp.HandlerMsg{Type: udp.HANDLER_RESP, Data: raw}
}

// HandleDatagram implements the udp.Handler interface
func (s Semtech) HandleDatagram(data []byte, chresp chan<- udp.HandlerMsg) {
	defer close(chresp)
	pkt := new(semtech.Packet)
	err := pkt.UnmarshalBinary(data)
	if err != nil {
		chresp <- udp.HandlerMsg{Type: udp.HANDLER_ERROR, Data: []byte(err.Error())}
		return
	}

	switch pkt.Identifier {
	case semtech.PULL_DATA: // PULL_DATA -> Respond to the recipient with an ACK
		stats.MarkMeter("semtech_adapter.pull_data")
		pullAck, err := semtech.Packet{
			Version:    semtech.VERSION,
			Token:      pkt.Token,
			Identifier: semtech.PULL_ACK,
		}.MarshalBinary()
		if err != nil {
			chresp <- udp.HandlerMsg{Type: udp.HANDLER_ERROR, Data: []byte(err.Error())}
			return
		}
		chresp <- udp.HandlerMsg{Type: udp.HANDLER_RESP, Data: pullAck}
	case semtech.PUSH_DATA: // PUSH_DATA -> Transfer all RXPK to the component
		stats.MarkMeter("semtech_adapter.push_data")
		pushAck, err := semtech.Packet{
			Version:    semtech.VERSION,
			Token:      pkt.Token,
			Identifier: semtech.PUSH_ACK,
		}.MarshalBinary()
		if err != nil {
			chresp <- udp.HandlerMsg{Type: udp.HANDLER_ERROR, Data: []byte(err.Error())}
			return
		}
		chresp <- udp.HandlerMsg{Type: udp.HANDLER_RESP, Data: pushAck}

		if pkt.Payload == nil {
			return
		}

		for _, rxpk := range pkt.Payload.RXPK {
			pktOut, err := rxpk2packet(rxpk)
			if err != nil {
				chresp <- udp.HandlerMsg{Type: udp.HANDLER_ERROR, Data: []byte(err.Error())}
				continue
			}
			rawPkt, err := pktOut.MarshalBinary()
			if err != nil {
				chresp <- udp.HandlerMsg{Type: udp.HANDLER_ERROR, Data: []byte(err.Error())}
				continue
			}
			chresp <- udp.HandlerMsg{Type: udp.HANDLER_OUT, Data: rawPkt}
		}
	default:
	}
}

func rxpk2packet(p semtech.RXPK) (core.Packet, error) {
	// First, we have to get the physical payload which is encoded in the Data field
	if p.Data == nil {
		return nil, errors.New(ErrInvalidStructure, "There's no data in the packet")
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
		return nil, errors.New(ErrInvalidStructure, err)
	}

	payload := lorawan.NewPHYPayload(true)
	if err = payload.UnmarshalBinary(raw); err != nil {
		return nil, errors.New(ErrInvalidStructure, err)
	}

	// Then, we interpret every other known field as a metadata and store them into an appropriate
	// metadata object.
	metadata := core.Metadata{}
	rxpkValue := reflect.ValueOf(p)
	rxpkStruct := rxpkValue.Type()
	metas := reflect.ValueOf(&metadata).Elem()
	for i := 0; i < rxpkStruct.NumField(); i += 1 {
		field := rxpkStruct.Field(i).Name
		if metas.FieldByName(field).CanSet() {
			metas.FieldByName(field).Set(rxpkValue.Field(i))
		}
	}

	// At the end, our converted packet hold the same metadata than the RXPK packet but the Data
	// which as been completely transformed into a lorawan Physical Payload.
	return core.NewRPacket(payload, metadata), nil
}

func packet2txpk(p core.Packet) (semtech.TXPK, error) {
	// Interpret the packet as a Router Packet.
	rPkt, ok := p.(core.RPacket)
	if !ok {
		return semtech.TXPK{}, errors.New(ErrInvalidStructure, "Unable to interpret packet as a RPacket")
	}

	// Step 1, convert the physical payload to a base64 string (without the padding)
	raw, err := rPkt.Payload().MarshalBinary()
	if err != nil {
		return semtech.TXPK{}, errors.New(ErrInvalidStructure, err)
	}

	data := strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")
	txpk := semtech.TXPK{Data: pointer.String(data)}

	// Step 2, copy every compatible metadata from the packet to the TXPK packet.
	// We are possibly loosing information here.
	metadataValue := reflect.ValueOf(rPkt.Metadata())
	metadataStruct := metadataValue.Type()
	txpkStruct := reflect.ValueOf(&txpk).Elem()
	for i := 0; i < metadataStruct.NumField(); i += 1 {
		field := metadataStruct.Field(i).Name
		if txpkStruct.FieldByName(field).CanSet() {
			txpkStruct.FieldByName(field).Set(metadataValue.Field(i))
		}
	}

	return txpk, nil
}

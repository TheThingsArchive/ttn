// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/brocaar/lorawan"
)

// DevAddr returns a lorawan device address associated to the packet if any
func (p Packet) DevAddr() (lorawan.DevAddr, error) {
	if p.Payload.MACPayload == nil {
		return lorawan.DevAddr{}, errors.NewFailure(ErrInvalidPacket, "MACPAyload should not be empty")
	}

	macpayload, ok := p.Payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return lorawan.DevAddr{}, errors.NewFailure(ErrInvalidPacket, "Packet does not carry a MACPayload")
	}

	return macpayload.FHDR.DevAddr, nil
}

// FCnt returns the frame counter of the given packet if any
func (p Packet) Fcnt() (uint32, error) {
	if p.Payload.MACPayload == nil {
		return 0, errors.NewFailure(ErrInvalidPacket, "MACPayload should not be empty")
	}

	macpayload, ok := p.Payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return 0, errors.NewFailure(ErrInvalidPacket, "Packet does not carry a MACPayload")
	}

	return macpayload.FHDR.FCnt, nil
}

// String returns a string representation of the packet. It implements the io.Stringer interface
func (p Packet) String() string {
	str := "Packet {"
	str += fmt.Sprintf("\n\t%s}", p.Metadata.String())
	str += fmt.Sprintf("\n\tPayload%+v\n}", p.Payload)
	return str
}

// ConvertRXPK create a core.Packet from a semtech.RXPK. It's an handy way to both decode the
// frame payload and retrieve associated metadata from that packet
func ConvertRXPK(p semtech.RXPK) (Packet, error) {
	// First, we have to get the physical payload which is encoded in the Data field
	packet := Packet{}
	if p.Data == nil {
		return packet, errors.NewFailure(ErrInvalidPacket, "There's no data in the packet")
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
		return packet, errors.NewFailure(ErrInvalidPacket, err)
	}

	payload := lorawan.NewPHYPayload(true)
	if err = payload.UnmarshalBinary(raw); err != nil {
		return packet, errors.NewFailure(ErrInvalidPacket, err)
	}

	// Then, we interpret every other known field as a metadata and store them into an appropriate
	// metadata object.
	metadata := Metadata{}
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
	return Packet{Metadata: metadata, Payload: payload}, nil
}

// ConvertToTXPK converts a core Packet to a semtech TXPK packet using compatible metadata.
func ConvertToTXPK(p Packet) (semtech.TXPK, error) {
	// Step 1, convert the physical payload to a base64 string (without the padding)
	raw, err := p.Payload.MarshalBinary()
	if err != nil {
		return semtech.TXPK{}, errors.NewFailure(ErrInvalidPacket, err)
	}
	data := strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")
	txpk := semtech.TXPK{Data: pointer.String(data)}

	// Step 2, copy every compatible metadata from the packet to the TXPK packet.
	// We are possibly loosing information here.
	metadataValue := reflect.ValueOf(p.Metadata)
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

// MarshalJSON implements the json.Marshaler interface
func (p Packet) MarshalJSON() ([]byte, error) {
	rawMetadata, err := json.Marshal(p.Metadata)
	if err != nil {
		return nil, errors.NewFailure(ErrInvalidPacket, err)
	}
	rawPayload, err := p.Payload.MarshalBinary()
	if err != nil {
		return nil, errors.NewFailure(ErrInvalidPacket, err)
	}
	strPayload := base64.StdEncoding.EncodeToString(rawPayload)
	return []byte(fmt.Sprintf(`{"payload":"%s","metadata":%s}`, strPayload, string(rawMetadata))), nil
}

// UnmarshalJSON impements the json.Marshaler interface
func (p *Packet) UnmarshalJSON(raw []byte) error {
	if p == nil {
		return errors.NewFailure(ErrInvalidPacket, "Cannot unmarshal a nil packet")
	}

	// The payload is a bit tricky to unmarshal as we do not know if its an uplink or downlink
	// packet. Thus, we'll assume it's an uplink packet (because that's the case most of the time)
	// and check whether or not the unmarshalling process was okay.
	var proxy struct {
		Payload  string `json:"payload"`
		Metadata Metadata
	}

	err := json.Unmarshal(raw, &proxy)
	if err != nil {
		return errors.NewFailure(ErrInvalidPacket, err)
	}

	rawPayload, err := base64.StdEncoding.DecodeString(proxy.Payload)
	if err != nil {
		return errors.NewFailure(ErrInvalidPacket, err)
	}

	payload := lorawan.NewPHYPayload(true) // true -> uplink
	if err := payload.UnmarshalBinary(rawPayload); err != nil {
		return errors.NewFailure(ErrInvalidPacket, err)
	}

	// Now, we check the nature of the decoded payload
	switch payload.MHDR.MType.String() {
	case "JoinAccept":
		fallthrough
	case "UnconfirmedDataDown":
		fallthrough
	case "ConfirmedDataDown":
		// JoinAccept, UnconfirmedDataDown and ConfirmedDataDown are all downlink messages.
		// We thus have to unmarshall properly
		payload = lorawan.NewPHYPayload(false) // false -> downlink
		if err := payload.UnmarshalBinary(rawPayload); err != nil {
			return errors.NewFailure(ErrInvalidPacket, err)
		}
	case "JoinRequest":
		fallthrough
	case "UnconfirmedDataUp":
		fallthrough
	case "ConfirmedDataUp":
		// JoinRequest, UnconfirmedDataUp and ConfirmedDataUp are all uplink messages.
		// There's nothing to do, we've already handled them.

	case "Proprietary":
		// Proprietary can be either downlink or uplink. Right now, we do not have any message of
		// that type and thus, we just don't know how to handle them. Let's throw an error.
		return errors.NewFailure(ErrInvalidPacket, "Unsupported MType 'Proprietary'")
	}

	// Packet = Payload + Metadata
	p.Payload = payload
	p.Metadata = proxy.Metadata
	return nil
}

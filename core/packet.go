// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/brocaar/lorawan"
)

var ErrImpossibleConversion error = fmt.Errorf("Illegal attempt to convert a packet")

// DevAddr return a lorawan device address associated to the packet if any
func (p Packet) DevAddr() (lorawan.DevAddr, error) {
	if p.Payload.MACPayload == nil {
		return lorawan.DevAddr{}, fmt.Errorf("lorawan: MACPayload should not be empty")
	}

	macpayload, ok := p.Payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return lorawan.DevAddr{}, fmt.Errorf("lorawan: unable to get address of a join message")
	}

	return macpayload.FHDR.DevAddr, nil
}

// String returns a string representation of the packet
func (p Packet) String() string {
	str := "Packet {"
	str += fmt.Sprintf("\n\t%s}", p.Metadata.String())
	str += fmt.Sprintf("\n\tPayload%+v\n}", p.Payload)
	return str
}

// ConvertRXPK create a core.Packet from a semtech.RXPK. It's an handy way to both decode the
// frame payload and retrieve associated metadata from that packet
func ConvertRXPK(p semtech.RXPK) (Packet, error) {
	packet := Packet{}
	if p.Data == nil {
		return packet, ErrImpossibleConversion
	}

	encoded := *p.Data
	switch len(encoded) % 4 {
	case 2:
		encoded += "=="
	case 3:
		encoded += "="
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return packet, err
	}

	payload := lorawan.NewPHYPayload(true)
	if err = payload.UnmarshalBinary(raw); err != nil {
		return packet, err
	}

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

	return Packet{Metadata: metadata, Payload: payload}, nil
}

// ConvertToTXPK converts a core Packet to a semtech TXPK packet using compatible metadata.
func ConvertToTXPK(p Packet) (semtech.TXPK, error) {
	raw, err := p.Payload.MarshalBinary()
	if err != nil {
		return semtech.TXPK{}, ErrImpossibleConversion
	}
	data := strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")

	txpk := semtech.TXPK{Data: pointer.String(data)}

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
		return nil, err
	}
	rawPayload, err := p.Payload.MarshalBinary()
	if err != nil {
		return nil, err
	}
	strPayload := base64.StdEncoding.EncodeToString(rawPayload)
	return []byte(fmt.Sprintf(`{"payload":"%s","metadata":%s}`, strPayload, string(rawMetadata))), nil
}

// UnmarshalJSON impements the json.Marshaler interface
func (p *Packet) UnmarshalJSON(raw []byte) error {
	if p == nil {
		return ErrImpossibleConversion
	}
	var proxy struct {
		Payload  string `json:"payload"`
		Metadata Metadata
	}
	err := json.Unmarshal(raw, &proxy)
	if err != nil {
		return err
	}
	rawPayload, err := base64.StdEncoding.DecodeString(proxy.Payload)
	if err != nil {
		return err
	}

	// Try first to unmarshal as an uplink payload
	payload := lorawan.NewPHYPayload(true)
	if err := payload.UnmarshalBinary(rawPayload); err != nil {
		return err
	}

	switch payload.MHDR.MType.String() {
	case "JoinAccept":
		fallthrough
	case "UnconfirmedDataDown":
		fallthrough
	case "ConfirmedDataDown":
		payload = lorawan.NewPHYPayload(false)
		if err := payload.UnmarshalBinary(rawPayload); err != nil {
			return err
		}
	case "JoinRequest":
		fallthrough
	case "UnconfirmedDataUp":
		fallthrough
	case "ConfirmedDataUp":
		// Nothing, we handle them by default

	case "Proprietary":
		return fmt.Errorf("Unsupported MType Proprietary")
	}

	p.Payload = payload
	p.Metadata = proxy.Metadata
	return nil
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"encoding/base64"
	"reflect"
	"strings"

	"github.com/KtorZ/rpc/core"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/brocaar/lorawan"
)

func (a adapter) newDataRouterReq(rxpk semtech.RXPK, gid []byte) (*core.DataRouterReq, error) {
	// First, we have to get the physical payload which is encoded in the Data field
	if rxpk.Data == nil {
		return nil, errors.New(errors.Structural, "There's no data in the packet")
	}

	// RXPK Data are base64 encoded, yet without the trailing "==" if any.....
	encoded := *rxpk.Data
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

	macpayload, ok := payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		// TODO OTAA join request payloads
		return nil, errors.New(errors.Implementation, "Unhandled Physical payload. Expected a MACPayload")
	}
	if len(macpayload.FRMPayload) != 1 {
		// TODO Handle pure MAC Commands payloads (FType = 0)
		return nil, errors.New(errors.Implementation, "Unhandled Physical payload. Expected a Data Payload")
	}
	frmpayload, err := macpayload.FRMPayload[0].MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	var fopts [][]byte
	for _, cmd := range macpayload.FHDR.FOpts {
		if data, err := cmd.MarshalBinary(); err == nil { // We just ignore invalid MAC Commands
			fopts = append(fopts, data)
		}
	}

	// At the end, our converted packet hold the same metadata than the RXPK packet but the Data
	// which as been completely transformed into a lorawan Physical Payload.
	return &core.DataRouterReq{
		Payload: &core.LoRaWANData{
			MHDR: &core.LoRaWANMHDR{
				MType: uint32(payload.MHDR.MType),
				Major: uint32(payload.MHDR.Major),
			},
			MACPayload: &core.LoRaWANMACPayload{
				FHDR: &core.LoRaWANFHDR{
					DevAddr: macpayload.FHDR.DevAddr[:],
					FCnt:    macpayload.FHDR.FCnt,
					FCtrl: &core.LoRaWANFCtrl{
						ADR:       macpayload.FHDR.FCtrl.ADR,
						ADRAckReq: macpayload.FHDR.FCtrl.ADRACKReq,
						Ack:       macpayload.FHDR.FCtrl.ACK,
						FPending:  macpayload.FHDR.FCtrl.FPending,
					},
					FOpts: fopts,
				},
				FPort:      uint32(macpayload.FPort),
				FRMPayload: frmpayload,
			},
			MIC: payload.MIC[:],
		},
		Metadata:  extractMetadata(rxpk, new(core.Metadata)).(*core.Metadata),
		GatewayID: gid,
	}, nil
}

func (a adapter) newTXPK(resp core.DataRouterRes) (semtech.TXPK, error) {
	// Step 0: validate the response
	mac, mhdr, fhdr, fctrl, err := core.ValidateLoRaWANData(resp.Payload)
	if err != nil {
		return semtech.TXPK{}, errors.New(errors.Structural, err)
	}
	if resp.Metadata == nil {
		return semtech.TXPK{}, errors.New(errors.Structural, "Missing mandatory Metadata")
	}

	// Step 1: create a new LoRaWAN payload
	macpayload := lorawan.NewMACPayload(false)
	macpayload.FPort = uint8(mac.FPort)
	copy(macpayload.FHDR.DevAddr[:], fhdr.DevAddr)
	macpayload.FHDR.FCnt = fhdr.FCnt
	for _, data := range fhdr.FOpts {
		cmd := new(lorawan.MACCommand)
		if err := cmd.UnmarshalBinary(data); err == nil { // We ignore invalid commands
			macpayload.FHDR.FOpts = append(macpayload.FHDR.FOpts, *cmd)
		}
	}
	macpayload.FHDR.FCtrl.ADR = fctrl.ADR
	macpayload.FHDR.FCtrl.ACK = fctrl.Ack
	macpayload.FHDR.FCtrl.ADRACKReq = fctrl.ADRAckReq
	macpayload.FHDR.FCtrl.FPending = fctrl.FPending
	macpayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{
		Bytes: mac.FRMPayload,
	}}
	payload := lorawan.NewPHYPayload(false)
	payload.MHDR.MType = lorawan.MType(mhdr.MType)
	payload.MHDR.Major = lorawan.Major(mhdr.Major)
	copy(payload.MIC[:], resp.Payload.MIC)
	payload.MACPayload = macpayload

	// Step2: Convert the physical payload to a base64 string (without the padding)
	raw, err := payload.MarshalBinary()
	if err != nil {
		return semtech.TXPK{}, errors.New(errors.Structural, err)
	}
	data := strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")
	txpk := semtech.TXPK{Data: pointer.String(data)}

	// Step 3, copy every compatible metadata from the packet to the TXPK packet.
	// We are possibly loosing information here.
	injectMetadata(&txpk, *resp.Metadata) // Validation::2
	return txpk, nil
}

func injectMetadata(xpk interface{}, src interface{}) interface{} {
	v := reflect.ValueOf(src)
	t := v.Type()
	d := reflect.ValueOf(xpk).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i).Name
		if d.FieldByName(field).CanSet() {
			d.FieldByName(field).Set(v.Field(i))
		}
	}
	return xpk
}

func extractMetadata(xpk interface{}, target interface{}) interface{} {
	v := reflect.ValueOf(xpk)
	t := v.Type()
	m := reflect.ValueOf(target).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i).Name
		if m.FieldByName(field).CanSet() {
			m.FieldByName(field).Set(v.Field(i))
		}
	}
	return target
}

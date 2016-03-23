// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"encoding"
	"encoding/base64"
	"reflect"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

func toLoRaWANPayload(rxpk semtech.RXPK, gid []byte, ctx log.Interface) (interface{}, error) {
	// First, we have to get the physical payload which is encoded in the Data field
	if rxpk.Data == nil {
		return nil, errors.New(errors.Structural, "There's no data in the packet")
	}

	// RXPK Data are base64 encoded
	raw, err := base64.RawStdEncoding.DecodeString(*rxpk.Data)
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	payload := lorawan.NewPHYPayload(true)
	if err = payload.UnmarshalBinary(raw); err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	// Switch over MType
	switch payload.MHDR.MType {
	case lorawan.ConfirmedDataUp:
		fallthrough
	case lorawan.UnconfirmedDataUp:
		macpayload, ok := payload.MACPayload.(*lorawan.MACPayload)
		if !ok {
			return nil, errors.New(errors.Structural, "Unhandled Physical payload. Expected a MACPayload")
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
					FPort:      uint32(*macpayload.FPort),
					FRMPayload: frmpayload,
				},
				MIC: payload.MIC[:],
			},
			Metadata:  extractMetadata(rxpk, new(core.Metadata)).(*core.Metadata),
			GatewayID: gid,
		}, nil
	case lorawan.JoinRequest:
		joinpayload, ok := payload.MACPayload.(*lorawan.JoinRequestPayload)
		if !ok {
			return nil, errors.New(errors.Structural, "Unhandled Physical payload. Expected a JoinRequest Payload")
		}
		return &core.JoinRouterReq{
			GatewayID: gid,
			AppEUI:    joinpayload.AppEUI[:],
			DevEUI:    joinpayload.DevEUI[:],
			DevNonce:  joinpayload.DevNonce[:],
			MIC:       payload.MIC[:],
			Metadata:  extractMetadata(rxpk, new(core.Metadata)).(*core.Metadata),
		}, nil
	default:
		return nil, errors.New(errors.Structural, "Unexpected and unhandled LoRaWAN MHDR Mtype")
	}
}

func newTXPK(payload encoding.BinaryMarshaler, metadata *core.Metadata, ctx log.Interface) (semtech.TXPK, error) {
	// Step 0: validate the response
	if metadata == nil {
		return semtech.TXPK{}, errors.New(errors.Structural, "Missing mandatory Metadata")
	}

	// Step2: Convert the physical payload to a base64 string (without the padding)
	raw, err := payload.MarshalBinary()
	if err != nil {
		return semtech.TXPK{}, errors.New(errors.Structural, err)
	}
	data := strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")
	txpk := semtech.TXPK{Data: pointer.String(data)}

	// Step 3, copy every compatible metadata from the packet to the TXPK packet.
	// We are possibly loosing information here.
	injectMetadata(&txpk, *metadata)
	return txpk, nil
}

// injectMetadata takes metadata from a Struct and inject them into an xpk Struct (rxpk or txpk).
// The xpk is expected to be a pointer to an existing struct. Struct-tag in the rxpk struct are used
// to indicate with field from the struct is bound to which field on the xpk.
//
// All fields in the src struct are expected to be non-pointer values.
//
// It eventually returns the initial xpk argument
func injectMetadata(xpk interface{}, src interface{}) interface{} {
	m := reflect.ValueOf(src)
	x := reflect.ValueOf(xpk).Elem()
	tx := x.Type()

	for i := 0; i < tx.NumField(); i++ {
		t := tx.Field(i).Tag.Get("full")
		f := m.FieldByName(t)
		if f.IsValid() && f.Interface() != reflect.Zero(f.Type()).Interface() {
			p := reflect.New(f.Type())
			p.Elem().Set(f)
			if p.Type().AssignableTo(x.Field(i).Type()) {
				x.Field(i).Set(p)
			}
		}
	}

	return xpk
}

// extractMetadata does the reverse operation than injectMetadata. It takes metadata from an xpk
// structure and inject all compatible field into a given Struct.
// This time, the xpk is expected to be a plain xpk struct (not a pointer) and the target struct
// should be a reference to that struct.
//
// It eventually returns the completed target element.
func extractMetadata(xpk interface{}, target interface{}) interface{} {
	x := reflect.ValueOf(xpk)
	tx := x.Type()
	m := reflect.ValueOf(target).Elem()

	for i := 0; i < tx.NumField(); i++ {
		t := tx.Field(i).Tag.Get("full")
		f := m.FieldByName(t)
		if f.IsValid() && !x.Field(i).IsNil() {
			e := x.Field(i).Elem()
			if e.Type().AssignableTo(m.FieldByName(t).Type()) {
				m.FieldByName(t).Set(e)
			} else if e.Type().AssignableTo(reflect.TypeOf(time.Time{})) {
				m.FieldByName(t).Set(reflect.ValueOf(e.Interface().(time.Time).Format(time.RFC3339Nano)))
			}
		}
	}

	return target
}

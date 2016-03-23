// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

//go:generate sh -c "find protos -name '*.proto' | xargs protoc --gofast_out=plugins=grpc:. -I=protos"

package core

import (
	"reflect"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

// ValidateLoRaWANData makes sure that the provided lorawanData reference is valid by checking if
// All required parameters are present and conform.
func ValidateLoRaWANData(data *LoRaWANData) (*LoRaWANMACPayload, *LoRaWANMHDR, *LoRaWANFHDR, *LoRaWANFCtrl, error) {
	// Validation::0 -> Data isn't nil
	if data == nil {
		return nil, nil, nil, nil, errors.New(errors.Structural, "No data")
	}

	// Validation::1 -> All required fields are there
	if data.MHDR == nil || data.MACPayload == nil || data.MACPayload.FHDR == nil || data.MACPayload.FHDR.FCtrl == nil {
		return nil, nil, nil, nil, errors.New(errors.Structural, "Invalid Payload Structure")
	}

	// Validation::2 -> The MIC is 4-bytes long
	if len(data.MIC) != 4 {
		return nil, nil, nil, nil, errors.New(errors.Structural, "Invalid MIC")
	}

	// Validation::3 -> Device address is 4-bytes long
	if len(data.MACPayload.FHDR.DevAddr) != 4 {
		return nil, nil, nil, nil, errors.New(errors.Structural, "Invalid Device Address")
	}

	return data.MACPayload, data.MHDR, data.MACPayload.FHDR, data.MACPayload.FHDR.FCtrl, nil
}

// NewLoRaWANData converts a LoRaWANData to a brocaar/lorawan.PHYPayload
func NewLoRaWANData(reqPayload *LoRaWANData, uplink bool) (lorawan.PHYPayload, error) {
	mac, mhdr, fhdr, fctrl, err := ValidateLoRaWANData(reqPayload)
	if err != nil {
		return lorawan.PHYPayload{}, errors.New(errors.Structural, err)
	}

	macpayload := lorawan.NewMACPayload(uplink)
	macpayload.FPort = new(uint8)
	*macpayload.FPort = uint8(mac.FPort)
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
	payload := lorawan.NewPHYPayload(uplink)
	payload.MHDR.MType = lorawan.MType(mhdr.MType)
	payload.MHDR.Major = lorawan.Major(mhdr.Major)
	copy(payload.MIC[:], reqPayload.MIC)
	payload.MACPayload = macpayload

	return payload, nil
}

// ProtoMetaToAppMeta converts a set of Metadata generate with Protobuf to a set of valid
// AppMetadata ready to be marshaled to json
func ProtoMetaToAppMeta(srcs ...*Metadata) []AppMetadata {
	var dest []AppMetadata

	for _, src := range srcs {
		if src == nil {
			continue
		}
		to := new(AppMetadata)
		v := reflect.ValueOf(src).Elem()
		t := v.Type()
		d := reflect.ValueOf(to).Elem()

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i).Name
			if d.FieldByName(field).CanSet() {
				d.FieldByName(field).Set(v.Field(i))
			}
		}

		dest = append(dest, *to)
	}

	return dest
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (m StatsMetadata) MarshalBinary() ([]byte, error) {
	return m.Marshal()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (m *StatsMetadata) UnmarshalBinary(data []byte) error {
	return m.Unmarshal(data)
}

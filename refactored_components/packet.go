// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"encoding/base64"
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
	"github.com/thethingsnetwork/core/semtech"
	"reflect"
)

var ErrImpossibleConversion = fmt.Errorf("The given packet can't be converted")

func ConvertRXPK(p semtech.RXPK) (core.Packet, error) {
	packet := core.Packet{}
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

	return core.Packet{Metadata: &metadata, Payload: payload}, nil
}

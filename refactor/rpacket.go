// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package refactor

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

// type RPacket materializes packets manipulated by the router and corresponding adapter handlers
type RPacket struct {
	metadata Metadata
	payload  lorawan.PHYPayload
}

func NewRPacket(payload lorawan.PHYPayload, metadata Metadata) RPacket {
	return RPacket{
		metadata: metadata,
		payload:  payload,
	}
}

// String implements the Stringer interface
func (p RPacket) String() string {
	str := "Packet {"
	str += fmt.Sprintf("\n\t%s}", p.metadata.String())
	str += fmt.Sprintf("\n\tPayload%+v\n}", p.payload)
	return str
}

// Payload implements the core.Packet interface
func (p RPacket) Payload() encoding.BinaryMarshaler {
	return p.payload
}

// Metadata implements the core.Packet interface
func (p RPacket) Metadata() Metadata {
	return p.metadata
}

// DevEUI implements the core.Addressable interface
func (p RPacket) DevEUI() (lorawan.EUI64, error) {
	if p.payload.MACPayload == nil {
		return lorawan.EUI64{}, errors.New(ErrInvalidStructure, "MACPAyload should not be empty")
	}

	macpayload, ok := p.payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return lorawan.EUI64{}, errors.New(ErrInvalidStructure, "Packet does not carry a MACPayload")
	}

	var devEUI lorawan.EUI64
	copy(devEUI[4:], macpayload.FHDR.DevAddr[:])
	return devEUI, nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p RPacket) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

// MarshalBinary implements the encoding.BinaryUnMarshaler interface
func (p *RPacket) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

// MarshalJSON implements the json.Marshaler interface
func (p RPacket) MarshalJSON() ([]byte, error) {
	rawMetadata, err := json.Marshal(p.metadata)
	if err != nil {
		return nil, errors.New(ErrInvalidStructure, err)
	}
	rawPayload, err := p.payload.MarshalBinary()
	if err != nil {
		return nil, errors.New(ErrInvalidStructure, err)
	}
	strPayload := base64.StdEncoding.EncodeToString(rawPayload)
	return []byte(fmt.Sprintf(`{"payload":"%s","metadata":%s}`, strPayload, string(rawMetadata))), nil
}

// UnmarshalJSON impements the json.Unmarshaler interface
func (p *RPacket) UnmarshalJSON(raw []byte) error {
	if p == nil {
		return errors.New(ErrInvalidStructure, "Cannot unmarshal a nil packet")
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
		return errors.New(ErrInvalidStructure, err)
	}

	rawPayload, err := base64.StdEncoding.DecodeString(proxy.Payload)
	if err != nil {
		return errors.New(ErrInvalidStructure, err)
	}

	payload := lorawan.NewPHYPayload(true) // true -> uplink
	if err := payload.UnmarshalBinary(rawPayload); err != nil {
		return errors.New(ErrInvalidStructure, err)
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
			return errors.New(ErrInvalidStructure, err)
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
		return errors.New(ErrInvalidStructure, "Unsupported MType 'Proprietary'")
	}

	// Packet = Payload + Metadata
	p.payload = payload
	p.metadata = proxy.Metadata
	return nil
}

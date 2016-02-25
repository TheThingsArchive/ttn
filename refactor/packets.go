// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package refactor

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	"github.com/brocaar/lorawan"
)

type RPacket interface {
	Packet
	Metadata() Metadata
	Payload() lorawan.PHYPayload
	DevEUI() lorawan.EUI64
}

// rpacket implements the core.RPacket interface
type rpacket struct {
	metadata Metadata
	payload  lorawan.PHYPayload
}

// NewRPacket construct a new router packet given a payload and metadata
func NewRPacket(payload lorawan.PHYPayload, metadata Metadata) (RPacket, error) {
	packet := rpacket{payload: payload, metadata: metadata}

	// Check and extract the devEUI
	if payload.MACPayload == nil {
		return nil, errors.New(errors.Structural, "MACPAyload should not be empty")
	}

	_, ok := payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return nil, errors.New(errors.Structural, "Packet does not carry a MACPayload")
	}

	return &packet, nil
}

// DevEUI implements the core.BPacket interface
func (p rpacket) DevEUI() lorawan.EUI64 {
	var devEUI lorawan.EUI64
	copy(devEUI[4:], p.payload.MACPayload.(*lorawan.MACPayload).FHDR.DevAddr[:])
	return devEUI
}

// Metadata implements the core.RPacket interface
func (p rpacket) Metadata() Metadata {
	return p.metadata
}

// Payload implements the core.RPacket interface
func (p rpacket) Payload() lorawan.PHYPayload {
	return p.payload
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p rpacket) MarshalBinary() ([]byte, error) {
	var mtype byte
	switch p.payload.MHDR.MType {
	case lorawan.JoinRequest:
		fallthrough
	case lorawan.UnconfirmedDataUp:
		fallthrough
	case lorawan.ConfirmedDataUp:
		mtype = 1 // Up
	case lorawan.JoinAccept:
		fallthrough
	case lorawan.UnconfirmedDataDown:
		fallthrough
	case lorawan.ConfirmedDataDown:
		mtype = 2 // Down
	default:
		msg := fmt.Sprintf("Unsupported mtype: %s", p.payload.MHDR.MType.String())
		return nil, errors.New(errors.Implementation, msg)
	}

	dataMetadata, err := p.metadata.MarshalJSON()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	dataPayload, err := p.payload.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	rw := readwriter.New(nil)
	rw.Write([]byte{mtype})
	rw.Write(dataMetadata)
	rw.Write(dataPayload)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *rpacket) UnmarshalBinary(data []byte) error {
	if p == nil {
		return errors.New(errors.Structural, "Cannot unmarshal nil packet")
	}
	var isUp bool
	rw := readwriter.New(data)
	rw.Read(func(data []byte) {
		if data[0] == 1 {
			isUp = true
		}
	})

	var dataMetadata []byte
	rw.Read(func(data []byte) { dataMetadata = data })

	var dataPayload []byte
	rw.Read(func(data []byte) { dataPayload = data })

	if rw.Err() != nil {
		return errors.New(errors.Structural, rw.Err())
	}

	p.metadata = Metadata{}
	if err := p.metadata.UnmarshalJSON(dataMetadata); err != nil {
		return errors.New(errors.Structural, err)
	}

	p.payload = lorawan.NewPHYPayload(isUp)
	if err := p.payload.UnmarshalBinary(dataPayload); err != nil {
		return errors.New(errors.Structural, err)
	}

	return nil
}

// String implements the Stringer interface
func (p rpacket) String() string {
	str := "Packet {"
	str += fmt.Sprintf("\n\t%s}", p.metadata.String())
	str += fmt.Sprintf("\n\tPayload%+v\n}", p.payload)
	return str
}

type BPacket interface {
	Commands() []lorawan.MACCommand
	DevEUI() lorawan.EUI64
	FCnt() uint32
	Metadata() Metadata
	Payload() []byte
	ValidateMIC(key lorawan.AES128Key) (bool, error)
}

// bpacket implements the core.BPacket interface
type bpacket struct {
	rpacket
}

// NewBPacket constructs a new broker packets given a payload and metadata
func NewBPacket(payload lorawan.PHYPayload, metadata Metadata) (BPacket, error) {
	packet, err := NewRPacket(payload, metadata)
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	macPayload := packet.Payload().MACPayload.(*lorawan.MACPayload)
	if len(macPayload.FRMPayload) != 1 {
		return nil, errors.New(errors.Structural, "Invalid frame payload. Expected exactly 1")
	}

	_, ok := macPayload.FRMPayload[0].(*lorawan.DataPayload)
	if !ok {
		return nil, errors.New(errors.Structural, "Invalid frame payload. Expected only data")
	}

	return bpacket{rpacket: packet.(rpacket)}, nil
}

// FCnt implements the core.BPacket interface
func (p bpacket) FCnt() uint32 {
	return p.payload.MACPayload.(*lorawan.MACPayload).FHDR.FCnt
}

// Payload implements the core.BPacket interface
func (p bpacket) Payload() []byte {
	macPayload := p.rpacket.payload.MACPayload.(*lorawan.MACPayload)
	return macPayload.FRMPayload[0].(*lorawan.DataPayload).Bytes
}

// ValidateMIC implements the core.BPacket interface
func (p bpacket) ValidateMIC(key lorawan.AES128Key) (bool, error) {
	return p.rpacket.payload.ValidateMIC(key)
}

// Commands implements the core.BPacket interface
func (p bpacket) Commands() []lorawan.MACCommand {
	return p.rpacket.payload.MACPayload.(*lorawan.MACPayload).FHDR.FOpts
}

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
}

// rpacket implements the core.RPacket interface
type rpacket struct {
	metadata Metadata
	payload  lorawan.PHYPayload
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

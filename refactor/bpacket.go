// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package refactor

import (
	"encoding"

	"github.com/brocaar/lorawan"
)

// type BPacket materializes packets manipulated by the broker and corresponding adapter handlers
type BPacket struct{}

// FCnt implements the core.Packet interface
func (p BPacket) FCnt() (uint32, error) {
	return 0, nil
}

// Payload implements the core.Packet interface
func (p BPacket) Payload() encoding.BinaryMarshaler {
	return nil
}

// Metadata implements the core.Packet interface
func (p BPacket) Metadata() []Metadata {
	return nil
}

// AppEUI implements the core.Packet interface
func (p BPacket) AppEUI() (lorawan.EUI64, error) {
	return lorawan.EUI64{}, nil
}

// DevEUI implements the core.Packet interface
func (p BPacket) DevEUI() (lorawan.EUI64, error) {
	return lorawan.EUI64{}, nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p BPacket) MarshalBinary() ([]byte, error) {
	return nil, nil
}

// MarshalBinary implements the encoding.BinaryUnMarshaler interface
func (p *BPacket) UnmarshalBinary(data []byte) error {
	return nil
}

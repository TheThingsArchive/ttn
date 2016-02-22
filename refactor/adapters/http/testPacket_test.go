// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"encoding"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	core "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

// type testPacket materializes packets manipulated by the broker and corresponding adapter handlers
type testPacket struct {
	devEUI  lorawan.EUI64
	payload string
}

// FCnt implements the core.Packet interface
func (p testPacket) FCnt() (uint32, error) {
	return 0, nil
}

// Payload implements the core.Packet interface
func (p testPacket) Payload() encoding.BinaryMarshaler {
	return nil
}

// Metadata implements the core.Packet interface
func (p testPacket) Metadata() []core.Metadata {
	return nil
}

// AppEUI implements the core.Packet interface
func (p testPacket) AppEUI() (lorawan.EUI64, error) {
	return lorawan.EUI64{}, nil
}

// DevEUI implements the core.Packet interface
func (p testPacket) DevEUI() (lorawan.EUI64, error) {
	return p.devEUI, nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p testPacket) MarshalBinary() ([]byte, error) {
	if p.payload == "" {
		return nil, errors.New(ErrInvalidStructure, "Fake error")
	}

	return []byte(p.payload), nil
}

// MarshalBinary implements the encoding.BinaryUnMarshaler interface
func (p *testPacket) UnmarshalBinary(data []byte) error {
	return nil
}

// String implements the core.Packet interface
func (p testPacket) String() string {
	return p.payload
}

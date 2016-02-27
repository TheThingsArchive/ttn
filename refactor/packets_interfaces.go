// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package refactor

import (
	"encoding"
	"fmt"

	"github.com/brocaar/lorawan"
)

type Packet interface {
	DevEUI() lorawan.EUI64
	encoding.BinaryMarshaler
	fmt.Stringer
}

type RPacket interface {
	Packet
	Metadata() Metadata
	Payload() lorawan.PHYPayload
}

type BPacket interface {
	Packet
	Commands() []lorawan.MACCommand
	FCnt() uint32
	Metadata() Metadata
	Payload() lorawan.PHYPayload
	ValidateMIC(key lorawan.AES128Key) (bool, error)
}

type HPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	FCnt() uint32
	Payload(appSKey lorawan.AES128Key) ([]byte, error) // Unencrypted FRMPayload
	Metadata() Metadata                                // TTL on down, DutyCycle + Rssi on Up
}

type APacket interface {
	Packet
	Payload() []byte
	Metadata() []Metadata
}

type JPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	DevNonce() [2]byte
	Metadata() Metadata // Rssi + DutyCycle
}

type CPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	Payload() []byte
	NwkSKey() lorawan.AES128Key
}

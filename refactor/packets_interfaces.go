// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package refactor

import (
	"github.com/brocaar/lorawan"
)

type RPacket interface {
	Packet
	Metadata() Metadata
	Payload() lorawan.PHYPayload
	DevEUI() lorawan.EUI64
}

type BPacket interface {
	Packet
	Commands() []lorawan.MACCommand
	DevEUI() lorawan.EUI64
	FCnt() uint32
	Metadata() Metadata
	Payload() []byte
	ValidateMIC(key lorawan.AES128Key) (bool, error)
}

type HPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	DevEUI() lorawan.EUI64
	Payload() []byte    // FRMPayload
	Metadata() Metadata // TTL on down, DutyCycle + Rssi on Up
}

type APacket interface {
	Packet
	Payload() []byte
	Metadata() []Metadata
}

type JPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	DevEUI() lorawan.EUI64
	DevNonce() [2]byte
	Metadata() Metadata // Rssi + DutyCycle
}

type CPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	DevEUI() lorawan.EUI64
	Payload() []byte
	NwkSKey() lorawan.AES128Key
}

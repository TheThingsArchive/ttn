// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"github.com/brocaar/lorawan"
)

type Component interface {
	Register(reg Registration, an AckNacker) error
	HandleUp(p []byte, an AckNacker, upAdapter Adapter) error
	HandleDown(p []byte, an AckNacker, downAdapter Adapter) error
}

type NetworkController interface {
	HandleCommands(packet BPacket) error
	UpdateFCntUp(appEUI lorawan.EUI64, devEUI lorawan.EUI64, fcnt uint32)
	UpdateFCntDown(appEUI lorawan.EUI64, devEUI lorawan.EUI64, fcnt uint32)
	MergeCommands(appEUI lorawan.EUI64, devEUI lorawan.EUI64, pkt BPacket) RPacket
}

type AckNacker interface {
	Ack(p Packet) error
	Nack() error
}

type Adapter interface {
	Send(p Packet, r ...Recipient) ([]byte, error)
	//Join(r JoinRequest, r ...Recipient) (JoinResponse, error)
	GetRecipient(raw []byte) (Recipient, error)
	Next() ([]byte, AckNacker, error)
	NextRegistration() (Registration, AckNacker, error)
}

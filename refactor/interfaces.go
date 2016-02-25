// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package refactor

import (
	"encoding"
	"fmt"

	"github.com/brocaar/lorawan"
)

type Component interface {
	Register(reg Registration, an AckNacker) error
	HandleUp(p []byte, an AckNacker, upAdapter Adapter) error
	HandleDown(p []byte, an AckNacker, downAdapter Adapter) error
}

type AckNacker interface {
	Ack(p Packet) error
	Nack() error
}

type Adapter interface {
	Send(p Packet, r ...Recipient) ([]byte, error)
	GetRecipient(raw []byte) (Recipient, error)
	Next() ([]byte, AckNacker, error)
	NextRegistration() (Registration, AckNacker, error)
	//Join(r JoinRequest, r ...Recipient) (JoinResponse, error)
}

type Packet interface {
	encoding.BinaryMarshaler
	fmt.Stringer
}

type Addressable interface {
	DevEUI() (lorawan.EUI64, error)
}

type Registration interface {
	Recipient() Recipient
	AppEUI() (lorawan.EUI64, error)
	AppSKey() (lorawan.AES128Key, error)
	DevEUI() (lorawan.EUI64, error)
	NwkSKey() (lorawan.AES128Key, error)
}

type Recipient interface {
	encoding.BinaryMarshaler
}

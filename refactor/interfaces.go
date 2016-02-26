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
	//Join(r JoinRequest, r ...Recipient) (JoinResponse, error)
	GetRecipient(raw []byte) (Recipient, error)
	Next() ([]byte, AckNacker, error)
	NextRegistration() (Registration, AckNacker, error)
}

type Packet interface {
	DevEUI() lorawan.EUI64
	encoding.BinaryMarshaler
	fmt.Stringer
}

type Registration interface {
	Recipient() Recipient
	DevEUI() lorawan.EUI64
}

type Recipient interface {
	encoding.BinaryMarshaler
}

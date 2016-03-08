// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

type Handler interface {
	Register(reg Registration, an AckNacker, s Subscriber) error
	HandleUp(p []byte, an AckNacker, upAdapter Adapter) error
	HandleDown(p []byte, an AckNacker, down Adapter) error
}

type Router interface {
	Register(reg Registration, an AckNacker) error
	HandleUp(p []byte, an AckNacker, up Adapter) error
}

type Broker interface {
	Register(reg Registration, an AckNacker) error
	HandleUp(p []byte, an AckNacker, up Adapter) error
}

// AckNacker represents an interface that allow adapters to decouple their protocol from the
// behaviour expected by the caller.
type AckNacker interface {
	Ack(p Packet) error
	Nack(err error) error
}

// Adapter handles communications between components. They implement a specific communication
// protocol which is completely hidden from the outside.
type Adapter interface {
	Send(p Packet, r ...Recipient) ([]byte, error)
	//Join(r JoinRequest, r ...Recipient) (JoinResponse, error)
	GetRecipient(raw []byte) (Recipient, error)
	Next() ([]byte, AckNacker, error)
	NextRegistration() (Registration, AckNacker, error)
}

type Subscriber interface {
	Subscribe(reg Registration) error
}

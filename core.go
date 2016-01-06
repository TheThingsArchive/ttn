package core

import (
	"encoding/json"
	"fmt"
	"github.com/thethingsnetwork/core/lorawan"
)

var ErrInvalidPacket error = fmt.Errorf("Invalid Packet")

type Packet struct {
	Metadata Metadata
	Payload  lorawan.PHYPayload
}

func (p Packet) DevAddr() *lorawan.DevAddr {
	addr, err := p.Payload.DevAddr()
	if err != nil {
		return nil
	}
	return &addr
}

func (p Packet) String() string {
	str := "Packet {"
	if p.Metadata != nil {
		str += fmt.Sprintf("\n\t%s}", p.Metadata.String())
	}
	str += fmt.Sprintf("\n\tPayload%+v\n}", p.Payload)
	return str
}

type Addressable interface {
	DevAddr() [4]byte
}

type Recipient struct {
	Address interface{}
	Id      interface{}
}

type Metadata interface {
	json.Marshaler
	json.Unmarshaler
	String() string
}

type AckNacker interface {
	Ack(p Packet) error
	Nack(p Packet) error
}

type Component interface {
	Handle(p Packet, an AckNacker) error
	NextUp() (*Packet, error)
	NextDown() (*Packet, error)
}

type Adapter interface {
	Send(p Packet, r ...Recipient) error
	Next() (*Packet, *AckNacker, error)
}

type Router Component
type Broker Component
type Handler Component
type NetworkController Component

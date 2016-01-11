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

func (p Packet) DevAddr() (lorawan.DevAddr, error) {
	return p.Payload.DevAddr()
}

func (p Packet) String() string {
	str := "Packet {"
	if p.Metadata != nil {
		str += fmt.Sprintf("\n\t%s}", p.Metadata.String())
	}
	str += fmt.Sprintf("\n\tPayload%+v\n}", p.Payload)
	return str
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
	Register(reg Registration) error
	HandleUp(p Packet, an AckNacker, upAdapter Adapter) error
	HandleDown(p Packet, an AckNacker, downAdapter Adapter) error
}

type Adapter interface {
	Send(p Packet, r ...Recipient) (Packet, error)
	Next() (Packet, AckNacker, error)
	NextRegistration() (Registration, AckNacker, error)
}

type Registration struct {
	DevAddr   lorawan.DevAddr
	Recipient Recipient
	Options   interface{}
}

type Router Component
type Broker Component
type Handler Component
type NetworkController Component

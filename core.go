package core

import (
	"encoding"
	"encoding/json"
)

type Packet struct {
	Addressable
	Metadata Metadata
	Payload  PHYPayload
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
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type PHYPayload interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	DecryptMACPayload(key [16]byte) error
	EncryptMACPayload(key [16]byte) error
	SetMIC(key [16]byte) error
	ValidateMic(key [16]byte) (bool, error)
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

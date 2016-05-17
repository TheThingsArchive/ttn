package types

import (
	"encoding/hex"
	"errors"
	"strings"
)

// DevAddr is a non-unique address for LoRaWAN devices.
type DevAddr [4]byte

// ParseDevAddr parses a 32-bit hex-encoded string to a DevAddr
func ParseDevAddr(input string) (addr DevAddr, err error) {
	bytes, err := parseHEX(input, 4)
	if err != nil {
		return
	}
	copy(addr[:], bytes)
	return
}

// Bytes returns the DevAddr as a byte slice
func (addr DevAddr) Bytes() []byte {
	return addr[:]
}

// String implements the Stringer interface.
func (addr DevAddr) String() string {
	return strings.ToUpper(hex.EncodeToString(addr.Bytes()))
}

// GoString implements the GoStringer interface.
func (addr DevAddr) GoString() string {
	return addr.String()
}

// MarshalText implements the TextMarshaler interface.
func (addr DevAddr) MarshalText() ([]byte, error) {
	return []byte(addr.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (addr *DevAddr) UnmarshalText(data []byte) error {
	parsed, err := ParseDevAddr(string(data))
	if err != nil {
		return err
	}
	*addr = DevAddr(parsed)
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (addr DevAddr) MarshalBinary() ([]byte, error) {
	return addr.Bytes(), nil
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (addr *DevAddr) UnmarshalBinary(data []byte) error {
	if len(data) != 4 {
		return errors.New("ttn/core: Invalid length for DevAddr")
	}
	copy(addr[:], data)
	return nil
}

// Marshal implements the Marshaler interface.
func (addr DevAddr) Marshal() ([]byte, error) {
	return addr.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (addr *DevAddr) Unmarshal(data []byte) error {
	*addr = [4]byte{} // Reset the receiver
	return addr.UnmarshalBinary(data)
}

var empty DevAddr

func (addr DevAddr) IsEmpty() bool {
	return addr == empty
}

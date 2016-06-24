// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import "errors"

// DeviceType is the type of a LoRaWAN device.
type DeviceType byte

const (
	// ABP is a LoRaWAN device that is activated by personalization.
	ABP DeviceType = iota
	// OTAA is an over-the-air activated LoRaWAN device.
	OTAA
)

// ParseDeviceType parses a string to a DeviceType.
func ParseDeviceType(input string) (devType DeviceType, err error) {
	switch input {
	case "ABP":
		devType = ABP
	case "OTAA":
		devType = OTAA
	default:
		err = errors.New("ttn/core: Invalid DeviceType")
	}
	return
}

// String implements the Stringer interface.
func (devType DeviceType) String() string {
	switch devType {
	case ABP:
		return "ABP"
	case OTAA:
		return "OTAA"
	}
	return ""
}

// GoString implements the GoStringer interface.
func (devType DeviceType) GoString() string {
	return devType.String()
}

// MarshalText implements the TextMarshaler interface.
func (devType DeviceType) MarshalText() ([]byte, error) {
	str := devType.String()
	if str == "" {
		return nil, errors.New("ttn/core: Invalid DeviceType")
	}
	return []byte(devType.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (devType *DeviceType) UnmarshalText(data []byte) error {
	parsed, err := ParseDeviceType(string(data))
	if err != nil {
		return err
	}
	*devType = DeviceType(parsed)
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (devType DeviceType) MarshalBinary() ([]byte, error) {
	switch devType {
	case ABP, OTAA:
		return []byte{byte(devType)}, nil
	default:
		return nil, errors.New("ttn/core: Invalid DeviceType")
	}
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (devType *DeviceType) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("ttn/core: Invalid length for DeviceType")
	}
	switch data[0] {
	case byte(ABP), byte(OTAA):
		*devType = DeviceType(data[0])
	default:
		return errors.New("ttn/core: Invalid DeviceType")
	}
	return nil
}

// MarshalTo is used by Protobuf
func (devType *DeviceType) MarshalTo(b []byte) (int, error) {
	copy(b, []byte(devType.String()))
	return len(devType.String()), nil
}

// Size is used by Protobuf
func (devType *DeviceType) Size() int {
	return len(devType.String())
}

// Marshal implements the Marshaler interface.
func (devType DeviceType) Marshal() ([]byte, error) {
	return devType.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (devType *DeviceType) Unmarshal(data []byte) error {
	return devType.UnmarshalBinary(data)
}

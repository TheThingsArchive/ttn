package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

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

// Marshal implements the Marshaler interface.
func (devType DeviceType) Marshal() ([]byte, error) {
	return devType.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (devType *DeviceType) Unmarshal(data []byte) error {
	return devType.UnmarshalBinary(data)
}

// parseHEX parses a string "input" to a byteslice with length "length".
func parseHEX(input string, length int) ([]byte, error) {
	pattern, err := regexp.Compile(fmt.Sprintf("[[:xdigit:]]{%d}", length*2))
	if err != nil {
		return nil, fmt.Errorf("Invalid pattern")
	}

	valid := pattern.MatchString(input)
	if !valid {
		return nil, fmt.Errorf("Invalid input: %s", input)
	}

	slice, err := hex.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("Could not decode input: %s", input)
	}

	return slice, nil
}

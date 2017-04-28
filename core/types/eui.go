// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"strings"

	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// EUI64 is used for AppEUIs and DevEUIs.
type EUI64 [8]byte

// AppEUI is a unique identifier for applications.
type AppEUI EUI64

// DevEUI is a unique identifier for devices.
type DevEUI EUI64

// ParseEUI64 parses a 64-bit hex-encoded string to an EUI64.
func ParseEUI64(input string) (eui EUI64, err error) {
	bytes, err := ParseHEX(input, 8)
	if err != nil {
		return
	}
	copy(eui[:], bytes)
	return
}

// Bytes returns the EUI64 as a byte slice
func (eui EUI64) Bytes() []byte {
	return eui[:]
}

func (eui EUI64) String() string {
	if eui.IsEmpty() {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(eui.Bytes()))
}

// GoString implements the GoStringer interface.
func (eui EUI64) GoString() string {
	return eui.String()
}

// MarshalText implements the TextMarshaler interface.
func (eui EUI64) MarshalText() ([]byte, error) {
	return []byte(eui.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (eui *EUI64) UnmarshalText(data []byte) error {
	parsed, err := ParseEUI64(string(data))
	if err != nil {
		return err
	}
	*eui = EUI64(parsed)
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (eui EUI64) MarshalBinary() ([]byte, error) {
	return eui.Bytes(), nil
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (eui *EUI64) UnmarshalBinary(data []byte) error {
	if len(data) != 8 {
		return errors.New("ttn/core: Invalid length for EUI64")
	}
	copy(eui[:], data)
	return nil
}

// MarshalTo is used by Protobuf
func (eui *EUI64) MarshalTo(b []byte) (int, error) {
	copy(b, eui.Bytes())
	return 8, nil
}

// Size is used by Protobuf
func (eui *EUI64) Size() int {
	return 8
}

// Marshal implements the Marshaler interface.
func (eui EUI64) Marshal() ([]byte, error) {
	return eui.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (eui *EUI64) Unmarshal(data []byte) error {
	*eui = [8]byte{} // Reset the receiver
	return eui.UnmarshalBinary(data)
}

// Equal returns whether eui is equal to other
func (eui EUI64) Equal(other EUI64) bool {
	return eui == other
}

// ParseAppEUI parses a 64-bit hex-encoded string to an AppEUI
func ParseAppEUI(input string) (eui AppEUI, err error) {
	eui64, err := ParseEUI64(input)
	if err != nil {
		return
	}
	eui = AppEUI(eui64)
	return
}

// Bytes returns the AppEUI as a byte slice
func (eui AppEUI) Bytes() []byte {
	return EUI64(eui).Bytes()
}

// String implements the Stringer interface.
func (eui AppEUI) String() string {
	return EUI64(eui).String()
}

// GoString implements the GoStringer interface.
func (eui AppEUI) GoString() string {
	return eui.String()
}

// MarshalText implements the TextMarshaler interface.
func (eui AppEUI) MarshalText() ([]byte, error) {
	return EUI64(eui).MarshalText()
}

// UnmarshalText implements the TextUnmarshaler interface.
func (eui *AppEUI) UnmarshalText(data []byte) error {
	e := EUI64(*eui)
	err := e.UnmarshalText(data)
	if err != nil {
		return err
	}
	*eui = AppEUI(e)
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (eui AppEUI) MarshalBinary() ([]byte, error) {
	return EUI64(eui).MarshalBinary()
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (eui *AppEUI) UnmarshalBinary(data []byte) error {
	e := EUI64(*eui)
	err := e.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	*eui = AppEUI(e)
	return nil
}

// MarshalTo is used by Protobuf
func (eui *AppEUI) MarshalTo(b []byte) (int, error) {
	copy(b, eui.Bytes())
	return 8, nil
}

// Size is used by Protobuf
func (eui *AppEUI) Size() int {
	return 8
}

// Marshal implements the Marshaler interface.
func (eui AppEUI) Marshal() ([]byte, error) {
	return eui.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (eui *AppEUI) Unmarshal(data []byte) error {
	*eui = [8]byte{} // Reset the receiver
	return eui.UnmarshalBinary(data)
}

// Equal returns whether eui is equal to other
func (eui AppEUI) Equal(other AppEUI) bool {
	return eui == other
}

// ParseDevEUI parses a 64-bit hex-encoded string to an DevEUI
func ParseDevEUI(input string) (eui DevEUI, err error) {
	eui64, err := ParseEUI64(input)
	if err != nil {
		return
	}
	eui = DevEUI(eui64)
	return
}

// Bytes returns the DevEUI as a byte slice
func (eui DevEUI) Bytes() []byte {
	return EUI64(eui).Bytes()
}

// String implements the Stringer interface.
func (eui DevEUI) String() string {
	return EUI64(eui).String()
}

// GoString implements the GoStringer interface.
func (eui DevEUI) GoString() string {
	return eui.String()
}

// MarshalText implements the TextMarshaler interface.
func (eui DevEUI) MarshalText() ([]byte, error) {
	return EUI64(eui).MarshalText()
}

// UnmarshalText implements the TextUnmarshaler interface.
func (eui *DevEUI) UnmarshalText(data []byte) error {
	e := EUI64(*eui)
	err := e.UnmarshalText(data)
	if err != nil {
		return err
	}
	*eui = DevEUI(e)
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (eui DevEUI) MarshalBinary() ([]byte, error) {
	return EUI64(eui).MarshalBinary()
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (eui *DevEUI) UnmarshalBinary(data []byte) error {
	e := EUI64(*eui)
	err := e.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	*eui = DevEUI(e)
	return nil
}

// MarshalTo is used by Protobuf
func (eui *DevEUI) MarshalTo(b []byte) (int, error) {
	copy(b, eui.Bytes())
	return 8, nil
}

// Size is used by Protobuf
func (eui *DevEUI) Size() int {
	return 8
}

// Marshal implements the Marshaler interface.
func (eui DevEUI) Marshal() ([]byte, error) {
	return eui.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (eui *DevEUI) Unmarshal(data []byte) error {
	*eui = [8]byte{} // Reset the receiver
	return eui.UnmarshalBinary(data)
}

// Equal returns whether eui is equal to other
func (eui DevEUI) Equal(other DevEUI) bool {
	return eui == other
}

var emptyEUI64 EUI64

func (eui EUI64) IsEmpty() bool {
	return eui == emptyEUI64
}

func (eui DevEUI) IsEmpty() bool {
	return EUI64(eui).IsEmpty()
}

func (eui AppEUI) IsEmpty() bool {
	return EUI64(eui).IsEmpty()
}

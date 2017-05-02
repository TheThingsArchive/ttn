// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"errors"
	"strings"
)

// Activation messages are used to notify application of a device activation
type Activation struct {
	AppID    string   `json:"app_id,omitempty"`
	DevID    string   `json:"dev_id,omitempty"`
	AppEUI   AppEUI   `json:"app_eui,omitempty"`
	DevEUI   DevEUI   `json:"dev_eui,omitempty"`
	DevAddr  DevAddr  `json:"dev_addr,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
}

// DevNonce for LoRaWAN
type DevNonce [2]byte

// Bytes returns the DevNonce as a byte slice
func (n DevNonce) Bytes() []byte {
	return n[:]
}

func (n DevNonce) String() string {
	if n == [2]byte{0, 0} {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(n.Bytes()))
}

// GoString implements the GoStringer interface.
func (n DevNonce) GoString() string {
	return n.String()
}

// MarshalText implements the TextMarshaler interface.
func (n DevNonce) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (n *DevNonce) UnmarshalText(data []byte) error {
	parsed, err := ParseHEX(string(data), 2)
	if err != nil {
		return err
	}
	copy(n[:], parsed[:])
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (n DevNonce) MarshalBinary() ([]byte, error) {
	return n.Bytes(), nil
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (n *DevNonce) UnmarshalBinary(data []byte) error {
	if len(data) != 2 {
		return errors.New("ttn/core: Invalid length for DevNonce")
	}
	copy(n[:], data)
	return nil
}

// MarshalTo is used by Protobuf
func (n *DevNonce) MarshalTo(b []byte) (int, error) {
	copy(b, n.Bytes())
	return 2, nil
}

// Size is used by Protobuf
func (n *DevNonce) Size() int {
	return 2
}

// Marshal implements the Marshaler interface.
func (n DevNonce) Marshal() ([]byte, error) {
	return n.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (n *DevNonce) Unmarshal(data []byte) error {
	*n = [2]byte{} // Reset the receiver
	return n.UnmarshalBinary(data)
}

// Equal returns whether n is equal to other
func (n DevNonce) Equal(other DevNonce) bool {
	return n == other
}

// AppNonce for LoRaWAN
type AppNonce [3]byte

// Bytes returns the AppNonce as a byte slice
func (n AppNonce) Bytes() []byte {
	return n[:]
}

func (n AppNonce) String() string {
	if n == [3]byte{0, 0, 0} {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(n.Bytes()))
}

// GoString implements the GoStringer interface.
func (n AppNonce) GoString() string {
	return n.String()
}

// MarshalText implements the TextMarshaler interface.
func (n AppNonce) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (n *AppNonce) UnmarshalText(data []byte) error {
	parsed, err := ParseHEX(string(data), 3)
	if err != nil {
		return err
	}
	copy(n[:], parsed[:])
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (n AppNonce) MarshalBinary() ([]byte, error) {
	return n.Bytes(), nil
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (n *AppNonce) UnmarshalBinary(data []byte) error {
	if len(data) != 3 {
		return errors.New("ttn/core: Invalid length for AppNonce")
	}
	copy(n[:], data)
	return nil
}

// MarshalTo is used by Protobuf
func (n *AppNonce) MarshalTo(b []byte) (int, error) {
	copy(b, n.Bytes())
	return 3, nil
}

// Size is used by Protobuf
func (n *AppNonce) Size() int {
	return 3
}

// Marshal implements the Marshaler interface.
func (n AppNonce) Marshal() ([]byte, error) {
	return n.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (n *AppNonce) Unmarshal(data []byte) error {
	*n = [3]byte{} // Reset the receiver
	return n.UnmarshalBinary(data)
}

// Equal returns whether n is equal to other
func (n AppNonce) Equal(other AppNonce) bool {
	return n == other
}

// NetID for LoRaWAN
type NetID [3]byte

var emptyNetID NetID

// Bytes returns the NetID as a byte slice
func (n NetID) Bytes() []byte {
	return n[:]
}

func (n NetID) String() string {
	if n == [3]byte{0, 0, 0} {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(n.Bytes()))
}

func (n NetID) IsEmpty() bool {
	return n == emptyNetID
}

// GoString implements the GoStringer interface.
func (n NetID) GoString() string {
	return n.String()
}

// MarshalText implements the TextMarshaler interface.
func (n NetID) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (n *NetID) UnmarshalText(data []byte) error {
	parsed, err := ParseHEX(string(data), 3)
	if err != nil {
		return err
	}
	copy(n[:], parsed[:])
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (n NetID) MarshalBinary() ([]byte, error) {
	return n.Bytes(), nil
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (n *NetID) UnmarshalBinary(data []byte) error {
	if len(data) != 3 {
		return errors.New("ttn/core: Invalid length for NetID")
	}
	copy(n[:], data)
	return nil
}

// MarshalTo is used by Protobuf
func (n *NetID) MarshalTo(b []byte) (int, error) {
	copy(b, n.Bytes())
	return 3, nil
}

// Size is used by Protobuf
func (n *NetID) Size() int {
	return 3
}

// Marshal implements the Marshaler interface.
func (n NetID) Marshal() ([]byte, error) {
	return n.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (n *NetID) Unmarshal(data []byte) error {
	*n = [3]byte{} // Reset the receiver
	return n.UnmarshalBinary(data)
}

// Equal returns whether n is equal to other
func (n NetID) Equal(other NetID) bool {
	return n == other
}

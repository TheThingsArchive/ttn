// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"strings"

	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// AES128Key is an 128 bit AES key.
type AES128Key [16]byte

// AppKey (Application Key) is used for LoRaWAN OTAA.
type AppKey AES128Key

// NwkSKey (Network Session Key) is used for LoRaWAN MIC calculation.
type NwkSKey AES128Key

// AppSKey (Application Session Key) is used for LoRaWAN payload encryption.
type AppSKey AES128Key

// ParseAES128Key parses a 128-bit hex-encoded string to an AES128Key
func ParseAES128Key(input string) (key AES128Key, err error) {
	bytes, err := ParseHEX(input, 16)
	if err != nil {
		return
	}
	copy(key[:], bytes)
	return
}

// Bytes returns the AES128Key as a byte slice
func (key AES128Key) Bytes() []byte {
	return key[:]
}

// String implements the Stringer interface.
func (key AES128Key) String() string {
	if key.IsEmpty() {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(key.Bytes()))
}

// GoString implements the GoStringer interface.
func (key AES128Key) GoString() string {
	return key.String()
}

// MarshalText implements the TextMarshaler interface.
func (key AES128Key) MarshalText() ([]byte, error) {
	return []byte(key.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (key *AES128Key) UnmarshalText(data []byte) error {
	parsed, err := ParseAES128Key(string(data))
	if err != nil {
		return err
	}
	*key = AES128Key(parsed)
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (key AES128Key) MarshalBinary() ([]byte, error) {
	return key.Bytes(), nil
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (key *AES128Key) UnmarshalBinary(data []byte) error {
	if len(data) != 16 {
		return errors.New("ttn/core: Invalid length for AES128Key")
	}
	copy(key[:], data)
	return nil
}

// MarshalTo is used by Protobuf
func (key *AES128Key) MarshalTo(b []byte) (int, error) {
	copy(b, key.Bytes())
	return 16, nil
}

// Size is used by Protobuf
func (key *AES128Key) Size() int {
	return 16
}

// Marshal implements the Marshaler interface.
func (key AES128Key) Marshal() ([]byte, error) {
	return key.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (key *AES128Key) Unmarshal(data []byte) error {
	*key = [16]byte{} // Reset the receiver
	return key.UnmarshalBinary(data)
}

// Equal returns whether key is equal to other
func (key AES128Key) Equal(other AES128Key) bool {
	return key == other
}

// ParseAppKey parses a 64-bit hex-encoded string to an AppKey
func ParseAppKey(input string) (key AppKey, err error) {
	aes128key, err := ParseAES128Key(input)
	if err != nil {
		return
	}
	key = AppKey(aes128key)
	return
}

// Bytes returns the AppKey as a byte slice
func (key AppKey) Bytes() []byte {
	return AES128Key(key).Bytes()
}

func (key AppKey) String() string {
	return AES128Key(key).String()
}

// GoString implements the GoStringer interface.
func (key AppKey) GoString() string {
	return key.String()
}

// MarshalText implements the TextMarshaler interface.
func (key AppKey) MarshalText() ([]byte, error) {
	return AES128Key(key).MarshalText()
}

// UnmarshalText implements the TextUnmarshaler interface.
func (key *AppKey) UnmarshalText(data []byte) error {
	e := AES128Key(*key)
	err := e.UnmarshalText(data)
	if err != nil {
		return err
	}
	*key = AppKey(e)
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (key AppKey) MarshalBinary() ([]byte, error) {
	return AES128Key(key).MarshalBinary()
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (key *AppKey) UnmarshalBinary(data []byte) error {
	e := AES128Key(*key)
	err := e.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	*key = AppKey(e)
	return nil
}

// MarshalTo is used by Protobuf
func (key *AppKey) MarshalTo(b []byte) (int, error) {
	copy(b, key.Bytes())
	return 16, nil
}

// Size is used by Protobuf
func (key *AppKey) Size() int {
	return 16
}

// Marshal implements the Marshaler interface.
func (key AppKey) Marshal() ([]byte, error) {
	return key.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (key *AppKey) Unmarshal(data []byte) error {
	*key = [16]byte{} // Reset the receiver
	return key.UnmarshalBinary(data)
}

// Equal returns whether key is equal to other
func (key AppKey) Equal(other AppKey) bool {
	return key == other
}

// ParseAppSKey parses a 64-bit hex-encoded string to an AppSKey
func ParseAppSKey(input string) (key AppSKey, err error) {
	aes128key, err := ParseAES128Key(input)
	if err != nil {
		return
	}
	key = AppSKey(aes128key)
	return
}

// Bytes returns the AppSKey as a byte slice
func (key AppSKey) Bytes() []byte {
	return AES128Key(key).Bytes()
}

func (key AppSKey) String() string {
	return AES128Key(key).String()
}

// GoString implements the GoStringer interface.
func (key AppSKey) GoString() string {
	return key.String()
}

// MarshalText implements the TextMarshaler interface.
func (key AppSKey) MarshalText() ([]byte, error) {
	return AES128Key(key).MarshalText()
}

// UnmarshalText implements the TextUnmarshaler interface.
func (key *AppSKey) UnmarshalText(data []byte) error {
	e := AES128Key(*key)
	err := e.UnmarshalText(data)
	if err != nil {
		return err
	}
	*key = AppSKey(e)
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (key AppSKey) MarshalBinary() ([]byte, error) {
	return AES128Key(key).MarshalBinary()
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (key *AppSKey) UnmarshalBinary(data []byte) error {
	e := AES128Key(*key)
	err := e.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	*key = AppSKey(e)
	return nil
}

// MarshalTo is used by Protobuf
func (key *AppSKey) MarshalTo(b []byte) (int, error) {
	copy(b, key.Bytes())
	return 16, nil
}

// Size is used by Protobuf
func (key *AppSKey) Size() int {
	return 16
}

// Marshal implements the Marshaler interface.
func (key AppSKey) Marshal() ([]byte, error) {
	return key.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (key *AppSKey) Unmarshal(data []byte) error {
	*key = [16]byte{} // Reset the receiver
	return key.UnmarshalBinary(data)
}

// Equal returns whether key is equal to other
func (key AppSKey) Equal(other AppSKey) bool {
	return key == other
}

// ParseNwkSKey parses a 64-bit hex-encoded string to an NwkSKey
func ParseNwkSKey(input string) (key NwkSKey, err error) {
	aes128key, err := ParseAES128Key(input)
	if err != nil {
		return
	}
	key = NwkSKey(aes128key)
	return
}

// Bytes returns the NwkSKey as a byte slice
func (key NwkSKey) Bytes() []byte {
	return AES128Key(key).Bytes()
}

// String implements the Stringer interface.
func (key NwkSKey) String() string {
	return AES128Key(key).String()
}

// GoString implements the GoStringer interface.
func (key NwkSKey) GoString() string {
	return key.String()
}

// MarshalText implements the TextMarshaler interface.
func (key NwkSKey) MarshalText() ([]byte, error) {
	return AES128Key(key).MarshalText()
}

// UnmarshalText implements the TextUnmarshaler interface.
func (key *NwkSKey) UnmarshalText(data []byte) error {
	e := AES128Key(*key)
	err := e.UnmarshalText(data)
	if err != nil {
		return err
	}
	*key = NwkSKey(e)
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (key NwkSKey) MarshalBinary() ([]byte, error) {
	return AES128Key(key).MarshalBinary()
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (key *NwkSKey) UnmarshalBinary(data []byte) error {
	e := AES128Key(*key)
	err := e.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	*key = NwkSKey(e)
	return nil
}

// MarshalTo is used by Protobuf
func (key *NwkSKey) MarshalTo(b []byte) (int, error) {
	copy(b, key.Bytes())
	return 16, nil
}

// Size is used by Protobuf
func (key *NwkSKey) Size() int {
	return 16
}

// Marshal implements the Marshaler interface.
func (key NwkSKey) Marshal() ([]byte, error) {
	return key.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (key *NwkSKey) Unmarshal(data []byte) error {
	*key = [16]byte{} // Reset the receiver
	return key.UnmarshalBinary(data)
}

// Equal returns whether key is equal to other
func (key NwkSKey) Equal(other NwkSKey) bool {
	return key == other
}

var emptyAES AES128Key

func (key AES128Key) IsEmpty() bool {
	return key == emptyAES
}

func (key AppKey) IsEmpty() bool {
	return AES128Key(key).IsEmpty()
}

func (key AppSKey) IsEmpty() bool {
	return AES128Key(key).IsEmpty()
}

func (key NwkSKey) IsEmpty() bool {
	return AES128Key(key).IsEmpty()
}

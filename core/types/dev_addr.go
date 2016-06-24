// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// DevAddr is a non-unique address for LoRaWAN devices.
type DevAddr [4]byte

// ParseDevAddr parses a 32-bit hex-encoded string to a DevAddr
func ParseDevAddr(input string) (addr DevAddr, err error) {
	bytes, err := ParseHEX(input, 4)
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

// MarshalTo is used by Protobuf
func (addr DevAddr) MarshalTo(b []byte) (int, error) {
	copy(b, addr.Bytes())
	return 4, nil
}

// Size is used by Protobuf
func (addr DevAddr) Size() int {
	return 4
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

// Mask returns a copy of the DevAddr with only the first "bits" bits
func (addr DevAddr) Mask(bits int) (masked DevAddr) {
	return empty.WithPrefix(addr, bits)
}

// WithPrefix returns the DevAddr, but with the first length bits replaced by the Prefix
func (addr DevAddr) WithPrefix(prefix DevAddr, length int) (prefixed DevAddr) {
	k := uint(length)
	for i := 0; i < 4; i++ {
		if k >= 8 {
			prefixed[i] = prefix[i] & 0xff
			k -= 8
			continue
		}
		prefixed[i] = (prefix[i] & ^byte(0xff>>k)) | (addr[i] & byte(0xff>>k))
		k = 0
	}
	return
}

// HasPrefix returns true if the DevAddr has a prefix of given length
func (addr DevAddr) HasPrefix(prefix DevAddr, length int) bool {
	return addr.Mask(length) == prefix.Mask(length)
}

// ParseDevAddrPrefix parses a DevAddr in prefix notation (01020304/24) to a prefix (01020300) and length (24)
func ParseDevAddrPrefix(prefix string) (addr DevAddr, length int, err error) {
	pattern := regexp.MustCompile("([[:xdigit:]]{8})/([[:digit:]]+)")
	matches := pattern.FindStringSubmatch(prefix)
	if len(matches) != 3 {
		err = errors.New("Invalid Prefix")
		return
	}
	addr, _ = ParseDevAddr(matches[1])   // errors handled in regexp
	length, _ = strconv.Atoi(matches[2]) // errors handled in regexp
	addr = addr.Mask(length)
	return
}

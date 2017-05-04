// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/TheThingsNetwork/ttn/utils/errors"
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
	if addr.IsEmpty() {
		return ""
	}
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

// Equal returns whether addr is equal to other
func (addr DevAddr) Equal(other DevAddr) bool {
	return addr == other
}

var empty DevAddr

func (addr DevAddr) IsEmpty() bool {
	return addr == empty
}

// DevAddrPrefix is a DevAddr with a prefix length
type DevAddrPrefix struct {
	DevAddr DevAddr
	Length  int
}

// ParseDevAddrPrefix parses a DevAddr in prefix notation (01020304/24) to a prefix
func ParseDevAddrPrefix(prefixString string) (prefix DevAddrPrefix, err error) {
	pattern := regexp.MustCompile("([[:xdigit:]]{8})/([[:digit:]]+)")
	matches := pattern.FindStringSubmatch(prefixString)
	if len(matches) != 3 {
		err = errors.New("Invalid Prefix")
		return
	}
	addr, _ := ParseDevAddr(matches[1])         // errors handled in regexp
	prefix.Length, _ = strconv.Atoi(matches[2]) // errors handled in regexp
	prefix.DevAddr = addr.Mask(prefix.Length)
	return
}

// Bytes returns the DevAddrPrefix as a byte slice
func (prefix DevAddrPrefix) Bytes() (bytes []byte) {
	bytes = append(bytes, byte(prefix.Length))
	bytes = append(bytes, prefix.DevAddr.Bytes()...)
	return bytes
}

// String implements the fmt.Stringer interface
func (prefix DevAddrPrefix) String() string {
	var addr string
	if prefix.DevAddr.IsEmpty() {
		addr = "00000000"
	} else {
		addr = prefix.DevAddr.String()
	}
	return fmt.Sprintf("%s/%d", addr, prefix.Length)
}

// MarshalText implements the TextMarshaler interface.
func (prefix DevAddrPrefix) MarshalText() ([]byte, error) {
	return []byte(prefix.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (prefix *DevAddrPrefix) UnmarshalText(data []byte) error {
	parsed, err := ParseDevAddrPrefix(string(data))
	if err != nil {
		return err
	}
	*prefix = DevAddrPrefix(parsed)
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (prefix DevAddrPrefix) MarshalBinary() ([]byte, error) {
	return prefix.Bytes(), nil
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (prefix *DevAddrPrefix) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return errors.New("ttn/core: Invalid length for DevAddrPrefix")
	}
	copy(prefix.DevAddr[:], data[1:])
	prefix.Length = int(data[0])
	prefix.DevAddr = prefix.DevAddr.Mask(prefix.Length)
	return nil
}

// MarshalTo is used by Protobuf
func (prefix DevAddrPrefix) MarshalTo(b []byte) (int, error) {
	copy(b, prefix.Bytes())
	return 4, nil
}

// Size is used by Protobuf
func (prefix DevAddrPrefix) Size() int {
	return 5
}

// Marshal implements the Marshaler interface.
func (prefix DevAddrPrefix) Marshal() ([]byte, error) {
	return prefix.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (prefix *DevAddrPrefix) Unmarshal(data []byte) error {
	prefix.DevAddr = [4]byte{} // Reset the receiver
	prefix.Length = 0
	return prefix.UnmarshalBinary(data)
}

// Mask returns a copy of the DevAddr with only the first "bits" bits
func (addr DevAddr) Mask(bits int) (masked DevAddr) {
	return empty.WithPrefix(DevAddrPrefix{addr, bits})
}

// WithPrefix returns the DevAddr, but with the first length bits replaced by the Prefix
func (addr DevAddr) WithPrefix(prefix DevAddrPrefix) (prefixed DevAddr) {
	k := uint(prefix.Length)
	for i := 0; i < 4; i++ {
		if k >= 8 {
			prefixed[i] = prefix.DevAddr[i] & 0xff
			k -= 8
			continue
		}
		prefixed[i] = (prefix.DevAddr[i] & ^byte(0xff>>k)) | (addr[i] & byte(0xff>>k))
		k = 0
	}
	return
}

// HasPrefix returns true if the DevAddr has a prefix of given length
func (addr DevAddr) HasPrefix(prefix DevAddrPrefix) bool {
	return addr.Mask(prefix.Length) == prefix.DevAddr.Mask(prefix.Length)
}

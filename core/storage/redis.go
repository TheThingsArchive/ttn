// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

// FormatFloat32 does what its name suggests
func FormatFloat32(value float32) string {
	return strconv.FormatFloat(float64(value), 'f', -1, 32)
}

// FormatFloat64 does what its name suggests
func FormatFloat64(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// FormatInt32 does what its name suggests
func FormatInt32(value int32) string {
	return FormatInt64(int64(value))
}

// FormatInt64 does what its name suggests
func FormatInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

// FormatUint32 does what its name suggests
func FormatUint32(value uint32) string {
	return FormatUint64(uint64(value))
}

// FormatUint64 does what its name suggests
func FormatUint64(value uint64) string {
	return strconv.FormatUint(value, 10)
}

// FormatBool does what its name suggests
func FormatBool(value bool) string {
	return strconv.FormatBool(value)
}

// FormatBytes does what its name suggests
func FormatBytes(value []byte) string {
	return fmt.Sprintf("%X", value)
}

// ParseFloat32 does what its name suggests
func ParseFloat32(val string) (float32, error) {
	res, err := strconv.ParseFloat(val, 32)
	return float32(res), err
}

// ParseFloat64 does what its name suggests
func ParseFloat64(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}

// ParseInt32 does what its name suggests
func ParseInt32(val string) (int32, error) {
	res, err := strconv.ParseInt(val, 10, 32)
	return int32(res), err
}

// ParseInt64 does what its name suggests
func ParseInt64(val string) (int64, error) {
	return strconv.ParseInt(val, 10, 64)
}

// ParseUint32 does what its name suggests
func ParseUint32(val string) (uint32, error) {
	res, err := strconv.ParseUint(val, 10, 32)
	return uint32(res), err
}

// ParseUint64 does what its name suggests
func ParseUint64(val string) (uint64, error) {
	return strconv.ParseUint(val, 10, 64)
}

// ParseBool does what its name suggests
func ParseBool(val string) (bool, error) {
	return strconv.ParseBool(val)
}

// ParseBytes does what its name suggests
func ParseBytes(val string) ([]byte, error) {
	return hex.DecodeString(val)
}

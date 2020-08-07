// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"errors"
)

var errInvalidLength = errors.New("wrong input length")

// ParseHEX parses a string "input" to a byteslice with length "length".
func ParseHEX(input string, length int) ([]byte, error) {
	if input == "" {
		return make([]byte, length), nil
	}

	if len(input) != length*2 {
		return nil, errInvalidLength
	}

	return hex.DecodeString(input)
}

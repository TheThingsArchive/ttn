// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"fmt"
	"regexp"
)

// ParseHEX parses a string "input" to a byteslice with length "length".
func ParseHEX(input string, length int) ([]byte, error) {
	if input == "" {
		return make([]byte, length), nil
	}

	pattern := regexp.MustCompile(fmt.Sprintf("^[[:xdigit:]]{%d}$", length*2))

	valid := pattern.MatchString(input)
	if !valid {
		return nil, fmt.Errorf("Invalid input: %s is not hex", input)
	}

	slice, _ := hex.DecodeString(input)

	return slice, nil
}

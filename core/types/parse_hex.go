// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"fmt"
	"regexp"
)

// ParseHEX parses a string "input" to a byteslice with length "length".
func ParseHEX(input string, length int) ([]byte, error) {
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

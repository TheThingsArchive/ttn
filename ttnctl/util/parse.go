// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"encoding/hex"
	"fmt"
	"regexp"
)

// ParseHEX parses a hexidecimal string to a byte slice
func ParseHEX(input string, length int) ([]byte, error) {
	pattern, err := regexp.Compile(fmt.Sprintf("[[:xdigit:]]{%d}", length))
	if err != nil {
		return nil, fmt.Errorf("Invalid pattern")
	}

	valid := pattern.MatchString(input)
	if !valid {
		return nil, fmt.Errorf("Invalid input")
	}

	devAddr, err := hex.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("Could not decode input")
	}

	return devAddr, nil
}

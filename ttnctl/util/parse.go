// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"encoding/hex"
	"fmt"
	"regexp"
)

func parseHEX(input string, length int) ([]byte, error) {
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

// Parse32 parses a 32-bit hex-encoded string
func Parse32(input string) ([]byte, error) {
	return parseHEX(input, 8)
}

// Parse64 parses a 64-bit hex-encoded string
func Parse64(input string) ([]byte, error) {
	return parseHEX(input, 16)
}

// Parse128 parses a 128-bit hex-encoded string
func Parse128(input string) ([]byte, error) {
	return parseHEX(input, 32)
}

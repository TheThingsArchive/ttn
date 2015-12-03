// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package pointer provides helper method to quickly define pointer from basic go types
package pointer

// String creates a pointer to a string from a string value
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

// Int creates a pointer to an int from an int value
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// Uint creates a pointer to an unsigned int from an unsigned int value
func Uint(v uint) *uint {
	p := new(uint)
	*p = v
	return p
}

// Float64 creates a pointer to a float64 from a float64 value
func Float64(v float64) *float64 {
	p := new(float64)
	*p = v
	return p
}

// Bool creates a pointer to a boolean from a boolean value
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

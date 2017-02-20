// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package random

import (
	"time"

	"github.com/TheThingsNetwork/go-utils/pseudorandom"
	"github.com/TheThingsNetwork/go-utils/random"
)

// TTNRandom is used as a wrapper around math/rand
type TTNRandom struct {
	random.Interface
}

// New returns a new TTNRandom, in most cases you can just use the global funcs
func New() *TTNRandom {
	return &TTNRandom{
		Interface: pseudorandom.New(time.Now().UnixNano()),
	}
}

var global = New()

// Intn returns a random int with max n
func Intn(n int) int {
	return global.Intn(n)
}

// String returns a random string of length n
func String(n int) string {
	return global.String(n)
}

// Bytes generates a random byte slice of length n
func Bytes(n int) []byte {
	return global.Bytes(n)
}

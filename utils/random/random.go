// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package random

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/TheThingsNetwork/go-utils/pseudorandom"
	"github.com/TheThingsNetwork/go-utils/random"
	"github.com/TheThingsNetwork/ttn/core/types"
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

// ID returns randomly generated ID
func (r *TTNRandom) ID() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.String(2 + rand.Intn(61))
}

// ByteArray generates a random byte array of length n
func (r *TTNRandom) ByteArray(n int) interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	var b byte
	v := reflect.New(reflect.ArrayOf(n, reflect.TypeOf(b)))
	for i, b := range global.Bytes(n) {
		v.Index(i).Set(reflect.ValueOf(b))
	}
	return v.Interface()
}

func (r *TTNRandom) Bool() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Int63()%2 == 0
}

func (r *TTNRandom) DevNonce() types.DevNonce {
	return types.DevNonce(r.ByteArray(2).([2]byte))
}
func (r *TTNRandom) AppNonce() types.AppNonce {
	return types.AppNonce(r.ByteArray(3).([3]byte))
}
func (r *TTNRandom) NetID() types.NetID {
	return types.NetID(r.ByteArray(3).([3]byte))
}
func (r *TTNRandom) DevAddr() types.DevAddr {
	return types.DevAddr(r.ByteArray(4).([4]byte))
}
func (r *TTNRandom) EUI64() types.EUI64 {
	return types.EUI64(r.ByteArray(8).([8]byte))
}
func (r *TTNRandom) DevEUI() types.DevEUI {
	return types.DevEUI(r.EUI64())
}
func (r *TTNRandom) AppEUI() types.AppEUI {
	return types.AppEUI(r.EUI64())
}

// Intn returns random int with max n
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

// ByteArray generates a random byte array of length n
func ByteArray(n int) interface{} {
	return global.ByteArray(n)
}

// Bool generates a random boolean
func Bool() bool {
	return global.Bool()
}

func ID() string {
	return global.ID()
}

func DevNonce() types.DevNonce {
	return global.DevNonce()
}
func AppNonce() types.AppNonce {
	return global.AppNonce()
}
func NetID() types.NetID {
	return global.NetID()
}
func DevAddr() types.DevAddr {
	return global.DevAddr()
}
func EUI64() types.EUI64 {
	return global.EUI64()
}
func DevEUI() types.DevEUI {
	return global.DevEUI()
}
func AppEUI() types.AppEUI {
	return global.AppEUI()
}

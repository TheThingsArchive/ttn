// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package random

import (
	"fmt"
	"math"
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

const validIDChars = "abcdefghijklmnopqrstuvwxyz1234567890"

func (r *TTNRandom) randomChar(alphabet string) byte { return alphabet[r.Intn(len(alphabet))] }

func (r *TTNRandom) randomChars(alphabet string, chars int) []byte {
	o := make([]byte, chars)
	for n := 0; n < chars; n++ {
		o[n] = r.randomChar(alphabet)
	}
	return o
}

func (r *TTNRandom) id(length int) string {
	o := r.randomChars(validIDChars, length)
	for n := 0; n < length/8; n++ { // max 1 out of 8 will be a dash/underscore
		l := 1 + r.Intn(length-2)
		if o[l-1] != '_' && o[l-1] != '-' && o[l+1] != '_' && o[l+1] != '-' {
			o[l] = r.randomChar("-_")
		}
	}
	return string(o)
}

// ID returns randomly generated ID
func (r *TTNRandom) ID() string {
	return r.id(2 + r.Intn(35))
}

// AppID returns randomly generated AppID
func (r *TTNRandom) AppID() string {
	return r.ID()
}

// DevID returns randomly generated DevID
func (r *TTNRandom) DevID() string {
	return r.ID()
}

// Bool return randomly generated bool value
func (r *TTNRandom) Bool() bool {
	return r.Interface.Intn(2) == 0
}

// RSSI generates RSSI signal between -120 < rssi < 0
func (r *TTNRandom) RSSI() int32 {
	// Generate RSSI. Tend towards generating great signal strength.
	x := float64(r.Interface.Intn(math.MaxInt32)) * float64(2e-9)
	return int32(-1.6 * math.Exp(x))
}

var freqs = []float32{
	// EU
	868.1,
	868.3,
	868.5,
	868.8,
	867.1,
	867.3,
	867.5,
	867.7,
	867.9,

	// US
	903.9,
	904.1,
	904.3,
	904.5,
	904.7,
	904.9,
	905.1,
	905.3,
	904.6,
}

// Freq generates a frequency between 867.1 and 905.3 Mhz
func (r *TTNRandom) Freq() float32 {
	return freqs[r.Interface.Intn(len(freqs))]
}

// Datr generates Datr for instance: SF4BW125
func (r *TTNRandom) Datr() string {
	// Spread Factor from 12 to 7
	sf := 12 - r.Interface.Intn(7)
	var bw int
	if sf == 6 {
		// DR6 -> SF7@250Khz
		sf = 7
		bw = 250
	} else {
		bw = 125
	}
	return fmt.Sprintf("SF%dBW%d", sf, bw)
}

// Codr generates Codr for instance: 4/6
func (r *TTNRandom) Codr() string {
	d := r.Interface.Intn(4) + 5
	return fmt.Sprintf("4/%d", d)
}

// LSNR generates LoRa SNR ratio in db. Tend towards generating good ratio with low noise
func (r *TTNRandom) LSNR() float32 {
	x := float64(r.Interface.Intn(math.MaxInt32)) * float64(2e-9)
	return float32(math.Floor((-0.1*math.Exp(x)+5.5)*10) / 10)
}

func (r *TTNRandom) DevNonce() (devNonce types.DevNonce) {
	r.Interface.FillBytes(devNonce[:])
	return
}
func (r *TTNRandom) AppNonce() (appNonce types.AppNonce) {
	r.Interface.FillBytes(appNonce[:])
	return
}
func (r *TTNRandom) NetID() (netID types.NetID) {
	r.Interface.FillBytes(netID[:])
	return
}
func (r *TTNRandom) DevAddr() (devAddr types.DevAddr) {
	r.Interface.FillBytes(devAddr[:])
	return
}
func (r *TTNRandom) EUI64() (eui types.EUI64) {
	r.Interface.FillBytes(eui[:])
	return
}
func (r *TTNRandom) DevEUI() (eui types.DevEUI) {
	return types.DevEUI(r.EUI64())
}
func (r *TTNRandom) AppEUI() (eui types.AppEUI) {
	return types.AppEUI(r.EUI64())
}

// RSSI generates RSSI signal between -120 < rssi < 0
func RSSI() int32 {
	return global.RSSI()
}

// Freq generates a frequency between 865.0 and 870.0 Mhz
func Freq() float32 {
	return global.Freq()
}

// Datr generates Datr for instance: SF4BW125
func Datr() string {
	return global.Datr()
}

// Codr generates Codr for instance: 4/6
func Codr() string {
	return global.Codr()
}

// LSNR generates LoRa SNR ratio in db. Tend towards generating good ratio with low noise
func LSNR() float32 {
	return global.LSNR()
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

// Bool generates a random boolean
func Bool() bool {
	return global.Bool()
}

func ID() string {
	return global.ID()
}
func AppID() string {
	return global.AppID()
}
func DevID() string {
	return global.DevID()
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

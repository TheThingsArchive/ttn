// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package random

import (
	"fmt"
	"math"
	"strings"
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

func (r *TTNRandom) randomChar() byte {
	return byte('a' + r.Intn('z'-'a'))
}
func (r *TTNRandom) randomCharString() string {
	return string(r.randomChar())
}

func (r *TTNRandom) formatID(s string) string {
	switch {
	case strings.HasPrefix(s, "-"):
		s = strings.TrimPrefix(s, "-") + r.randomCharString()
	case strings.HasPrefix(s, "_"):
		s = strings.TrimPrefix(s, "_") + r.randomCharString()
	case strings.HasSuffix(s, "-"):
		s = strings.TrimSuffix(s, "-") + r.randomCharString()
	case strings.HasSuffix(s, "_"):
		s = strings.TrimSuffix(s, "_") + r.randomCharString()
	}
	return strings.ToLower(strings.NewReplacer(
		"_-", r.randomCharString()+r.randomCharString(),
		"-_", r.randomCharString()+r.randomCharString(),
		"__", r.randomCharString()+r.randomCharString(),
		"--", r.randomCharString()+r.randomCharString(),
	).Replace(s))
}

// ID returns randomly generated ID
func (r *TTNRandom) ID() string {
	return r.formatID(r.Interface.String(2 + r.Interface.Intn(61)))
}

// AppID returns randomly generated AppID
func (r *TTNRandom) AppID() string {
	return r.formatID(r.Interface.String(2 + r.Interface.Intn(34)))
}

// DevID returns randomly generated DevID
func (r *TTNRandom) DevID() string {
	return r.formatID(r.Interface.String(2 + r.Interface.Intn(34)))
}

// Bool return randomly generated bool value
func (r *TTNRandom) Bool() bool {
	return r.Interface.Intn(2) == 0
}

// Rssi generates RSSI signal between -120 < rssi < 0
func (r *TTNRandom) Rssi() int32 {
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

// Lsnr generates LoRa SNR ratio in db. Tend towards generating good ratio with low noise
func (r *TTNRandom) Lsnr() float32 {
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

// Rssi generates RSSI signal between -120 < rssi < 0
func Rssi() int32 {
	return global.Rssi()
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

// Lsnr generates LoRa SNR ratio in db. Tend towards generating good ratio with low noise
func Lsnr() float32 {
	return global.Lsnr()
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

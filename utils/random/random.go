// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package random

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Source: http://stackoverflow.com/a/31832326

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.New(rand.NewSource(time.Now().UnixNano()))

// String returns random string of length n
func String(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// Token generate a random 2-bytes token
func Token() []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, src.Uint32())
	return b[0:2]
}

// Rssi generates RSSI signal between -120 < rssi < 0
func Rssi() int32 {
	// Generate RSSI. Tend towards generating great signal strength.
	x := float64(src.Int31()) * float64(2e-9)
	return int32(-1.6 * math.Exp(x))
}

// Freq generates a frequency between 865.0 and 870.0 Mhz
func Freq() float32 {
	// EU 865-870MHz
	return float32(src.Float64()*5 + 865.0)
}

// Datr generates Datr for instance: SF4BW125
func Datr() string {
	// Spread Factor from 12 to 7
	sf := 12 - src.Intn(7)
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
func Codr() string {
	d := src.Intn(4) + 5
	return fmt.Sprintf("4/%d", d)
}

// Lsnr generates LoRa SNR ratio in db. Tend towards generating good ratio with low noise
func Lsnr() float32 {
	x := float64(src.Int31()) * float64(2e-9)
	return float32(math.Floor((-0.1*math.Exp(x)+5.5)*10) / 10)
}

// Bytes generates a random byte slice of length n
func Bytes(n int) []byte {
	p := make([]byte, n)
	src.Read(p)
	return p
}

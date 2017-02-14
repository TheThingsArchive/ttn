// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package random

import (
	crypto "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Source: http://stackoverflow.com/a/31832326

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// TTNRandom is used as a wrapper around math/rand
type TTNRandom struct {
	mu   sync.Mutex
	rand *rand.Rand
}

// New returns a new TTNRandom, in most cases you can just use the global funcs
func New() *TTNRandom {
	return &TTNRandom{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

var global = New()

// Intn wraps rand.Intn
func (r *TTNRandom) Intn(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Intn(n)
}

// String returns random string of length n
func (r *TTNRandom) String(n int) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	b := make([]byte, n)
	// A Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, r.rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = r.rand.Int63(), letterIdxMax
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
func (r *TTNRandom) Token() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, r.rand.Uint32())
	return b[0:2]
}

// Rssi generates RSSI signal between -120 < rssi < 0
func (r *TTNRandom) Rssi() int32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Generate RSSI. Tend towards generating great signal strength.
	x := float64(r.rand.Int31()) * float64(2e-9)
	return int32(-1.6 * math.Exp(x))
}

var euFreqs = []float32{
	868.1,
	868.3,
	868.5,
	868.8,
	867.1,
	867.3,
	867.5,
	867.7,
	867.9,
}

var usFreqs = []float32{
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

// Freq generates a frequency between 865.0 and 870.0 Mhz
func (r *TTNRandom) Freq() float32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return usFreqs[r.rand.Intn(len(usFreqs))]
}

// Datr generates Datr for instance: SF4BW125
func (r *TTNRandom) Datr() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Spread Factor from 12 to 7
	sf := 12 - r.rand.Intn(7)
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
	r.mu.Lock()
	defer r.mu.Unlock()
	d := r.rand.Intn(4) + 5
	return fmt.Sprintf("4/%d", d)
}

// Lsnr generates LoRa SNR ratio in db. Tend towards generating good ratio with low noise
func (r *TTNRandom) Lsnr() float32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	x := float64(r.rand.Int31()) * float64(2e-9)
	return float32(math.Floor((-0.1*math.Exp(x)+5.5)*10) / 10)
}

// Intn returns random int with max n
func Intn(n int) int {
	return global.Intn(n)
}

// String returns random string of length n
func String(n int) string {
	return global.String(n)
}

// Token generate a random 2-bytes token
func Token() []byte {
	return global.Token()
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

// Bytes generates a random byte slice of length n
func Bytes(n int) []byte {
	p := make([]byte, n)
	_, err := crypto.Read(p)
	if err != nil {
		panic(fmt.Errorf("random.Bytes: %s", err))
	}
	return p
}

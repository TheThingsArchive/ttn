// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/brocaar/lorawan"
)

func genToken() []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, rand.Uint32())
	return b[0:2]
}

func ackToken(index int, packet semtech.Packet) [4]byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(index))
	id := buf.Bytes()[0]

	var kind byte
	switch packet.Identifier {
	case semtech.PUSH_ACK, semtech.PUSH_DATA:
		kind = 0x1
	case semtech.PULL_ACK, semtech.PULL_DATA, semtech.PULL_RESP:
		kind = 0x2
	}

	return [4]byte{id, kind, packet.Token[0], packet.Token[1]}
}

// Generates RSSI signal between -120 < rssi < 0
func generateRssi() int {
	// Generate RSSI. Tend towards generating great signal strength.
	x := float64(rand.Int31()) * float64(2e-9)
	return int(-1.6 * math.Exp(x))
}

// Generates a frequency between 863.0 and 870.0 Mhz
func generateFreq() float64 {
	// EU 863-870MHz
	return rand.Float64()*7 + 863.0
}

// Generates Datr for instance: SF4BW125
func generateDatr() string {
	// Spread Factor from 12 to 7
	sf := 12 - rand.Intn(7)
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

// Generates Codr for instance: 4/6
func generateCodr() string {
	d := rand.Intn(4) + 5
	return fmt.Sprintf("4/%d", d)
}

// Generates LoRa SNR ratio in db. Tend towards generating good ratio with low noise
func generateLsnr() float64 {
	x := float64(rand.Int31()) * float64(2e-9)
	return math.Floor((-0.1*math.Exp(x)+5.5)*10) / 10
}

// Generates fake data from a device
func generateData(frmData string, devAddr lorawan.DevAddr) string {
	macPayload := lorawan.NewMACPayload(true)
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: devAddr,
		FCtrl:   lorawan.FCtrl{},
		FCnt:    0,
	}
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{
		Bytes: []byte(frmData),
	}}
	macPayload.FPort = 14

	phyPayload := lorawan.NewPHYPayload(true)
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	phyPayload.MACPayload = macPayload
	phyPayload.SetMIC(lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}))

	raw, err := phyPayload.MarshalBinary()
	if err != nil { // Shouldn't be
		panic(err)
	}
	return strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")
}

// Generate a random device address
func generateDevAddr() lorawan.DevAddr {
	devAddr := [4]byte{}
	token := new(bytes.Buffer)
	binary.Write(token, binary.LittleEndian, time.Now().UnixNano())
	copy(devAddr[:], token.Bytes()[:4])
	return lorawan.DevAddr(devAddr)
}

func generateRXPK(data string, devAddr lorawan.DevAddr) semtech.RXPK {
	now := time.Now().In(time.UTC)
	return semtech.RXPK{
		Time: &now,
		Tmst: pointer.Uint(uint(now.UnixNano())),
		Freq: pointer.Float64(generateFreq()),
		Chan: pointer.Uint(0),                             // Irrelevant
		Rfch: pointer.Uint(0),                             // Irrelevant
		Stat: pointer.Int(1),                              // Assuming CRC was ok
		Modu: pointer.String("LORA"),                      // For now, only consider LORA modulation
		Datr: pointer.String(generateDatr()),              // Arbitrary
		Codr: pointer.String("4/5"),                       // Arbitrary
		Rssi: pointer.Int(generateRssi()),                 // Arbitrary
		Lsnr: pointer.Float64(generateLsnr()),             // Arbitrary
		Data: pointer.String(generateData(data, devAddr)), // Arbitrary
	}
}

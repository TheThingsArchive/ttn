// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/udp"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

// ----- build utilities
type mockServer struct {
	conn     *net.UDPConn
	response chan semtech.Packet
}

// Generate a mock server that will send packet through a udp connection and communicate back
// received packet.
func genMockServer(port uint) mockServer {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	response := make(chan semtech.Packet)
	go func() {
		for {
			buf := make([]byte, 5000)
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				panic(err)
			}
			packet := new(semtech.Packet)
			if err = packet.UnmarshalBinary(buf[:n]); err != nil {
				panic(err)
			}
			response <- *packet
		}
	}()

	return mockServer{conn: conn, response: response}
}

// Send a packet through the udp mock server toward the adapter
func (s mockServer) send(p semtech.Packet) semtech.Packet {
	raw, err := p.MarshalBinary()
	if err != nil {
		panic(err)
	}
	s.conn.Write(raw)
	select {
	case packet := <-s.response:
		return packet
	case <-time.After(100 * time.Millisecond):
		return semtech.Packet{}
	}
}

// Generates an adapter as well as a channel that behaves like the Next() methods (but can be used
// in a select for timeout)
func genAdapter(t *testing.T, port uint) (*udp.Adapter, chan interface{}) {
	// Logging
	ctx := GetLogger(t, "Adapter")

	net := fmt.Sprintf("0.0.0.0:%d", port)
	adapter, err := udp.NewAdapter(net, ctx)
	if err != nil {
		panic(err)
	}
	adapter.Bind(Semtech{})
	next := make(chan interface{})
	go func() {
		for {
			packet, _, err := adapter.Next()
			next <- struct {
				err    error
				packet []byte
			}{err: err, packet: packet}
		}
	}()
	return adapter, next
}

// Generate a core packet from a semtech packet that has one RXPK
func genCorePacket(p semtech.Packet) core.Packet {
	if p.Payload == nil || len(p.Payload.RXPK) != 1 {
		panic("Expected a payload with one rxpk")
	}
	packet, err := rxpk2packet(p.Payload.RXPK[0], p.GatewayId)
	if err != nil {
		panic(err)
	}
	return packet
}

func genPUSHDATANoRXPK(token []byte) semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		GatewayId:  []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		Token:      token,
		Identifier: semtech.PUSH_DATA,
	}
}

func genPUSHDATANoPayload(token []byte) semtech.Packet {
	packet := genPUSHDATAWithRXPK(token)
	packet.Payload.RXPK[0].Data = nil
	return packet
}

func genPUSHDATAWithRXPK(token []byte) semtech.Packet {
	packet := genPUSHDATANoRXPK(token)
	packet.Payload = &semtech.Payload{
		RXPK: []semtech.RXPK{
			semtech.RXPK{
				Rssi: pointer.Int(-60),
				Codr: pointer.String("4/7"),
				Data: pointer.String(genRXPKData()),
			},
		},
	}
	return packet
}

func genPULLACK(token []byte) semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		Token:      token,
		Identifier: semtech.PULL_ACK,
	}
}

func genPUSHACK(token []byte) semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		Token:      token,
		Identifier: semtech.PUSH_ACK,
	}
}

func genPULLDATA(token []byte) semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		Token:      token,
		GatewayId:  []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		Identifier: semtech.PULL_DATA,
	}
}

func genPHYPayload(uplink bool) lorawan.PHYPayload {
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	macPayload := lorawan.NewMACPayload(uplink)
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: lorawan.DevAddr([4]byte{1, 2, 3, 4}),
		FCtrl: lorawan.FCtrl{
			ADR:       false,
			ADRACKReq: false,
			ACK:       false,
		},
		FCnt: 0,
	}
	macPayload.FPort = 10
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{1, 2, 3, 4}}}

	if err := macPayload.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	payload := lorawan.NewPHYPayload(uplink)
	payload.MHDR = lorawan.MHDR{
		MType: lorawan.ConfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	payload.MACPayload = macPayload

	if err := payload.SetMIC(nwkSKey); err != nil {
		panic(err)
	}

	return payload
}

func genRXPKData() string {
	// 1. Generate a physical payload
	payload := genPHYPayload(true)

	// 2. Generate a JSON payload received by the server
	raw, err := payload.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

// Generate a Metadata object that matches RXPK metadata
func genMetadata(RXPK semtech.RXPK) core.Metadata {
	return core.Metadata{
		Chan: RXPK.Chan,
		Codr: RXPK.Codr,
		Freq: RXPK.Freq,
		Lsnr: RXPK.Lsnr,
		Modu: RXPK.Modu,
		Rfch: RXPK.Rfch,
		Rssi: RXPK.Rssi,
		Size: RXPK.Size,
		Stat: RXPK.Stat,
		Time: RXPK.Time,
		Tmst: RXPK.Tmst,
	}
}

// Generates a Metadata object with all field completed with relevant values
func genFullMetadata() core.Metadata {
	timeRef := time.Date(2016, 1, 13, 14, 11, 28, 207288421, time.UTC)
	return core.Metadata{
		Chan: pointer.Uint(2),
		Codr: pointer.String("4/6"),
		Datr: pointer.String("LORA"),
		Fdev: pointer.Uint(3),
		Freq: pointer.Float64(863.125),
		Imme: pointer.Bool(false),
		Ipol: pointer.Bool(false),
		Lsnr: pointer.Float64(5.2),
		Modu: pointer.String("LORA"),
		Ncrc: pointer.Bool(true),
		Powe: pointer.Uint(3),
		Prea: pointer.Uint(8),
		Rfch: pointer.Uint(2),
		Rssi: pointer.Int(-27),
		Size: pointer.Uint(14),
		Stat: pointer.Int(0),
		Time: pointer.Time(timeRef),
		Tmst: pointer.Uint(uint(timeRef.UnixNano())),
	}
}

// Generate an RXPK packet using the given payload as Data
func genRXPK(phyPayload lorawan.PHYPayload) semtech.RXPK {
	raw, err := phyPayload.MarshalBinary()
	if err != nil {
		panic(err)
	}
	data := strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")

	return semtech.RXPK{
		Chan: pointer.Uint(2),
		Codr: pointer.String("4/6"),
		Data: pointer.String(data),
		Freq: pointer.Float64(863.125),
		Lsnr: pointer.Float64(5.2),
		Modu: pointer.String("LORA"),
		Rfch: pointer.Uint(2),
		Rssi: pointer.Int(-27),
		Size: pointer.Uint(uint(len([]byte(data)))),
		Stat: pointer.Int(0),
		Time: pointer.Time(time.Now()),
		Tmst: pointer.Uint(uint(time.Now().UnixNano())),
	}
}

// Generates a TXPK packet using the given payload and the given metadata
func genTXPK(phyPayload lorawan.PHYPayload, metadata core.Metadata) semtech.TXPK {
	raw, err := phyPayload.MarshalBinary()
	if err != nil {
		panic(err)
	}
	data := strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")
	return semtech.TXPK{
		Codr: metadata.Codr,
		Data: pointer.String(data),
		Datr: metadata.Datr,
		Fdev: metadata.Fdev,
		Freq: metadata.Freq,
		Imme: metadata.Imme,
		Ipol: metadata.Ipol,
		Modu: metadata.Modu,
		Ncrc: metadata.Ncrc,
		Powe: metadata.Powe,
		Prea: metadata.Prea,
		Rfch: metadata.Rfch,
		Size: metadata.Size,
		Time: metadata.Time,
		Tmst: metadata.Tmst,
	}
}

// Generates a test suite where the RXPK is fully complete
func genRXPKWithFullMetadata(test *convertRXPKTest) convertRXPKTest {
	phyPayload := genPHYPayload(true)
	rxpk := genRXPK(phyPayload)
	metadata := genMetadata(rxpk)
	test.CorePacket, _ = core.NewRPacket(phyPayload, []byte{1, 2, 3, 4, 5, 6, 7, 8}, metadata)
	test.RXPK = rxpk
	return *test
}

// Generates a test suite where the RXPK contains partial metadata
func genRXPKWithPartialMetadata(test *convertRXPKTest) convertRXPKTest {
	phyPayload := genPHYPayload(true)
	rxpk := genRXPK(phyPayload)
	rxpk.Codr = nil
	rxpk.Rfch = nil
	rxpk.Rssi = nil
	rxpk.Time = nil
	rxpk.Size = nil
	metadata := genMetadata(rxpk)
	test.CorePacket, _ = core.NewRPacket(phyPayload, []byte{1, 2, 3, 4, 5, 6, 7, 8}, metadata)
	test.RXPK = rxpk
	return *test
}

// Generates a test suite where the RXPK contains no data
func genRXPKWithNoData(test *convertRXPKTest) convertRXPKTest {
	rxpk := genRXPK(genPHYPayload(true))
	rxpk.Data = nil
	test.RXPK = rxpk
	return *test
}

// Generates a test suite where the core packet has all txpk metadata
func genCoreFullMetadata(test *convertToTXPKTest) convertToTXPKTest {
	phyPayload := genPHYPayload(false)
	metadata := genFullMetadata()
	metadata.Chan = nil
	metadata.Lsnr = nil
	metadata.Rssi = nil
	metadata.Stat = nil
	test.TXPK = genTXPK(phyPayload, metadata)
	test.CorePacket, _ = core.NewRPacket(phyPayload, []byte{1, 2, 3, 4, 5, 6, 7, 8}, metadata)
	return *test
}

// Generates a test suite where the core packet has no metadata
func genCoreNoMetadata(test *convertToTXPKTest) convertToTXPKTest {
	phyPayload := genPHYPayload(false)
	metadata := core.Metadata{}
	test.TXPK = genTXPK(phyPayload, metadata)
	test.CorePacket, _ = core.NewRPacket(phyPayload, []byte{1, 2, 3, 4, 5, 6, 7, 8}, metadata)
	return *test
}

// Generates a test suite where the core packet has partial metadata but all supported
func genCorePartialMetadata(test *convertToTXPKTest) convertToTXPKTest {
	phyPayload := genPHYPayload(false)
	metadata := genFullMetadata()
	metadata.Chan = nil
	metadata.Lsnr = nil
	metadata.Rssi = nil
	metadata.Stat = nil
	metadata.Modu = nil
	metadata.Fdev = nil
	metadata.Time = nil
	test.TXPK = genTXPK(phyPayload, metadata)
	test.CorePacket, _ = core.NewRPacket(phyPayload, []byte{1, 2, 3, 4, 5, 6, 7, 8}, metadata)
	return *test
}

// Generates a test suite where the core packet has extra metadata not supported by txpk
func genCoreExtraMetadata(test *convertToTXPKTest) convertToTXPKTest {
	phyPayload := genPHYPayload(false)
	metadata := genFullMetadata()
	test.TXPK = genTXPK(phyPayload, metadata)
	test.CorePacket, _ = core.NewRPacket(phyPayload, []byte{1, 2, 3, 4, 5, 6, 7, 8}, metadata)
	return *test
}

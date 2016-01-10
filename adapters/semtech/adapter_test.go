// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"encoding/base64"
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
	components "github.com/thethingsnetwork/core/refactored_components"
	"github.com/thethingsnetwork/core/semtech"
	"github.com/thethingsnetwork/core/utils/log"
	"github.com/thethingsnetwork/core/utils/pointer"
	. "github.com/thethingsnetwork/core/utils/testing"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestNewAdapter(t *testing.T) {
	Ok(t, "pending")
}

func TestSend(t *testing.T) {
	Desc(t, "Send is not supported")
	adapter, err := NewAdapter(33000)
	if err != nil {
		panic(err)
	}
	err = adapter.Send(core.Packet{})
	checkErrors(t, ErrNotSupported, err)
}

func TestNextRegistration(t *testing.T) {
	Desc(t, "Next registration is not supported")
	adapter, err := NewAdapter(33001)
	if err != nil {
		panic(err)
	}
	_, _, err = adapter.NextRegistration()
	checkErrors(t, ErrNotSupported, err)
}

func TestNext(t *testing.T) {
	adapter, err := NewAdapter(33002, log.TestLogger{Tag: "Adapter", T: t})
	if err != nil {
		panic(err)
	}
	server := genMockServer(33002)

	tests := []struct {
		Adapter   *Adapter
		Packet    semtech.Packet
		WantAck   semtech.Packet
		WantNext  core.Packet
		WantError error
	}{
		{ // Valid uplink PUSH_DATA
			Adapter:   adapter,
			Packet:    genPUSH_DATAWithRXPK([]byte{0x14, 0x42}),
			WantAck:   genPUSH_ACK([]byte{0x14, 0x42}),
			WantNext:  genCorePacket(genPUSH_DATAWithRXPK([]byte{0x14, 0x42})),
			WantError: nil,
		},
		{ // Invalid uplink packet
			Adapter:   adapter,
			Packet:    genPUSH_ACK([]byte{0x22, 0x35}),
			WantAck:   semtech.Packet{},
			WantNext:  core.Packet{},
			WantError: nil,
		},
		{ // Uplink PUSH_DATA with no RXPK
			Adapter:   adapter,
			Packet:    genPUSH_DATANoRXPK([]byte{0x22, 0x35}),
			WantAck:   genPUSH_ACK([]byte{0x22, 0x35}),
			WantNext:  core.Packet{},
			WantError: nil,
		},
		{ // Uplink PULL_DATA
			Adapter:   adapter,
			Packet:    genPULL_DATA([]byte{0x62, 0xfa}),
			WantAck:   genPULL_ACK([]byte{0x62, 0xfa}),
			WantNext:  core.Packet{},
			WantError: nil,
		},
		{ // Uplink PUSH_DATA with no encoded payload
			Adapter:   adapter,
			Packet:    genPUSH_DATANoPayload([]byte{0x22, 0x35}),
			WantAck:   semtech.Packet{},
			WantNext:  core.Packet{},
			WantError: ErrInvalidPacket,
		},
	}

	for _, test := range tests {
		// Describe
		Desc(t, "Sending packet through adapter: %v", test.Packet)

		// Operate
		ack := server.send(test.Packet)
		packet, err := getNextPacket(adapter)

		// Check
		checkErrors(t, test.WantError, err)
		checkCorePackets(t, test.WantNext, packet)
		checkResponses(t, test.WantAck, ack)
	}
}

// ----- build utilities
type mockServer struct {
	conn *net.UDPConn
}

func genMockServer(port uint) mockServer {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	return mockServer{conn: conn}
}

func (s mockServer) send(p semtech.Packet) semtech.Packet {
	raw, err := semtech.Marshal(p)
	if err != nil {
		panic(err)
	}
	response := make(chan semtech.Packet)
	go func() {
		buf := make([]byte, 256)
		n, _, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		packet, err := semtech.Unmarshal(buf[:n])
		if err != nil {
			panic(err)
		}
		response <- *packet
	}()
	s.conn.Write(raw)
	select {
	case packet := <-response:
		return packet
	case <-time.After(100 * time.Millisecond):
		return semtech.Packet{}
	}
}

func genCorePacket(p semtech.Packet) core.Packet {
	if p.Payload == nil || len(p.Payload.RXPK) != 1 {
		panic("Expected a payload with one rxpk")
	}
	packet, err := components.ConvertRXPK(p.Payload.RXPK[0])
	if err != nil {
		panic(err)
	}
	return packet
}

func genPUSH_DATANoRXPK(token []byte) semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		GatewayId:  []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		Token:      token,
		Identifier: semtech.PUSH_DATA,
	}
}

func genPUSH_DATANoPayload(token []byte) semtech.Packet {
	packet := genPUSH_DATAWithRXPK(token)
	packet.Payload.RXPK[0].Data = nil
	return packet
}

func genPUSH_DATAWithRXPK(token []byte) semtech.Packet {
	packet := genPUSH_DATANoRXPK(token)
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

func genPULL_ACK(token []byte) semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		Token:      token,
		Identifier: semtech.PULL_ACK,
	}
}

func genPUSH_ACK(token []byte) semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		Token:      token,
		Identifier: semtech.PUSH_ACK,
	}
}

func genPULL_DATA(token []byte) semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		Token:      token,
		GatewayId:  []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		Identifier: semtech.PULL_DATA,
	}
}

func genRXPKData() string {
	// 1. Generate a PHYPayload
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	macPayload := lorawan.NewMACPayload(true)
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
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte("My Data")}}

	if err := macPayload.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	payload := lorawan.NewPHYPayload(true)
	payload.MHDR = lorawan.MHDR{
		MType: lorawan.ConfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	payload.MACPayload = macPayload

	if err := payload.SetMIC(nwkSKey); err != nil {
		panic(err)
	}

	// 2. Generate a JSON payload received by the server
	raw, err := payload.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

// ----- operate utilities
func getNextPacket(a *Adapter) (core.Packet, error) {
	next := make(chan struct {
		err    error
		packet core.Packet
	})
	go func() {
		packet, _, err := a.Next()
		next <- struct {
			err    error
			packet core.Packet
		}{err: err, packet: packet}
	}()

	select {
	case res := <-next:
		return res.packet, res.err
	case <-time.After(100 * time.Millisecond):
		return core.Packet{}, nil
	}
}

// ----- check utilities
func checkErrors(t *testing.T, want error, got error) {
	if want == got {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be %v but got %v", want, got)
}

func checkCorePackets(t *testing.T, want core.Packet, got core.Packet) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check core packets")
		return
	}
	Ko(t, "Received core packet does not match expecatations.\nWant: %v\nGot:  %v", want, got)
}

func checkResponses(t *testing.T, want semtech.Packet, got semtech.Packet) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check responses")
		return
	}
	Ko(t, "Received response does not match expecatations.\nWant: %v\nGot:  %v", want, got)
}

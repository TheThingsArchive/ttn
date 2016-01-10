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
	"net"
	"testing"
	"time"
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
			buf := make([]byte, 256)
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				panic(err)
			}
			packet, err := semtech.Unmarshal(buf[:n])
			if err != nil {
				panic(err)
			}
			response <- *packet
		}
	}()

	return mockServer{conn: conn, response: response}
}

// Send a packet through the udp mock server toward the adapter
func (s mockServer) send(p semtech.Packet) semtech.Packet {
	raw, err := semtech.Marshal(p)
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
func genAdapter(t *testing.T, port uint) (*Adapter, chan interface{}) {
	adapter, err := NewAdapter(port, log.TestLogger{Tag: "Adapter", T: t})
	if err != nil {
		panic(err)
	}
	next := make(chan interface{})
	go func() {
		for {
			packet, _, err := adapter.Next()
			next <- struct {
				err    error
				packet core.Packet
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

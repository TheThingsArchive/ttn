// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"encoding/base64"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func listenPackets(conn *net.UDPConn) chan semtech.Packet {
	chresp := make(chan semtech.Packet, 2)
	go func() {
		for {
			buf := make([]byte, 5000)
			n, err := conn.Read(buf)
			if err == nil {
				pkt := new(semtech.Packet)
				if err := pkt.UnmarshalBinary(buf[:n]); err == nil {
					chresp <- *pkt
				}
			}
		}
	}()
	return chresp
}

func TestUDPAdapter(t *testing.T) {
	port := 10000
	newAddr := func() string {
		port++
		return fmt.Sprintf("0.0.0.0:%d", port)
	}

	{
		Desc(t, "Send a valid packet through udp, no downlink")

		// Build
		payload := lorawan.NewPHYPayload(true)
		payload.MHDR.MType = lorawan.UnconfirmedDataUp
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		macpayload := lorawan.NewMACPayload(true)
		macpayload.FPort = new(uint8)
		*macpayload.FPort = 1

		macpayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{1, 2, 3}}}
		payload.MACPayload = macpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)

		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
					},
				},
			},
		}
		data, err = packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantDataRouterReq = &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(payload.MHDR.MType),
					Major: uint32(payload.MHDR.Major),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: macpayload.FHDR.DevAddr[:],
						FCnt:    macpayload.FHDR.FCnt,
						FCtrl: &core.LoRaWANFCtrl{
							ADR:       macpayload.FHDR.FCtrl.ADR,
							ADRAckReq: macpayload.FHDR.FCtrl.ADRACKReq,
							Ack:       macpayload.FHDR.FCtrl.ACK,
							FPending:  macpayload.FHDR.FCtrl.FPending,
							FOptsLen:  nil,
						},
						FOpts: nil,
					},
					FPort:      uint32(*macpayload.FPort),
					FRMPayload: macpayload.FRMPayload[0].(*lorawan.DataPayload).Bytes,
				},
				MIC: payload.MIC[:],
			},
			Metadata:  new(core.Metadata),
			GatewayID: packet.GatewayId,
		}
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a valid packet through udp, with valid downlink")

		// Build
		payload := lorawan.NewPHYPayload(true)
		payload.MHDR.MType = lorawan.UnconfirmedDataUp
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		macpayload := lorawan.NewMACPayload(true)
		macpayload.FPort = new(uint8)
		*macpayload.FPort = 1

		macpayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{1, 2, 3}}}
		payload.MACPayload = macpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)

		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
					},
				},
			},
		}
		data, err = packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()
		router.OutHandleData.Res = &core.DataRouterRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataDown),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    2,
						FCtrl:   new(core.LoRaWANFCtrl),
						FOpts:   nil,
					},
					FPort:      2,
					FRMPayload: []byte{1, 2, 3, 4},
				},
				MIC: []byte{1, 2, 3, 4},
			},
			Metadata: new(core.Metadata),
		}

		payloadDown := lorawan.NewPHYPayload(false)
		payloadDown.MHDR.MType = lorawan.MType(router.OutHandleData.Res.Payload.MHDR.MType)
		payloadDown.MHDR.Major = lorawan.Major(router.OutHandleData.Res.Payload.MHDR.Major)
		copy(payloadDown.MIC[:], router.OutHandleData.Res.Payload.MIC)
		macpayloadDown := lorawan.NewMACPayload(false)
		macpayloadDown.FPort = new(uint8)
		*macpayloadDown.FPort = uint8(router.OutHandleData.Res.Payload.MACPayload.FPort)
		macpayloadDown.FHDR.FCnt = router.OutHandleData.Res.Payload.MACPayload.FHDR.FCnt
		copy(macpayloadDown.FHDR.DevAddr[:], router.OutHandleData.Res.Payload.MACPayload.FHDR.DevAddr)
		macpayloadDown.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{
			Bytes: router.OutHandleData.Res.Payload.MACPayload.FRMPayload,
		}}
		payloadDown.MACPayload = macpayloadDown
		dataDown, err := payloadDown.MarshalBinary()
		FatalUnless(t, err)

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantDataRouterReq = &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(payload.MHDR.MType),
					Major: uint32(payload.MHDR.Major),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: macpayload.FHDR.DevAddr[:],
						FCnt:    macpayload.FHDR.FCnt,
						FCtrl: &core.LoRaWANFCtrl{
							ADR:       macpayload.FHDR.FCtrl.ADR,
							ADRAckReq: macpayload.FHDR.FCtrl.ADRACKReq,
							Ack:       macpayload.FHDR.FCtrl.ACK,
							FPending:  macpayload.FHDR.FCtrl.FPending,
							FOptsLen:  nil,
						},
						FOpts: nil,
					},
					FPort:      uint32(*macpayload.FPort),
					FRMPayload: macpayload.FRMPayload[0].(*lorawan.DataPayload).Bytes,
				},
				MIC: payload.MIC[:],
			},
			Metadata:  new(core.Metadata),
			GatewayID: packet.GatewayId,
		}
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{
				Version:    semtech.VERSION,
				Identifier: semtech.PULL_RESP,
				Payload: &semtech.Payload{
					TXPK: &semtech.TXPK{
						Data: pointer.String(base64.RawStdEncoding.EncodeToString(dataDown)),
						Rfch: pointer.Uint32(0),
					},
				},
			},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a valid packet through udp, with invalid downlink")

		// Build
		payload := lorawan.NewPHYPayload(true)
		payload.MHDR.MType = lorawan.UnconfirmedDataUp
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		macpayload := lorawan.NewMACPayload(true)
		macpayload.FPort = new(uint8)
		*macpayload.FPort = 1

		macpayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{1, 2, 3}}}
		payload.MACPayload = macpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)

		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
					},
				},
			},
		}
		data, err = packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()
		router.OutHandleData.Res = &core.DataRouterRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataDown),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: nil,
				MIC:        []byte{1, 2, 3, 4},
			},
			Metadata: new(core.Metadata),
		}

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantDataRouterReq = &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(payload.MHDR.MType),
					Major: uint32(payload.MHDR.Major),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: macpayload.FHDR.DevAddr[:],
						FCnt:    macpayload.FHDR.FCnt,
						FCtrl: &core.LoRaWANFCtrl{
							ADR:       macpayload.FHDR.FCtrl.ADR,
							ADRAckReq: macpayload.FHDR.FCtrl.ADRACKReq,
							Ack:       macpayload.FHDR.FCtrl.ACK,
							FPending:  macpayload.FHDR.FCtrl.FPending,
							FOptsLen:  nil,
						},
						FOpts: nil,
					},
					FPort:      uint32(*macpayload.FPort),
					FRMPayload: macpayload.FRMPayload[0].(*lorawan.DataPayload).Bytes,
				},
				MIC: payload.MIC[:],
			},
			Metadata:  new(core.Metadata),
			GatewayID: packet.GatewayId,
		}
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a packet through udp, no payload")

		// Build
		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload:    nil,
		}
		data, err := packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantDataRouterReq *core.DataRouterReq
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a packet through udp, empty payload")

		// Build
		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload:    new(semtech.Payload),
		}
		data, err := packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantDataRouterReq *core.DataRouterReq
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a valid packet through udp, no downlink, router fails")

		// Build
		payload := lorawan.NewPHYPayload(true)
		payload.MHDR.MType = lorawan.UnconfirmedDataUp
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		macpayload := lorawan.NewMACPayload(true)
		macpayload.FPort = new(uint8)
		*macpayload.FPort = 1

		macpayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{1, 2, 3}}}
		payload.MACPayload = macpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)

		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
					},
				},
			},
		}
		data, err = packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()
		router.Failures["HandleData"] = fmt.Errorf("Mock Error")

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantDataRouterReq = &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(payload.MHDR.MType),
					Major: uint32(payload.MHDR.Major),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: macpayload.FHDR.DevAddr[:],
						FCnt:    macpayload.FHDR.FCnt,
						FCtrl: &core.LoRaWANFCtrl{
							ADR:       macpayload.FHDR.FCtrl.ADR,
							ADRAckReq: macpayload.FHDR.FCtrl.ADRACKReq,
							Ack:       macpayload.FHDR.FCtrl.ACK,
							FPending:  macpayload.FHDR.FCtrl.FPending,
							FOptsLen:  nil,
						},
						FOpts: nil,
					},
					FPort:      uint32(*macpayload.FPort),
					FRMPayload: macpayload.FRMPayload[0].(*lorawan.DataPayload).Bytes,
				},
				MIC: payload.MIC[:],
			},
			Metadata:  new(core.Metadata),
			GatewayID: packet.GatewayId,
		}
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a valid packet through udp with stats, no downlink")

		// Build
		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload: &semtech.Payload{
				Stat: &semtech.Stat{
					Alti: pointer.Int32(14),
					Long: pointer.Float32(44.7654),
					Lati: pointer.Float32(23.56),
				},
			},
		}
		data, err := packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()
		router.Failures["HandleData"] = fmt.Errorf("Mock Error")

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantDataRouterReq *core.DataRouterReq
		var wantStats = &core.StatsReq{
			GatewayID: packet.GatewayId,
			Metadata: &core.StatsMetadata{
				Altitude:  *packet.Payload.Stat.Alti,
				Longitude: *packet.Payload.Stat.Long,
				Latitude:  *packet.Payload.Stat.Lati,
			},
		}
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a packet through udp, no data in rxpk")

		// Build
		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Codr: pointer.String("4/6"),
						Datr: pointer.String("SF8BW125"),
					},
				},
			},
		}
		data, err := packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantDataRouterReq *core.DataRouterReq
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// -------------------

	{
		Desc(t, "Send a PULL_DATA through udp")

		// Build
		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PULL_DATA,
		}
		data, err := packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantDataRouterReq *core.DataRouterReq
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PULL_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// -------------------

	{
		Desc(t, "Invalid options NetAddr")

		// Build
		router := mocks.NewRouterServer()

		// Expectations
		var wantErrStart = ErrOperational
		var wantDataRouterReq *core.DataRouterReq
		var wantStats *core.StatsReq

		// Operate
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: "patate", MaxReconnectionDelay: 25 * time.Millisecond},
		)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
	}

	// -------------------

	{
		Desc(t, "Send an invalid semtech as uplink")

		// Build
		packet := semtech.Packet{
			Version:    semtech.VERSION,
			Token:      []byte{1, 2},
			Identifier: semtech.PULL_ACK,
		}
		data, err := packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantDataRouterReq *core.DataRouterReq
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantDataRouterReq, router.InHandleData.Req, "Data Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a valid join through udp, with valid join-accept")

		// Build
		payload := lorawan.NewPHYPayload(true)
		payload.MHDR.MType = lorawan.JoinRequest
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		joinpayload := &lorawan.JoinRequestPayload{
			AppEUI:   [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:   [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
			DevNonce: [2]byte{14, 42},
		}
		payload.MACPayload = joinpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)

		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
					},
				},
			},
		}
		data, err = packet.MarshalBinary()
		FatalUnless(t, err)

		dataDown := []byte("accept")
		router := mocks.NewRouterServer()
		router.OutHandleJoin.Res = &core.JoinRouterRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: dataDown,
			},
			Metadata: new(core.Metadata),
		}

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantJoinRouterReq = &core.JoinRouterReq{
			GatewayID: packet.GatewayId,
			DevEUI:    joinpayload.DevEUI[:],
			AppEUI:    joinpayload.AppEUI[:],
			DevNonce:  joinpayload.DevNonce[:],
			MIC:       payload.MIC[:],
			Metadata:  new(core.Metadata),
		}
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{
				Version:    semtech.VERSION,
				Identifier: semtech.PULL_RESP,
				Payload: &semtech.Payload{
					TXPK: &semtech.TXPK{
						Data: pointer.String(base64.RawStdEncoding.EncodeToString(dataDown)),
						Rfch: pointer.Uint32(0),
					},
				},
			},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantJoinRouterReq, router.InHandleJoin.Req, "Join Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a valid join through udp, no payload in accept")

		// Build
		payload := lorawan.NewPHYPayload(true)
		payload.MHDR.MType = lorawan.JoinRequest
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		joinpayload := &lorawan.JoinRequestPayload{
			AppEUI:   [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:   [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
			DevNonce: [2]byte{14, 42},
		}
		payload.MACPayload = joinpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)

		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
					},
				},
			},
		}
		data, err = packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()
		router.OutHandleJoin.Res = &core.JoinRouterRes{
			Payload:  nil,
			Metadata: new(core.Metadata),
		}

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantJoinRouterReq = &core.JoinRouterReq{
			GatewayID: packet.GatewayId,
			DevEUI:    joinpayload.DevEUI[:],
			AppEUI:    joinpayload.AppEUI[:],
			DevNonce:  joinpayload.DevNonce[:],
			MIC:       payload.MIC[:],
			Metadata:  new(core.Metadata),
		}
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantJoinRouterReq, router.InHandleJoin.Req, "Join Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a valid join through udp, no metadata in response")

		// Build
		payload := lorawan.NewPHYPayload(true)
		payload.MHDR.MType = lorawan.JoinRequest
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		joinpayload := &lorawan.JoinRequestPayload{
			AppEUI:   [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:   [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
			DevNonce: [2]byte{14, 42},
		}
		payload.MACPayload = joinpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)

		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
					},
				},
			},
		}
		data, err = packet.MarshalBinary()
		FatalUnless(t, err)

		dataDown := []byte("accept")
		router := mocks.NewRouterServer()
		router.OutHandleJoin.Res = &core.JoinRouterRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: dataDown,
			},
			Metadata: nil,
		}

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantJoinRouterReq = &core.JoinRouterReq{
			GatewayID: packet.GatewayId,
			DevEUI:    joinpayload.DevEUI[:],
			AppEUI:    joinpayload.AppEUI[:],
			DevNonce:  joinpayload.DevNonce[:],
			MIC:       payload.MIC[:],
			Metadata:  new(core.Metadata),
		}
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantJoinRouterReq, router.InHandleJoin.Req, "Join Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}

	// --------------------

	{
		Desc(t, "Send a valid join through udp, router fails to handle")

		// Build
		payload := lorawan.NewPHYPayload(true)
		payload.MHDR.MType = lorawan.JoinRequest
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		joinpayload := &lorawan.JoinRequestPayload{
			AppEUI:   [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:   [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
			DevNonce: [2]byte{14, 42},
		}
		payload.MACPayload = joinpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)

		packet := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PUSH_DATA,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
					},
				},
			},
		}
		data, err = packet.MarshalBinary()
		FatalUnless(t, err)

		router := mocks.NewRouterServer()
		router.Failures["HandleJoin"] = fmt.Errorf("Mock Error")

		netAddr := newAddr()
		addr, err := net.ResolveUDPAddr("udp", netAddr)
		FatalUnless(t, err)
		conn, err := net.DialUDP("udp", nil, addr)
		FatalUnless(t, err)

		// Expectations
		var wantErrStart *string
		var wantJoinRouterReq = &core.JoinRouterReq{
			GatewayID: packet.GatewayId,
			DevEUI:    joinpayload.DevEUI[:],
			AppEUI:    joinpayload.AppEUI[:],
			DevNonce:  joinpayload.DevNonce[:],
			MIC:       payload.MIC[:],
			Metadata:  new(core.Metadata),
		}
		var wantStats *core.StatsReq
		var wantSemtechResp = []semtech.Packet{
			{
				Version:    semtech.VERSION,
				Token:      packet.Token,
				Identifier: semtech.PUSH_ACK,
			},
			{},
		}

		// Operate
		chpkt := listenPackets(conn)
		errStart := Start(
			Components{Router: router, Ctx: GetLogger(t, "Adapter")},
			Options{NetAddr: netAddr, MaxReconnectionDelay: 25 * time.Millisecond},
		)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		_, err = conn.Write(data)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 50)
		close(chpkt)

		// Check
		CheckErrors(t, wantErrStart, errStart)
		Check(t, wantJoinRouterReq, router.InHandleJoin.Req, "Join Router Requests")
		Check(t, wantStats, router.InHandleStats.Req, "Data Router Stats")
		Check(t, wantSemtechResp[0], <-chpkt, "Acknowledgements")
		Check(t, wantSemtechResp[1], <-chpkt, "Downlinks")
	}
}

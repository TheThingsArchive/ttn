// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"reflect"
	"testing"
	"time"

	core "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/refactor/adapters/udp"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestSend(t *testing.T) {
	Desc(t, "Send is not supported")
	adapter, _ := genAdapter(t, 33000)
	_, err := adapter.Send(core.RPacket{})
	checkErrors(t, pointer.String(string(errors.Implementation)), err)
}

func TestNextRegistration(t *testing.T) {
	Desc(t, "Next registration is not supported")
	adapter, _ := genAdapter(t, 33001)
	_, _, err := adapter.NextRegistration()
	checkErrors(t, pointer.String(string(errors.Implementation)), err)
}

func TestNext(t *testing.T) {
	adapter, next := genAdapter(t, 33002)
	server := genMockServer(33002)

	tests := []struct {
		Adapter   *udp.Adapter
		Packet    semtech.Packet
		WantAck   semtech.Packet
		WantNext  core.Packet
		WantError *string
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
			WantNext:  core.RPacket{},
			WantError: nil,
		},
		{ // Uplink PUSH_DATA with no RXPK
			Adapter:   adapter,
			Packet:    genPUSH_DATANoRXPK([]byte{0x22, 0x35}),
			WantAck:   genPUSH_ACK([]byte{0x22, 0x35}),
			WantNext:  core.RPacket{},
			WantError: nil,
		},
		{ // Uplink PULL_DATA
			Adapter:   adapter,
			Packet:    genPULL_DATA([]byte{0x62, 0xfa}),
			WantAck:   genPULL_ACK([]byte{0x62, 0xfa}),
			WantNext:  core.RPacket{},
			WantError: nil,
		},
		{ // Uplink PUSH_DATA with no encoded payload
			Adapter:   adapter,
			Packet:    genPUSH_DATANoPayload([]byte{0x22, 0x35}),
			WantAck:   genPUSH_ACK([]byte{0x22, 0x35}),
			WantNext:  core.RPacket{},
			WantError: nil,
		},
	}

	for _, test := range tests {
		// Describe
		Desc(t, "Sending packet through adapter: %v", test.Packet)
		<-time.After(time.Millisecond * 100)

		// Operate
		ack := server.send(test.Packet)
		packet, err := getNextPacket(next)

		// Check
		checkErrors(t, test.WantError, err)
		checkCorePackets(t, test.WantNext, packet)
		checkResponses(t, test.WantAck, ack)
	}
}

// ----- OPERATE utilities
func getNextPacket(next chan interface{}) (core.Packet, error) {
	select {
	case i := <-next:
		res := i.(struct {
			err    error
			packet []byte
		})
		var packet core.RPacket
		err := packet.UnmarshalBinary(res.packet)
		if err != nil {
			panic(err)
		}
		return packet, res.err
	case <-time.After(100 * time.Millisecond):
		return core.RPacket{}, nil
	}
}

// ----- CHECK utilities
func checkErrors(t *testing.T, want *string, got error) {
	if got == nil {
		if want == nil {
			Ok(t, "Check errors")
			return
		}
		Ko(t, "Expected error to be {%s} but got nothing", *want)
		return
	}

	if want == nil {
		Ko(t, "Expected no error but got {%v}", got)
		return
	}

	if got.(errors.Failure).Nature == errors.Nature(*want) {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be {%s} but got {%v}", *want, got)
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
	Ko(t, "Received response does not match expecatations.\nWant: %v\nGot:  %v", want.String(), got.String())
}

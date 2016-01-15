// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"reflect"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/semtech"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestNewAdapter(t *testing.T) {
	Ok(t, "pending")
}

func TestSend(t *testing.T) {
	Desc(t, "Send is not supported")
	adapter, _ := genAdapter(t, 33000)
	_, err := adapter.Send(core.Packet{})
	checkErrors(t, ErrNotSupported, err)
}

func TestNextRegistration(t *testing.T) {
	Desc(t, "Next registration is not supported")
	adapter, _ := genAdapter(t, 33001)
	_, _, err := adapter.NextRegistration()
	checkErrors(t, ErrNotSupported, err)
}

func TestNext(t *testing.T) {
	adapter, next := genAdapter(t, 33002)
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
			WantAck:   genPUSH_ACK([]byte{0x22, 0x35}),
			WantNext:  core.Packet{},
			WantError: ErrInvalidPacket,
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

// ----- operate utilities
func getNextPacket(next chan interface{}) (core.Packet, error) {
	select {
	case i := <-next:
		res := i.(struct {
			err    error
			packet core.Packet
		})
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

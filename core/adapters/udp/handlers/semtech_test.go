// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"reflect"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/udp"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestSend(t *testing.T) {
	Desc(t, "Send is not supported")
	adapter, _ := genAdapter(t, 33000)
	_, err := adapter.Send(nil)
	CheckErrors(t, pointer.String(string(errors.Implementation)), err)
}

func TestNextRegistration(t *testing.T) {
	Desc(t, "Next registration is not supported")
	adapter, _ := genAdapter(t, 33001)
	_, _, err := adapter.NextRegistration()
	CheckErrors(t, pointer.String(string(errors.Implementation)), err)
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
			Packet:    genPUSHDATAWithRXPK([]byte{0x14, 0x42}),
			WantAck:   genPUSHACK([]byte{0x14, 0x42}),
			WantNext:  genCorePacket(genPUSHDATAWithRXPK([]byte{0x14, 0x42})),
			WantError: nil,
		},
		{ // Invalid uplink packet
			Adapter:   adapter,
			Packet:    genPUSHACK([]byte{0x22, 0x35}),
			WantAck:   semtech.Packet{},
			WantNext:  nil,
			WantError: nil,
		},
		{ // Uplink PUSH_DATA with no RXPK
			Adapter:   adapter,
			Packet:    genPUSHDATANoRXPK([]byte{0x22, 0x35}),
			WantAck:   genPUSHACK([]byte{0x22, 0x35}),
			WantNext:  nil,
			WantError: nil,
		},
		{ // Uplink PULL_DATA
			Adapter:   adapter,
			Packet:    genPULLDATA([]byte{0x62, 0xfa}),
			WantAck:   genPULLACK([]byte{0x62, 0xfa}),
			WantNext:  nil,
			WantError: nil,
		},
		{ // Uplink PUSH_DATA with no encoded payload
			Adapter:   adapter,
			Packet:    genPUSHDATANoPayload([]byte{0x22, 0x35}),
			WantAck:   genPUSHACK([]byte{0x22, 0x35}),
			WantNext:  nil,
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
		CheckErrors(t, test.WantError, err)
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
		itf, err := core.UnmarshalPacket(res.packet)
		if err != nil {
			panic(err)
		}
		pkt, ok := itf.(core.RPacket)
		if !ok {
			return itf.(core.Packet), res.err
		}
		return pkt, res.err
	case <-time.After(100 * time.Millisecond):
		return nil, nil
	}
}

// ----- CHECK utilities
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

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gtw_rtr_udp

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/testing/mock_components"
	"github.com/thethingsnetwork/core/utils/log"
	. "github.com/thethingsnetwork/core/utils/testing"
	"net"
	"reflect"
	"testing"
	"time"
)

// ----- The adapter should be able to create a udp connection given a valid udp port
func TestListenOptions(t *testing.T) {
	tests := []listenOptionsTest{
		{uint(3000), nil},
		{uint(3000), core.ErrBadOptions}, // Already used now
		{int(14), core.ErrBadOptions},
		{"somethingElse", core.ErrBadOptions},
	}

	for _, test := range tests {
		test.run(t)
	}
}

type listenOptionsTest struct {
	options interface{}
	want    error
}

func (test listenOptionsTest) run(t *testing.T) {
	Desc(t, "Run Listen(router, %T %v)", test.options, test.options)
	adapter, router := generateAdapterAndRouter(t)
	got := adapter.Listen(router, test.options)
	test.check(t, got)
}

func (test listenOptionsTest) check(t *testing.T, got error) {
	// 1. Check if errors match
	if got != test.want {
		t.Errorf("expected {%v} to be {%v}\n", got, test.want)
		Ko(t)
		return
	}
	Ok(t)
}

// ----- The adapter should catch from the connection and forward valid semtech.Packet to the router
func TestPacketProcessing(t *testing.T) {
	tests := []packetProcessingTest{
		{generatePUSH_DATA(), 1, 3001},
		{[]byte{0x14, 0xff}, 0, 3003},
	}

	for _, test := range tests {
		test.run(t)
	}
}

type packetProcessingTest struct {
	in   interface{} // Could be raw []byte or plain semtech.Packet
	want uint        // 0 or 1 depending whether or not we expect a packet to has been transmitted
	port uint        // Probably temporary, just because goroutine and connection are still living between tests
}

func (test packetProcessingTest) run(t *testing.T) {
	Desc(t, "Simulate incoming datagram: %+v", test.in)
	adapter, router := generateAdapterAndRouter(t)
	conn, gateway := listen(adapter, router, test.port)
	send(conn, test.in)
	test.check(t, router, gateway) // Check whether or not packet has been forwarded to core router
}

func (test packetProcessingTest) check(t *testing.T, router core.Router, gateway core.GatewayAddress) {
	<-time.After(time.Millisecond * 50)
	mockRouter := router.(*mock_components.Router)

	// 1. Check if we expect a packet
	packets := mockRouter.Packets[gateway]
	if nb := len(packets); uint(nb) != test.want {
		t.Errorf("Received %d packets whereas expected %d", nb, test.want)
		Ko(t)
		return
	}

	// 2. If a packet was expected, check that it has been forwarded to the router
	if test.want > 0 {
		if !reflect.DeepEqual(packets[0], test.in) {
			t.Errorf("Expected %+v to match %+v", packets[0], test.in)
			Ko(t)
			return
		}
	}

	Ok(t)
}

// ----- Build Utilities
func generateAdapterAndRouter(t *testing.T) (Adapter, core.Router) {
	return Adapter{
		Logger: log.TestLogger{
			Tag: "Adapter",
			T:   t,
		},
	}, mock_components.NewRouter()
}

func generatePUSH_DATA() semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		GatewayId:  []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		Token:      []byte{0x14, 0x42},
		Identifier: semtech.PUSH_DATA,
	}
}

// ----- Operate Utilities
func listen(adapter Adapter, router core.Router, port uint) (*net.UDPConn, core.GatewayAddress) {
	var err error

	// 1. Start the adapter watching procedure
	if err = adapter.Listen(router, port); err != nil {
		panic(err)
	}

	// 2. Create a UDP connection on the same port the adapter is listening
	var addr *net.UDPAddr
	var conn *net.UDPConn
	if addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port)); err != nil {
		panic(err)
	}
	if conn, err = net.DialUDP("udp", nil, addr); err != nil {
		panic(err)
	}

	// 3. Return the UDP connection and the corresponding simulated gateway address
	return conn, core.GatewayAddress(conn.LocalAddr().String())
}

func send(conn *net.UDPConn, data interface{}) {
	// 1. Send the packet or the raw sequence of bytes passed as argument
	var raw []byte
	var err error
	switch data.(type) {
	case []byte:
		raw = data.([]byte)
	case semtech.Packet:
		if raw, err = semtech.Marshal(data.(semtech.Packet)); err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("Unexpected data type to be send : %T", data))
	}
	if _, err = conn.Write(raw); err != nil {
		panic(err)
	}
}

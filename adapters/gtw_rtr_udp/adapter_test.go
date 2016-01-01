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
		{uint(3000), core.ErrBadGatewayAddress}, // Already used now
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
		{[]byte{0x14, 0xff}, 0, 3002},
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
	conn, gateway := createConnection(&adapter, router, test.port)
	defer conn.Close()
	sendDatagram(conn, test.in)
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

// ----- The adapter should send packet via back to an existing address through an opened connection
func TestSendAck(t *testing.T) {
	// 1. Initialize test data
	adapter, router := generateAdapterAndRouter(t)
	adapter2, router2 := generateAdapterAndRouter(t)
	conn, gateway := createConnection(&adapter, router, 3003)
	defer conn.Close()

	tests := []sendAckTest{
		{adapter, router, conn, gateway, generatePUSH_ACK(), nil},
		{adapter, router, conn, core.GatewayAddress("patate"), generatePUSH_ACK(), core.ErrBadGatewayAddress},
		{adapter, router, conn, gateway, semtech.Packet{}, core.ErrInvalidPacket},
		{adapter2, router2, nil, gateway, generatePUSH_ACK(), core.ErrMissingConnection},
	}

	// 2. Run tests
	for _, test := range tests {
		test.run(t)
	}
}

type sendAckTest struct {
	adapter Adapter
	router  core.Router
	conn    *net.UDPConn
	gateway core.GatewayAddress
	packet  semtech.Packet
	want    error
}

func (test sendAckTest) run(t *testing.T) {
	Desc(t, "Send ack packet %v to %v via %v", test.packet, test.conn, test.gateway)
	// Starts a goroutine that will redirect udp message to a dedicated channel
	cmsg := listenFromConnection(test.conn)
	defer close(cmsg)
	got := test.adapter.Ack(test.router, test.packet, test.gateway)
	test.check(t, cmsg, got) // Check the error or the packet if no error
}

func (test sendAckTest) check(t *testing.T, cmsg chan semtech.Packet, got error) {
	// 1. Check if an error was expected
	if test.want != nil {
		if got != test.want {
			t.Errorf("Expected %+v error but got %+v", test.want, got)
			Ko(t)
			return
		}
		Ok(t)
		return
	}

	// 2. Ensure the ack packet has been sent correctly
	packet := <-cmsg
	if !reflect.DeepEqual(test.packet, packet) {
		t.Errorf("Expected %+v to equal %+v", test.packet, packet)
		Ko(t)
		return
	}
	Ok(t)
}

// ----- In case of issue, the connection can be re-established
func TestConnectionRecovering(t *testing.T) {
	adapter, router := generateAdapterAndRouter(t)
	if err := adapter.Listen(router, uint(3004)); err != nil {
		panic(err)
	}
	err := adapter.Listen(router, uint(3005))

	if err != nil {
		t.Errorf("No error was expected but got: %+v", err)
		Ko(t)
		return
	}

	// Now try to send a packet on a switched connection
	err = adapter.Ack(router, generatePUSH_ACK(), core.GatewayAddress("0.0.0.0:3005"))
	if err != nil {
		t.Errorf("No error was expected but got: %+v", err)
		Ko(t)
		return
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

func generatePUSH_ACK() semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		Token:      []byte{0x14, 0x42},
		Identifier: semtech.PUSH_ACK,
	}
}

// ----- Operate Utilities
func createConnection(adapter *Adapter, router core.Router, port uint) (*net.UDPConn, core.GatewayAddress) {
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

func sendDatagram(conn *net.UDPConn, data interface{}) {
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

func listenFromConnection(conn *net.UDPConn) (cmsg chan semtech.Packet) {
	cmsg = make(chan semtech.Packet)

	// We won't listen on a nil connection
	if conn == nil {
		return
	}

	// Otherwise, wait for a packet
	go func() {
		for {
			buf := make([]byte, 128)
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			packet, err := semtech.Unmarshal(buf[:n])
			if err == nil {
				cmsg <- *packet
			}
		}
	}()

	return
}

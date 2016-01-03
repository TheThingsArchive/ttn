// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/testing/mock_adapters/gtw_rtr_mock"
	"github.com/thethingsnetwork/core/testing/mock_adapters/rtr_brk_mock"
	. "github.com/thethingsnetwork/core/utils/testing"
	"testing"
)

// ----- A new router instance can be created an obtained from a constuctor
func TestNewRouter(t *testing.T) {
	tests := []newRouterTest{
		{genBrokers(), nil},
		{[]core.BrokerAddress{}, core.ErrBadOptions},
	}

	for _, test := range tests {
		test.run(t)
	}
}

type newRouterTest struct {
	in   []core.BrokerAddress
	want error
}

func (test newRouterTest) run(t *testing.T) {
	Desc(t, "Create new router with params: %v", test.in)
	router, err := NewRouter(test.in...)
	checkErrors(t, test.want, err, router)
}

// ----- A router can handle uplink packets
func TestHandleUplink(t *testing.T) {
	tests := []handleUplinkTest{
		{genPULL_DATA(), core.GatewayAddress("a1"), 1, 0, 0},
	}

	for _, test := range tests {
		test.run(t)
	}
}

type handleUplinkTest struct {
	packet        semtech.Packet
	gateway       core.GatewayAddress
	wantAck       int
	wantForward   int
	wantBroadcast int
}

func (test handleUplinkTest) run(t *testing.T) {
	// Describe
	Desc(t, "Handle uplink packet %v from gateway %v", test.packet, test.gateway)

	// Build
	router, upAdapter, downAdapter := genAdaptersAndRouter(t)

	// Operate
	router.HandleUplink(test.packet, test.gateway)

	// Check
	checkUplink(t, upAdapter, downAdapter, test.wantAck, test.wantForward, test.wantBroadcast)
}

// ----- Build Utilities
func genBrokers() []core.BrokerAddress {
	return []core.BrokerAddress{
		core.BrokerAddress("0.0.0.0:3000"),
		core.BrokerAddress("0.0.0.0:3001"),
	}
}

func genAdaptersAndRouter(t *testing.T) (core.Router, core.GatewayRouterAdapter, core.RouterBrokerAdapter) {
	brokers := genBrokers()
	router, err := NewRouter(brokers...)
	if err != nil {
		panic(err)
	}

	upAdapter := gtw_rtr_mock.New()
	downAdapter := rtr_brk_mock.New()

	upAdapter.Listen(router, nil)
	downAdapter.Listen(router, brokers)
	router.Connect(upAdapter, downAdapter)

	return router, upAdapter, downAdapter
}

func genPULL_DATA() semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		Identifier: semtech.PULL_DATA,
		Token:      []byte{0x14, 0xba},
		GatewayId:  []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
	}
}

// ----- Check Utilities
func checkErrors(t *testing.T, want error, got error, router core.Router) {
	if want != got {
		Ko(t, "Expected error {%v} but got {%v}", want, got)
		return
	}

	if want == nil && router == nil {
		Ko(t, "Expected no error but got a nil router")
		return
	}

	Ok(t)
}

func checkUplink(t *testing.T, upAdapter core.GatewayRouterAdapter, downAdapter core.RouterBrokerAdapter, wantAck int, wantForward int, wantBroadcast int) {
	mockUp := upAdapter.(*gtw_rtr_mock.Adapter)
	mockDown := downAdapter.(*rtr_brk_mock.Adapter)

	if len(mockDown.Broadcasts) != wantBroadcast {
		Ko(t, "Expected %d broadcast(s) but %d has/have been done", wantBroadcast, len(mockDown.Broadcasts))
		return
	}

	if len(mockDown.Forwards) != wantForward {
		Ko(t, "Expected %d forward(s) but %d has/have been done", wantForward, len(mockDown.Forwards))
	}

	if len(mockUp.Acks) != wantAck {
		Ko(t, "Expected %d ack(s) but got %d", wantAck, len(mockUp.Acks))
		return
	}

	Ok(t)
}

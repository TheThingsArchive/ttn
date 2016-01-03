// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/testing/mock_adapters/gtw_rtr_mock"
	"github.com/thethingsnetwork/core/testing/mock_adapters/rtr_brk_mock"
	"github.com/thethingsnetwork/core/utils/log"
	"github.com/thethingsnetwork/core/utils/pointer"
	. "github.com/thethingsnetwork/core/utils/testing"
	"testing"
	"time"
)

// ----- A new router instance can be created an obtained from a constuctor
func TestNewRouter(t *testing.T) {
	tests := []newRouterTest{
		{genBrokers(2), nil},
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
	// Build tests
	nbBrokers := 2
	brokers := genBrokers(nbBrokers)
	router, upAdapter, downAdapter := genAdaptersAndRouter(t, brokers)
	down := downAdapter.(*rtr_brk_mock.Adapter)
	pushData := genPUSH_DATA()
	down.Relations[*pushData.Payload.RXPK[0].DevAddr()] = brokers[1:] // Register the second broker as being in charge of device #1

	tests := []handleUplinkTest{
		{genPULL_DATA(), 1, 0, 0},
		{pushData, 1, 0, 3}, // PUSH_DATA generate a packet with 4 RXPK from 3 different devices
		{pushData, 1, 1, 2}, // Now device #1 should be handled by broker #2
	}

	for _, test := range tests {
		test.run(t, router, upAdapter, downAdapter)
	}
}

type handleUplinkTest struct {
	packet        semtech.Packet
	wantAck       int
	wantForward   int
	wantBroadcast int
}

func (test handleUplinkTest) run(t *testing.T, router core.Router, upAdapter core.GatewayRouterAdapter, downAdapter core.RouterBrokerAdapter) {
	// Describe
	Desc(t, "Handle uplink packet %v", test.packet)

	// Build
	mockDown := downAdapter.(*rtr_brk_mock.Adapter)
	mockDown.Forwards = make(map[semtech.DeviceAddress][]semtech.Payload)
	mockDown.Broadcasts = make(map[semtech.DeviceAddress][]semtech.Payload)

	// Operate
	router.HandleUplink(test.packet, core.GatewayAddress("Gateway"))
	<-time.After(time.Millisecond * 100)

	// Check
	checkUplink(t, upAdapter, downAdapter, test.wantAck, test.wantForward, test.wantBroadcast)
}

// ----- Build Utilities
func genBrokers(n int) []core.BrokerAddress {
	var brokers []core.BrokerAddress
	for i := 0; i < n; i += 1 {
		brokers = append(brokers, core.BrokerAddress(fmt.Sprintf("0.0.0.0:%d", 3000+i)))
	}
	return brokers
}

func genAdaptersAndRouter(t *testing.T, brokers []core.BrokerAddress) (core.Router, core.GatewayRouterAdapter, core.RouterBrokerAdapter) {
	router, err := NewRouter(brokers...)
	if err != nil {
		panic(err)
	}

	router.Logger = log.TestLogger{Tag: "Router", T: t}
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

// genPUSH_DATA generate a a PUSH_DATA packet with 4 RXPK in payload coming from 3 different
// devices.
func genPUSH_DATA() semtech.Packet {
	return semtech.Packet{
		Version:    semtech.VERSION,
		Identifier: semtech.PUSH_DATA,
		Token:      []byte{0x14, 0xba},
		GatewayId:  []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		Payload: &semtech.Payload{
			RXPK: []semtech.RXPK{
				semtech.RXPK{Data: pointer.String("/xRC/zcBAAABqqq7uw==")}, // Device #1
				semtech.RXPK{Data: pointer.String("/7zN3u8BAAABqqv3RA==")}, // Device #2
				semtech.RXPK{Data: pointer.String("/xRC/zcBAAABqmcSRQ==")}, // Device #1
				semtech.RXPK{Data: pointer.String("/wASo3YBAAAB+qpFeQ==")}, // Device #3
			},
		},
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

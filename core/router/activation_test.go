// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleActivation(t *testing.T) {
	a := New(t)

	gatewayEUI := types.GatewayEUI{0, 1, 2, 3, 4, 5, 6, 7}

	r := &router{
		Component: &core.Component{
			Ctx: GetLogger(t, "TestHandleActivation"),
		},
		gateways: map[types.GatewayEUI]*gateway.Gateway{
			gatewayEUI: newReferenceGateway(t, "EU_863_870"),
		},
		brokerDiscovery: &mockBrokerDiscovery{},
	}

	appEUI := types.AppEUI{0, 1, 2, 3, 4, 5, 6, 7}
	devEUI := types.DevEUI{0, 1, 2, 3, 4, 5, 6, 7}

	uplink := newReferenceUplink()
	activation := &pb.DeviceActivationRequest{
		Payload:          []byte{},
		ProtocolMetadata: uplink.ProtocolMetadata,
		GatewayMetadata:  uplink.GatewayMetadata,
		AppEui:           &appEUI,
		DevEui:           &devEUI,
	}
	gtwEUI := types.GatewayEUI{0, 1, 2, 3, 4, 5, 6, 7}

	res, err := r.HandleActivation(gtwEUI, activation)
	a.So(res, ShouldBeNil)
	a.So(err, ShouldNotBeNil)
	utilization := r.getGateway(gtwEUI).Utilization
	utilization.Tick()
	rx, _ := utilization.Get()
	a.So(rx, ShouldBeGreaterThan, 0)

	// TODO: Integration test that checks broker forward
}

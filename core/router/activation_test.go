// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	pb "github.com/TheThingsNetwork/api/router"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleActivation(t *testing.T) {
	a := New(t)

	gtwID := "eui-0102030405060708"

	r := &router{
		Component: &component.Component{
			Ctx: GetLogger(t, "TestHandleActivation"),
		},
		gateways: map[string]*gateway.Gateway{
			gtwID: newReferenceGateway(t, "EU_863_870"),
		},
	}
	r.InitStatus()

	appEUI := types.AppEUI{0, 1, 2, 3, 4, 5, 6, 7}
	devEUI := types.DevEUI{0, 1, 2, 3, 4, 5, 6, 7}

	uplink := newReferenceUplink()
	activation := &pb.DeviceActivationRequest{
		Payload:          []byte{},
		ProtocolMetadata: uplink.ProtocolMetadata,
		GatewayMetadata:  uplink.GatewayMetadata,
		AppEUI:           &appEUI,
		DevEUI:           &devEUI,
	}

	res, err := r.HandleActivation(gtwID, activation)
	a.So(res, ShouldBeNil)
	a.So(err, ShouldNotBeNil)

	utilization := r.getGateway(gtwID).Utilization
	utilization.Tick()
	rx, _ := utilization.Get()
	a.So(rx, ShouldBeGreaterThan, 0)

	// TODO: Integration test that checks broker forward
}

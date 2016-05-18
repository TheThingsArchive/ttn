package router

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestHandleActivation(t *testing.T) {
	a := New(t)

	r := &router{
		gateways:        map[types.GatewayEUI]*gateway.Gateway{},
		brokerDiscovery: &mockBrokerDiscovery{},
	}

	uplink := newReferenceUplink()
	activation := &pb.DeviceActivationRequest{
		Payload:          []byte{},
		ProtocolMetadata: uplink.ProtocolMetadata,
		GatewayMetadata:  uplink.GatewayMetadata,
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

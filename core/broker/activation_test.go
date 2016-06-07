package broker

import (
	"sync"
	"testing"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/broker/application"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleActivation(t *testing.T) {
	a := New(t)

	gtwEUI := types.GatewayEUI([8]byte{0, 1, 2, 3, 4, 5, 6, 7})
	devEUI := types.DevEUI([8]byte{0, 1, 2, 3, 4, 5, 6, 7})
	appEUI := types.AppEUI([8]byte{0, 1, 2, 3, 4, 5, 6, 7})

	b := &broker{
		Component: &core.Component{
			Ctx: GetLogger(t, "TestHandleActivation"),
		},
		activationDeduplicator: NewDeduplicator(10 * time.Millisecond),
		applications:           application.NewApplicationStore(),
		ns:                     &mockNetworkServer{},
	}

	// Non-existing Application
	res, err := b.HandleActivation(&pb.DeviceActivationRequest{
		Payload:          []byte{},
		DevEui:           &devEUI,
		AppEui:           &appEUI,
		GatewayMetadata:  &gateway.RxMetadata{Snr: 1.2, GatewayEui: &gtwEUI},
		ProtocolMetadata: &protocol.RxMetadata{},
	})
	a.So(err, ShouldNotBeNil)
	a.So(res, ShouldBeNil)

	// Non-existing Broker
	b.applications.Set(&application.Application{
		AppEUI:            appEUI,
		HandlerNetAddress: "localhost:1234",
	})
	res, err = b.HandleActivation(&pb.DeviceActivationRequest{
		Payload:          []byte{},
		DevEui:           &devEUI,
		AppEui:           &appEUI,
		GatewayMetadata:  &gateway.RxMetadata{Snr: 1.2, GatewayEui: &gtwEUI},
		ProtocolMetadata: &protocol.RxMetadata{},
	})
	a.So(err, ShouldNotBeNil)
	a.So(res, ShouldBeNil)

	// TODO: Integration test with Handler
}

func TestDeduplicateActivation(t *testing.T) {
	a := New(t)
	d := NewDeduplicator(10 * time.Millisecond).(*deduplicator)

	payload := []byte{0x01, 0x02, 0x03}
	protocolMetadata := &protocol.RxMetadata{}
	activation1 := &pb.DeviceActivationRequest{Payload: payload, GatewayMetadata: &gateway.RxMetadata{Snr: 1.2}, ProtocolMetadata: protocolMetadata}
	activation2 := &pb.DeviceActivationRequest{Payload: payload, GatewayMetadata: &gateway.RxMetadata{Snr: 3.4}, ProtocolMetadata: protocolMetadata}
	activation3 := &pb.DeviceActivationRequest{Payload: payload, GatewayMetadata: &gateway.RxMetadata{Snr: 5.6}, ProtocolMetadata: protocolMetadata}
	activation4 := &pb.DeviceActivationRequest{Payload: payload, GatewayMetadata: &gateway.RxMetadata{Snr: 7.8}, ProtocolMetadata: protocolMetadata}

	b := &broker{activationDeduplicator: d}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		res := b.deduplicateActivation(activation1)
		a.So(res, ShouldResemble, []*pb.DeviceActivationRequest{activation1, activation2, activation3})
		wg.Done()
	}()

	<-time.After(5 * time.Millisecond)

	a.So(b.deduplicateActivation(activation2), ShouldBeNil)
	a.So(b.deduplicateActivation(activation3), ShouldBeNil)

	wg.Wait()

	<-time.After(20 * time.Millisecond)

	a.So(b.deduplicateActivation(activation4), ShouldNotBeNil)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"sync"
	"testing"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/core"
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
		handlerDiscovery:       &mockHandlerDiscovery{},
		activationDeduplicator: NewDeduplicator(10 * time.Millisecond),
		ns: &mockNetworkServer{},
	}

	res, err := b.HandleActivation(&pb.DeviceActivationRequest{
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
	d := NewDeduplicator(20 * time.Millisecond).(*deduplicator)

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

	<-time.After(10 * time.Millisecond)

	a.So(b.deduplicateActivation(activation2), ShouldBeNil)
	a.So(b.deduplicateActivation(activation3), ShouldBeNil)

	<-time.After(50 * time.Millisecond)

	a.So(b.deduplicateActivation(activation4), ShouldNotBeNil)

	wg.Wait()
}

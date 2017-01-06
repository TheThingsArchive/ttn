// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"sync"
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/assertions"
)

func TestHandleActivation(t *testing.T) {
	a := New(t)

	gtwID := "eui-0102030405060708"
	devEUI := types.DevEUI([8]byte{0, 1, 2, 3, 4, 5, 6, 7})
	appEUI := types.AppEUI([8]byte{0, 1, 2, 3, 4, 5, 6, 7})

	b := getTestBroker(t)
	b.ns.EXPECT().PrepareActivation(gomock.Any(), gomock.Any()).Return(&pb_broker.DeduplicatedDeviceActivationRequest{
		Payload: []byte{},
		DevEui:  &devEUI,
		AppEui:  &appEUI,
		AppId:   "appid",
		DevId:   "devid",
		GatewayMetadata: []*gateway.RxMetadata{
			&gateway.RxMetadata{Snr: 1.2, GatewayId: gtwID},
		},
		ProtocolMetadata: &protocol.RxMetadata{},
	}, nil)
	b.discovery.EXPECT().GetAllHandlersForAppID("appid").Return([]*pb_discovery.Announcement{}, nil)

	res, err := b.HandleActivation(&pb_broker.DeviceActivationRequest{
		Payload:          []byte{},
		DevEui:           &devEUI,
		AppEui:           &appEUI,
		GatewayMetadata:  &gateway.RxMetadata{Snr: 1.2, GatewayId: gtwID},
		ProtocolMetadata: &protocol.RxMetadata{},
	})
	a.So(err, ShouldNotBeNil)
	a.So(res, ShouldBeNil)

	b.ctrl.Finish()

	// TODO: Integration test with Handler
}

func TestDeduplicateActivation(t *testing.T) {
	a := New(t)

	payload := []byte{0x01, 0x02, 0x03}
	protocolMetadata := &protocol.RxMetadata{}
	activation1 := &pb_broker.DeviceActivationRequest{Payload: payload, GatewayMetadata: &gateway.RxMetadata{Snr: 1.2}, ProtocolMetadata: protocolMetadata}
	activation2 := &pb_broker.DeviceActivationRequest{Payload: payload, GatewayMetadata: &gateway.RxMetadata{Snr: 3.4}, ProtocolMetadata: protocolMetadata}
	activation3 := &pb_broker.DeviceActivationRequest{Payload: payload, GatewayMetadata: &gateway.RxMetadata{Snr: 5.6}, ProtocolMetadata: protocolMetadata}
	activation4 := &pb_broker.DeviceActivationRequest{Payload: payload, GatewayMetadata: &gateway.RxMetadata{Snr: 7.8}, ProtocolMetadata: protocolMetadata}

	b := getTestBroker(t)
	b.activationDeduplicator = NewDeduplicator(20 * time.Millisecond).(*deduplicator)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		res := b.deduplicateActivation(activation1)
		a.So(res, ShouldResemble, []*pb_broker.DeviceActivationRequest{activation1, activation2, activation3})
		wg.Done()
	}()

	<-time.After(10 * time.Millisecond)

	a.So(b.deduplicateActivation(activation2), ShouldBeNil)
	a.So(b.deduplicateActivation(activation3), ShouldBeNil)

	<-time.After(50 * time.Millisecond)

	a.So(b.deduplicateActivation(activation4), ShouldNotBeNil)

	wg.Wait()
}

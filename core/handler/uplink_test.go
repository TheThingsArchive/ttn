// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleUplink(t *testing.T) {
	a := New(t)
	var err error
	var wg WaitGroup
	appEUI := types.AppEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	appID := "appid"
	devEUI := types.DevEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	devID := "devid"
	h := &handler{
		Component:    &core.Component{Ctx: GetLogger(t, "TestHandleUplink")},
		devices:      device.NewRedisDeviceStore(GetRedisClient(), "handler-test-handle-uplink"),
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-handle-uplink"),
	}
	dev := &device.Device{
		AppID:  appID,
		DevID:  devID,
		AppEUI: appEUI,
		DevEUI: devEUI,
	}
	h.devices.Set(dev)
	defer func() {
		h.devices.Delete(appID, devID)
	}()
	h.applications.Set(&application.Application{
		AppID: appID,
	})
	defer func() {
		h.applications.Delete(appID)
	}()
	h.mqttUp = make(chan *types.UplinkMessage)
	h.mqttEvent = make(chan *mqttEvent, 10)
	h.downlink = make(chan *pb_broker.DownlinkMessage)

	uplink, _ := buildLorawanUplink([]byte{0x40, 0x04, 0x03, 0x02, 0x01, 0x00, 0x01, 0x00, 0x0A, 0x4D, 0xDA, 0x23, 0x99, 0x61, 0xD4})

	downlinkEmpty := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x00, 0x0A, 0x21, 0xEA, 0x8B, 0x0E}
	downlinkACK := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x20, 0x00, 0x00, 0x0A, 0x3B, 0x3F, 0x77, 0x0B}
	downlinkMAC := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x05, 0x00, 0x00, 0x03, 0x30, 0x00, 0x00, 0x00, 0x0A, 0x4D, 0x11, 0x55, 0x01}
	expected := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x00, 0x0A, 0x66, 0xE6, 0x1D, 0x49, 0x82, 0x84}

	downlink := &pb_broker.DownlinkMessage{
		DownlinkOption: &pb_broker.DownlinkOption{
			ProtocolConfig: &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{Lorawan: &pb_lorawan.TxConfiguration{
				FCnt: 0,
			}}},
		},
	}

	// Test Uplink, no downlink option available
	wg.Add(1)
	go func() {
		<-h.mqttUp
		wg.Done()
	}()
	err = h.HandleUplink(uplink)
	a.So(err, ShouldBeNil)
	wg.WaitFor(50 * time.Millisecond)

	uplink.ResponseTemplate = downlink

	// Test Uplink, no downlink needed
	wg.Add(1)
	go func() {
		<-h.mqttUp
		wg.Done()
	}()
	downlink.Payload = downlinkEmpty
	err = h.HandleUplink(uplink)
	a.So(err, ShouldBeNil)
	wg.WaitFor(50 * time.Millisecond)

	// Test Uplink, ACK downlink needed
	wg.Add(2)
	go func() {
		<-h.mqttUp
		wg.Done()
	}()
	go func() {
		<-h.downlink
		wg.Done()
	}()
	downlink.Payload = downlinkACK
	err = h.HandleUplink(uplink)
	a.So(err, ShouldBeNil)
	wg.WaitFor(50 * time.Millisecond)

	// Test Uplink, MAC downlink needed
	wg.Add(2)
	go func() {
		<-h.mqttUp
		wg.Done()
	}()
	go func() {
		<-h.downlink
		wg.Done()
	}()
	downlink.Payload = downlinkMAC
	err = h.HandleUplink(uplink)
	a.So(err, ShouldBeNil)
	wg.WaitFor(50 * time.Millisecond)

	dev.StartUpdate()
	dev.NextDownlink = &types.DownlinkMessage{
		PayloadRaw: []byte{0xaa, 0xbc},
	}

	// Test Uplink, Data downlink needed
	h.devices.Set(dev)
	wg.Add(2)
	go func() {
		<-h.mqttUp
		wg.Done()
	}()
	go func() {
		dl := <-h.downlink
		a.So(dl.Payload, ShouldResemble, expected)
		wg.Done()
	}()
	downlink.Payload = downlinkEmpty
	err = h.HandleUplink(uplink)
	a.So(err, ShouldBeNil)
	wg.WaitFor(50 * time.Millisecond)

	dev, _ = h.devices.Get(appID, devID)
	a.So(dev.NextDownlink, ShouldBeNil)
}

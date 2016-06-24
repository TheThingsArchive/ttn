// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"sync"
	"testing"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleUplink(t *testing.T) {
	a := New(t)
	var err error
	var wg sync.WaitGroup
	appEUI := types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}
	appID := "AppID-1"
	devEUI := types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8}
	h := &handler{
		Component:    &core.Component{Ctx: GetLogger(t, "TestHandleUplink")},
		devices:      device.NewDeviceStore(),
		applications: application.NewApplicationStore(),
	}
	h.devices.Set(&device.Device{
		AppID:  appID,
		AppEUI: appEUI,
		DevEUI: devEUI,
	})
	h.applications.Set(&application.Application{
		AppID: appID,
	})
	h.mqttUp = make(chan *mqtt.UplinkMessage)
	h.downlink = make(chan *pb_broker.DownlinkMessage)

	uplink, _ := buildLorawanUplink([]byte{0x80, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x00, 0x0A, 0x9C, 0x15, 0x7C, 0xEB, 0x09, 0x80})
	downlinkEmpty := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00}
	downlinkACK := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x20, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00}
	downlinkMAC := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x05, 0x00, 0x00, 0x03, 0x30, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00}
	expected := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x00, 0x0A, 0x66, 0xE6, 0x00, 0x00, 0x00, 0x00}
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
	wg.Wait()

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
	wg.Wait()

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
	wg.Wait()

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
	wg.Wait()

	// Test Uplink, Data downlink needed
	h.devices.Set(&device.Device{
		AppEUI: appEUI,
		DevEUI: devEUI,
		NextDownlink: &mqtt.DownlinkMessage{
			Payload: []byte{0xaa, 0xbc},
		},
	})
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
	wg.Wait()

	dev, _ := h.devices.Get(appEUI, devEUI)
	a.So(dev.NextDownlink, ShouldBeNil)
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_protocol "github.com/TheThingsNetwork/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/component"
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
		Component:    &component.Component{Ctx: GetLogger(t, "TestHandleUplink")},
		devices:      device.NewRedisDeviceStore(GetRedisClient(), "handler-test-handle-uplink"),
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-handle-uplink"),
	}
	h.InitStatus()
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
	h.qUp = make(chan *types.UplinkMessage, 10)
	h.qEvent = make(chan *types.DeviceEvent, 10)
	h.downlink = make(chan *pb_broker.DownlinkMessage)

	downlinkEmpty := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x00, 0x0A, 0x21, 0xEA, 0x8B, 0x0E}
	downlinkACK := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x20, 0x00, 0x00, 0x0A, 0x3B, 0x3F, 0x77, 0x0B}
	downlinkMAC := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x05, 0x00, 0x00, 0x03, 0x30, 0x00, 0x00, 0x00, 0x0A, 0x4D, 0x11, 0x55, 0x01}
	expected := []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x00, 0x0A, 0x66, 0xE6, 0x1D, 0x49, 0x82, 0x84}

	downlink := &pb_broker.DownlinkMessage{
		DownlinkOption: &pb_broker.DownlinkOption{
			ProtocolConfiguration: pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_LoRaWAN{LoRaWAN: &pb_lorawan.TxConfiguration{
				FCnt: 0,
			}}},
		},
	}

	getUplink := func() *pb_broker.DeduplicatedUplinkMessage {
		uplink, _ := buildLoRaWANUplink([]byte{0x40, 0x04, 0x03, 0x02, 0x01, 0x00, 0x01, 0x00, 0x0A, 0x4D, 0xDA, 0x23, 0x99, 0x61, 0xD4})
		uplink.ResponseTemplate = downlink
		return uplink
	}

	// Test Uplink, no downlink option available
	{
		wg.Add(1)
		go func() {
			<-h.qUp
			wg.Done()
		}()
		uplink := getUplink()
		uplink.ResponseTemplate = nil
		err = h.HandleUplink(uplink)
		a.So(err, ShouldBeNil)
		wg.WaitFor(50 * time.Millisecond)
	}

	// Test Uplink, no downlink needed
	{
		wg.Add(1)
		go func() {
			<-h.qUp
			wg.Done()
		}()
		downlink.Payload = downlinkEmpty
		err = h.HandleUplink(getUplink())
		a.So(err, ShouldBeNil)
		wg.WaitFor(50 * time.Millisecond)
	}

	// Test Uplink, ACK downlink needed
	{
		wg.Add(2)
		go func() {
			<-h.qUp
			wg.Done()
		}()
		go func() {
			<-h.downlink
			wg.Done()
		}()
		downlink.Payload = downlinkACK
		err = h.HandleUplink(getUplink())
		a.So(err, ShouldBeNil)
		wg.WaitFor(50 * time.Millisecond)
	}

	// Test Uplink, MAC downlink needed
	{
		wg.Add(2)
		go func() {
			<-h.qUp
			wg.Done()
		}()
		go func() {
			<-h.downlink
			wg.Done()
		}()
		downlink.Payload = downlinkMAC
		err = h.HandleUplink(getUplink())
		a.So(err, ShouldBeNil)
		wg.WaitFor(50 * time.Millisecond)
	}

	queue, _ := h.devices.DownlinkQueue(appID, devID)
	queue.PushFirst(&types.DownlinkMessage{PayloadRaw: []byte{0xaa, 0xbc}})

	// Test Uplink, Data downlink needed
	{
		h.devices.Set(dev)
		wg.Add(2)
		go func() {
			<-h.qUp
			wg.Done()
		}()
		go func() {
			dl := <-h.downlink
			a.So(dl.Payload, ShouldResemble, expected)
			wg.Done()
		}()
		downlink.Payload = downlinkEmpty
		err = h.HandleUplink(getUplink())
		a.So(err, ShouldBeNil)
		wg.WaitFor(50 * time.Millisecond)
	}

	dev, _ = h.devices.Get(appID, devID)
	qLen, _ := queue.Length()
	a.So(qLen, ShouldEqual, 0)
	a.So(dev.CurrentDownlink, ShouldNotBeNil)
	a.So(dev.CurrentDownlink.PayloadRaw, ShouldResemble, []byte{0xaa, 0xbc})

	dev.StartUpdate()
	dev.CurrentDownlink = &types.DownlinkMessage{PayloadRaw: []byte{0xaa, 0xbc}, Confirmed: true}
	queue.PushFirst(&types.DownlinkMessage{PayloadRaw: []byte{0x12, 0x34}})

	// Test Uplink, Data downlink needed
	{
		h.devices.Set(dev)
		wg.Add(2)
		go func() {
			<-h.qUp
			wg.Done()
		}()
		go func() {
			dl := <-h.downlink
			a.So(dl.Payload, ShouldResemble, []byte{160, 4, 3, 2, 1, 16, 0, 0, 10, 102, 230, 154, 218, 17, 187}) // The confirmed downlink with FPending on
			wg.Done()
		}()
		downlink.Payload = downlinkEmpty
		err = h.HandleUplink(getUplink())
		a.So(err, ShouldBeNil)
		wg.WaitFor(50 * time.Millisecond)
	}

	dev, _ = h.devices.Get(appID, devID)
	next, _ := queue.Next()
	a.So(next, ShouldNotBeNil)
	a.So(next.PayloadRaw, ShouldResemble, []byte{0x12, 0x34})
	a.So(dev.CurrentDownlink, ShouldNotBeNil)
	a.So(dev.CurrentDownlink.PayloadRaw, ShouldResemble, []byte{0xaa, 0xbc})
}

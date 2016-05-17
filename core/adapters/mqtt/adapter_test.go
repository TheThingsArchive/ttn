// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"sync"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/core/types"
	ttnMQTT "github.com/TheThingsNetwork/ttn/mqtt"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestNewAdapter(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestNewAdapter")
	client := ttnMQTT.NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	adapter := NewAdapter(ctx, client)

	a.So(adapter.(*defaultAdapter).client, ShouldEqual, client)
}

func TestPublishUplink(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestPublishUplink")
	client := ttnMQTT.NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	client.Connect()

	adapter := NewAdapter(ctx, client)

	appEUI := types.AppEUI{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	devEUI := types.DevEUI{0x0a, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	req := core.DataUpAppReq{
		Payload: []byte{0x01, 0x02},
		Metadata: []core.AppMetadata{
			core.AppMetadata{DataRate: "SF7BW125"},
		},
		DevEUI: devEUI.String(),
		FPort:  14,
		FCnt:   200,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	client.SubscribeDeviceUplink(appEUI, devEUI, func(client ttnMQTT.Client, rappEUI types.AppEUI, rdevEUI types.DevEUI, dataUp core.DataUpAppReq) {
		a.So(rappEUI, ShouldEqual, appEUI)
		a.So(rdevEUI, ShouldEqual, devEUI)
		a.So(dataUp.FPort, ShouldEqual, 14)
		a.So(dataUp.FCnt, ShouldEqual, 200)
		a.So(dataUp.Payload, ShouldResemble, []byte{0x01, 0x02})
		a.So(dataUp.Metadata[0].DataRate, ShouldEqual, "SF7BW125")
		wg.Done()
	}).Wait()

	err := adapter.PublishUplink(appEUI, devEUI, req)
	a.So(err, ShouldBeNil)

	wg.Wait()
}

func TestHandleJoin(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestHandleJoin")
	client := ttnMQTT.NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	client.Connect()

	adapter := NewAdapter(ctx, client)

	appEUI := types.AppEUI{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	devEUI := types.DevEUI{0x0a, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	req := core.OTAAAppReq{
		Metadata: []core.AppMetadata{
			core.AppMetadata{DataRate: "SF7BW125"},
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)

	client.SubscribeDeviceActivations(appEUI, devEUI, func(client ttnMQTT.Client, rappEUI types.AppEUI, rdevEUI types.DevEUI, activation core.OTAAAppReq) {
		a.So(rappEUI, ShouldResemble, appEUI)
		a.So(rdevEUI, ShouldResemble, devEUI)
		a.So(activation.Metadata[0].DataRate, ShouldEqual, "SF7BW125")
		wg.Done()
	}).Wait()

	err := adapter.PublishActivation(appEUI, devEUI, req)
	a.So(err, ShouldBeNil)

	wg.Wait()
}

func TestSubscribeDownlink(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestSubscribeDownlink")
	client := ttnMQTT.NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	client.Connect()

	adapter := NewAdapter(ctx, client)
	handler := mocks.NewHandlerServer()

	err := adapter.SubscribeDownlink(handler)
	a.So(err, ShouldBeNil)

	appEUI := types.AppEUI{0x04, 0x03, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	devEUI := types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
	client.PublishDownlink(appEUI, devEUI, core.DataDownAppReq{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()

	<-time.After(50 * time.Millisecond)

	expected := &core.DataDownHandlerReq{
		AppEUI:  appEUI.Bytes(),
		DevEUI:  devEUI.Bytes(),
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}

	a.So(handler.InHandleDataDown.Req, ShouldResemble, expected)
}

func TestSubscribeInvalidDownlink(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestSubscribeInvalidDownlink")
	client := ttnMQTT.NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	client.Connect()

	adapter := NewAdapter(ctx, client)
	handler := mocks.NewHandlerServer()

	err := adapter.SubscribeDownlink(handler)
	a.So(err, ShouldBeNil)

	appEUI := types.AppEUI{0x04, 0x03, 0x03, 0x09, 0x05, 0x06, 0x07, 0x08}
	devEUI := types.DevEUI{0x08, 0x07, 0x06, 0x09, 0x04, 0x03, 0x02, 0x01}
	client.PublishDownlink(appEUI, devEUI, core.DataDownAppReq{Payload: []byte{}}).Wait()

	<-time.After(50 * time.Millisecond)

	a.So(handler.InHandleDataDown.Req, ShouldBeNil)
}

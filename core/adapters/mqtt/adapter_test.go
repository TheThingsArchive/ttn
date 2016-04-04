// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
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

func TestHandleData(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestHandleData")
	client := ttnMQTT.NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	client.Connect()

	adapter := NewAdapter(ctx, client)

	eui := []byte{0x0a, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	req := core.DataAppReq{
		Payload: []byte{0x01, 0x02},
		Metadata: []*core.Metadata{
			&core.Metadata{DataRate: "SF7BW125"},
		},
		AppEUI: eui,
		DevEUI: eui,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	client.SubscribeDeviceUplink(eui, eui, func(client ttnMQTT.Client, appEUI []byte, devEUI []byte, dataUp core.DataUpAppReq) {
		a.So(appEUI, ShouldResemble, eui)
		a.So(devEUI, ShouldResemble, eui)
		a.So(dataUp.Payload, ShouldResemble, []byte{0x01, 0x02})
		a.So(dataUp.Metadata[0].DataRate, ShouldEqual, "SF7BW125")
		wg.Done()
	}).Wait()

	res, err := adapter.HandleData(context.Background(), &req)
	a.So(err, ShouldBeNil)
	a.So(res, ShouldResemble, new(core.DataAppRes))

	wg.Wait()

}

func TestHandleInvalidData(t *testing.T) {
	a := New(t)
	client := ttnMQTT.NewClient(nil, "test", "", "", "tcp://localhost:1883")
	adapter := NewAdapter(nil, client)

	// nil Request
	_, err := adapter.HandleData(context.Background(), nil)
	a.So(err, ShouldNotBeNil)

	// Invalid Payload
	_, err = adapter.HandleData(context.Background(), &core.DataAppReq{
		Payload: []byte{},
	})
	a.So(err, ShouldNotBeNil)

	// Invalid DevEUI
	_, err = adapter.HandleData(context.Background(), &core.DataAppReq{
		Payload: []byte{0x00},
		DevEUI:  []byte{},
	})
	a.So(err, ShouldNotBeNil)

	// Invalid AppEUI
	_, err = adapter.HandleData(context.Background(), &core.DataAppReq{
		Payload: []byte{0x00},
		DevEUI:  []byte{0x0b, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		AppEUI:  []byte{},
	})
	a.So(err, ShouldNotBeNil)

	// Missing Metadata
	_, err = adapter.HandleData(context.Background(), &core.DataAppReq{
		Payload: []byte{0x00},
		DevEUI:  []byte{0x0b, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		AppEUI:  []byte{0x0b, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	})
	a.So(err, ShouldNotBeNil)

	// Not Connected
	_, err = adapter.HandleData(context.Background(), &core.DataAppReq{
		Payload:  []byte{0x00},
		DevEUI:   []byte{0x0b, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		AppEUI:   []byte{0x0b, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Metadata: []*core.Metadata{},
	})
	a.So(err, ShouldNotBeNil)
}

func TestHandleJoin(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestHandleJoin")
	client := ttnMQTT.NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	client.Connect()

	adapter := NewAdapter(ctx, client)

	eui := []byte{0x0a, 0x03, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	req := core.JoinAppReq{
		AppEUI: eui,
		DevEUI: eui,
		Metadata: []*core.Metadata{
			&core.Metadata{DataRate: "SF7BW125"},
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)

	client.SubscribeDeviceActivations(eui, eui, func(client ttnMQTT.Client, appEUI []byte, devEUI []byte, activation core.OTAAAppReq) {
		a.So(appEUI, ShouldResemble, eui)
		a.So(devEUI, ShouldResemble, eui)
		a.So(activation.Metadata[0].DataRate, ShouldEqual, "SF7BW125")
		wg.Done()
	}).Wait()

	res, err := adapter.HandleJoin(context.Background(), &req)
	a.So(err, ShouldBeNil)
	a.So(res, ShouldResemble, new(core.JoinAppRes))

	wg.Wait()
}

func TestHandleInvalidJoin(t *testing.T) {
	a := New(t)
	client := ttnMQTT.NewClient(nil, "test", "", "", "tcp://localhost:1883")
	adapter := NewAdapter(nil, client)

	// nil Request
	_, err := adapter.HandleJoin(context.Background(), nil)
	a.So(err, ShouldNotBeNil)

	// Invalid DevEUI
	_, err = adapter.HandleJoin(context.Background(), &core.JoinAppReq{
		DevEUI: []byte{},
	})
	a.So(err, ShouldNotBeNil)

	// Invalid AppEUI
	_, err = adapter.HandleJoin(context.Background(), &core.JoinAppReq{
		DevEUI: []byte{0x0c, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		AppEUI: []byte{},
	})
	a.So(err, ShouldNotBeNil)

	// Missing Metadata
	_, err = adapter.HandleJoin(context.Background(), &core.JoinAppReq{
		DevEUI: []byte{0x0c, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		AppEUI: []byte{0x0c, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	})
	a.So(err, ShouldNotBeNil)

	// Not Connected
	_, err = adapter.HandleJoin(context.Background(), &core.JoinAppReq{
		DevEUI:   []byte{0x0c, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		AppEUI:   []byte{0x0c, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Metadata: []*core.Metadata{},
	})
	a.So(err, ShouldNotBeNil)
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

	appEUI := []byte{0x04, 0x03, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	devEUI := []byte{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
	client.PublishDownlink(appEUI, devEUI, core.DataDownAppReq{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()

	<-time.After(50 * time.Millisecond)

	expected := &core.DataDownHandlerReq{
		AppEUI:  appEUI,
		DevEUI:  devEUI,
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

	appEUI := []byte{0x04, 0x03, 0x03, 0x09, 0x05, 0x06, 0x07, 0x08}
	devEUI := []byte{0x08, 0x07, 0x06, 0x09, 0x04, 0x03, 0x02, 0x01}
	client.PublishDownlink(appEUI, devEUI, core.DataDownAppReq{Payload: []byte{}}).Wait()

	<-time.After(50 * time.Millisecond)

	a.So(handler.InHandleDataDown.Req, ShouldBeNil)
}

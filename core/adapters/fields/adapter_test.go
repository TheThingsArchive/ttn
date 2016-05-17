// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package fields

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

type mockPublisher struct {
	appEUI        *types.AppEUI
	devEUI        *types.DevEUI
	uplinkReq     *core.DataUpAppReq
	activationReq *core.OTAAAppReq
}

func TestHandleData(t *testing.T) {
	a := New(t)

	ctx := GetLogger(t, "TestHandleData")
	storage, err := ConnectRedis("localhost:6379", 0)
	a.So(err, ShouldBeNil)
	defer storage.Close()
	defer storage.Reset()

	publisher := &mockPublisher{}
	adapter := NewAdapter(ctx, storage, publisher)

	appEUI, _ := types.ParseAppEUI("0102030405060708")
	devEUI, _ := types.ParseDevEUI("00000000AABBCCDD")

	req := &core.DataAppReq{
		AppEUI:  appEUI.Bytes(),
		DevEUI:  devEUI.Bytes(),
		FPort:   1,
		Payload: []byte{0x08, 0x70},
	}

	// No functions
	res, err := adapter.HandleData(context.Background(), req)
	a.So(res, ShouldEqual, new(core.DataAppRes))
	a.So(err, ShouldBeNil)
	a.So(publisher.appEUI, ShouldResemble, &appEUI)
	a.So(publisher.devEUI, ShouldResemble, &devEUI)
	a.So(publisher.uplinkReq, ShouldNotBeNil)
	a.So(publisher.uplinkReq.Fields, ShouldBeEmpty)

	// Normal flow
	functions := &Functions{
		Decoder: `function(data) { return { temperature: ((data[0] << 8) | data[1]) / 100 }; }`,
	}
	err = storage.SetFunctions(appEUI, functions)
	a.So(err, ShouldBeNil)
	res, err = adapter.HandleData(context.Background(), req)
	a.So(res, ShouldEqual, new(core.DataAppRes))
	a.So(err, ShouldBeNil)
	a.So(publisher.appEUI, ShouldResemble, &appEUI)
	a.So(publisher.devEUI, ShouldResemble, &devEUI)
	a.So(publisher.uplinkReq, ShouldNotBeNil)
	a.So(publisher.uplinkReq.Fields, ShouldResemble, map[string]interface{}{
		"temperature": 21.6,
	})

	// Invalidate data
	functions.Validator = `function(data) { return false; }`
	err = storage.SetFunctions(appEUI, functions)
	a.So(err, ShouldBeNil)
	res, err = adapter.HandleData(context.Background(), req)
	a.So(res, ShouldEqual, new(core.DataAppRes))
	a.So(err, ShouldNotBeNil)

	// Function error
	functions.Converter = `function(data) { throw "expected"; }`
	err = storage.SetFunctions(appEUI, functions)
	a.So(err, ShouldBeNil)
	res, err = adapter.HandleData(context.Background(), req)
	a.So(res, ShouldEqual, new(core.DataAppRes))
	a.So(err, ShouldBeNil)
	a.So(publisher.appEUI, ShouldResemble, &appEUI)
	a.So(publisher.devEUI, ShouldResemble, &devEUI)
	a.So(publisher.uplinkReq, ShouldNotBeNil)
	a.So(publisher.uplinkReq.Fields, ShouldResemble, *new(map[string]interface{}))
}

func TestHandleJoin(t *testing.T) {
	a := New(t)

	ctx := GetLogger(t, "TestHandleJoin")
	storage, err := ConnectRedis("localhost:6379", 0)
	a.So(err, ShouldBeNil)
	defer storage.Close()
	defer storage.Reset()

	publisher := &mockPublisher{}
	adapter := NewAdapter(ctx, storage, publisher)

	appEUI, _ := types.ParseAppEUI("0102030405060708")
	devEUI, _ := types.ParseDevEUI("00000000AABBCCDD")

	req := &core.JoinAppReq{
		AppEUI: appEUI.Bytes(),
		DevEUI: devEUI.Bytes(),
	}

	res, err := adapter.HandleJoin(context.Background(), req)
	a.So(res, ShouldEqual, new(core.JoinAppRes))
	a.So(err, ShouldBeNil)

	a.So(publisher.appEUI, ShouldResemble, &appEUI)
	a.So(publisher.devEUI, ShouldResemble, &devEUI)
	a.So(publisher.activationReq, ShouldNotBeNil)
}

func (p *mockPublisher) PublishUplink(appEUI types.AppEUI, devEUI types.DevEUI, req core.DataUpAppReq) error {
	p.appEUI = &appEUI
	p.devEUI = &devEUI
	p.uplinkReq = &req
	return nil
}

func (p *mockPublisher) PublishActivation(appEUI types.AppEUI, devEUI types.DevEUI, req core.OTAAAppReq) error {
	p.appEUI = &appEUI
	p.devEUI = &devEUI
	p.activationReq = &req
	return nil
}

func (p *mockPublisher) SubscribeDownlink(handler core.HandlerServer) error {
	return nil
}

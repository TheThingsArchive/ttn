// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/log/test"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	. "github.com/smartystreets/assertions"
)

func TestSubscribeUnsubscribeDownlink(t *testing.T) {
	a := New(t)

	log := test.NewLogger()
	ttnlog.Set(log)
	defer log.Print(t)

	r := &router{
		Component: &component.Component{Ctx: log},
		gateways:  make(map[string]*gateway.Gateway),
	}

	downlink, err := r.SubscribeDownlink("test")
	a.So(err, ShouldBeNil)
	a.So(downlink, ShouldNotBeNil)

	time.Sleep(10 * time.Millisecond)

	err = r.getGateway("test").ScheduleDownlink("", &pb_router.DownlinkMessage{
		Payload: []byte{1, 2, 3, 4},
	})
	a.So(err, ShouldBeNil)

	select {
	case msg := <-downlink:
		a.So(msg.Payload, ShouldResemble, []byte{1, 2, 3, 4})
	case <-time.After(time.Second):
		t.Fatal("Did not receive on downlink within a second")
	}

	err = r.UnsubscribeDownlink("test")
	a.So(err, ShouldBeNil)

	select {
	case _, ok := <-downlink:
		a.So(ok, ShouldBeFalse)
	case <-time.After(time.Second):
		t.Fatal("Downlink channel was not closed")
	}
}

func TestHandleDownlink(t *testing.T) {
	a := New(t)

	log := test.NewLogger()
	ttnlog.Set(log)
	defer log.Print(t)

	r := &router{
		Component: &component.Component{Ctx: log},
		gateways:  make(map[string]*gateway.Gateway),
	}
	r.InitStatus()

	err := r.HandleDownlink(&pb_broker.DownlinkMessage{
		DownlinkOption: &pb_broker.DownlinkOption{},
	})
	a.So(err, ShouldNotBeNil)

	err = r.HandleDownlink(&pb_broker.DownlinkMessage{
		DownlinkOption: &pb_broker.DownlinkOption{
			Identifier: "incorrect",
		},
	})
	a.So(err, ShouldNotBeNil)

}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"testing"
	"time"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestNewGateway(t *testing.T) {
	a := New(t)
	gtw := NewGateway(GetLogger(t, "TestNewGateway"), "eui-0102030405060708", nil)
	a.So(gtw, ShouldNotBeNil)

	// GatewayTrusted should be determined by the router only
	a.So(gtw.Status().GatewayTrusted, ShouldBeFalse)
	err := gtw.HandleStatus(&pb_gateway.Status{GatewayTrusted: true})
	a.So(err, ShouldBeNil)
	a.So(gtw.Status().GatewayTrusted, ShouldBeFalse)
	gtw.SetAuth("token", true)
	gtw.HandleStatus(&pb_gateway.Status{})
	a.So(gtw.Status().GatewayTrusted, ShouldBeTrue)

	// HandleStatus should update the lastSeen
	a.So(gtw.LastSeen(), ShouldHappenAfter, time.Now().Add(-1*time.Second))
	gtw.lastSeen = time.Unix(0, 0)

	msg := pb_router.RandomLorawanUnconfirmedUplink()

	err = gtw.HandleUplink(msg)
	a.So(err, ShouldBeNil)
}

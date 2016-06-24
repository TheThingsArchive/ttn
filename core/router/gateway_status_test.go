// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleGatewayStatus(t *testing.T) {
	a := New(t)
	eui := types.GatewayEUI{0, 0, 0, 0, 0, 0, 0, 2}

	router := &router{
		Component: &core.Component{
			Ctx: GetLogger(t, "TestHandleGatewayStatus"),
		},
		gateways: map[types.GatewayEUI]*gateway.Gateway{},
	}

	// Handle
	statusMessage := &pb_gateway.Status{Description: "Fake Gateway"}
	err := router.HandleGatewayStatus(eui, statusMessage)
	a.So(err, ShouldBeNil)

	// Check storage
	status, err := router.getGateway(eui).Status.Get()
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, *statusMessage)
}

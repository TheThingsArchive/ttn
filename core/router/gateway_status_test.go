// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleGatewayStatus(t *testing.T) {
	a := New(t)
	gtwID := "eui-0102030405060708"

	router := &router{
		Component: &component.Component{
			Ctx:      GetLogger(t, "TestHandleGatewayStatus"),
			Identity: &pb_discovery.Announcement{},
		},
		gateways: map[string]*gateway.Gateway{},
	}

	// Handle
	statusMessage := &pb_gateway.Status{Description: "Fake Gateway"}
	err := router.HandleGatewayStatus(gtwID, statusMessage)
	a.So(err, ShouldBeNil)

	// Check storage
	status, err := router.getGateway(gtwID).Status.Get()
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, *statusMessage)
}

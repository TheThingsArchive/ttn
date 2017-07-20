// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	pb_discovery "github.com/TheThingsNetwork/api/discovery"
	pb_gateway "github.com/TheThingsNetwork/api/gateway"
	"github.com/TheThingsNetwork/api/monitor/monitorclient"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
	"golang.org/x/net/context"
)

func TestHandleGatewayStatus(t *testing.T) {
	a := New(t)
	gtwID := "eui-0102030405060708"

	logger := GetLogger(t, "TestHandleGatewayStatus")
	router := &router{
		Component: &component.Component{
			Context:  context.Background(),
			Ctx:      logger,
			Identity: &pb_discovery.Announcement{},
			Monitor:  monitorclient.NewMonitorClient(),
		},
		gateways: map[string]*gateway.Gateway{},
	}
	router.InitStatus()

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

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/monitor"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/golang/mock/gomock"
)

type testRouter struct {
	*router
	ctrl      *gomock.Controller
	discovery *discovery.MockClient
}

func getTestRouter(t *testing.T) *testRouter {
	ctrl := gomock.NewController(t)
	discovery := discovery.NewMockClient(ctrl)
	logger := GetLogger(t, "TestRouter")
	r := &testRouter{
		router: &router{
			Component: &component.Component{
				Discovery: discovery,
				Ctx:       logger,
				Monitor:   monitor.NewClient(monitor.DefaultClientConfig),
			},
			gateways: map[string]*gateway.Gateway{},
		},
		ctrl:      ctrl,
		discovery: discovery,
	}
	r.InitStatus()
	return r
}

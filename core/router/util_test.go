// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	"github.com/TheThingsNetwork/api/discovery/discoveryclient"
	"github.com/TheThingsNetwork/api/monitor/monitorclient"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/golang/mock/gomock"
	"golang.org/x/net/context"
)

type testRouter struct {
	*router
	ctrl      *gomock.Controller
	discovery *discoveryclient.MockClient
}

func getTestRouter(t *testing.T) *testRouter {
	ctrl := gomock.NewController(t)
	discovery := discoveryclient.NewMockClient(ctrl)
	logger := GetLogger(t, "TestRouter")
	r := &testRouter{
		router: &router{
			Component: &component.Component{
				Context:   context.Background(),
				Discovery: discovery,
				Ctx:       logger,
				Monitor:   monitorclient.NewMonitorClient(),
			},
			gateways: map[string]*gateway.Gateway{},
		},
		ctrl:      ctrl,
		discovery: discovery,
	}
	r.InitStatus()
	return r
}

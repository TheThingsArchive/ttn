// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"testing"
	"time"

	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/monitor"
	pb_networkserver "github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/core/component"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/golang/mock/gomock"
)

type testBroker struct {
	*broker
	ctrl      *gomock.Controller
	discovery *pb_discovery.MockClient
	ns        *pb_networkserver.MockNetworkServerClient
}

func getTestBroker(t *testing.T) *testBroker {
	ctrl := gomock.NewController(t)
	discovery := pb_discovery.NewMockClient(ctrl)
	ns := pb_networkserver.NewMockNetworkServerClient(ctrl)
	logger := GetLogger(t, "TestBroker")
	b := &testBroker{
		broker: &broker{
			Component: &component.Component{
				Discovery: discovery,
				Ctx:       logger,
				Monitor:   monitor.NewClient(monitor.DefaultClientConfig),
			},
			handlers:               make(map[string]*handler),
			activationDeduplicator: NewDeduplicator(10 * time.Millisecond),
			uplinkDeduplicator:     NewDeduplicator(10 * time.Millisecond),
			ns:                     ns,
		},
		ns:        ns,
		ctrl:      ctrl,
		discovery: discovery,
	}
	b.InitStatus()
	return b
}

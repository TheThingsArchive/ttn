// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"sort"
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_monitor "github.com/TheThingsNetwork/ttn/api/monitor"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestDownlinkScoring(t *testing.T) {
	a := New(t)
	refMD := pb_gateway.RxMetadata{Snr: 10, Rssi: -50}
	refOption := pb.DownlinkOption{PossibleConflicts: 1, DutyCycle: 0.1, Utilization: 0.1}

	var options ByScore
	sort := func(md pb_gateway.RxMetadata, option pb.DownlinkOption) {
		options = ByScore{
			downlinkOption{uplinkMetadata: &refMD, option: &refOption},
			downlinkOption{uplinkMetadata: &md, option: &option},
		}
		sort.Sort(options)
	}

	{
		md, option := refMD, refOption
		md.Snr += 2
		sort(md, option)
		a.So(*options[0].uplinkMetadata, ShouldResemble, md)
	}

	{
		md, option := refMD, refOption
		md.Snr -= 2
		sort(md, option)
		a.So(*options[0].uplinkMetadata, ShouldResemble, refMD)
	}

	{
		md, option := refMD, refOption
		md.Rssi += 10
		sort(md, option)
		a.So(*options[0].uplinkMetadata, ShouldResemble, md)
	}

	{
		md, option := refMD, refOption
		md.Rssi -= 10
		sort(md, option)
		a.So(*options[0].uplinkMetadata, ShouldResemble, refMD)
	}

	{
		md, option := refMD, refOption
		option.Utilization -= 0.05
		sort(md, option)
		a.So(*options[0].option, ShouldResemble, option)
	}

	{
		md, option := refMD, refOption
		option.Utilization += 0.05
		sort(md, option)
		a.So(*options[0].option, ShouldResemble, refOption)
	}

}

func TestDownlink(t *testing.T) {
	a := New(t)

	appEUI := types.AppEUI{0, 1, 2, 3, 4, 5, 6, 7}
	devEUI := types.DevEUI{0, 1, 2, 3, 4, 5, 6, 7}

	dlch := make(chan *pb.DownlinkMessage, 2)
	logger := GetLogger(t, "TestDownlink")
	b := &broker{
		Component: &component.Component{
			Ctx:     logger,
			Monitor: pb_monitor.NewClient(pb_monitor.DefaultClientConfig),
		},
		ns: &mockNetworkServer{},
		routers: map[string]chan *pb.DownlinkMessage{
			"routerID": dlch,
		},
	}
	b.InitStatus()

	err := b.HandleDownlink(&pb.DownlinkMessage{
		DevEui: &devEUI,
		AppEui: &appEUI,
		DownlinkOption: &pb.DownlinkOption{
			Identifier: "fakeID",
		},
	})
	a.So(err, ShouldNotBeNil)

	err = b.HandleDownlink(&pb.DownlinkMessage{
		DevEui: &devEUI,
		AppEui: &appEUI,
		DownlinkOption: &pb.DownlinkOption{
			Identifier: "nonExistentRouterID:scheduleID",
		},
	})
	a.So(err, ShouldNotBeNil)

	err = b.HandleDownlink(&pb.DownlinkMessage{
		DevEui: &devEUI,
		AppEui: &appEUI,
		DownlinkOption: &pb.DownlinkOption{
			Identifier: "routerID:scheduleID",
		},
	})
	a.So(err, ShouldBeNil)
	a.So(len(dlch), ShouldEqual, 1)
}

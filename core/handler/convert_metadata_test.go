// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestConvertMetadata(t *testing.T) {
	a := New(t)
	h := &handler{
		Component: &core.Component{Ctx: GetLogger(t, "TestConvertMetadata")},
	}

	ttnUp := &pb_broker.DeduplicatedUplinkMessage{}
	appUp := &types.UplinkMessage{}

	err := h.ConvertMetadata(h.Ctx, ttnUp, appUp)
	a.So(err, ShouldBeNil)

	gtwID := "eui-0102030405060708"
	ttnUp.GatewayMetadata = []*pb_gateway.RxMetadata{
		&pb_gateway.RxMetadata{
			GatewayId: gtwID,
		},
		&pb_gateway.RxMetadata{
			GatewayId: gtwID,
		},
	}

	err = h.ConvertMetadata(h.Ctx, ttnUp, appUp)
	a.So(err, ShouldBeNil)
	a.So(appUp.Metadata.Gateways, ShouldHaveLength, 2)

	ttnUp.ProtocolMetadata = &pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_Lorawan{
		Lorawan: &pb_lorawan.Metadata{
			DataRate: "SF7BW125",
		},
	}}

	err = h.ConvertMetadata(h.Ctx, ttnUp, appUp)
	a.So(err, ShouldBeNil)
	a.So(appUp.Metadata.DataRate, ShouldEqual, "SF7BW125")

	ttnUp.GatewayMetadata[0].Time = 1465831736000000000
	ttnUp.GatewayMetadata[0].Gps = &pb_gateway.GPSMetadata{
		Latitude: 42,
	}

	err = h.ConvertMetadata(h.Ctx, ttnUp, appUp)
	a.So(err, ShouldBeNil)
	a.So(appUp.Metadata.Gateways[0].Latitude, ShouldEqual, 42)
	a.So(time.Time(appUp.Metadata.Gateways[0].Time).UTC(), ShouldResemble, time.Date(2016, 06, 13, 15, 28, 56, 0, time.UTC))

}

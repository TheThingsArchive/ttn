// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_gateway "github.com/TheThingsNetwork/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestConvertMetadata(t *testing.T) {
	a := New(t)
	h := &handler{
		Component: &component.Component{Ctx: GetLogger(t, "TestConvertMetadata")},
	}

	ttnUp := &pb_broker.DeduplicatedUplinkMessage{
		Payload: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
	}
	appUp := &types.UplinkMessage{}
	device := &device.Device{
		Latitude: 12.34,
	}

	err := h.ConvertMetadata(h.Ctx, ttnUp, appUp, device)
	a.So(err, ShouldBeNil)
	a.So(appUp.Metadata.Latitude, ShouldEqual, 12.34)

	gtwID := "eui-0102030405060708"
	ttnUp.GatewayMetadata = []*pb_gateway.RxMetadata{
		&pb_gateway.RxMetadata{
			GatewayID: gtwID,
		},
		&pb_gateway.RxMetadata{
			GatewayID: gtwID,
			Antennas: []*pb_gateway.RxMetadata_Antenna{
				&pb_gateway.RxMetadata_Antenna{},
				&pb_gateway.RxMetadata_Antenna{},
			},
		},
	}

	err = h.ConvertMetadata(h.Ctx, ttnUp, appUp, device)
	a.So(err, ShouldBeNil)
	a.So(appUp.Metadata.Gateways, ShouldHaveLength, 3)

	ttnUp.ProtocolMetadata = pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_LoRaWAN{
		LoRaWAN: &pb_lorawan.Metadata{
			DataRate:   "SF7BW125",
			CodingRate: "4/5",
		},
	}}

	err = h.ConvertMetadata(h.Ctx, ttnUp, appUp, device)
	a.So(err, ShouldBeNil)
	a.So(appUp.Metadata.DataRate, ShouldEqual, "SF7BW125")

	ttnUp.GatewayMetadata[0].Time = 1465831736000000000
	ttnUp.GatewayMetadata[0].Location = &pb_gateway.LocationMetadata{
		Latitude: 42,
	}

	err = h.ConvertMetadata(h.Ctx, ttnUp, appUp, device)
	a.So(err, ShouldBeNil)
	a.So(appUp.Metadata.Gateways[0].Latitude, ShouldEqual, 42)
	a.So(time.Time(appUp.Metadata.Gateways[0].Time).UTC(), ShouldResemble, time.Date(2016, 06, 13, 15, 28, 56, 0, time.UTC))
	a.So(appUp.Metadata.Airtime, ShouldEqual, time.Duration(46336)*time.Microsecond)
}

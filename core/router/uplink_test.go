// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	"github.com/TheThingsNetwork/api/discovery"
	pb_gateway "github.com/TheThingsNetwork/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/api/router"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/assertions"
)

// newReferenceGateway returns a default gateway
func newReferenceGateway(t *testing.T, frequencyPlan string) *gateway.Gateway {
	gtw := gateway.NewGateway(GetLogger(t, "ReferenceGateway"), "eui-0102030405060708")
	gtw.Status.Update(&pb_gateway.Status{
		FrequencyPlan: frequencyPlan,
	})
	return gtw
}

// newReferenceUplink returns a default uplink message
func newReferenceUplink() *pb.UplinkMessage {
	gtwID := "eui-0102030405060708"

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.UnconfirmedDataUp,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: lorawan.DevAddr([4]byte{1, 2, 3, 4}),
			},
		},
	}
	bytes, _ := phy.MarshalBinary()

	up := &pb.UplinkMessage{
		Payload: bytes,
		ProtocolMetadata: &pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_LoRaWAN{LoRaWAN: &pb_lorawan.Metadata{
			CodingRate: "4/5",
			DataRate:   "SF7BW125",
			Modulation: pb_lorawan.Modulation_LORA,
		}}},
		GatewayMetadata: &pb_gateway.RxMetadata{
			GatewayID: gtwID,
			Timestamp: 100,
			Frequency: 868100000,
			RSSI:      -25.0,
			SNR:       5.0,
		},
	}
	return up
}

func TestHandleUplink(t *testing.T) {
	a := New(t)

	r := getTestRouter(t)
	r.discovery.EXPECT().GetAllBrokersForDevAddr(types.DevAddr([4]byte{1, 2, 3, 4})).Return([]*discovery.Announcement{}, nil)

	uplink := newReferenceUplink()
	gtwID := "eui-0102030405060708"

	err := r.HandleUplink(gtwID, uplink)
	a.So(err, ShouldBeNil)
	utilization := r.getGateway(gtwID).Utilization
	utilization.Tick()
	rx, _ := utilization.Get()
	a.So(rx, ShouldBeGreaterThan, 0)

	// TODO: Integration test that checks broker forward
}

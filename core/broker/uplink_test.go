// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"sync"
	"testing"
	"time"

	pb "github.com/TheThingsNetwork/api/broker"
	pb_discovery "github.com/TheThingsNetwork/api/discovery"
	"github.com/TheThingsNetwork/api/gateway"
	pb_networkserver "github.com/TheThingsNetwork/api/networkserver"
	"github.com/TheThingsNetwork/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/assertions"
)

func TestHandleUplink(t *testing.T) {
	a := New(t)

	b := getTestBroker(t)

	gtwID := "eui-0102030405060708"

	// Invalid Payload
	err := b.HandleUplink(&pb.UplinkMessage{
		Payload:          []byte{0x01, 0x02, 0x03},
		GatewayMetadata:  &gateway.RxMetadata{SNR: 1.2, GatewayID: gtwID},
		ProtocolMetadata: &protocol.RxMetadata{},
	})
	a.So(err, ShouldNotBeNil)

	// Valid Payload
	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.UnconfirmedDataUp,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: lorawan.DevAddr([4]byte{1, 2, 3, 4}),
				FCnt:    1,
			},
		},
	}
	bytes, _ := phy.MarshalBinary()

	// Device not found
	b.uplinkDeduplicator = NewDeduplicator(10 * time.Millisecond)
	b.ns.EXPECT().GetDevices(gomock.Any(), gomock.Any()).Return(&pb_networkserver.DevicesResponse{
		Results: []*pb_lorawan.Device{},
	}, nil)
	err = b.HandleUplink(&pb.UplinkMessage{
		Payload:          bytes,
		GatewayMetadata:  &gateway.RxMetadata{SNR: 1.2, GatewayID: gtwID},
		ProtocolMetadata: &protocol.RxMetadata{Protocol: &protocol.RxMetadata_LoRaWAN{LoRaWAN: &pb_lorawan.Metadata{}}},
	})
	a.So(err, ShouldHaveSameTypeAs, &errors.ErrNotFound{})

	devEUI := types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8}
	wrongDevEUI := types.DevEUI{1, 2, 3, 4, 5, 6, 7, 9}
	appEUI := types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}
	appID := "appid-1"
	nwkSKey := types.NwkSKey{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}

	// Add devices
	b = getTestBroker(t)
	nsResponse := &pb_networkserver.DevicesResponse{
		Results: []*pb_lorawan.Device{
			&pb_lorawan.Device{
				DevEUI:  &wrongDevEUI,
				AppEUI:  &appEUI,
				AppID:   appID,
				NwkSKey: &nwkSKey,
				FCntUp:  4,
			},
			&pb_lorawan.Device{
				DevEUI:  &devEUI,
				AppEUI:  &appEUI,
				AppID:   appID,
				NwkSKey: &nwkSKey,
				FCntUp:  3,
			},
		},
	}
	b.handlers["handlerID"] = &handler{uplink: make(chan *pb.DeduplicatedUplinkMessage, 10)}

	// Device doesn't match
	b.uplinkDeduplicator = NewDeduplicator(10 * time.Millisecond)
	b.ns.EXPECT().GetDevices(gomock.Any(), gomock.Any()).Return(nsResponse, nil)
	err = b.HandleUplink(&pb.UplinkMessage{
		Payload:          bytes,
		GatewayMetadata:  &gateway.RxMetadata{SNR: 1.2, GatewayID: gtwID},
		ProtocolMetadata: &protocol.RxMetadata{Protocol: &protocol.RxMetadata_LoRaWAN{LoRaWAN: &pb_lorawan.Metadata{}}},
	})
	a.So(err, ShouldHaveSameTypeAs, &errors.ErrNotFound{})

	phy.SetMIC(lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8})
	bytes, _ = phy.MarshalBinary()

	// Wrong FCnt
	b.uplinkDeduplicator = NewDeduplicator(10 * time.Millisecond)
	b.ns.EXPECT().GetDevices(gomock.Any(), gomock.Any()).Return(nsResponse, nil)
	err = b.HandleUplink(&pb.UplinkMessage{
		Payload:          bytes,
		GatewayMetadata:  &gateway.RxMetadata{SNR: 1.2, GatewayID: gtwID},
		ProtocolMetadata: &protocol.RxMetadata{Protocol: &protocol.RxMetadata_LoRaWAN{LoRaWAN: &pb_lorawan.Metadata{}}},
	})
	a.So(err, ShouldHaveSameTypeAs, &errors.ErrInvalidArgument{})

	// Disable FCnt Check
	b.uplinkDeduplicator = NewDeduplicator(10 * time.Millisecond)
	nsResponse.Results[0].DisableFCntCheck = true
	b.ns.EXPECT().GetDevices(gomock.Any(), gomock.Any()).Return(nsResponse, nil)
	b.ns.EXPECT().Uplink(gomock.Any(), gomock.Any()).Return(&pb.DeduplicatedUplinkMessage{}, nil)
	b.discovery.EXPECT().GetAllHandlersForAppID("appid-1").Return([]*pb_discovery.Announcement{
		&pb_discovery.Announcement{
			ID: "handlerID",
		},
	}, nil)
	err = b.HandleUplink(&pb.UplinkMessage{
		Payload:          bytes,
		GatewayMetadata:  &gateway.RxMetadata{SNR: 1.2, GatewayID: gtwID},
		ProtocolMetadata: &protocol.RxMetadata{Protocol: &protocol.RxMetadata_LoRaWAN{LoRaWAN: &pb_lorawan.Metadata{}}},
	})
	a.So(err, ShouldBeNil)

	// OK FCnt
	b.uplinkDeduplicator = NewDeduplicator(10 * time.Millisecond)
	nsResponse.Results[0].FCntUp = 0
	nsResponse.Results[0].DisableFCntCheck = false
	b.ns.EXPECT().GetDevices(gomock.Any(), gomock.Any()).Return(nsResponse, nil)
	b.ns.EXPECT().Uplink(gomock.Any(), gomock.Any()).Return(&pb.DeduplicatedUplinkMessage{}, nil)
	b.discovery.EXPECT().GetAllHandlersForAppID("appid-1").Return([]*pb_discovery.Announcement{
		&pb_discovery.Announcement{
			ID: "handlerID",
		},
	}, nil)
	err = b.HandleUplink(&pb.UplinkMessage{
		Payload:          bytes,
		GatewayMetadata:  &gateway.RxMetadata{SNR: 1.2, GatewayID: gtwID},
		ProtocolMetadata: &protocol.RxMetadata{Protocol: &protocol.RxMetadata_LoRaWAN{LoRaWAN: &pb_lorawan.Metadata{}}},
	})
	a.So(err, ShouldBeNil)
}

func TestDeduplicateUplink(t *testing.T) {
	a := New(t)

	payload := []byte{0x01, 0x02, 0x03}
	protocolMetadata := &protocol.RxMetadata{}
	uplink1 := &pb.UplinkMessage{Payload: payload, GatewayMetadata: &gateway.RxMetadata{SNR: 1.2}, ProtocolMetadata: protocolMetadata}
	uplink2 := &pb.UplinkMessage{Payload: payload, GatewayMetadata: &gateway.RxMetadata{SNR: 3.4}, ProtocolMetadata: protocolMetadata}
	uplink3 := &pb.UplinkMessage{Payload: payload, GatewayMetadata: &gateway.RxMetadata{SNR: 5.6}, ProtocolMetadata: protocolMetadata}
	uplink4 := &pb.UplinkMessage{Payload: payload, GatewayMetadata: &gateway.RxMetadata{SNR: 7.8}, ProtocolMetadata: protocolMetadata}

	b := getTestBroker(t)
	b.uplinkDeduplicator = NewDeduplicator(20 * time.Millisecond).(*deduplicator)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		res := b.deduplicateUplink(uplink1)
		a.So(res, ShouldResemble, []*pb.UplinkMessage{uplink1, uplink2, uplink3})
		a.So(res, ShouldNotContain, uplink4)
		wg.Done()
	}()

	<-time.After(10 * time.Millisecond)

	a.So(b.deduplicateUplink(uplink2), ShouldBeNil)
	a.So(b.deduplicateUplink(uplink3), ShouldBeNil)

	<-time.After(50 * time.Millisecond)

	a.So(b.deduplicateUplink(uplink4), ShouldNotBeNil)

	wg.Wait()
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

type payloadMarshalerUnmarshaler interface {
	UnmarshalPayload() error
	MarshalPayload() error
}

func TestMarshalUnmarshalPayload(t *testing.T) {
	a := New(t)

	var subjects []payloadMarshalerUnmarshaler

	// Do nothing when message and payload are nil
	subjects = []payloadMarshalerUnmarshaler{
		&UplinkMessage{},
		&DownlinkMessage{},
		&DeviceActivationRequest{},
	}

	for _, sub := range subjects {
		a.So(sub.MarshalPayload(), ShouldEqual, nil)
		a.So(sub.UnmarshalPayload(), ShouldEqual, nil)
	}

	rxMeta := &pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_Lorawan{Lorawan: &pb_lorawan.Metadata{}}}
	txConf := &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{Lorawan: &pb_lorawan.TxConfiguration{}}}

	macMsg := &pb_protocol.Message{Protocol: &pb_protocol.Message_Lorawan{Lorawan: &pb_lorawan.Message{
		MHDR: pb_lorawan.MHDR{
			Major: 1,
			MType: pb_lorawan.MType_UNCONFIRMED_UP,
		},
		Payload: &pb_lorawan.Message_MacPayload{MacPayload: &pb_lorawan.MACPayload{
			FHDR: pb_lorawan.FHDR{
				DevAddr: types.DevAddr([4]byte{1, 2, 3, 4}),
				FCnt:    1,
			},
		}},
		Mic: []byte{1, 2, 3, 4},
	}}}
	macBin := []byte{65, 4, 3, 2, 1, 0, 1, 0, 0, 1, 2, 3, 4}
	joinReqMsg := &pb_protocol.Message{Protocol: &pb_protocol.Message_Lorawan{Lorawan: &pb_lorawan.Message{
		MHDR: pb_lorawan.MHDR{
			Major: 1,
			MType: pb_lorawan.MType_JOIN_REQUEST,
		},
		Payload: &pb_lorawan.Message_JoinRequestPayload{JoinRequestPayload: &pb_lorawan.JoinRequestPayload{
			AppEui:   types.AppEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
			DevEui:   types.DevEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
			DevNonce: types.DevNonce([2]byte{1, 2}),
		}},
		Mic: []byte{1, 2, 3, 4},
	}}}
	joinReqBin := []byte{1, 8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1, 2, 1, 1, 2, 3, 4}

	// Only Marshal
	subjects = []payloadMarshalerUnmarshaler{
		&UplinkMessage{
			ProtocolMetadata: rxMeta,
			Message:          macMsg,
		},
		&DownlinkMessage{
			ProtocolConfiguration: txConf,
			Message:               macMsg,
		},
		&DeviceActivationRequest{
			ProtocolMetadata: rxMeta,
			Message:          joinReqMsg,
		},
	}

	for _, sub := range subjects {
		a.So(sub.UnmarshalPayload(), ShouldEqual, nil)
		a.So(sub.MarshalPayload(), ShouldEqual, nil)
	}

	// Only Unmarshal
	subjects = []payloadMarshalerUnmarshaler{
		&UplinkMessage{
			ProtocolMetadata: rxMeta,
			Payload:          macBin,
		},
		&DownlinkMessage{
			ProtocolConfiguration: txConf,
			Payload:               macBin,
		},
		&DeviceActivationRequest{
			ProtocolMetadata: rxMeta,
			Payload:          joinReqBin,
		},
	}

	for _, sub := range subjects {
		a.So(sub.MarshalPayload(), ShouldEqual, nil)
		a.So(sub.UnmarshalPayload(), ShouldEqual, nil)
	}

}

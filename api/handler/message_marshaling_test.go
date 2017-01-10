// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
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
		&DeviceActivationResponse{},
	}

	for _, sub := range subjects {
		a.So(sub.MarshalPayload(), ShouldEqual, nil)
		a.So(sub.UnmarshalPayload(), ShouldEqual, nil)
	}

	txConf := &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{Lorawan: &pb_lorawan.TxConfiguration{}}}

	joinAccMsg := &pb_protocol.Message{Protocol: &pb_protocol.Message_Lorawan{Lorawan: &pb_lorawan.Message{
		MHDR: pb_lorawan.MHDR{
			Major: 1,
			MType: pb_lorawan.MType_JOIN_ACCEPT,
		},
		Payload: &pb_lorawan.Message_JoinAcceptPayload{JoinAcceptPayload: &pb_lorawan.JoinAcceptPayload{
			AppNonce: types.AppNonce([3]byte{1, 2, 3}),
			NetId:    types.NetID([3]byte{1, 2, 3}),
			DevAddr:  types.DevAddr([4]byte{1, 2, 3, 4}),
			DLSettings: pb_lorawan.DLSettings{
				Rx2Dr: 3,
			},
		}},
		Mic: []byte{1, 2, 3, 4},
	}}}
	joinAccBin := []byte{33, 3, 2, 1, 3, 2, 1, 4, 3, 2, 1, 3, 0, 1, 2, 3, 4}

	// Only Marshal
	subjects = []payloadMarshalerUnmarshaler{
		&DeviceActivationResponse{
			DownlinkOption: &pb_broker.DownlinkOption{ProtocolConfig: txConf},
			Message:        joinAccMsg,
		},
	}

	for _, sub := range subjects {
		a.So(sub.UnmarshalPayload(), ShouldEqual, nil)
		a.So(sub.MarshalPayload(), ShouldEqual, nil)
	}

	// Only Unmarshal
	subjects = []payloadMarshalerUnmarshaler{
		&DeviceActivationResponse{
			DownlinkOption: &pb_broker.DownlinkOption{ProtocolConfig: txConf},
			Payload:        joinAccBin,
		},
	}

	for _, sub := range subjects {
		a.So(sub.MarshalPayload(), ShouldEqual, nil)
		a.So(sub.UnmarshalPayload(), ShouldEqual, nil)
	}

}

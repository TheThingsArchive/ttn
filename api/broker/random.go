// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"math/rand"
	"time"

	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/utils/random"
)

func randomDeduplicatedUplinkMessage() *DeduplicatedUplinkMessage {
	appEui := random.AppEUI()
	devEui := random.DevEUI()
	md := make([]*gateway.RxMetadata, random.Intn(10))
	for i := range md {
		md[i] = gateway.RandomRxMetadata()
	}
	return &DeduplicatedUplinkMessage{
		AppEui:          &appEui,
		DevEui:          &devEui,
		AppId:           random.AppID(),
		DevId:           random.DevID(),
		GatewayMetadata: md,
	}
}

// RandomLorawanDeduplicatedJoinRequest returns randomly generated lorawan join request.
// Used for testing.
func RandomLorawanDeduplicatedJoinRequest(modulation ...lorawan.Modulation) *DeduplicatedUplinkMessage {
	msg := randomDeduplicatedUplinkMessage()
	msg.ProtocolMetadata = protocol.RandomLorawanRxMetadata(modulation...)
	msg.Payload = lorawan.RandomPayload(lorawan.MType_JOIN_REQUEST)
	return msg
}

// RandomLorawanConfirmedDeduplicatedUplink returns randomly generated confirmed lorawan uplink message.
// Used for testing.
func RandomLorawanConfirmedDeduplicatedUplink(modulation ...lorawan.Modulation) *DeduplicatedUplinkMessage {
	msg := randomDeduplicatedUplinkMessage()
	msg.ProtocolMetadata = protocol.RandomLorawanRxMetadata(modulation...)
	msg.Payload = lorawan.RandomPayload(lorawan.MType_CONFIRMED_UP)
	return msg
}

// RandomLorawanUnconfirmedDeduplicatedUplink returns randomly generated unconfirmed lorawan uplink message.
// Used for testing.
func RandomLorawanUnconfirmedDeduplicatedUplink(modulation ...lorawan.Modulation) *DeduplicatedUplinkMessage {
	msg := randomDeduplicatedUplinkMessage()
	msg.ProtocolMetadata = protocol.RandomLorawanRxMetadata(modulation...)
	msg.Payload = lorawan.RandomPayload(lorawan.MType_UNCONFIRMED_UP)
	return msg
}

// RandomLorawanDeduplicatedUplinkMessage returns randomly generated lorawan uplink message(join request, confirmed or unconfirmed uplink).
// Used for testing.
func RandomLorawanDeduplicatedUplinkMessage(modulation ...lorawan.Modulation) *DeduplicatedUplinkMessage {
	switch rand.Intn(3) {
	case 0:
		return RandomLorawanDeduplicatedJoinRequest(modulation...)
	case 1:
		return RandomLorawanConfirmedDeduplicatedUplink(modulation...)
	default:
		return RandomLorawanUnconfirmedDeduplicatedUplink(modulation...)
	}
}

func randomDownlinkMessage() *DownlinkMessage {
	appEui := random.AppEUI()
	devEui := random.DevEUI()
	return &DownlinkMessage{
		AppEui: &appEui,
		DevEui: &devEui,
		AppId:  random.AppID(),
		DevId:  random.DevID(),
		DownlinkOption: &DownlinkOption{
			GatewayConfig:  gateway.RandomTxConfiguration(),
			ProtocolConfig: protocol.RandomTxConfiguration(),
			Score:          uint32(random.Intn(100)),
			Deadline:       int64(time.Now().Nanosecond() + random.Intn(10000)),
			GatewayId:      random.ID(),
			Identifier:     random.ID(),
		},
	}
}

// RandomLorawanJoinAccept returns randomly generated lorawan join request.
// Used for testing.
func RandomLorawanJoinAccept() *DownlinkMessage {
	msg := randomDownlinkMessage()
	msg.Payload = lorawan.RandomPayload(lorawan.MType_JOIN_ACCEPT)
	return msg
}

// RandomLorawanConfirmedDownlink returns randomly generated confirmed lorawan uplink message.
// Used for testing.
func RandomLorawanConfirmedDownlink() *DownlinkMessage {
	msg := randomDownlinkMessage()
	msg.Payload = lorawan.RandomPayload(lorawan.MType_CONFIRMED_DOWN)
	return msg
}

// RandomLorawanUnconfirmedDownlink returns randomly generated unconfirmed lorawan uplink message.
// Used for testing.
func RandomLorawanUnconfirmedDownlink() *DownlinkMessage {
	msg := randomDownlinkMessage()
	msg.Payload = lorawan.RandomPayload(lorawan.MType_UNCONFIRMED_DOWN)
	return msg
}

// RandomLorawanDownlinkMessage returns randomly generated lorawan downlink message(join request, confirmed or unconfirmed downlink).
// Used for testing.
func RandomLorawanDownlinkMessage() *DownlinkMessage {
	switch rand.Intn(3) {
	case 0:
		return RandomLorawanJoinAccept()
	case 1:
		return RandomLorawanConfirmedDownlink()
	default:
		return RandomLorawanUnconfirmedDownlink()
	}
}

// RandomDeduplicatedUplinkMessage returns randomly generated uplink message.
// Used for testing.
func RandomDeduplicatedUplinkMessage() *DeduplicatedUplinkMessage {
	return RandomLorawanDeduplicatedUplinkMessage()
}

// RandomDownlinkMessage returns randomly generated downlink message.
// Used for testing.
func RandomDownlinkMessage() *DownlinkMessage {
	return RandomLorawanDownlinkMessage()
}

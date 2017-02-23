// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"math/rand"

	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
)

// RandomLorawanJoinRequest returns randomly generated lorawan join request.
// Used for testing.
func RandomLorawanJoinRequest(modulation ...lorawan.Modulation) *UplinkMessage {
	return &UplinkMessage{
		GatewayMetadata:  gateway.RandomRxMetadata(),
		ProtocolMetadata: protocol.RandomLorawanRxMetadata(modulation...),
		Payload:          lorawan.RandomPayload(lorawan.MType_JOIN_REQUEST),
	}
}

// RandomLorawanConfirmedUplink returns randomly generated confirmed lorawan uplink message.
// Used for testing.
func RandomLorawanConfirmedUplink(modulation ...lorawan.Modulation) *UplinkMessage {
	return &UplinkMessage{
		GatewayMetadata:  gateway.RandomRxMetadata(),
		ProtocolMetadata: protocol.RandomLorawanRxMetadata(modulation...),
		Payload:          lorawan.RandomPayload(lorawan.MType_CONFIRMED_UP),
	}
}

// RandomLorawanUnconfirmedUplink returns randomly generated unconfirmed lorawan uplink message.
// Used for testing.
func RandomLorawanUnconfirmedUplink(modulation ...lorawan.Modulation) *UplinkMessage {
	return &UplinkMessage{
		GatewayMetadata:  gateway.RandomRxMetadata(),
		ProtocolMetadata: protocol.RandomLorawanRxMetadata(modulation...),
		Payload:          lorawan.RandomPayload(lorawan.MType_UNCONFIRMED_UP),
	}
}

// RandomLorawanUplinkMessage returns randomly generated lorawan uplink message(join request, confirmed or unconfirmed uplink).
// Used for testing.
func RandomLorawanUplinkMessage(modulation ...lorawan.Modulation) *UplinkMessage {
	switch rand.Intn(3) {
	case 0:
		return RandomLorawanJoinRequest(modulation...)
	case 1:
		return RandomLorawanConfirmedUplink(modulation...)
	default:
		return RandomLorawanUnconfirmedUplink(modulation...)
	}
}

// RandomLorawanJoinAccept returns randomly generated lorawan join request.
// Used for testing.
func RandomLorawanJoinAccept(modulation ...lorawan.Modulation) *DownlinkMessage {
	return &DownlinkMessage{
		GatewayConfiguration:  gateway.RandomTxConfiguration(),
		ProtocolConfiguration: protocol.RandomLorawanTxConfiguration(modulation...),
		Payload:               lorawan.RandomPayload(lorawan.MType_JOIN_ACCEPT),
	}
}

// RandomLorawanConfirmedDownlink returns randomly generated confirmed lorawan uplink message.
// Used for testing.
func RandomLorawanConfirmedDownlink(modulation ...lorawan.Modulation) *DownlinkMessage {
	return &DownlinkMessage{
		GatewayConfiguration:  gateway.RandomTxConfiguration(),
		ProtocolConfiguration: protocol.RandomLorawanTxConfiguration(modulation...),
		Payload:               lorawan.RandomPayload(lorawan.MType_CONFIRMED_DOWN),
	}
}

// RandomLorawanUnconfirmedDownlink returns randomly generated unconfirmed lorawan uplink message.
// Used for testing.
func RandomLorawanUnconfirmedDownlink(modulation ...lorawan.Modulation) *DownlinkMessage {
	return &DownlinkMessage{
		GatewayConfiguration:  gateway.RandomTxConfiguration(),
		ProtocolConfiguration: protocol.RandomLorawanTxConfiguration(modulation...),
		Payload:               lorawan.RandomPayload(lorawan.MType_UNCONFIRMED_DOWN),
	}
}

// RandomLorawanDownlinkMessage returns randomly generated lorawan downlink message(join request, confirmed or unconfirmed downlink).
// Used for testing.
func RandomLorawanDownlinkMessage(modulation ...lorawan.Modulation) *DownlinkMessage {
	switch rand.Intn(3) {
	case 0:
		return RandomLorawanJoinAccept(modulation...)
	case 1:
		return RandomLorawanConfirmedDownlink(modulation...)
	default:
		return RandomLorawanUnconfirmedDownlink(modulation...)
	}
}

// RandomUplinkMessage returns randomly generated uplink message.
// Used for testing.
func RandomUplinkMessage() *UplinkMessage {
	return RandomLorawanUplinkMessage()
}

// RandomDownlinkMessage returns randomly generated downlink message.
// Used for testing.
func RandomDownlinkMessage() *DownlinkMessage {
	return RandomLorawanDownlinkMessage()
}

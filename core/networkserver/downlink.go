// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

func (n *networkServer) HandleDownlink(message *pb_broker.DownlinkMessage) (*pb_broker.DownlinkMessage, error) {
	// Get Device
	dev, err := n.devices.Get(*message.AppEui, *message.DevEui)
	if err != nil {
		return nil, err
	}

	n.status.downlink.Mark(1)

	dev.StartUpdate()

	if dev.AppID != message.AppId || dev.DevID != message.DevId {
		return nil, errors.NewErrInvalidArgument("Downlink", "AppID and DevID do not match AppEUI and DevEUI")
	}

	// Unmarshal LoRaWAN Payload
	var phyPayload lorawan.PHYPayload
	err = phyPayload.UnmarshalBinary(message.Payload)
	if err != nil {
		return nil, err
	}
	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return nil, errors.NewErrInvalidArgument("Downlink", "does not contain a MAC payload")
	}

	// Set DevAddr
	macPayload.FHDR.DevAddr = lorawan.DevAddr(dev.DevAddr)

	// FIRST set and THEN increment FCntDown
	// TODO: For confirmed downlink, FCntDown should be incremented AFTER ACK
	macPayload.FHDR.FCnt = dev.FCntDown
	dev.FCntDown++
	err = n.devices.Set(dev)
	if err != nil {
		return nil, err
	}

	// Sign MIC
	phyPayload.SetMIC(lorawan.AES128Key(dev.NwkSKey))

	// Update message
	bytes, err := phyPayload.MarshalBinary()
	if err != nil {
		return nil, err
	}
	message.Payload = bytes

	return message, nil
}

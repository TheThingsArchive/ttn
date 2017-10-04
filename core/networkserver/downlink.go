// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	pb_broker "github.com/TheThingsNetwork/api/broker"
	"github.com/TheThingsNetwork/api/trace"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

func (n *networkServer) HandleDownlink(message *pb_broker.DownlinkMessage) (*pb_broker.DownlinkMessage, error) {
	err := message.UnmarshalPayload()
	if err != nil {
		return nil, err
	}
	lorawanDownlinkMAC := message.Message.GetLoRaWAN().GetMACPayload()
	if lorawanDownlinkMAC == nil {
		return nil, errors.NewErrInvalidArgument("Downlink", "does not contain a MAC payload")
	}

	n.status.downlink.Mark(1)

	// Get Device
	dev, err := n.devices.Get(message.AppEUI, message.DevEUI)
	if err != nil {
		return nil, err
	}

	if dev.AppID != message.AppID || dev.DevID != message.DevID {
		return nil, errors.NewErrInvalidArgument("Downlink", "AppID and DevID do not match AppEUI and DevEUI")
	}

	message.Trace = message.Trace.WithEvent(trace.UpdateStateEvent)

	dev.StartUpdate()
	defer func() {
		setErr := n.devices.Set(dev)
		if setErr != nil {
			n.Ctx.WithError(setErr).Error("Could not update device state")
		}
		if err == nil {
			err = setErr
		}
	}()

	if lorawanDownlinkMAC.DevAddr != dev.DevAddr {
		return nil, errors.NewErrInvalidArgument("Downlink", "DevAddr does not match device")
	}

	err = n.handleDownlinkMAC(message, dev)
	if err != nil {
		return nil, err
	}

	lorawanDownlinkMAC.FCnt = dev.FCntDown // Use full 32-bit FCnt for setting MIC
	dev.FCntDown++                         // TODO: For confirmed downlink, FCntDown should be incremented AFTER ACK

	phyPayload := message.Message.GetLoRaWAN().PHYPayload()
	phyPayload.SetMIC(lorawan.AES128Key(dev.NwkSKey))
	bytes, err := phyPayload.MarshalBinary()
	if err != nil {
		return nil, err
	}
	message.Payload = bytes

	return message, nil
}

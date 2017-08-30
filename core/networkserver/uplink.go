// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"time"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	"github.com/TheThingsNetwork/api/trace"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

func (n *networkServer) HandleUplink(message *pb_broker.DeduplicatedUplinkMessage) (*pb_broker.DeduplicatedUplinkMessage, error) {
	err := message.UnmarshalPayload()
	if err != nil {
		return nil, err
	}
	lorawanUplinkMAC := message.Message.GetLoRaWAN().GetMACPayload()
	if lorawanUplinkMAC == nil {
		return nil, errors.NewErrInvalidArgument("Uplink", "does not contain a MAC payload")
	}

	n.status.uplink.Mark(1)

	// Get Device
	dev, err := n.devices.Get(*message.AppEUI, *message.DevEUI)
	if err != nil {
		return nil, err
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

	dev.FCntUp = lorawanUplinkMAC.FCnt
	dev.LastSeen = time.Now()

	// Prepare Downlink
	message.InitResponseTemplate()
	lorawanDownlinkMsg := message.ResponseTemplate.Message.InitLoRaWAN()
	lorawanDownlinkMAC := lorawanDownlinkMsg.InitDownlink()
	lorawanDownlinkMAC.FPort = lorawanUplinkMAC.FPort
	lorawanDownlinkMAC.DevAddr = lorawanUplinkMAC.DevAddr
	lorawanDownlinkMAC.FCnt = dev.FCntDown
	if lorawan := message.ResponseTemplate.GetDownlinkOption().GetProtocolConfiguration().GetLoRaWAN(); lorawan != nil {
		lorawan.FCnt = dev.FCntDown
	}

	err = n.handleUplinkMAC(message, dev)
	if err != nil {
		return nil, err
	}

	message.ResponseTemplate.Payload, err = lorawanDownlinkMsg.PHYPayload().MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Unset response if no downlink option
	if message.ResponseTemplate.DownlinkOption == nil {
		message.ResponseTemplate = nil
	}

	return message, nil
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"fmt"

	"github.com/TheThingsNetwork/go-utils/log"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/brocaar/lorawan"
)

func (n *networkServer) handleUplinkMAC(message *pb_broker.DeduplicatedUplinkMessage, dev *device.Device) error {
	lorawanUplinkMsg := message.GetMessage().GetLorawan()
	lorawanUplinkMac := lorawanUplinkMsg.GetMacPayload()
	lorawanDownlinkMsg := message.GetResponseTemplate().GetMessage().GetLorawan()
	lorawanDownlinkMac := lorawanDownlinkMsg.GetMacPayload()

	ctx := n.Ctx.WithFields(log.Fields{
		"AppEUI": dev.AppEUI,
		"DevEUI": dev.DevEUI,
		"AppID":  dev.AppID,
		"DevID":  dev.DevID,
	})

	// Confirmed Uplink
	if lorawanUplinkMsg.IsConfirmed() {
		message.Trace = message.Trace.WithEvent("set ack")
		lorawanDownlinkMac.Ack = true
	}

	// Adaptive DataRate
	if err := n.handleUplinkADR(message, dev); err != nil {
		return err
	}

	// MAC Commands
	for _, cmd := range lorawanUplinkMac.FOpts {
		switch cmd.Cid {
		case uint32(lorawan.LinkCheckReq):
			response := &lorawan.LinkCheckAnsPayload{
				Margin: uint8(linkMargin(message.GetProtocolMetadata().GetLorawan().DataRate, bestSNR(message.GetGatewayMetadata()))),
				GwCnt:  uint8(len(message.GatewayMetadata)),
			}
			responsePayload, _ := response.MarshalBinary()
			lorawanDownlinkMac.FOpts = append(lorawanDownlinkMac.FOpts, pb_lorawan.MACCommand{
				Cid:     uint32(lorawan.LinkCheckAns),
				Payload: responsePayload,
			})
			message.Trace = message.Trace.WithEvent(trace.HandleMACEvent, macCMD, "link-check")
		case uint32(lorawan.LinkADRAns):
			var answer lorawan.LinkADRAnsPayload
			if err := answer.UnmarshalBinary(cmd.Payload); err != nil {
				break
			}
			message.Trace = message.Trace.WithEvent(trace.HandleMACEvent, macCMD, "link-adr",
				"data-rate-ack", answer.DataRateACK,
				"power-ack", answer.PowerACK,
				"channel-mask-ack", answer.ChannelMaskACK,
			)
			if answer.DataRateACK && answer.PowerACK && answer.ChannelMaskACK {
				dev.ADR.Failed = 0
				dev.ADR.SendReq = false
			} else {
				dev.ADR.Failed++
				ctx.
					WithField("Answer", fmt.Sprintf("%v/%v/%v", answer.DataRateACK, answer.PowerACK, answer.ChannelMaskACK)).
					Warn("Negative LinkADRAns")
			}
		default:
		}
	}

	// We can't send MAC on port 0; send them on port 1
	if len(lorawanDownlinkMac.FOpts) != 0 && lorawanDownlinkMac.FPort == 0 {
		lorawanDownlinkMac.FPort = 1
	}

	return nil
}

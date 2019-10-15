// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/api/trace"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/band"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/brocaar/lorawan"
	"github.com/spf13/viper"
)

func (n *networkServer) handleUplinkMAC(message *pb_broker.DeduplicatedUplinkMessage, dev *device.Device) error {
	lorawanUplinkMsg := message.GetMessage().GetLoRaWAN()
	lorawanUplinkMAC := lorawanUplinkMsg.GetMACPayload()
	lorawanDownlinkMsg := message.GetResponseTemplate().GetMessage().GetLoRaWAN()
	lorawanDownlinkMAC := lorawanDownlinkMsg.GetMACPayload()

	ctx := n.Ctx.WithFields(log.Fields{
		"AppEUI":  dev.AppEUI,
		"DevEUI":  dev.DevEUI,
		"AppID":   dev.AppID,
		"DevID":   dev.DevID,
		"DevAddr": dev.DevAddr,
	})

	// Confirmed Uplink
	if lorawanUplinkMsg.IsConfirmed() {
		message.Trace = message.Trace.WithEvent("set ack")
		lorawanDownlinkMAC.Ack = true
	}

	md := message.GetProtocolMetadata()

	// MAC Commands
	for _, cmd := range lorawanUplinkMAC.FOpts {
		switch cmd.CID {
		case uint32(lorawan.LinkCheckReq):
			response := &lorawan.LinkCheckAnsPayload{
				Margin: uint8(linkMargin(md.GetLoRaWAN().DataRate, bestSNR(message.GetGatewayMetadata()))),
				GwCnt:  uint8(len(message.GatewayMetadata)),
			}
			responsePayload, _ := response.MarshalBinary()
			lorawanDownlinkMAC.FOpts = append(lorawanDownlinkMAC.FOpts, pb_lorawan.MACCommand{
				CID:     uint32(lorawan.LinkCheckAns),
				Payload: responsePayload,
			})
			message.Trace = message.Trace.WithEvent(trace.HandleMACEvent, macCMD, "link-check")
		case uint32(lorawan.LinkADRAns):
			var answer lorawan.LinkADRAnsPayload
			if err := answer.UnmarshalBinary(cmd.Payload); err != nil {
				break
			}
			dev.ADR.ExpectRes = false
			message.Trace = message.Trace.WithEvent(trace.HandleMACEvent, macCMD, "link-adr",
				"data-rate-ack", answer.DataRateACK,
				"power-ack", answer.PowerACK,
				"channel-mask-ack", answer.ChannelMaskACK,
			)
			if answer.DataRateACK && answer.PowerACK && answer.ChannelMaskACK {
				dev.ADR.Failed = 0
				dev.ADR.SendReq = false
				dev.ADR.ConfirmedInitial = true
				ctx.WithFields(log.Fields{
					"DataRate": dev.ADR.DataRate,
					"TxPower":  dev.ADR.TxPower,
					"NbTrans":  dev.ADR.NbTrans,
				}).Debug("Positive LinkADRAns")
			} else {
				dev.ADR.Failed++
				ctx.WithFields(log.Fields{
					"DataRate":       dev.ADR.DataRate,
					"TxPower":        dev.ADR.TxPower,
					"NbTrans":        dev.ADR.NbTrans,
					"DataRateACK":    answer.DataRateACK,
					"PowerACK":       answer.PowerACK,
					"ChannelMaskACK": answer.ChannelMaskACK,
					"FailedReqs":     dev.ADR.Failed,
				}).Warn("Negative LinkADRAns")
			}
		default:
		}
	}

	if dev.ADR.ExpectRes {
		ctx.Warn("Expected LinkADRAns but did not receive any")
		if md.GetLoRaWAN().DataRate == dev.ADR.DataRate {
			dev.ADR.ExpectRes = false
			ctx.WithFields(log.Fields{
				"DataRate": dev.ADR.DataRate,
			}).Debug("DataRate is as expected, assuming LinkADRReq succeeded")
		}
	}

	// We did not receive an ADR response, the device may have the wrong RX2 settings
	if dev.ADR.ExpectRes && dev.ADR.Band == "EU_863_870" && viper.GetInt("eu-rx2-dr") != 0 && dev.ActivatedAt.IsZero() {
		settings := message.GetResponseTemplate().GetDownlinkOption()
		if settings.GetGatewayConfiguration().Frequency == 869525000 {
			if loraSettings := settings.ProtocolConfiguration.GetLoRaWAN(); loraSettings != nil {
				loraSettings.DataRate = "SF12BW125"

				band, _ := band.Get("EU_863_870")
				payload := lorawan.RX2SetupReqPayload{
					Frequency: uint32(band.RX2Frequency),
					DLSettings: lorawan.DLSettings{
						RX2DataRate: uint8(band.RX2DataRate),
					},
				}
				responsePayload, _ := payload.MarshalBinary()
				lorawanDownlinkMAC.FOpts = append(lorawanDownlinkMAC.FOpts, pb_lorawan.MACCommand{
					CID:     uint32(lorawan.RXParamSetupReq),
					Payload: responsePayload,
				})
			}
		}
	}

	// Adaptive DataRate
	if err := n.handleUplinkADR(message, dev); err != nil {
		return err
	}

	// We can't send MAC on port 0; send them on port 1
	if len(lorawanDownlinkMAC.FOpts) != 0 && lorawanDownlinkMAC.FPort == 0 {
		lorawanDownlinkMAC.FPort = 1
	}

	return nil
}

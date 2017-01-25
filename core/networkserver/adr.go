// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/band"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/brocaar/lorawan"
)

// DefaultADRMargin is the default SNR margin for ADR
var DefaultADRMargin = 15

func maxSNR(frames []*device.Frame) float32 {
	if len(frames) == 0 {
		return 0
	}
	max := frames[0].SNR
	for _, frame := range frames {
		if frame.SNR > max {
			max = frame.SNR
		}
	}
	return max
}

func (n *networkServer) handleUplinkADR(message *pb_broker.DeduplicatedUplinkMessage, dev *device.Device) error {
	lorawanUplinkMac := message.GetMessage().GetLorawan().GetMacPayload()
	lorawanDownlinkMac := message.GetResponseTemplate().GetMessage().GetLorawan().GetMacPayload()

	if lorawanUplinkMac.Adr {
		if err := n.devices.PushFrame(dev.AppEUI, dev.DevEUI, &device.Frame{
			FCnt:         lorawanUplinkMac.FCnt,
			SNR:          bestSNR(message.GetGatewayMetadata()),
			GatewayCount: uint32(len(message.GatewayMetadata)),
		}); err != nil {
			n.Ctx.WithError(err).Error("Could not push frame for device")
		}
		if dev.ADR.Band == "" {
			dev.ADR.Band = message.GetProtocolMetadata().GetLorawan().GetRegion().String()
		}
		dev.ADR.DataRate = message.GetProtocolMetadata().GetLorawan().GetDataRate()
		if lorawanUplinkMac.AdrAckReq {
			dev.ADR.SendReq = true
			lorawanDownlinkMac.Ack = true
		} else {
			dev.ADR.SendReq = false
		}
	} else {
		// Clear history and reset settings
		n.devices.ClearFrames(dev.AppEUI, dev.DevEUI)
		dev.ADR.SendReq = false
		dev.ADR.DataRate = ""
		dev.ADR.TxPower = 0
		dev.ADR.NbTrans = 0
	}

	return nil
}

func (n *networkServer) handleDownlinkADR(message *pb_broker.DownlinkMessage, dev *device.Device) error {
	if !dev.ADR.SendReq {
		return nil
	}

	frames, err := n.devices.GetFrames(dev.AppEUI, dev.DevEUI)
	if err != nil {
		return err
	}
	if len(frames) >= device.FramesHistorySize {
		// Check settings
		if dev.ADR.DataRate == "" {
			return nil
		}
		if dev.ADR.Margin == 0 {
			dev.ADR.Margin = DefaultADRMargin
		}
		if dev.ADR.Band == "" {
			return nil
		}
		fp, err := band.Get(dev.ADR.Band)
		if err != nil {
			return err
		}
		if dev.ADR.TxPower == 0 {
			dev.ADR.TxPower = fp.DefaultTXPower
		}

		// Calculate ADR settings
		dataRate, txPower, err := fp.ADRSettings(dev.ADR.DataRate, dev.ADR.TxPower, maxSNR(frames), float32(dev.ADR.Margin))
		if err == band.ErrADRUnavailable {
			return nil
		}
		if err != nil {
			return err
		}
		drIdx, err := fp.GetDataRateIndexFor(dataRate)
		if err != nil {
			return err
		}
		powerIdx, err := fp.GetTxPowerIndexFor(txPower)
		if err != nil {
			powerIdx, _ = fp.GetTxPowerIndexFor(fp.DefaultTXPower)
		}

		// Set MAC command
		lorawanDownlinkMac := message.GetMessage().GetLorawan().GetMacPayload()
		response := &lorawan.LinkADRReqPayload{
			DataRate: uint8(drIdx),
			TXPower:  uint8(powerIdx),
			Redundancy: lorawan.Redundancy{
				ChMaskCntl: 0, // Different for US/AU
				NbRep:      1,
			},
		}
		for i, ch := range fp.UplinkChannels { // Different for US/AU
			for _, dr := range ch.DataRates {
				if dr == drIdx {
					response.ChMask[i] = true
				}
			}
		}
		responsePayload, _ := response.MarshalBinary()
		lorawanDownlinkMac.FOpts = append(lorawanDownlinkMac.FOpts, pb_lorawan.MACCommand{
			Cid:     uint32(lorawan.LinkADRReq),
			Payload: responsePayload,
		})
	}

	return nil
}

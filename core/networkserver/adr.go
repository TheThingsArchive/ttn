// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"math"
	"sort"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/band"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/utils/errors"
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

func lossPercentage(frames []*device.Frame) int {
	if len(frames) == 0 {
		return 0
	}
	sentPackets := frames[0].FCnt - frames[len(frames)-1].FCnt + 1
	loss := sentPackets - uint32(len(frames))
	return int(math.Floor((float64(loss) / float64(sentPackets) * 100) + .5))
}

const ScheduleMACEvent = "schedule mac command"

func (n *networkServer) handleUplinkADR(message *pb_broker.DeduplicatedUplinkMessage, dev *device.Device) error {
	lorawanUplinkMAC := message.GetMessage().GetLoRaWAN().GetMACPayload()
	lorawanDownlinkMAC := message.GetResponseTemplate().GetMessage().GetLoRaWAN().GetMACPayload()

	history, err := n.devices.Frames(dev.AppEUI, dev.DevEUI)
	if err != nil {
		return err
	}

	if !lorawanUplinkMAC.ADR {
		// Clear history and reset settings
		if err := history.Clear(); err != nil {
			return err
		}
		dev.ADR.SentInitial = false
		dev.ADR.SendReq = false
		dev.ADR.DataRate = ""
		dev.ADR.TxPower = 0
		dev.ADR.NbTrans = 0
		return nil
	}

	if err := history.Push(&device.Frame{
		FCnt:         lorawanUplinkMAC.FCnt,
		SNR:          bestSNR(message.GetGatewayMetadata()),
		GatewayCount: uint32(len(message.GatewayMetadata)),
	}); err != nil {
		n.Ctx.WithError(err).Error("Could not push frame for device")
	}

	frames, _ := history.Get()

	if dev.ADR.Failed <= maxADRFails && len(frames) >= device.FramesHistorySize {
		lorawanDownlinkMAC.ADR = true
	}

	md := message.GetProtocolMetadata()
	if dev.ADR.Band == "" {
		dev.ADR.Band = md.GetLoRaWAN().GetFrequencyPlan().String()
	}

	var scheduleADR, forceADR bool

	switch dev.ADR.Band {
	case pb_lorawan.FrequencyPlan_US_902_928.String(), pb_lorawan.FrequencyPlan_AU_915_928.String():
		if !dev.ADR.SentInitial {
			scheduleADR = true
			forceADR = true
			message.Trace = message.Trace.WithEvent(ScheduleMACEvent, macCMD, "link-adr", "reason", "initial")
		}
	}

	dataRate := md.GetLoRaWAN().GetDataRate()
	if dev.ADR.DataRate != dataRate {
		dev.ADR.DataRate = dataRate
		scheduleADR = true
		message.Trace = message.Trace.WithEvent(ScheduleMACEvent, macCMD, "link-adr", "reason", "optimize")
	}

	if lorawanUplinkMAC.ADRAckReq {
		scheduleADR = true
		lorawanDownlinkMAC.Ack = true
		message.Trace = message.Trace.WithEvent("set ack", "reason", "adr-ack-req")
	}

	if scheduleADR {
		if fp, err := band.Get(dev.ADR.Band); err == nil {
			if drIdx, err := fp.GetDataRateIndexFor(dataRate); err == nil && drIdx == 0 {
				if len(frames) >= device.FramesHistorySize {
					forceADR = true
					message.Trace = message.Trace.WithEvent(ScheduleMACEvent, macCMD, "link-adr", "reason", "avoid high sf")
				}
			}
		}
	}

	dev.ADR.SendReq = scheduleADR
	if forceADR {
		err := n.setADR(lorawanDownlinkMAC, dev)
		if err != nil {
			message.Trace = message.Trace.WithEvent("mac error", macCMD, "link-adr", "error", err.Error())
			err = nil
		}
	}

	return nil
}

const maxADRFails = 3

func (n *networkServer) setADR(mac *pb_lorawan.MACPayload, dev *device.Device) error {
	if !dev.ADR.SendReq {
		return nil
	}
	if dev.ADR.Failed > maxADRFails {
		dev.ADR.ExpectRes = false // stop trying
		dev.ADR.SendReq = false
		return errors.New("too many failed ADR requests")
	}

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
	if dev.ADR.NbTrans == 0 {
		dev.ADR.NbTrans = 1
	}

	// Get history
	history, err := n.devices.Frames(dev.AppEUI, dev.DevEUI)
	if err != nil {
		return err
	}
	frames, err := history.Get()
	if err != nil {
		return err
	}
	switch {
	case len(frames) == 0:
		return nil
	case len(frames) >= device.FramesHistorySize:
		frames = frames[:device.FramesHistorySize]
	}

	// Add extra margin if we don't have enough data yet
	adrMargin := float32(dev.ADR.Margin)
	if len(frames) < device.FramesHistorySize {
		adrMargin += 2.5
	}

	// Calculate desired ADR settings
	dataRate, txPower, err := fp.ADRSettings(dev.ADR.DataRate, dev.ADR.TxPower, maxSNR(frames), adrMargin)
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
	var nbTrans = dev.ADR.NbTrans
	if dev.ADR.DataRate == dataRate && dev.ADR.TxPower == txPower && !dev.Options.DisableFCntCheck {
		lossPercentage := lossPercentage(frames)
		switch {
		case lossPercentage <= 5:
			nbTrans--
		case lossPercentage <= 10:
			// don't change
		case lossPercentage <= 30:
			nbTrans++
		default:
			nbTrans += 2
		}
		if nbTrans < 1 {
			nbTrans = 1
		}
		if nbTrans > 3 {
			nbTrans = 3
		}
	}

	if dev.ADR.SentInitial && dev.ADR.DataRate == dataRate && dev.ADR.TxPower == txPower && dev.ADR.NbTrans == nbTrans {
		return nil // Nothing to do
	}
	dev.ADR.DataRate, dev.ADR.TxPower, dev.ADR.NbTrans = dataRate, txPower, nbTrans

	payloads := getAdrReqPayloads(dev, &fp, drIdx, powerIdx)
	if len(payloads) == 0 {
		return nil
	}

	n.Ctx.WithFields(log.Fields{
		"AppEUI":  dev.AppEUI,
		"DevEUI":  dev.DevEUI,
		"DevAddr": dev.DevAddr,
		"AppID":   dev.AppID,
		"DevID":   dev.DevID,
	}).Debug("Sending ADR")

	dev.ADR.SentInitial = true
	dev.ADR.ExpectRes = true

	mac.ADR = true

	fOpts := make([]pb_lorawan.MACCommand, 0, len(mac.FOpts)+len(payloads))
	for _, existing := range mac.FOpts {
		if existing.CID != uint32(lorawan.LinkADRReq) {
			fOpts = append(fOpts, existing)
		}
	}
	for _, payload := range payloads {
		responsePayload, _ := payload.MarshalBinary()
		fOpts = append(fOpts, pb_lorawan.MACCommand{
			CID:     uint32(lorawan.LinkADRReq),
			Payload: responsePayload,
		})
	}
	mac.FOpts = fOpts

	return nil
}

func (n *networkServer) handleDownlinkADR(message *pb_broker.DownlinkMessage, dev *device.Device) error {
	err := n.setADR(message.GetMessage().GetLoRaWAN().GetMACPayload(), dev)
	if err != nil {
		message.Trace = message.Trace.WithEvent("mac error", macCMD, "link-adr", "error", err.Error())
		err = nil
	}

	return nil
}

func getAdrReqPayloads(dev *device.Device, frequencyPlan *band.FrequencyPlan, drIdx int, powerIdx int) []lorawan.LinkADRReqPayload {
	payloads := []lorawan.LinkADRReqPayload{}
	switch dev.ADR.Band {
	case pb_lorawan.FrequencyPlan_EU_863_870.String():
		if dev.ADR.Failed > 0 && powerIdx > 5 {
			// fall back to txPower 5 for LoRaWAN 1.0
			powerIdx = 5
		}
		payloads = []lorawan.LinkADRReqPayload{
			{
				DataRate: uint8(drIdx),
				TXPower:  uint8(powerIdx),
				Redundancy: lorawan.Redundancy{
					ChMaskCntl: 0,
					NbRep:      uint8(dev.ADR.NbTrans),
				},
			},
		}
		if dev.ADR.Failed > 0 {
			// Fall back to the mandatory LoRaWAN channels
			payloads[0].ChMask[0] = true
			payloads[0].ChMask[1] = true
			payloads[0].ChMask[2] = true
		} else {
			for i, ch := range frequencyPlan.UplinkChannels {
				for _, dr := range ch.DataRates {
					if dr == drIdx {
						payloads[0].ChMask[i] = true
					}
				}
			}
		}
	case pb_lorawan.FrequencyPlan_US_902_928.String(), pb_lorawan.FrequencyPlan_AU_915_928.String():
		var dr500 uint8
		switch dev.ADR.Band {
		case pb_lorawan.FrequencyPlan_US_902_928.String():
			dr500 = 4
		case pb_lorawan.FrequencyPlan_AU_915_928.String():
			dr500 = 6
		default:
			panic("could not determine 500kHz channel data rate index")
		}

		// Adapted from https://github.com/brocaar/lorawan/blob/master/band/band_us902_928.go
		payloads = []lorawan.LinkADRReqPayload{
			{
				DataRate: dr500, // fixed settings for 500kHz channel
				TXPower:  0,     // fixed settings for 500kHz channel
				Redundancy: lorawan.Redundancy{
					ChMaskCntl: 7,
					NbRep:      uint8(dev.ADR.NbTrans),
				},
			}, // All 125 kHz OFF ChMask applies to channels 64 to 71
		}
		channels := frequencyPlan.GetEnabledUplinkChannels()
		sort.Ints(channels)

		chMaskCntl := -1
		for _, c := range channels {
			// use the ChMask of the first LinkADRReqPayload, besides
			// turning off all 125 kHz this payload contains the ChMask
			// for the last block of channels.
			if c >= 64 {
				payloads[0].ChMask[c%16] = true
				continue
			}

			if c/16 != chMaskCntl {
				chMaskCntl = c / 16
				pl := lorawan.LinkADRReqPayload{
					DataRate: uint8(drIdx),
					TXPower:  uint8(powerIdx),
					Redundancy: lorawan.Redundancy{
						ChMaskCntl: uint8(chMaskCntl),
						NbRep:      uint8(dev.ADR.NbTrans),
					},
				}

				// set the channel mask for this block
				for _, ec := range channels {
					if ec >= chMaskCntl*16 && ec < (chMaskCntl+1)*16 {
						pl.ChMask[ec%16] = true
					}
				}
				payloads = append(payloads, pl)
			}
		}
	}
	return payloads
}

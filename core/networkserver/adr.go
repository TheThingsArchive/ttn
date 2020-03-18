// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"sort"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	"github.com/TheThingsNetwork/api/logfields"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/band"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
	"github.com/spf13/viper"
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

const ScheduleMACEvent = "schedule mac command"

func (n *networkServer) handleUplinkADR(message *pb_broker.DeduplicatedUplinkMessage, dev *device.Device) error {
	ctx := n.Ctx.WithFields(logfields.ForMessage(message))
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
		dev.ADR.ConfirmedInitial = false
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
		ctx.WithError(err).Error("Could not push frame for device")
	}

	md := message.GetProtocolMetadata()
	if dev.ADR.Band == "" {
		dev.ADR.Band = md.GetLoRaWAN().GetFrequencyPlan().String()
	}
	if dev.ADR.Margin == 0 {
		dev.ADR.Margin = DefaultADRMargin
	}

	fp, err := band.Get(dev.ADR.Band)
	if err != nil {
		return err
	}
	dev.ADR.DataRate = md.GetLoRaWAN().GetDataRate()
	if dev.ADR.TxPower == 0 {
		dev.ADR.TxPower = fp.DefaultTXPower
	}
	if dev.ADR.NbTrans == 0 {
		dev.ADR.NbTrans = 1
	}
	dev.ADR.SendReq = false

	adrMargin := float32(dev.ADR.Margin)
	frames, _ := history.Get()
	if len(frames) >= device.FramesHistorySize {
		frames = frames[:device.FramesHistorySize]
	} else {
		adrMargin += 2.5
	}

	desiredDataRate, desiredTxPower, err := fp.ADRSettings(dev.ADR.DataRate, dev.ADR.TxPower, maxSNR(frames), adrMargin)
	if err == band.ErrADRUnavailable {
		ctx.Debugf("ADR not available in %s", dev.ADR.Band)
		return nil
	}
	if err != nil {
		return err
	}

	var forceADR bool

	if !dev.ADR.ConfirmedInitial && (dev.ADR.Band == pb_lorawan.FrequencyPlan_US_902_928.String() || dev.ADR.Band == pb_lorawan.FrequencyPlan_AU_915_928.String()) {
		dev.ADR.SendReq = true
		forceADR = true
		message.Trace = message.Trace.WithEvent(ScheduleMACEvent, macCMD, "link-adr", "reason", "initial")
		ctx.Debug("Schedule ADR [initial]")
	} else if lorawanUplinkMAC.ADRAckReq {
		dev.ADR.SendReq = true
		forceADR = true
		message.Trace = message.Trace.WithEvent(ScheduleMACEvent, macCMD, "link-adr", "reason", "adr-ack-req")
		lorawanDownlinkMAC.Ack = true
		ctx.Debug("Schedule ADR [adr-ack-req]")
	} else if dev.ADR.DataRate != desiredDataRate || dev.ADR.TxPower != desiredTxPower {
		dev.ADR.SendReq = true
		if drIdx, err := fp.GetDataRateIndexFor(dev.ADR.DataRate); err == nil && drIdx == 0 {
			forceADR = true
		} else {
			forceADR = viper.GetBool("networkserver.force-adr-optimize")
		}
		message.Trace = message.Trace.WithEvent(ScheduleMACEvent, macCMD, "link-adr", "reason", "optimize")
		ctx.Debugf("Schedule ADR [optimize] %s->%s", dev.ADR.DataRate, desiredDataRate)
	}

	if !dev.ADR.SendReq {
		return nil
	}

	dev.ADR.DataRate, dev.ADR.TxPower, dev.ADR.NbTrans = desiredDataRate, desiredTxPower, 1

	if forceADR {
		err := n.setADR(lorawanDownlinkMAC, dev)
		if err != nil {
			message.Trace = message.Trace.WithEvent("mac error", macCMD, "link-adr", "error", err.Error())
			ctx.WithError(err).Warn("Could not set ADR")
			err = nil
		}
	}

	return nil
}

const maxADRFails = 10

func (n *networkServer) setADR(mac *pb_lorawan.MACPayload, dev *device.Device) error {
	if !dev.ADR.SendReq {
		return nil
	}
	if dev.ADR.Failed > maxADRFails {
		dev.ADR.ExpectRes = false // stop trying
		dev.ADR.SendReq = false
		return errors.New("too many failed ADR requests")
	}

	ctx := n.Ctx.WithFields(log.Fields{
		"AppEUI":   dev.AppEUI,
		"DevEUI":   dev.DevEUI,
		"DevAddr":  dev.DevAddr,
		"AppID":    dev.AppID,
		"DevID":    dev.DevID,
		"DataRate": dev.ADR.DataRate,
		"TxPower":  dev.ADR.TxPower,
		"NbTrans":  dev.ADR.NbTrans,
	})

	// Check settings
	if dev.ADR.DataRate == "" {
		ctx.Debug("Empty ADR DataRate")
		return nil
	}

	fp, err := band.Get(dev.ADR.Band)
	if err != nil {
		return err
	}
	drIdx, err := fp.GetDataRateIndexFor(dev.ADR.DataRate)
	if err != nil {
		return err
	}
	powerIdx, err := fp.GetTxPowerIndexFor(dev.ADR.TxPower)
	if err != nil {
		powerIdx, _ = fp.GetTxPowerIndexFor(fp.DefaultTXPower)
	}

	payloads := getAdrReqPayloads(dev, &fp, drIdx, powerIdx)
	if len(payloads) == 0 {
		ctx.Debug("No ADR payloads")
		return nil
	}

	dev.ADR.SentInitial = true
	dev.ADR.ExpectRes = true

	mac.ADR = true

	var hadADR bool
	fOpts := make([]pb_lorawan.MACCommand, 0, len(mac.FOpts)+len(payloads))
	for _, existing := range mac.FOpts {
		if existing.CID == uint32(lorawan.LinkADRReq) {
			hadADR = true
			continue
		}
		fOpts = append(fOpts, existing)
	}
	for _, payload := range payloads {
		responsePayload, _ := payload.MarshalBinary()
		fOpts = append(fOpts, pb_lorawan.MACCommand{
			CID:     uint32(lorawan.LinkADRReq),
			Payload: responsePayload,
		})
	}
	mac.FOpts = fOpts

	if !hadADR {
		ctx.Info("Sending ADR Request in Downlink")
	} else {
		ctx.Debug("Updating ADR Request in Downlink")
	}

	return nil
}

func (n *networkServer) handleDownlinkADR(message *pb_broker.DownlinkMessage, dev *device.Device) error {
	err := n.setADR(message.GetMessage().GetLoRaWAN().GetMACPayload(), dev)
	if err != nil {
		message.Trace = message.Trace.WithEvent("mac error", macCMD, "link-adr", "error", err.Error())
		n.Ctx.WithFields(logfields.ForMessage(message)).WithError(err).Warn("Could not set ADR")
		err = nil
	}

	return nil
}

func getAdrReqPayloads(dev *device.Device, frequencyPlan *band.FrequencyPlan, drIdx int, powerIdx int) []lorawan.LinkADRReqPayload {
	payloads := []lorawan.LinkADRReqPayload{}
	switch dev.ADR.Band {

	// Frequency plans with three mandatory channels:
	case pb_lorawan.FrequencyPlan_EU_863_870.String(),
		pb_lorawan.FrequencyPlan_EU_433.String(),
		pb_lorawan.FrequencyPlan_KR_920_923.String(),
		pb_lorawan.FrequencyPlan_IN_865_867.String():

		if dev.ADR.Band == pb_lorawan.FrequencyPlan_EU_863_870.String() && dev.ADR.Failed > 0 && powerIdx > 5 {
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
					if dr == drIdx && i < 8 { // We can enable up to 8 channels.
						payloads[0].ChMask[i] = true
					}
				}
			}
		}

	// Frequency plans with two default channels:
	case pb_lorawan.FrequencyPlan_AS_923.String(),
		pb_lorawan.FrequencyPlan_AS_920_923.String(),
		pb_lorawan.FrequencyPlan_AS_923_925.String(),
		pb_lorawan.FrequencyPlan_RU_864_870.String():
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
		} else {
			for i, ch := range frequencyPlan.UplinkChannels {
				for _, dr := range ch.DataRates {
					if dr == drIdx && i < 7 { // We can enable up to 7 channels.
						payloads[0].ChMask[i] = true
					}
				}
			}
		}

	// Frequency plans with 8 FSBs:
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

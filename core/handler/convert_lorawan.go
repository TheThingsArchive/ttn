// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/api/trace"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

func (h *handler) ConvertFromLoRaWAN(ctx ttnlog.Interface, ttnUp *pb_broker.DeduplicatedUplinkMessage, appUp *types.UplinkMessage, dev *device.Device) (err error) {
	if err := ttnUp.UnmarshalPayload(); err != nil {
		return err
	}
	if ttnUp.GetMessage().GetLoRaWAN() == nil {
		return errors.NewErrInvalidArgument("Uplink", "does not contain a LoRaWAN payload")
	}

	phyPayload := ttnUp.GetMessage().GetLoRaWAN()
	macPayload := phyPayload.GetMACPayload()
	if macPayload == nil {
		return errors.NewErrInvalidArgument("Uplink", "does not contain a MAC payload")
	}

	ttnUp.Trace = ttnUp.Trace.WithEvent(trace.CheckMICEvent)
	err = phyPayload.ValidateMIC(dev.NwkSKey)
	if err != nil {
		return err
	}

	appUp.HardwareSerial = dev.DevEUI.String()

	appUp.FCnt = macPayload.FCnt
	if dev.FCntUp == appUp.FCnt {
		appUp.IsRetry = true
	}
	dev.FCntUp = appUp.FCnt

	if phyPayload.MType == pb_lorawan.MType_CONFIRMED_UP {
		appUp.Confirmed = true
	}

	appUp.FPort = uint8(macPayload.FPort)
	if macPayload.FPort > 0 {
		if err := phyPayload.DecryptFRMPayload(dev.AppSKey); err != nil {
			return errors.NewErrInternal("Could not decrypt payload")
		}
		appUp.PayloadRaw = macPayload.FRMPayload
	}

	if dev.CurrentDownlink != nil && !appUp.IsRetry {
		// We have a downlink pending
		if dev.CurrentDownlink.Confirmed {
			// If it's confirmed, we can only unset it if we receive an ack.
			if macPayload.Ack {
				// Send event over MQTT
				h.qEvent <- &types.DeviceEvent{
					AppID: appUp.AppID,
					DevID: appUp.DevID,
					Event: types.DownlinkAckEvent,
					Data: types.DownlinkEventData{
						Message: dev.CurrentDownlink,
					},
				}
				dev.CurrentDownlink = nil
			}
		} else {
			// If it's unconfirmed, we can unset it.
			dev.CurrentDownlink = nil
		}
	}

	return nil
}

func (h *handler) ConvertToLoRaWAN(ctx ttnlog.Interface, appDown *types.DownlinkMessage, ttnDown *pb_broker.DownlinkMessage, dev *device.Device) (err error) {
	if err := ttnDown.UnmarshalPayload(); err != nil {
		return err
	}
	if ttnDown.GetMessage().GetLoRaWAN() == nil {
		return errors.NewErrInvalidArgument("Downlink", "does not contain a LoRaWAN payload")
	}

	phyPayload := ttnDown.GetMessage().GetLoRaWAN()
	macPayload := phyPayload.GetMACPayload()
	if macPayload == nil {
		return errors.NewErrInvalidArgument("Downlink", "does not contain a MAC payload")
	}

	// Abort when downlink not needed
	if len(appDown.PayloadRaw) == 0 && !macPayload.Ack && len(macPayload.FOpts) == 0 {
		return ErrNotNeeded
	}

	if appDown.FPort > 0 {
		macPayload.FPort = int32(appDown.FPort)
	}

	if appDown.Confirmed {
		phyPayload.MType = pb_lorawan.MType_CONFIRMED_DOWN
	}

	if queue, err := h.devices.DownlinkQueue(dev.AppID, dev.DevID); err == nil {
		if length, _ := queue.Length(); length > 0 {
			macPayload.FPending = true
		}
	}

	// Set Payload
	if len(appDown.PayloadRaw) > 0 {
		ttnDown.Trace = ttnDown.Trace.WithEvent("set payload")
		macPayload.FRMPayload = appDown.PayloadRaw
		if macPayload.FPort <= 0 {
			macPayload.FPort = 1
		}
		err = phyPayload.EncryptFRMPayload(dev.AppSKey)
		if err != nil {
			return err
		}
	} else {
		ttnDown.Trace = ttnDown.Trace.WithEvent("set empty payload")
		macPayload.FRMPayload = []byte{}
		macPayload.FPort = 0
	}

	// Set MIC
	err = phyPayload.SetMIC(dev.NwkSKey)
	if err != nil {
		return err
	}

	ttnDown.Payload = phyPayload.PHYPayloadBytes()

	return nil
}

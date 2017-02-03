// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/brocaar/lorawan"
)

func (h *handler) ConvertFromLoRaWAN(ctx ttnlog.Interface, ttnUp *pb_broker.DeduplicatedUplinkMessage, appUp *types.UplinkMessage, dev *device.Device) error {
	// Check for LoRaWAN
	if lorawan := ttnUp.ProtocolMetadata.GetLorawan(); lorawan == nil {
		return errors.NewErrInvalidArgument("Uplink", "does not contain LoRaWAN metadata")
	}

	// LoRaWAN: Unmarshal Uplink
	var phyPayload lorawan.PHYPayload
	err := phyPayload.UnmarshalBinary(ttnUp.Payload)
	if err != nil {
		return err
	}
	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return errors.NewErrInvalidArgument("Uplink", "does not contain a MAC payload")
	}

	// LoRaWAN: Validate MIC
	macPayload.FHDR.FCnt = ttnUp.ProtocolMetadata.GetLorawan().FCnt
	ttnUp.Trace = ttnUp.Trace.WithEvent(trace.CheckMICEvent)
	ok, err = phyPayload.ValidateMIC(lorawan.AES128Key(dev.NwkSKey))
	if err != nil {
		return err
	}
	if !ok {
		return errors.NewErrNotFound("device that validates MIC")
	}

	appUp.HardwareSerial = dev.DevEUI.String()
	appUp.FCnt = macPayload.FHDR.FCnt
	ctx = ctx.WithField("FCnt", appUp.FCnt)
	if dev.FCntUp == appUp.FCnt {
		appUp.IsRetry = true
	}
	dev.FCntUp = appUp.FCnt

	// LoRaWAN: Decrypt
	if macPayload.FPort != nil {
		appUp.FPort = *macPayload.FPort
		if *macPayload.FPort != 0 && len(macPayload.FRMPayload) == 1 {
			ctx = ctx.WithField("FCnt", appUp.FPort)
			if err := phyPayload.DecryptFRMPayload(lorawan.AES128Key(dev.AppSKey)); err != nil {
				return errors.NewErrInternal("Could not decrypt payload")
			}
			payload, ok := macPayload.FRMPayload[0].(*lorawan.DataPayload)
			if !ok {
				return errors.NewErrInvalidArgument("Uplink FRMPayload", "must be of type *lorawan.DataPayload")
			}
			appUp.PayloadRaw = payload.Bytes
		}
	}

	if dev.CurrentDownlink != nil && !appUp.IsRetry {
		// We have a downlink pending
		if dev.CurrentDownlink.Confirmed {
			// If it's confirmed, we can only unset it if we receive an ack.
			if macPayload.FHDR.FCtrl.ACK {
				// Send event over MQTT
				h.mqttEvent <- &types.DeviceEvent{
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

func (h *handler) ConvertToLoRaWAN(ctx ttnlog.Interface, appDown *types.DownlinkMessage, ttnDown *pb_broker.DownlinkMessage, dev *device.Device) error {
	// LoRaWAN: Unmarshal Downlink
	var phyPayload lorawan.PHYPayload
	err := phyPayload.UnmarshalBinary(ttnDown.Payload)
	if err != nil {
		return err
	}
	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return errors.NewErrInvalidArgument("Downlink", "does not contain a MAC payload")
	}
	if ttnDown.DownlinkOption != nil && ttnDown.DownlinkOption.ProtocolConfig.GetLorawan() != nil {
		macPayload.FHDR.FCnt = ttnDown.DownlinkOption.ProtocolConfig.GetLorawan().FCnt
	}

	// Abort when downlink not needed
	if len(appDown.PayloadRaw) == 0 && !macPayload.FHDR.FCtrl.ACK && len(macPayload.FHDR.FOpts) == 0 {
		return ErrNotNeeded
	}

	// Set FPort
	if appDown.FPort != 0 {
		macPayload.FPort = &appDown.FPort
	}

	if appDown.Confirmed {
		phyPayload.MHDR.MType = lorawan.ConfirmedDataDown
	}

	// Set Payload
	if len(appDown.PayloadRaw) > 0 {
		ttnDown.Trace = ttnDown.Trace.WithEvent("set payload")
		macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: appDown.PayloadRaw}}
		if macPayload.FPort == nil || *macPayload.FPort == 0 {
			macPayload.FPort = pointer.Uint8(1)
		}
	} else {
		ttnDown.Trace = ttnDown.Trace.WithEvent("set empty payload")
		macPayload.FRMPayload = []lorawan.Payload{}
	}

	// Encrypt
	err = phyPayload.EncryptFRMPayload(lorawan.AES128Key(dev.AppSKey))
	if err != nil {
		return err
	}

	// Set MIC
	err = phyPayload.SetMIC(lorawan.AES128Key(dev.NwkSKey))
	if err != nil {
		return err
	}

	// Marshal
	phyPayloadBytes, err := phyPayload.MarshalBinary()
	if err != nil {
		return err
	}

	ttnDown.Payload = phyPayloadBytes

	return nil
}

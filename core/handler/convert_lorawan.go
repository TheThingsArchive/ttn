// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

func (h *handler) ConvertFromLoRaWAN(ctx log.Interface, ttnUp *pb_broker.DeduplicatedUplinkMessage, appUp *mqtt.UplinkMessage) error {
	// Find Device
	dev, err := h.devices.Get(ttnUp.AppId, ttnUp.DevId)
	if err != nil {
		return err
	}

	// Check for LoRaWAN
	if lorawan := ttnUp.ProtocolMetadata.GetLorawan(); lorawan == nil {
		return core.NewErrInvalidArgument("Activation", "does not contain LoRaWAN metadata")
	}

	// LoRaWAN: Unmarshal Uplink
	var phyPayload lorawan.PHYPayload
	err = phyPayload.UnmarshalBinary(ttnUp.Payload)
	if err != nil {
		return err
	}
	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return core.NewErrInvalidArgument("Uplink", "does not contain a MAC payload")
	}
	macPayload.FHDR.FCnt = ttnUp.ProtocolMetadata.GetLorawan().FCnt
	appUp.FCnt = macPayload.FHDR.FCnt

	// LoRaWAN: Validate MIC
	ok, err = phyPayload.ValidateMIC(lorawan.AES128Key(dev.NwkSKey))
	if err != nil {
		return err
	}
	if !ok {
		return core.NewErrNotFound("device that validates MIC")
	}

	ctx = ctx.WithField("FCnt", appUp.FCnt)

	// LoRaWAN: Decrypt
	if macPayload.FPort != nil && *macPayload.FPort != 0 && len(macPayload.FRMPayload) == 1 {
		appUp.FPort = *macPayload.FPort
		ctx = ctx.WithField("FCnt", appUp.FPort)
		if err := phyPayload.DecryptFRMPayload(lorawan.AES128Key(dev.AppSKey)); err != nil {
			return core.NewErrInternal("Could not decrypt payload")
		}
		if len(macPayload.FRMPayload) == 1 {
			payload, ok := macPayload.FRMPayload[0].(*lorawan.DataPayload)
			if !ok {
				return core.NewErrInvalidArgument("Uplink FRMPayload", "must be of type *lorawan.DataPayload")
			}
			appUp.Payload = payload.Bytes
		}
	} else {
		return core.NewErrInvalidArgument("Uplink MACPayload", "could not get frame payload")
	}

	return nil
}

func (h *handler) ConvertToLoRaWAN(ctx log.Interface, appDown *mqtt.DownlinkMessage, ttnDown *pb_broker.DownlinkMessage) error {
	// Find Device
	dev, err := h.devices.Get(appDown.AppID, appDown.DevID)
	if err != nil {
		return err
	}

	// LoRaWAN: Unmarshal Downlink
	var phyPayload lorawan.PHYPayload
	err = phyPayload.UnmarshalBinary(ttnDown.Payload)
	if err != nil {
		return err
	}
	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return core.NewErrInvalidArgument("Downlink", "does not contain a MAC payload")
	}
	if ttnDown.DownlinkOption != nil {
		macPayload.FHDR.FCnt = ttnDown.DownlinkOption.ProtocolConfig.GetLorawan().FCnt
	}

	// Abort when downlink not needed
	if len(appDown.Payload) == 0 && !macPayload.FHDR.FCtrl.ACK && len(macPayload.FHDR.FOpts) == 0 {
		return ErrNotNeeded
	}

	// Set FPort
	if appDown.FPort != 0 {
		macPayload.FPort = &appDown.FPort
	}

	// Set Payload
	if len(appDown.Payload) > 0 {
		macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: appDown.Payload}}
		if macPayload.FPort == nil || *macPayload.FPort == 0 {
			macPayload.FPort = pointer.Uint8(1)
		}
	} else {
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

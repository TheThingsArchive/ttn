// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"errors"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/otaa"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

func (h *handler) HandleActivation(activation *pb_broker.DeduplicatedDeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	var appEUI types.AppEUI
	if activation.AppEui != nil {
		appEUI = *activation.AppEui
	}
	var devEUI types.DevEUI
	if activation.DevEui != nil {
		devEUI = *activation.DevEui
	}
	ctx := h.Ctx.WithFields(log.Fields{
		"DevEUI": devEUI,
		"AppEUI": appEUI,
		"AppID":  activation.AppId,
	})
	var err error
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle activation")
		}
	}()

	// Find Device
	var dev *device.Device
	dev, err = h.devices.Get(appEUI, devEUI)
	if err != nil {
		return nil, err
	}

	// Check for LoRaWAN
	if lorawan := activation.ActivationMetadata.GetLorawan(); lorawan == nil {
		err = errors.New("ttn/handler: Can not activate non-LoRaWAN device")
		return nil, err
	}

	// Unmarshal LoRaWAN
	var reqPHY lorawan.PHYPayload
	if err = reqPHY.UnmarshalBinary(activation.Payload); err != nil {
		return nil, err
	}
	reqMAC, ok := reqPHY.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		err = errors.New("MACPayload must be a *JoinRequestPayload")
		return nil, err
	}

	// Validate MIC
	if ok, err = reqPHY.ValidateMIC(lorawan.AES128Key(dev.AppKey)); err != nil || !ok {
		err = errors.New("ttn/handler: Invalid MIC")
		return nil, err
	}

	// Validate DevNonce
	var alreadyUsed bool
	for _, usedNonce := range dev.UsedDevNonces {
		if usedNonce == device.DevNonce(reqMAC.DevNonce) {
			alreadyUsed = true
			break
		}
	}
	if alreadyUsed {
		err = errors.New("ttn/handler: DevNonce already used")
		return nil, err
	}

	ctx.Debug("Accepting Join Request")

	// Publish Activation
	h.mqttActivation <- &mqtt.Activation{
		AppEUI: *activation.AppEui,
		DevEUI: *activation.DevEui,
	}

	// Prepare Device Activation Response
	var resPHY lorawan.PHYPayload
	if err = resPHY.UnmarshalBinary(activation.ResponseTemplate.Payload); err != nil {
		return nil, err
	}
	resMAC, ok := resPHY.MACPayload.(*lorawan.DataPayload)
	if !ok {
		err = errors.New("MACPayload must be a *DataPayload")
		return nil, err
	}
	joinAccept := &lorawan.JoinAcceptPayload{}
	if err = joinAccept.UnmarshalBinary(false, resMAC.Bytes); err != nil {
		return nil, err
	}
	resPHY.MACPayload = joinAccept

	// Generate random AppNonce
	var appNonce device.AppNonce
	for {
		// NOTE: As DevNonces are only 2 bytes, we will start rejecting those before we run out of AppNonces.
		// It might just take some time to get one we didn't use yet...
		alreadyUsed = false
		copy(appNonce[:], random.New().Bytes(3))
		for _, usedNonce := range dev.UsedAppNonces {
			if usedNonce == appNonce {
				alreadyUsed = true
				break
			}
		}
		if !alreadyUsed {
			break
		}
	}
	joinAccept.AppNonce = appNonce

	// Calculate session keys
	var appSKey types.AppSKey
	var nwkSKey types.NwkSKey
	appSKey, nwkSKey, err = otaa.CalculateSessionKeys(dev.AppKey, joinAccept.AppNonce, joinAccept.NetID, reqMAC.DevNonce)
	if err != nil {
		return nil, err
	}

	// Update Device
	dev.DevAddr = types.DevAddr(joinAccept.DevAddr)
	dev.AppSKey = appSKey
	dev.NwkSKey = nwkSKey
	dev.UsedAppNonces = append(dev.UsedAppNonces, appNonce)
	dev.UsedDevNonces = append(dev.UsedDevNonces, reqMAC.DevNonce)
	err = h.devices.Set(dev, "dev_addr", "app_key", "app_s_key", "nwk_s_key", "used_app_nonces", "used_dev_nonces") // app_key is only needed when the default app_key is used to activate the device
	if err != nil {
		return nil, err
	}

	if err = resPHY.SetMIC(lorawan.AES128Key(dev.AppKey)); err != nil {
		return nil, err
	}
	if err = resPHY.EncryptJoinAcceptPayload(lorawan.AES128Key(dev.AppKey)); err != nil {
		return nil, err
	}

	var resBytes []byte
	resBytes, err = resPHY.MarshalBinary()
	if err != nil {
		return nil, err
	}

	metadata := activation.ActivationMetadata
	metadata.GetLorawan().NwkSKey = &dev.NwkSKey
	metadata.GetLorawan().DevAddr = &dev.DevAddr
	res := &pb.DeviceActivationResponse{
		Payload:            resBytes,
		DownlinkOption:     activation.ResponseTemplate.DownlinkOption,
		ActivationMetadata: metadata,
		AppId:              dev.AppID,
	}

	return res, nil
}

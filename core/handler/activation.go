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
	"github.com/brocaar/lorawan"
)

func (h *handler) HandleActivation(activation *pb_broker.DeduplicatedDeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	// Find Device
	dev, err := h.devices.Get(*activation.AppEui, *activation.DevEui)
	if err != nil {
		// Find application
		app, err := h.applications.Get(*activation.AppEui)
		if err != nil || app.DefaultAppKey.IsEmpty() {
			return nil, err
		}
		// Use Default AppKey
		dev = &device.Device{
			AppEUI: *activation.AppEui,
			DevEUI: *activation.DevEui,
			AppKey: app.DefaultAppKey,
		}
	}

	// Check for LoRaWAN
	if lorawan := activation.ActivationMetadata.GetLorawan(); lorawan == nil {
		return nil, errors.New("ttn/handler: Can not activate non-LoRaWAN device")
	}

	// Unmarshal LoRaWAN
	var reqPHY lorawan.PHYPayload
	if err := reqPHY.UnmarshalBinary(activation.Payload); err != nil {
		return nil, err
	}
	reqMAC, ok := reqPHY.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		return nil, errors.New("MACPayload must be a *JoinRequestPayload")
	}

	// Prepare Device Activation Response
	var resPHY lorawan.PHYPayload
	if err := resPHY.UnmarshalBinary(activation.ResponseTemplate.Payload); err != nil {
		return nil, err
	}
	resMAC, ok := resPHY.MACPayload.(*lorawan.DataPayload)
	if !ok {
		return nil, errors.New("MACPayload must be a *DataPayload")
	}
	joinAccept := &lorawan.JoinAcceptPayload{}
	if err := joinAccept.UnmarshalBinary(false, resMAC.Bytes); err != nil {
		return nil, err
	}
	resPHY.MACPayload = joinAccept

	// Validate MIC
	if ok, err := reqPHY.ValidateMIC(lorawan.AES128Key(dev.AppKey)); err != nil || !ok {
		return nil, errors.New("ttn/handler: Invalid MIC")
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
		return nil, errors.New("ttn/handler: DevNonce already used")
	}

	// Publish Activation
	h.mqttActivation <- &mqtt.Activation{
		AppEUI: *activation.AppEui,
		DevEUI: *activation.DevEui,
	}

	// Generate random AppNonce
	var appNonce device.AppNonce
	for {
		// NOTE: As DevNonces are only 2 bytes, we will start rejecting those before we run out of AppNonces.
		// It might just take some time to get one we didn't use yet...
		alreadyUsed = false
		copy(appNonce[:], random.Bytes(3))
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
	appSKey, nwkSKey, err := otaa.CalculateSessionKeys(dev.AppKey, joinAccept.AppNonce, joinAccept.NetID, reqMAC.DevNonce)
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

	if err := resPHY.SetMIC(lorawan.AES128Key(dev.AppKey)); err != nil {
		return nil, err
	}
	if err := resPHY.EncryptJoinAcceptPayload(lorawan.AES128Key(dev.AppKey)); err != nil {
		return nil, err
	}

	resBytes, err := resPHY.MarshalBinary()
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
	}

	return res, nil
}

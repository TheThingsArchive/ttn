// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/random"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/fields"
	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/otaa"
	"github.com/brocaar/lorawan"
)

func (h *handler) getActivationMetadata(ctx ttnlog.Interface, activation *pb_broker.DeduplicatedDeviceActivationRequest, device *device.Device) (types.Metadata, error) {
	ttnUp := &pb_broker.DeduplicatedUplinkMessage{
		ProtocolMetadata: activation.ProtocolMetadata,
		GatewayMetadata:  activation.GatewayMetadata,
		ServerTime:       activation.ServerTime,
	}
	appUp := &types.UplinkMessage{}
	err := h.ConvertMetadata(ctx, ttnUp, appUp, device)
	if err != nil {
		return types.Metadata{}, err
	}
	return appUp.Metadata, nil
}

func (h *handler) HandleActivationChallenge(challenge *pb_broker.ActivationChallengeRequest) (*pb_broker.ActivationChallengeResponse, error) {
	// Find Device
	dev, err := h.devices.Get(challenge.AppId, challenge.DevId)
	if err != nil {
		return nil, err
	}

	if dev.AppKey.IsEmpty() {
		err = errors.NewErrNotFound(fmt.Sprintf("AppKey for device %s", challenge.DevId))
		return nil, err
	}

	// Unmarshal LoRaWAN
	var reqPHY lorawan.PHYPayload
	if err = reqPHY.UnmarshalBinary(challenge.Payload); err != nil {
		return nil, err
	}

	// Set MIC
	if err := reqPHY.SetMIC(lorawan.AES128Key(dev.AppKey)); err != nil {
		return nil, errors.NewErrNotFound("Could not set MIC")
	}

	// Marshal
	bytes, err := reqPHY.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &pb_broker.ActivationChallengeResponse{
		Payload: bytes,
	}, nil
}

func (h *handler) HandleActivation(activation *pb_broker.DeduplicatedDeviceActivationRequest) (res *pb.DeviceActivationResponse, err error) {
	appID, devID := activation.AppId, activation.DevId
	ctx := h.Ctx.WithFields(fields.Get(activation))
	start := time.Now()
	defer func() {
		if err != nil {
			h.qEvent <- &types.DeviceEvent{
				AppID: appID,
				DevID: devID,
				Event: types.ActivationErrorEvent,
				Data: types.ActivationEventData{
					AppEUI:         *activation.AppEui,
					DevEUI:         *activation.DevEui,
					ErrorEventData: types.ErrorEventData{Error: err.Error()},
				},
			}
			activation.Trace = activation.Trace.WithEvent(trace.DropEvent, "reason", err)
			ctx.WithError(err).Warn("Could not handle activation")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled activation")
		}
		if activation != nil && h.monitorStream != nil {
			h.monitorStream.Send(activation)
		}
	}()
	h.status.activations.Mark(1)

	activation.Trace = activation.Trace.WithEvent(trace.ReceiveEvent)

	if activation.ResponseTemplate == nil || activation.ResponseTemplate.DownlinkOption == nil {
		return nil, errors.NewErrInvalidArgument("Activation", "No gateways available for downlink")
	}

	// Find Device
	dev, err := h.devices.Get(appID, devID)
	if err != nil {
		return nil, err
	}

	if dev.AppKey.IsEmpty() {
		return nil, errors.NewErrNotFound(fmt.Sprintf("AppKey for device %s", devID))
	}

	// Check for LoRaWAN
	metadata := activation.ActivationMetadata.GetLorawan()
	if metadata == nil {
		return nil, errors.NewErrInvalidArgument("Activation", "does not contain LoRaWAN metadata")
	}
	if metadata.AppEui == nil || metadata.DevEui == nil || metadata.DevAddr == nil {
		return nil, errors.NewErrInvalidArgument("Activation Metadata", "incomplete")
	}
	if *metadata.AppEui != *activation.AppEui || *metadata.DevEui != *activation.DevEui {
		return nil, errors.NewErrInvalidArgument("Activation Metadata", "inconsistent")
	}

	// Unmarshal LoRaWAN
	var reqPHY lorawan.PHYPayload
	if err = reqPHY.UnmarshalBinary(activation.Payload); err != nil {
		return nil, err
	}
	reqMAC, ok := reqPHY.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		return nil, errors.NewErrInvalidArgument("Activation", "does not contain a JoinRequestPayload")
	}
	if types.AppEUI(reqMAC.AppEUI) != *activation.AppEui || types.DevEUI(reqMAC.DevEUI) != *activation.DevEui {
		return nil, errors.NewErrInvalidArgument("Activation Payload", "inconsistent")
	}

	// Validate MIC
	activation.Trace = activation.Trace.WithEvent(trace.CheckMICEvent)
	if ok, err = reqPHY.ValidateMIC(lorawan.AES128Key(dev.AppKey)); err != nil || !ok {
		return nil, errors.NewErrNotFound("device that validates MIC")
	}

	if dev.DevEUI.IsEmpty() {
		activation.Trace = activation.Trace.WithEvent("registering on join")
		dev, err = h.registerDeviceOnJoin(dev, activation)
		if err != nil {
			return nil, err
		}
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
		err = errors.NewErrInvalidArgument("Activation DevNonce", "already used")
		return nil, err
	}

	ctx.Debug("Accepting Join Request")
	activation.Trace = activation.Trace.WithEvent(trace.AcceptEvent)

	// Prepare Device Activation Response
	var resPHY lorawan.PHYPayload
	if err = resPHY.UnmarshalBinary(activation.ResponseTemplate.Payload); err != nil {
		return nil, err
	}
	resMAC, ok := resPHY.MACPayload.(*lorawan.DataPayload)
	if !ok {
		err = errors.NewErrInvalidArgument("Activation ResponseTemplate", "MACPayload must be a *DataPayload")
		return nil, err
	}
	joinAccept := &lorawan.JoinAcceptPayload{}
	if err = joinAccept.UnmarshalBinary(false, resMAC.Bytes); err != nil {
		return nil, err
	}
	resPHY.MACPayload = joinAccept

	// Publish Activation
	mqttMetadata, _ := h.getActivationMetadata(ctx, activation, dev)
	h.qEvent <- &types.DeviceEvent{
		AppID: appID,
		DevID: devID,
		Event: types.ActivationEvent,
		Data: types.ActivationEventData{
			AppEUI:   *activation.AppEui,
			DevEUI:   *activation.DevEui,
			DevAddr:  types.DevAddr(joinAccept.DevAddr),
			Metadata: mqttMetadata,
		},
	}

	// Generate random AppNonce
	var appNonce device.AppNonce
	for {
		// NOTE: As DevNonces are only 2 bytes, we will start rejecting those before we run out of AppNonces.
		// It might just take some time to get one we didn't use yet...
		alreadyUsed = false
		random.FillBytes(appNonce[:])
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
	dev.StartUpdate()
	dev.DevAddr = types.DevAddr(joinAccept.DevAddr)
	dev.AppSKey = appSKey
	dev.NwkSKey = nwkSKey
	dev.UsedAppNonces = append(dev.UsedAppNonces, appNonce)
	dev.UsedDevNonces = append(dev.UsedDevNonces, reqMAC.DevNonce)
	err = h.devices.Set(dev)
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

	metadata.NwkSKey = &dev.NwkSKey
	metadata.DevAddr = &dev.DevAddr
	res = &pb.DeviceActivationResponse{
		Payload:            resBytes,
		DownlinkOption:     activation.ResponseTemplate.DownlinkOption,
		ActivationMetadata: activation.ActivationMetadata,
		Trace:              activation.Trace,
	}

	return res, nil
}

func (h *handler) registerDeviceOnJoin(base *device.Device, activation *pb_broker.DeduplicatedDeviceActivationRequest) (*device.Device, error) {
	clone := base.Clone()
	clone.DevID = strings.ToLower(fmt.Sprintf("%s-%s", base.DevID, activation.DevEui.String()))
	clone.DevEUI = *activation.DevEui
	clone.Description = fmt.Sprintf("Registered on join on %s", time.Now().UTC().Format("02 Jan 06 15:04"))

	app, err := h.applications.Get(base.AppID)
	if err != nil {
		return nil, errors.Wrap(err, "Application not registered to this Handler")
	}

	if app.RegisterOnJoinAccessKey == "" {
		return nil, errors.NewErrInvalidArgument("Application", "Does not have Access Key configured for device registration on join")
	}

	token, err := h.ExchangeAppKeyForToken(app.AppID, app.RegisterOnJoinAccessKey)
	if err != nil {
		return nil, err
	}

	_, err = h.ttnDeviceManager.SetDevice(ttnctx.OutgoingContextWithToken(context.Background(), token), clone.GetLoRaWAN())
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Broker did not set device")
	}

	err = h.devices.Set(clone)
	if err != nil {
		return nil, err
	}

	h.qEvent <- &types.DeviceEvent{
		AppID: clone.AppID,
		DevID: clone.DevID,
		Event: types.CreateEvent,
		Data:  nil, // Don't send potentially sensitive details over MQTT
	}

	return clone, nil
}

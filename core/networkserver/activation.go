// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"fmt"
	"strings"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/brocaar/lorawan"
)

func (n *networkServer) getDevAddr(constraints ...string) (types.DevAddr, error) {
	// Generate random DevAddr bytes
	var devAddr types.DevAddr
	copy(devAddr[:], random.Bytes(4))

	// Get a random prefix that matches the constraints
	prefixes := n.GetPrefixesFor(constraints...)
	if len(prefixes) == 0 {
		return types.DevAddr{}, errors.NewErrNotFound(fmt.Sprintf("DevAddr prefix with constraints %v", constraints))
	}

	// Select a prefix
	prefix := prefixes[random.Intn(len(prefixes))]

	// Apply the prefix
	devAddr = devAddr.WithPrefix(prefix)

	return devAddr, nil
}

func (n *networkServer) HandlePrepareActivation(activation *pb_broker.DeduplicatedDeviceActivationRequest) (*pb_broker.DeduplicatedDeviceActivationRequest, error) {
	if activation.AppEui == nil || activation.DevEui == nil {
		return nil, errors.NewErrInvalidArgument("Activation", "missing AppEUI or DevEUI")
	}
	dev, err := n.devices.Get(*activation.AppEui, *activation.DevEui)
	if err != nil {
		return nil, err
	}
	activation.AppId = dev.AppID
	activation.DevId = dev.DevID

	// Don't take any action if there is no response possible
	if pld := activation.GetResponseTemplate(); pld == nil {
		return activation, nil
	}

	// Get activation constraints (for DevAddr prefix selection)
	activationConstraints := strings.Split(dev.Options.ActivationConstraints, ",")
	if len(activationConstraints) == 1 && activationConstraints[0] == "" {
		activationConstraints = []string{}
	}
	activationConstraints = append(activationConstraints, "otaa")

	// We can only activate LoRaWAN devices
	lorawanMeta := activation.GetActivationMetadata().GetLorawan()
	if lorawanMeta == nil {
		return nil, errors.NewErrInvalidArgument("Activation", "missing LoRaWAN metadata")
	}

	// Allocate a  device address
	activation.Trace = activation.Trace.WithEvent("allocate devaddr")
	devAddr, err := n.getDevAddr(activationConstraints...)
	if err != nil {
		return nil, err
	}

	// Set the DevAddr in the Activation Metadata
	lorawanMeta.DevAddr = &devAddr

	// Build JoinAccept Payload
	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			NetID:      n.netID,
			DLSettings: lorawan.DLSettings{RX2DataRate: uint8(lorawanMeta.Rx2Dr), RX1DROffset: uint8(lorawanMeta.Rx1DrOffset)},
			RXDelay:    uint8(lorawanMeta.RxDelay),
			DevAddr:    lorawan.DevAddr(devAddr),
		},
	}
	if lorawanMeta.CfList != nil {
		var cfList lorawan.CFList
		for i, cfListItem := range lorawanMeta.CfList.Freq {
			cfList[i] = cfListItem
		}
		phy.MACPayload.(*lorawan.JoinAcceptPayload).CFList = &cfList
	}

	// Set the Payload
	phyBytes, err := phy.MarshalBinary()
	if err != nil {
		return nil, err
	}
	activation.ResponseTemplate.Payload = phyBytes

	return activation, nil
}

func (n *networkServer) HandleActivate(activation *pb_handler.DeviceActivationResponse) (*pb_handler.DeviceActivationResponse, error) {
	meta := activation.GetActivationMetadata()
	if meta == nil {
		return nil, errors.NewErrInvalidArgument("Activation", "missing ActivationMetadata")
	}
	lorawan := meta.GetLorawan()
	if lorawan == nil {
		return nil, errors.NewErrInvalidArgument("Activation", "missing LoRaWAN ActivationMetadata")
	}
	n.status.activations.Mark(1)

	dev, err := n.devices.Get(*lorawan.AppEui, *lorawan.DevEui)
	if err != nil {
		return nil, err
	}

	activation.Trace = activation.Trace.WithEvent(trace.UpdateStateEvent)
	dev.StartUpdate()

	dev.LastSeen = time.Now()
	dev.UpdatedAt = time.Now()
	dev.DevAddr = *lorawan.DevAddr
	dev.NwkSKey = *lorawan.NwkSKey
	dev.FCntUp = 0
	dev.FCntDown = 0
	dev.ADR = device.ADRSettings{Band: dev.ADR.Band, Margin: dev.ADR.Margin}

	if band := meta.GetLorawan().GetRegion().String(); band != "" {
		dev.ADR.Band = band
	}

	err = n.devices.Set(dev)
	if err != nil {
		return nil, err
	}

	frames, err := n.devices.Frames(dev.AppEUI, dev.DevEUI)
	if err != nil {
		return nil, err
	}
	err = frames.Clear()
	if err != nil {
		return nil, err
	}

	return activation, nil
}

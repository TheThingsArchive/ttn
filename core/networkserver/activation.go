// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"fmt"
	"strings"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/brocaar/lorawan"
)

func (n *networkServer) getDevAddr(constraints ...string) (types.DevAddr, error) {
	// Instantiate a new random source
	random := random.New()

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

	// Get activation constraints (for DevAddr prefix selection)
	activationConstraints := strings.Split(dev.Options.ActivationConstraints, ",")
	if len(activationConstraints) == 1 && activationConstraints[0] == "" {
		activationConstraints = []string{}
	}
	activationConstraints = append(activationConstraints, "otaa")

	// Build activation metadata if not present
	if meta := activation.GetActivationMetadata(); meta == nil {
		activation.ActivationMetadata = &pb_protocol.ActivationMetadata{}
	}
	// Build lorawan metadata if not present
	if lorawan := activation.ActivationMetadata.GetLorawan(); lorawan == nil {
		return nil, errors.NewErrInvalidArgument("Activation", "missing LoRaWAN metadata")
	}

	// Build response template if not present
	if pld := activation.GetResponseTemplate(); pld == nil {
		return nil, errors.NewErrInvalidArgument("Activation", "missing response template")
	}
	lorawanMeta := activation.ActivationMetadata.GetLorawan()

	// Get a random device address
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
	if len(lorawanMeta.CfList) == 5 {
		var cfList lorawan.CFList
		for i, cfListItem := range lorawanMeta.CfList {
			cfList[i] = uint32(cfListItem)
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
	err := n.devices.Activate(*lorawan.AppEui, *lorawan.DevEui, *lorawan.DevAddr, *lorawan.NwkSKey)
	if err != nil {
		return nil, err
	}
	return activation, nil
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

func (n *networkServer) HandleUplink(message *pb_broker.DeduplicatedUplinkMessage) (*pb_broker.DeduplicatedUplinkMessage, error) {
	// Get Device
	dev, err := n.devices.Get(*message.AppEui, *message.DevEui)
	if err != nil {
		return nil, err
	}

	n.status.uplink.Mark(1)

	message.Trace = message.Trace.WithEvent(trace.UpdateStateEvent)

	dev.StartUpdate()

	// Unmarshal LoRaWAN Payload
	var phyPayload lorawan.PHYPayload
	err = phyPayload.UnmarshalBinary(message.Payload)
	if err != nil {
		return nil, err
	}
	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return nil, errors.NewErrInvalidArgument("Uplink", "does not contain a MAC payload")
	}

	// Update FCntUp (from metadata if possible, because only 16lsb are marshaled in FHDR)
	if lorawan := message.GetProtocolMetadata().GetLorawan(); lorawan != nil && lorawan.FCnt != 0 {
		dev.FCntUp = lorawan.FCnt
	} else {
		dev.FCntUp = macPayload.FHDR.FCnt
	}
	dev.LastSeen = time.Now()
	err = n.devices.Set(dev)
	if err != nil {
		return nil, err
	}

	// Prepare Downlink
	if message.ResponseTemplate == nil {
		return message, nil
	}
	message.ResponseTemplate.AppEui = message.AppEui
	message.ResponseTemplate.DevEui = message.DevEui
	message.ResponseTemplate.AppId = message.AppId
	message.ResponseTemplate.DevId = message.DevId

	// Add Full FCnt (avoiding nil pointer panics)
	if option := message.ResponseTemplate.DownlinkOption; option != nil {
		if protocol := option.ProtocolConfig; protocol != nil {
			if lorawan := protocol.GetLorawan(); lorawan != nil {
				lorawan.FCnt = dev.FCntDown
			}
		}
	}

	mac := &lorawan.MACPayload{
		FHDR: lorawan.FHDR{
			DevAddr: macPayload.FHDR.DevAddr,
			FCnt:    dev.FCntDown,
		},
	}

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.UnconfirmedDataDown,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: mac,
	}

	// Confirmed Uplink
	if phyPayload.MHDR.MType == lorawan.ConfirmedDataUp {
		message.Trace = message.Trace.WithEvent("set ack")
		mac.FHDR.FCtrl.ACK = true
	}

	// Adaptive DataRate
	if macPayload.FHDR.FCtrl.ADR {
		if macPayload.FHDR.FCtrl.ADRACKReq {
			message.Trace = message.Trace.WithEvent("set adr ack")
			mac.FHDR.FCtrl.ACK = true
		}
	}

	// MAC Commands
	for _, cmd := range macPayload.FHDR.FOpts {
		switch cmd.CID {
		case lorawan.LinkCheckReq:
			mac.FHDR.FOpts = append(mac.FHDR.FOpts, lorawan.MACCommand{
				CID: lorawan.LinkCheckAns,
				Payload: &lorawan.LinkCheckAnsPayload{
					Margin: uint8(linkMargin(message.GetProtocolMetadata().GetLorawan().DataRate, bestSNR(message.GetGatewayMetadata()))),
					GwCnt:  uint8(len(message.GatewayMetadata)),
				},
			})
			message.Trace = message.Trace.WithEvent(trace.HandleMACEvent, macCMD, "link-check")
		default:
		}
	}

	phyBytes, err := phy.MarshalBinary()
	if err != nil {
		return nil, err
	}

	message.ResponseTemplate.Payload = phyBytes

	return message, nil
}

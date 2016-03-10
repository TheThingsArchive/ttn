// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/hex"
	"strings"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/adapters/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

// Downlink handles communication between a handler and an application via MQTT
type Downlink struct{}

// Topic implements the mqtt.Handler interface
func (a Downlink) Topic() string {
	return "+/devices/+/down"
}

// Handle implements the mqtt.Handler interface
func (a Downlink) Handle(client Client, chpkt chan<- PktReq, chreg chan<- RegReq, msg MQTT.Message) error {
	infos := strings.Split(msg.Topic(), "/")

	if len(infos) != 4 {
		return errors.New(errors.Structural, "Unexpect (and invalid) mqtt topic")
	}

	appEUIRaw, erra := hex.DecodeString(infos[0])
	devEUIRaw, errd := hex.DecodeString(infos[2])
	if erra != nil || errd != nil || len(appEUIRaw) != 8 || len(devEUIRaw) != 8 {
		return errors.New(errors.Structural, "Topic constituted of invalid AppEUI or DevEUI")
	}

	var appEUI lorawan.EUI64
	copy(appEUI[:], appEUIRaw)
	var devEUI lorawan.EUI64
	copy(devEUI[:], devEUIRaw)

	apacket, err := core.NewAPacket(appEUI, devEUI, msg.Payload(), nil)
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	data, err := apacket.MarshalBinary()
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	chpkt <- PktReq{
		Packet: data,
		Chresp: nil,
	}
	return nil
}

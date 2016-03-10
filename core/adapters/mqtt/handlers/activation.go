// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/hex"
	"fmt"
	"strings"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	. "github.com/TheThingsNetwork/ttn/core/adapters/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

// Activation handles communication between a handler and an application via MQTT
type Activation struct{}

// Topic implements the mqtt.Handler interface
func (a Activation) Topic() string {
	return "+/devices/+/activations"
}

// Handle implements the mqtt.Handler interface
func (a Activation) Handle(client Client, chpkt chan<- PktReq, chreg chan<- RegReq, msg MQTT.Message) error {
	topicInfos := strings.Split(msg.Topic(), "/")

	if len(topicInfos) != 4 {
		return errors.New(errors.Structural, "Invalid given topic")
	}

	appEUIStr := topicInfos[0]
	devEUIStr := topicInfos[2]

	if devEUIStr != "personalized" {
		return errors.New(errors.Implementation, "OTAA not yet supported. Unable to register device")
	}

	payload := msg.Payload()
	if len(payload) != 36 {
		return errors.New(errors.Structural, "Invalid registration payload")
	}

	var appEUI lorawan.EUI64
	var devEUI lorawan.EUI64
	var nwkSKey lorawan.AES128Key
	var appSKey lorawan.AES128Key
	copy(devEUI[4:], msg.Payload()[:4])
	copy(nwkSKey[:], msg.Payload()[4:20])
	copy(appSKey[:], msg.Payload()[20:])

	data, err := hex.DecodeString(appEUIStr)
	if err != nil || len(data) != 8 {
		return errors.New(errors.Structural, "Invalid application EUI")
	}
	copy(appEUI[:], data[:])

	devEUIStr = hex.EncodeToString(devEUI[:])
	topic := fmt.Sprintf("%s/%s/%s/up", appEUIStr, "devices", devEUIStr)

	chreg <- RegReq{
		Registration: activationRegistration{
			recipient: NewRecipient(topic, ""),
			devEUI:    devEUI,
			appEUI:    appEUI,
			nwkSKey:   nwkSKey,
			appSKey:   appSKey,
		},
		Chresp: nil,
	}
	return nil
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/hex"
	"fmt"
	"strings"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	. "github.com/TheThingsNetwork/ttn/refactor/adapters/mqtt"
	"github.com/brocaar/lorawan"
)

type Activation struct{}

func (a Activation) Topic() string {
	return "+/devices/+/activations"
}

func (a Activation) Handle(client Client, chpkt chan<- PktReq, chreg chan<- RegReq, msg MQTT.Message) error {
	topicInfos := strings.Split(msg.Topic(), "/")
	appEUIStr := topicInfos[0]
	devEUIStr := topicInfos[2]

	if devEUIStr != "personalized" {
		// TODO Log warning
		//a.ctx.WithField("Device Address", devEUI).Warn("OTAA not yet supported. Unable to register device")
		return nil
	}

	payload := msg.Payload()
	if len(payload) != 36 {
		// TODO Log warning
		//a.ctx.WithField("Payload", payload).Error("Invalid registration payload")
		return nil
	}

	var appEUI lorawan.EUI64
	var devEUI lorawan.EUI64
	var nwkSKey lorawan.AES128Key
	var appSKey lorawan.AES128Key
	copy(appEUI[:], []byte(appEUIStr))
	copy(devEUI[4:], msg.Payload()[:4])
	copy(nwkSKey[:], msg.Payload()[4:20])
	copy(appSKey[:], msg.Payload()[20:])

	devEUIStr = hex.EncodeToString(devEUI[:])
	topic := fmt.Sprintf("%s/%s/%s/%s", appEUIStr, "devices", devEUIStr, "up")
	token := client.Subscribe(topic, 2, a.handleReception(chpkt))
	if token.Wait() && token.Error() != nil {
		// TODO Log Error
		// a.ctx.WithError(token.Error()).Error("Unable to subscribe")
		return nil
	}

	chreg <- RegReq{
		Registration: activationRegistration{
			recipient: NewRecipient(topic, "DO_NOT_USE_THIS_TOPIC"),
			devEUI:    devEUI,
			appEUI:    appEUI,
			nwkSKey:   nwkSKey,
			appSKey:   appSKey,
		},
		Chresp: nil,
	}
	return nil
}

func (a Activation) handleReception(chpkt chan<- PktReq) func(client Client, msg MQTT.Message) {
	return func(client Client, msg MQTT.Message) {
		// TODO
	}
}

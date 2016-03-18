// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/apex/log"
)

var (
	Client *MQTT.Client
	ctx    log.Interface
)

func Setup(broker string, _ctx log.Interface) {
	if Client != nil {
		_ctx.Fatal("MQTT Client already set up.")
	}
	ctx = _ctx

	mqttOpts := MQTT.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", broker))
	clientID := fmt.Sprintf("ttnctl-%s", util.RandString(16))
	mqttOpts.SetClientID(clientID)

	mqttOpts.SetKeepAlive(20)

	mqttOpts.SetDefaultPublishHandler(func(client *MQTT.Client, msg MQTT.Message) {
		ctx.WithField("message", msg).Debug("Received message")
	})

	mqttOpts.SetConnectionLostHandler(func(client *MQTT.Client, err error) {
		ctx.WithError(err).Warn("Connection Lost. Reconnecting...")
	})

	Client = MQTT.NewClient(mqttOpts)
}

func Connect() {
	if Client.IsConnected() {
		return
	}

	ctx.Infof("Connecting to The Things Network...")
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		ctx.WithError(token.Error()).Fatal("Could not connect.")
	}
}

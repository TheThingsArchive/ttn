// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
	"github.com/spf13/viper"
)

// ConnectMQTTClient connects a new MQTT clients with the specified credentials
func ConnectMQTTClient(ctx log.Interface, appEui []byte, accessKey string) mqtt.Client {
	username := fmt.Sprintf("%X", appEui)
	password := accessKey

	broker := fmt.Sprintf("tcp://%s", viper.GetString("mqtt-broker"))
	client := mqtt.NewClient(ctx, "ttnctl", username, password, broker)

	err := client.Connect()
	if err != nil {
		ctx.WithError(err).Fatal("Could not connect")
	}

	return client
}

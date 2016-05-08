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
func ConnectMQTTClient(ctx log.Interface) mqtt.Client {
	appEUI := GetAppEUI(ctx)

	apps, err := GetApplications(ctx)
	if err != nil {
		ctx.WithError(err).Fatal("Failed to get applications")
	}

	var app *App
	for _, a := range apps {
		if a.EUI == appEUI {
			app = a
		}
	}
	if app == nil {
		ctx.Fatal("Application not found")
	}

	broker := fmt.Sprintf("tcp://%s", viper.GetString("mqtt-broker"))
	// Don't care about which access key here
	client := mqtt.NewClient(ctx, "ttnctl", app.EUI.String(), app.AccessKeys[0], broker)

	if err := client.Connect(); err != nil {
		ctx.WithError(err).Fatal("Could not connect")
	}

	return client
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
	"github.com/gosuri/uitable"
	"github.com/spf13/viper"
)

// GetMQTT connects a new MQTT clients with the specified credentials
func GetMQTT(ctx log.Interface) mqtt.Client {
	username, password, err := getMQTTCredentials(ctx)
	if err != nil {
		ctx.WithError(err).Fatal("Failed to get MQTT credentials")
	}

	mqttProto := "tcp"
	if strings.HasSuffix(viper.GetString("mqtt-broker"), ":8883") {
		mqttProto = "ssl"
		ctx.Fatal("TLS connections are not yet supported by ttnctl")
	}
	broker := fmt.Sprintf("%s://%s", mqttProto, viper.GetString("mqtt-broker"))
	client := mqtt.NewClient(ctx, "ttnctl", username, password, broker)

	ctx.WithFields(log.Fields{
		"MQTT Broker": broker,
		"Username":    username,
	}).Info("Connecting to MQTT...")

	if err := client.Connect(); err != nil {
		ctx.WithError(err).Fatal("Could not connect")
	}

	return client
}

func getMQTTCredentials(ctx log.Interface) (username string, password string, err error) {
	username = viper.GetString("mqtt-username")
	password = viper.GetString("mqtt-password")
	if username != "" {
		return
	}

	// Do not use authentication on local MQTT
	if strings.HasPrefix(viper.GetString("mqtt-broker"), "localhost") {
		return
	}

	return getAppMQTTCredentials(ctx)
}

func getAppMQTTCredentials(ctx log.Interface) (string, string, error) {
	appID := GetAppID(ctx)

	account := GetAccount(ctx)
	app, err := account.FindApplication(appID)
	if err != nil {
		return "", "", err
	}

	var keyIdx int
	switch len(app.AccessKeys) {
	case 0:
		return "", "", errors.New("Can not connect to MQTT. Your application does not have any access keys.")
	case 1:
	default:
		ctx.Infof("Found %d access keys for your application:", len(app.AccessKeys))

		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("", "Name", "Rights")
		for i, key := range app.AccessKeys {
			rightStrings := make([]string, 0, len(key.Rights))
			for _, i := range key.Rights {
				rightStrings = append(rightStrings, string(i))
			}
			table.AddRow(i+1, key.Name, strings.Join(rightStrings, ","))
		}

		fmt.Println()
		fmt.Println(table)
		fmt.Println()

		fmt.Println("Which one do you want to use?")
		fmt.Printf("Enter the number (1 - %d) > ", len(app.AccessKeys))
		fmt.Scanf("%d", &keyIdx)
		keyIdx--
	}

	if keyIdx < 0 || keyIdx >= len(app.AccessKeys) {
		return "", "", errors.New("Invalid choice for access key")
	}
	return appID, app.AccessKeys[keyIdx].Key, nil
}

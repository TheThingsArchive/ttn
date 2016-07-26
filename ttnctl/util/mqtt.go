// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
	"github.com/gosuri/uitable"
	"github.com/spf13/viper"
)

// GetMQTT connects a new MQTT clients with the specified credentials
func GetMQTT(ctx log.Interface) mqtt.Client {
	appID := GetAppID(ctx)

	account := GetAccount(ctx)
	app, err := account.FindApplication(appID)
	if err != nil {
		ctx.WithError(err).Fatal("Failed to get application")
	}

	var keyIdx int
	switch len(app.AccessKeys) {
	case 0:
		ctx.Fatal("Can not connect to MQTT. Your application does not have any access keys.")
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
		ctx.Fatal("Invalid choice for access key")
	}
	key := app.AccessKeys[keyIdx]

	broker := fmt.Sprintf("tcp://%s", viper.GetString("mqtt-broker"))
	client := mqtt.NewClient(ctx, "ttnctl", appID, key.Key, broker)

	if err := client.Connect(); err != nil {
		ctx.WithError(err).Fatal("Could not connect")
	}

	return client
}

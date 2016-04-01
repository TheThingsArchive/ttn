// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
	"github.com/howeyc/gopass"
	"github.com/spf13/viper"
)

func GetMQTTClient(ctx log.Interface) mqtt.Client {
	user, err := LoadAuth(viper.GetString("ttn-account-server"))
	if err != nil {
		ctx.WithError(err).Fatal("Failed to load authentication token")
	}
	if user == nil {
		ctx.Fatal("No login found. Please login with ttnctl user login [e-mail]")
	}

	// NOTE: until the MQTT server supports access tokens, we'll have to ask for a password.
	fmt.Print("Password: ")
	password, err := gopass.GetPasswd()
	if err != nil {
		ctx.Fatal(err.Error())
	}

	broker := fmt.Sprintf("tcp://%s", viper.GetString("mqtt-broker"))
	client := mqtt.NewClient(ctx, "ttnctl", user.Email, string(password), broker)

	err = client.Connect()
	if err != nil {
		ctx.WithError(err).Fatal("Could not connect")
	}

	return client
}

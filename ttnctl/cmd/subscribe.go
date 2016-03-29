// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"regexp"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/api"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var subscribeCmd = &cobra.Command{
	Use:   "subscribe [DevEUI]",
	Short: "Subscribe to uplink messages from a device",
	Long: `ttnctl subscribe prints out uplink messages from a device as they arrive.

The optional DevEUI argument can be used to only receive messages from a
specific device. By default you will receive messages from all devices of your
application.`,
	Run: func(cmd *cobra.Command, args []string) {
		appEUI, err := util.Parse64(viper.GetString("app-eui"))
		if err != nil {
			ctx.Fatalf("Invalid AppEUI: %s", err)
		}

		var devEUI = "+"
		if len(args) > 0 {
			eui, err := util.Parse64(args[0])
			if err != nil {
				ctx.Fatalf("Invalid DevEUI: %s", err)
			}
			devEUI = fmt.Sprintf("%X", eui)
			ctx.Infof("Subscribing uplink messages from device %s", devEUI)
		} else {
			ctx.Infof("Subscribing to uplink messages from all devices in application %x", appEUI)
		}

		t, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication token")
		}
		if t == nil {
			ctx.Fatal("No login found. Please login with ttnctl user login [e-mail]")
		}

		// NOTE: until the MQTT server supports access tokens, we'll have to ask for a password.
		fmt.Print("Password: ")
		password, err := gopass.GetPasswd()
		if err != nil {
			ctx.Fatal(err.Error())
		}

		broker := fmt.Sprintf("tcp://%s", viper.GetString("mqtt-broker"))
		opts := MQTT.NewClientOptions().AddBroker(broker)

		clientID := fmt.Sprintf("ttntool-%s", util.RandString(15))
		opts.SetClientID(clientID)

		opts.SetUsername(t.Email)
		opts.SetPassword(string(password))

		opts.SetKeepAlive(20)

		opts.SetOnConnectHandler(func(client *MQTT.Client) {
			ctx.Info("Connected to The Things Network.")
		})

		opts.SetDefaultPublishHandler(func(client *MQTT.Client, msg MQTT.Message) {
			t, err := api.DecodeDeviceTopic(msg.Topic())
			if err != nil {
				ctx.WithError(err).Warn("There's something wrong with the MQTT topic.")
			}

			ctx := ctx.WithField("DevEUI", t.DevEUI)

			dataUp := &core.DataUpAppReq{}
			_, err = dataUp.UnmarshalMsg(msg.Payload())
			if err != nil {
				ctx.WithError(err).Warn("Could not unmarshal uplink.")
			}

			// TODO: Find out what Metadata people want to see here

			unprintable, _ := regexp.Compile(`[^[:print:]]`)
			if unprintable.Match(dataUp.Payload) {
				ctx.Infof("%X", dataUp.Payload)
			} else {
				ctx.Infof("%s", dataUp.Payload)
				ctx.Warn("Sending data as plain text is bad practice. We recommend to transmit data in a binary format.")
			}

			if l := len(dataUp.Payload); l > 12 {
				ctx.Warnf("Your payload has a size of %d bytes. We recommend to send no more than 12 bytes.", l)
			}

			// TODO: Add warnings for airtime / duty-cycle / fair-use

		})

		opts.SetConnectionLostHandler(func(client *MQTT.Client, err error) {
			ctx.WithError(err).Error("Connection Lost. Reconnecting...")
		})

		mqttClient := MQTT.NewClient(opts)

		ctx.WithField("mqtt-broker", broker).Info("Connecting to The Things Network...")
		if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		topic := fmt.Sprintf("%X/devices/%s/up", appEUI, devEUI)
		ctx.WithField("topic", topic).Debug("Subscribing...")

		if token := mqttClient.Subscribe(topic, 2, nil); token.Wait() && token.Error() != nil {
			ctx.WithField("topic", topic).WithError(token.Error()).Fatal("Could not subscribe.")
		}

		ctx.WithField("topic", topic).Debug("Subscribed.")

		<-make(chan bool)

	},
}

func init() {
	RootCmd.AddCommand(subscribeCmd)
}

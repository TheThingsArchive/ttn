// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/ttnctl/mqtt"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// downlinkCmd represents the `downlink` command
var downlinkCmd = &cobra.Command{
	Use:   "downlink [DevEUI] [Payload] [TTL]",
	Short: "Send downlink messages to the network",
	Long: `ttnctl downlink sends a downlink message to the network

The DevEUI should be an 8-byte long hex-encoded string (16 chars), whereas the TTL is
expected to define a Time To Live in a handy format, for instance: "1h" for one hour.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			ctx.Fatal("Insufficient arguments")
		}

		devEUI, err := util.Parse64(args[0])
		if err != nil {
			ctx.Fatalf("Invalid DevEUI: %s", err)
		}

		req := core.DataDownAppReq{
			Payload: []byte(args[1]),
			TTL:     args[2],
		}

		payload, err := req.MarshalMsg(nil)

		if err != nil {
			ctx.WithError(err).Fatal("Unable to create downlink payload")
		}

		mqtt.Setup(viper.GetString("handler.mqtt-broker"), ctx)
		mqtt.Connect()

		ctx.WithFields(log.Fields{
			"DevEUI":  hex.EncodeToString(devEUI),
			"Payload": string(payload),
		}).Info("Pushing downlink...")

		token := mqtt.Client.Publish(fmt.Sprintf("%s/devices/%x/down", viper.GetString("handler.app-eui"), devEUI), 2, false, payload)
		if token.Wait() && token.Error() != nil {
			ctx.WithError(token.Error()).Fatal("Downlink failed.")
		} else {
			// Although we can't be sure whether it actually succeeded, we can know when the command is published to the MQTT.
			ctx.Info("Downlink sent.")
		}
	},
}

func init() {
	RootCmd.AddCommand(downlinkCmd)

	downlinkCmd.Flags().String("mqtt-broker", "localhost:1883", "The address of the MQTT broker")
	viper.BindPFlag("handler.mqtt-broker", downlinkCmd.Flags().Lookup("mqtt-broker"))

	downlinkCmd.Flags().String("app-eui", "0102030405060708", "The app EUI to use")
	viper.BindPFlag("handler.app-eui", downlinkCmd.Flags().Lookup("app-eui"))
}

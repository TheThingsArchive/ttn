// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"os"
	"os/signal"
	"regexp"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var subscribeCmd = &cobra.Command{
	Use:   "subscribe [DevEUI]",
	Short: "Subscribe to uplink messages from a device",
	Long: `ttnctl subscribe prints out uplink messages from a device as they
arrive.

The optional DevEUI argument can be used to only receive messages from a
specific device. By default you will receive messages from all devices of your
application.`,
	Run: func(cmd *cobra.Command, args []string) {

		appEUI := util.GetAppEUI(ctx)

		var devEUI []byte
		if len(args) > 0 {
			devEUI, err := util.Parse64(args[0])
			if err != nil {
				ctx.Fatalf("Invalid DevEUI: %s", err)
			}
			ctx.Infof("Subscribing uplink messages from device %s", devEUI)
		} else {
			ctx.Infof("Subscribing to uplink messages from all devices in application %x", appEUI)
		}

		client := util.ConnectMQTTClient(ctx)

		token := client.SubscribeDeviceUplink(appEUI, devEUI, func(client mqtt.Client, appEUI []byte, devEUI []byte, dataUp core.DataUpAppReq) {
			ctx := ctx.WithField("DevEUI", devEUI)

			// TODO: Find out what Metadata people want to see here

			// NOTE: This is a race condition; binary values may be printable
			unprintable, _ := regexp.Compile(`[^[:print:]]`)
			if unprintable.Match(dataUp.Payload) {
				ctx.Infof("%X", dataUp.Payload)
			} else {
				ctx.Infof("%s", dataUp.Payload)
				ctx.Warn("Sending data as plain text is bad practice. We recommend to transmit data in a binary format.")
			}

			if l := len(dataUp.Payload); l > 20 {
				ctx.Warnf("Your payload has a size of %d bytes. We recommend to send no more than 20 bytes.", l)
			}

			// TODO: Add warnings for airtime / duty-cycle / fair-use
		})

		if token.Wait(); token.Error() != nil {
			ctx.WithError(token.Error()).Fatal("Could not subscribe")
		}
		ctx.Info("Subscribed. Waiting for messages...")

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		// Block until a signal is received.
		<-c

		client.Disconnect()

	},
}

func init() {
	RootCmd.AddCommand(subscribeCmd)
}

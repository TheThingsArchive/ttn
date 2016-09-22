// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var subscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "Subscribe to events for this application",
	Long:  `ttnctl subscribe can be used to subscribe to events for this application.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := util.GetMQTT(ctx)
		defer client.Disconnect()

		token := client.SubscribeActivations(func(client mqtt.Client, appID string, devID string, req mqtt.Activation) {
			ctx.Info("Activation")
			printKV("AppID", appID)
			printKV("DevID", devID)
			printKV("AppEUI", req.AppEUI)
			printKV("DevEUI", req.DevEUI)
			printKV("DevAddr", req.DevAddr)
			fmt.Println()
		})
		token.Wait()
		if err := token.Error(); err != nil {
			ctx.WithError(err).Fatal("Could not subscribe to activations")
		}
		ctx.Info("Subscribed to activations")

		token = client.SubscribeUplink(func(client mqtt.Client, appID string, devID string, req mqtt.UplinkMessage) {
			ctx.Info("Uplink Message")
			printKV("AppID", appID)
			printKV("DevID", devID)
			printKV("Port", req.FPort)
			printKV("FCnt", req.FCnt)
			printKV("Payload (hex)", req.Payload)
			if len(req.Fields) > 0 {
				ctx.Info("Decoded fields")
				for k, v := range req.Fields {
					printKV(k, v)
				}
			}
			fmt.Println()
		})
		token.Wait()
		if err := token.Error(); err != nil {
			ctx.WithError(err).Fatal("Could not subscribe to uplink")
		}
		ctx.Info("Subscribed to uplink")

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		ctx.WithField("signal", <-sigChan).Info("signal received")
	},
}

func init() {
	RootCmd.AddCommand(subscribeCmd)
}

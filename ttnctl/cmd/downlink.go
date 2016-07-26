// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

// downlinkCmd represents the `downlink` command
var downlinkCmd = &cobra.Command{
	Use:   "downlink [DevID] [Payload]",
	Short: "Send a downlink message to a device",
	Long:  `ttnctl downlink can be used to send a downlink message to a device.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := util.GetMQTT(ctx)
		defer client.Disconnect()

		appID := util.GetAppID(ctx)
		ctx = ctx.WithField("AppID", appID)

		devID := args[0]
		if !api.ValidID(devID) {
			ctx.Fatalf("Invalid Device ID") // TODO: Add link to wiki explaining device IDs
		}
		ctx = ctx.WithField("DevID", devID)

		payload, err := types.ParseHEX(args[1], len(args[1])/2)
		if err != nil {
			ctx.WithError(err).Fatal("Invalid Payload")
		}

		token := client.PublishDownlink(mqtt.DownlinkMessage{
			AppID:   appID,
			DevID:   devID,
			Payload: payload,
		})
		token.Wait()
		if token.Error() != nil {
			ctx.WithError(token.Error()).Fatal("Could not enqueue downlink")
		}

		ctx.Info("Enqueued downlink")

	},
}

func init() {
	RootCmd.AddCommand(downlinkCmd)
}

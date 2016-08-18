// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"encoding/json"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var downlinkCmd = &cobra.Command{
	Use:   "downlink [DevID] [Payload]",
	Short: "Send a downlink message to a device",
	Long:  `ttnctl downlink can be used to send a downlink message to a device.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := util.GetMQTT(ctx)
		defer client.Disconnect()

		if len(args) < 2 {
			ctx.Info("Not enough arguments. Please, provide a devID and a Payload")
			cmd.UsageFunc()(cmd)
			return
		}

		appID := util.GetAppID(ctx)
		ctx = ctx.WithField("AppID", appID)

		devID := args[0]
		if !api.ValidID(devID) {
			ctx.Fatalf("Invalid Device ID") // TODO: Add link to wiki explaining device IDs
		}
		ctx = ctx.WithField("DevID", devID)

		jsonFlag, err := cmd.Flags().GetBool("JSON")

		if err != nil {
			ctx.WithError(err).Fatal("Failed to read JSON flag")
		}

		fPort, err := cmd.Flags().GetInt("fport")

		if err != nil {
			ctx.WithError(err).Fatal("Failed to read port flag")
		}

		message := mqtt.DownlinkMessage{
			AppID: appID,
			DevID: devID,
			FPort: uint8(fPort),
		}

		if args[1] != "" {
			if jsonFlag {
				// Valid payload provided + json flag
				_, err := types.ParseHEX(args[1], len(args[1])/2)
				if err == nil {
					ctx.WithError(err).Fatal("You are providing a valid payload using the --JSON flag.")
				}

				err = json.Unmarshal([]byte(args[1]), &message.Fields)

				if err != nil {
					ctx.WithError(err).Fatal("Invalid JSON string")
					return
				}
			} else { // Payload provided
				payload, err := types.ParseHEX(args[1], len(args[1])/2)
				if err != nil {
					ctx.WithError(err).Fatal("Invalid Payload")
				}

				message.Payload = payload
			}
			token := client.PublishDownlink(message)
			token.Wait()
			if token.Error() != nil {
				ctx.WithError(token.Error()).Fatal("Could not enqueue downlink")
			}
			ctx.Info("Enqueued downlink")
		} else {
			ctx.Info("Invalid command")
			cmd.UsageFunc()(cmd)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(downlinkCmd)
	downlinkCmd.Flags().Int("fport", 1, "FPort for downlink")
	downlinkCmd.Flags().Bool("JSON", false, "Send JSON to the handler (MQTT)")
}

// Example of json string : "{\"foo\":{\"key\": [1,2,3]}}"

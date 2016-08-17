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

		JSON, err := cmd.Flags().GetBool("json")

		if err != nil {
			ctx.WithError(err).Fatal("Error: value of the json flag")
		}

		if JSON {
			fields := args[1]
			if fields == "" {
				ctx.WithError(err).Fatal("No fields or payload provided")
			}

			// Valid payload provided + json flag
			_, err := types.ParseHEX(args[1], len(args[1])/2)
			if err == nil {
				ctx.WithError(err).Fatal("You are providing a valid payload using the --json flag.")
			}

			fPort, _ := cmd.Flags().GetInt("fport")

			token := client.PublishDownlink(mqtt.DownlinkMessage{
				AppID:  appID,
				DevID:  devID,
				FPort:  uint8(fPort),
				Fields: parseJSON(fields),
			})
			token.Wait()
			if token.Error() != nil {
				ctx.WithError(token.Error()).Fatal("Could not enqueue downlink")
			}

			ctx.Info("Enqueued downlink")
		} else { // Payload provided
			payload, err := types.ParseHEX(args[1], len(args[1])/2)
			if err != nil {
				ctx.WithError(err).Fatal("Invalid Payload")
			}

			fPort, _ := cmd.Flags().GetInt("fport")

			token := client.PublishDownlink(mqtt.DownlinkMessage{
				AppID:   appID,
				DevID:   devID,
				FPort:   uint8(fPort),
				Payload: payload,
			})
			token.Wait()
			if token.Error() != nil {
				ctx.WithError(token.Error()).Fatal("Could not enqueue downlink")
			}

			ctx.Info("Enqueued downlink")
		}
	},
}

func init() {
	RootCmd.AddCommand(downlinkCmd)
	downlinkCmd.Flags().Int("fport", 1, "FPort for downlink")
	downlinkCmd.Flags().Bool("json", false, "Send Json to the handler (MQTT)")
}

// Example of json string : "{\"foo\":{\"key\": [1,2,3]}}"
func parseJSON(s string) map[string]interface{} {
	var res map[string]interface{}
	json.Unmarshal([]byte(s), &res)
	return res
}

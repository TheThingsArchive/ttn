// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"encoding/json"
	"strings"

	"github.com/TheThingsNetwork/api"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var downlinkCmd = &cobra.Command{
	Use:   "downlink [DevID] [Payload]",
	Short: "Send a downlink message to a device",
	Long:  `ttnctl downlink can be used to send a downlink message to a device.`,
	Example: `$ ttnctl downlink test aabc
  INFO Connecting to MQTT...
  INFO Connected to MQTT
  INFO Enqueued downlink                        AppID=test DevID=test

$ ttnctl downlink test --json '{"led":"on"}'
  INFO Connecting to MQTT...
  INFO Connected to MQTT
  INFO Enqueued downlink                        AppID=test DevID=test
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)

		appID := util.GetAppID(ctx)
		ctx = ctx.WithField("AppID", appID)

		devID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(devID, "Device ID"); err != nil {
			ctx.Fatal(err.Error())
		}
		ctx = ctx.WithField("DevID", devID)

		jsonflag, err := cmd.Flags().GetBool("json")
		if err != nil {
			ctx.WithError(err).Fatal("Failed to read json flag")
		}

		fPort, err := cmd.Flags().GetInt("fport")
		if err != nil {
			ctx.WithError(err).Fatal("Failed to read fport flag")
		}

		confirmed, err := cmd.Flags().GetBool("confirmed")
		if err != nil {
			ctx.WithError(err).Fatal("Failed to read confirmed flag")
		}

		accessKey, err := cmd.Flags().GetString("access-key")
		if err != nil {
			ctx.WithError(err).Fatal("Failed to read access-key flag")
		}

		client := util.GetMQTT(ctx, accessKey)
		defer client.Disconnect()

		message := types.DownlinkMessage{
			AppID:     appID,
			DevID:     devID,
			FPort:     uint8(fPort),
			Confirmed: confirmed,
		}

		if args[1] == "" {
			ctx.Info("Invalid command")
			cmd.UsageFunc()(cmd)
			return
		}

		if jsonflag {
			// Valid payload provided + json flag
			_, err := types.ParseHEX(args[1], len(args[1])/2)
			if err == nil {
				ctx.WithError(err).Fatal("You are providing a valid HEX payload while sending payload in JSON.")
			}

			err = json.Unmarshal([]byte(args[1]), &message.PayloadFields)

			if err != nil {
				ctx.WithError(err).Fatal("Invalid json string")
				return
			}
		} else { // Payload provided
			payload, err := types.ParseHEX(args[1], len(args[1])/2)
			if err != nil {
				ctx.WithError(err).Fatal("Invalid Payload")
			}

			message.PayloadRaw = payload
		}
		token := client.PublishDownlink(message)
		token.Wait()
		if token.Error() != nil {
			ctx.WithError(token.Error()).Fatal("Could not enqueue downlink")
		}
		ctx.Info("Enqueued downlink")
	},
}

func init() {
	RootCmd.AddCommand(downlinkCmd)
	downlinkCmd.Flags().Int("fport", 1, "FPort for downlink")
	downlinkCmd.Flags().Bool("confirmed", false, "Confirmed downlink")
	downlinkCmd.Flags().Bool("json", false, "Provide the payload as JSON")
	downlinkCmd.Flags().String("access-key", "", "The access key to use")
}

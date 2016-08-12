// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"

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

		// Flag JSON
		fJSON, err := cmd.Flags().GetBool("fjson")

		if err != nil {
			ctx.WithError(err).Fatal("Error: value of the json flag")
		}

		if fJSON { // No Payload but fields
			fields := args[1]
			if fields == "" {
				ctx.WithError(err).Fatal("No fields or payload provided")
			}

			// In case we provide a valid payload using the json flag
			_, err := types.ParseHEX(args[1], len(args[1])/2)
			if err == nil {
				ctx.WithError(err).Fatal("You are providing a valid payload using the --fjson flag.")
			}

			fPort, _ := cmd.Flags().GetInt("fport")

			token := client.PublishDownlink(mqtt.DownlinkMessage{
				AppID:  appID,
				DevID:  devID,
				FPort:  uint8(fPort),
				Fields: convertStringIntoMapStringInterface(fields),
			})
			token.Wait()
			if token.Error() != nil {
				ctx.WithError(token.Error()).Fatal("Could not enqueue downlink")
			}

			ctx.Info("Enqueued downlink")
		} else { // Payload
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
	// New flag for json fields. Using the flag means that we send json, otherwise we send []byte (false by default)
	downlinkCmd.Flags().Bool("fjson", false, "Send Json to the handler (MQTT)")
}

// example of json string : "{"key" : 1}"
func convertStringIntoMapStringInterface(s string) map[string]interface{} {
	fmt.Println("La chaine entree est : ", s)

	res := make(map[string]interface{})
	s = strings.Replace(s, "{", "", -1)
	s = strings.Replace(s, "}", "", -1)
	arrayIndex := strings.Split(s, ":")
	res[arrayIndex[0]] = arrayIndex[1]
	return res
}

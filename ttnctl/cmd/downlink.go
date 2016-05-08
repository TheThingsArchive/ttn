// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

// downlinkCmd represents the `downlink` command
var downlinkCmd = &cobra.Command{
	Use:   "downlink [DevEUI] [Payload] [TTL]",
	Short: "Send downlink messages to the network",
	Long: `ttnctl downlink sends a downlink message to the network

The DevEUI should be an 8-byte long hex-encoded string (16 chars), the Payload
is a hex-encoded string and the TTL defines the time-to-live of this downlink,
formatted as "1h2m" for one hour and two minutes. The default TTL is one hour.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			ctx.Fatal("Insufficient arguments")
		}
		if len(args) < 3 {
			args = append(args, "1h")
		}

		appEUI := util.GetAppEUI(ctx)

		devEUI, err := types.ParseDevEUI(args[0])
		if err != nil {
			ctx.Fatalf("Invalid DevEUI: %s", err)
		}

		var payload []byte
		if plain, _ := cmd.Flags().GetBool("plain"); plain {
			ctx.Warn("Sending data as plain text is bad practice. We recommend to transmit data in a binary format.")
			payload = []byte(args[1])
		} else {
			payload, err = util.ParseHEX(args[1], len(args[1]))
			if err != nil {
				ctx.Fatalf("Invalid hexadecimal payload. If you are trying to send a plain-text payload, use the --plain flag.")
			}
		}

		dataDown := core.DataDownAppReq{
			Payload: payload,
			TTL:     args[2],
		}

		if err != nil {
			ctx.WithError(err).Fatal("Unable to create downlink payload")
		}

		client := util.ConnectMQTTClient(ctx)
		defer client.Disconnect()

		token := client.PublishDownlink(appEUI, devEUI, dataDown)

		if token.Wait(); token.Error() != nil {
			ctx.WithError(token.Error()).Fatal("Could not publish downlink")
		}
		ctx.Info("Scheduled downlink")
	},
}

func init() {
	RootCmd.AddCommand(downlinkCmd)
	downlinkCmd.Flags().Bool("plain", false, "send payload as plain-text")
}

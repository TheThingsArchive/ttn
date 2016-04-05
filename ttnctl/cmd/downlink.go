// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
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

		appEUI, err := util.Parse64(viper.GetString("app-eui"))
		if err != nil {
			ctx.Fatalf("Invalid AppEUI: %s", err)
		}

		apps, err := util.GetApplications(ctx)
		if err != nil {
			ctx.WithError(err).Fatal("Failed to get applications")
		}

		var appAccessKey string
		for _, a := range apps {
			if a.EUI == fmt.Sprintf("%X", appEUI) {
				// Don't care about which access key in this cli
				appAccessKey = a.AccessKeys[0]
			}
		}
		if appAccessKey == "" {
			ctx.Fatal("Application not found")
		}

		devEUI, err := util.Parse64(args[0])
		if err != nil {
			ctx.Fatalf("Invalid DevEUI: %s", err)
		}

		dataDown := core.DataDownAppReq{
			Payload: []byte(args[1]),
			TTL:     args[2],
		}

		if err != nil {
			ctx.WithError(err).Fatal("Unable to create downlink payload")
		}

		client := util.ConnectMQTTClient(ctx, appEUI, appAccessKey)

		token := client.PublishDownlink(appEUI, devEUI, dataDown)

		if token.Wait(); token.Error() != nil {
			ctx.WithError(token.Error()).Fatal("Could not publish downlink")
		}

		client.Disconnect()

	},
}

func init() {
	RootCmd.AddCommand(downlinkCmd)
}

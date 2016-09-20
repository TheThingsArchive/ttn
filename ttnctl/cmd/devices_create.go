// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

// devicesCreateCmd represents the `device create` command
var devicesCreateCmd = &cobra.Command{
	Use:   "create [Device ID] [DevEUI] [AppKey]",
	Short: "Create a new device",
	Long:  `ttnctl devices create can be used to create a new device.`,
	Example: `$ ttnctl devices create test
  INFO Using Application                        AppEUI=70B3D57EF0000024 AppID=test
  INFO Generating random DevEUI...
  INFO Generating random AppKey...
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Created device                           AppEUI=70B3D57EF0000024 AppID=test AppKey=EBD2E2810A4307263FE5EF78E2EF589D DevEUI=0001D544B2936FCE DevID=test
`,
	Run: func(cmd *cobra.Command, args []string) {

		var err error

		if len(args) == 0 {
			ctx.Fatalf("Device ID is required")
		}

		devID := args[0]
		if !api.ValidID(devID) {
			ctx.Fatalf("Invalid Device ID") // TODO: Add link to wiki explaining device IDs
		}

		appID := util.GetAppID(ctx)
		appEUI := util.GetAppEUI(ctx)

		var devEUI types.DevEUI
		if len(args) > 1 {
			devEUI, err = types.ParseDevEUI(args[1])
			if err != nil {
				ctx.Fatalf("Invalid DevEUI: %s", err)
			}
		} else {
			ctx.Info("Generating random DevEUI...")
			copy(devEUI[1:], random.Bytes(7))
		}

		var appKey types.AppKey
		if len(args) > 2 {
			appKey, err = types.ParseAppKey(args[2])
			if err != nil {
				ctx.Fatalf("Invalid AppKey: %s", err)
			}
		} else {
			ctx.Info("Generating random AppKey...")
			copy(appKey[:], random.Bytes(16))
		}

		conn, manager := util.GetHandlerManager(ctx)
		defer conn.Close()

		err = manager.SetDevice(&handler.Device{
			AppId: appID,
			DevId: devID,
			Device: &handler.Device_LorawanDevice{LorawanDevice: &lorawan.Device{
				AppId:  appID,
				DevId:  devID,
				AppEui: &appEUI,
				DevEui: &devEUI,
				AppKey: &appKey,
			}},
		})
		if err != nil {
			ctx.WithError(err).Fatal("Could not create Device")
		}

		ctx.WithFields(log.Fields{
			"AppID":  appID,
			"DevID":  devID,
			"AppEUI": appEUI,
			"DevEUI": devEUI,
			"AppKey": appKey,
		}).Info("Created device")
	},
}

func init() {
	devicesCmd.AddCommand(devicesCreateCmd)
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"strings"

	"github.com/TheThingsNetwork/api"
	"github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/api/protocol/lorawan"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/pseudorandom"
	"github.com/TheThingsNetwork/go-utils/random"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var devicesRegisterCmd = &cobra.Command{
	Use:   "register [Device ID] [DevEUI] [AppKey] [Lat,Long]",
	Short: "Register a new device",
	Long:  `ttnctl devices register can be used to register a new device.`,
	Example: `$ ttnctl devices register test
  INFO Using Application                        AppEUI=70B3D57EF0000024 AppID=test
  INFO Generating random DevEUI...
  INFO Generating random AppKey...
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Registered device                        AppEUI=70B3D57EF0000024 AppID=test AppKey=EBD2E2810A4307263FE5EF78E2EF589D DevEUI=0001D544B2936FCE DevID=test
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 4)

		var err error

		devID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(devID, "Device ID"); err != nil {
			ctx.Fatal(err.Error())
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
			pseudorandom.FillBytes(devEUI[1:])
		}

		var appKey types.AppKey
		if len(args) > 2 {
			appKey, err = types.ParseAppKey(args[2])
			if err != nil {
				ctx.Fatalf("Invalid AppKey: %s", err)
			}
		} else {
			ctx.Info("Generating random AppKey...")
			random.FillBytes(appKey[:])
		}

		device := &handler.Device{
			AppID: appID,
			DevID: devID,
			Device: &handler.Device_LoRaWANDevice{LoRaWANDevice: &lorawan.Device{
				AppID:         appID,
				DevID:         devID,
				AppEUI:        &appEUI,
				DevEUI:        &devEUI,
				AppKey:        &appKey,
				Uses32BitFCnt: true,
			}},
		}

		if len(args) > 3 {
			location, err := util.ParseLocation(args[3])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid location")
			}
			device.Latitude = float32(location.Latitude)
			device.Longitude = float32(location.Longitude)
		}

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		err = manager.SetDevice(device)
		if err != nil {
			ctx.WithError(err).Fatal("Could not register Device")
		}

		ctx.WithFields(ttnlog.Fields{
			"AppID":  appID,
			"DevID":  devID,
			"AppEUI": appEUI,
			"DevEUI": devEUI,
			"AppKey": appKey,
		}).Info("Registered device")
	},
}

func init() {
	devicesCmd.AddCommand(devicesRegisterCmd)
}

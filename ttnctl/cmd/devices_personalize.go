// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"strings"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/random"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var devicesPersonalizeCmd = &cobra.Command{
	Use:   "personalize [Device ID] [NwkSKey] [AppSKey]",
	Short: "Personalize a device",
	Long:  `ttnctl devices personalize can be used to personalize a device (ABP).`,
	Example: `$ ttnctl devices personalize test
  INFO Using Application                        AppEUI=70B3D57EF0000024 AppID=test
  INFO Generating random NwkSKey...
  INFO Generating random AppSKey...
  INFO Discovering Handler...                   Handler=ttn-handler-eu
  INFO Connecting with Handler...               Handler=eu.thethings.network:1904
  INFO Requesting DevAddr for device...
  INFO Personalized device                      AppID=test AppSKey=D8DD37B4B709BA76C6FEC62CAD0CCE51 DevAddr=26001ADA DevID=test NwkSKey=3382A3066850293421ED8D392B9BF4DF
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 3)

		var err error

		devID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(devID, "Device ID"); err != nil {
			ctx.Fatal(err.Error())
		}

		appID := util.GetAppID(ctx)

		var nwkSKey types.NwkSKey
		if len(args) > 1 {
			nwkSKey, err = types.ParseNwkSKey(args[1])
			if err != nil {
				ctx.Fatalf("Invalid NwkSKey: %s", err)
			}
		} else {
			ctx.Info("Generating random NwkSKey...")
			random.FillBytes(nwkSKey[:])
		}

		var appSKey types.AppSKey
		if len(args) > 2 {
			appSKey, err = types.ParseAppSKey(args[2])
			if err != nil {
				ctx.Fatalf("Invalid AppSKey: %s", err)
			}
		} else {
			ctx.Info("Generating random AppSKey...")
			random.FillBytes(appSKey[:])
		}

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		dev, err := manager.GetDevice(appID, devID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get existing device.")
		}

		ctx.Info("Requesting DevAddr for device...")

		var constraints []string
		if lorawan := dev.GetLorawanDevice(); lorawan != nil && lorawan.ActivationConstraints != "" {
			constraints = strings.Split(lorawan.ActivationConstraints, ",")
		}
		constraints = append(constraints, "abp")

		devAddr, err := manager.GetDevAddr(constraints...)
		if err != nil {
			ctx.WithError(err).Fatal("Could not request device address")
		}

		var emptyAppKey types.AppKey
		dev.GetLorawanDevice().AppKey = &emptyAppKey
		dev.GetLorawanDevice().DevAddr = &devAddr
		dev.GetLorawanDevice().NwkSKey = &nwkSKey
		dev.GetLorawanDevice().AppSKey = &appSKey
		dev.GetLorawanDevice().FCntUp = 0
		dev.GetLorawanDevice().FCntDown = 0

		err = manager.SetDevice(dev)
		if err != nil {
			ctx.WithError(err).Fatal("Could not update Device")
		}

		ctx.WithFields(ttnlog.Fields{
			"AppID":   appID,
			"DevID":   devID,
			"DevAddr": devAddr,
			"NwkSKey": nwkSKey,
			"AppSKey": appSKey,
		}).Info("Personalized device")
	},
}

func init() {
	devicesCmd.AddCommand(devicesPersonalizeCmd)
}

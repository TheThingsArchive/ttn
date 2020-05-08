// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"strings"

	"github.com/TheThingsNetwork/api"
	"github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/api/handler/handlerclient"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func convertToABP(manager *handlerclient.ManagerClient, dev *handler.Device, flags *pflag.FlagSet) {
	// Do all updates
	if dev.GetLoRaWANDevice().AppKey.String() != "" {
		old := dev.GetLoRaWANDevice().AppKey.String()
		dev.GetLoRaWANDevice().AppKey = &types.AppKey{}

		attr, _ := flags.GetString("save-to-attribute")
		if attr != "" {
			if dev.Attributes == nil {
				dev.Attributes = make(map[string]string)
			}
			dev.Attributes[attr] = old
		}

		err := manager.SetDevice(dev)
		if err != nil {
			ctx.WithError(err).Fatal("Could not update Device")
		}

		ctx.WithFields(ttnlog.Fields{
			"AppID":          dev.AppID,
			"DevID":          dev.DevID,
			"OriginalAppKey": old,
		}).Info("Remove device AppKey")
	} else {
		ctx.WithFields(ttnlog.Fields{
			"AppID": dev.AppID,
			"DevID": dev.DevID,
		}).Info("Device has no AppKey")
	}
}

var devicesConvertToABPCmd = &cobra.Command{
	Use:   "convert-to-abp [Device ID]",
	Short: "Remove AppKey for an OTAA device",
	Long:  `ttnctl devices convert-to-abp can be used to remove the AppKey of an OTAA device.`,
	Example: `$ ttnctl devices disable DeviceID
  INFO Using Application                        AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Remove device AppKey                     AppID=test DevID=test OriginalAppKey=XXXX
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		devID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(devID, "Device ID"); err != nil {
			ctx.Fatal(err.Error())
		}

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		dev, err := manager.GetDevice(appID, devID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get existing device.")
		}

		convertToABP(manager, dev, cmd.Flags())
	},
}

var devicesConvertAllToABPCmd = &cobra.Command{
	Use:   "convert-all-to-abp",
	Short: "Remove AppKey for all OTAA devices",
	Long:  `ttnctl devices convert-all-to-abp can be used to remove the AppKey of all OTAA devices of an application.`,
	Example: `$ ttnctl devices disable-all
  INFO Using Application                        AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Remove device AppKey                     AppID=test DevID=test1 OriginalAppKey=XXXX
  INFO Remove device AppKey                     AppID=test DevID=test2 OriginalAppKey=XXXX
  INFO Remove device AppKey                     AppID=test DevID=test3 OriginalAppKey=XXXX
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 0, 0)

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		devs, err := manager.GetDevicesForApplication(appID, 0, 0)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get existing device.")
		}

		for _, dev := range devs {
			convertToABP(manager, dev, cmd.Flags())
		}
	},
}

func convertToABPFlagSet() *pflag.FlagSet {
	set := &pflag.FlagSet{}
	set.String("save-to-attribute", "", "Save original AppKey as a device attribute")
	return set
}

func init() {
	devicesCmd.AddCommand(devicesConvertToABPCmd)
	devicesCmd.AddCommand(devicesConvertAllToABPCmd)
	devicesConvertToABPCmd.Flags().AddFlagSet(convertToABPFlagSet())
	devicesConvertAllToABPCmd.Flags().AddFlagSet(convertToABPFlagSet())
}

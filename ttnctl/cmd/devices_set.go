// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var devicesSetCmd = &cobra.Command{
	Use:   "set [Device ID]",
	Short: "Set properties of a device",
	Long:  `ttnctl devices set can be used to set properties of a device.`,
	Example: `$ ttnctl devices set test --fcnt-up 0 --fcnt-down 0
  INFO Using Application                        AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Updated device                           AppID=test DevID=test
`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			cmd.UsageFunc()(cmd)
			return
		}

		devID := args[0]
		if !api.ValidID(devID) {
			ctx.Fatalf("Invalid Device ID") // TODO: Add link to wiki explaining device IDs
		}

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx)
		defer conn.Close()

		dev, err := manager.GetDevice(appID, devID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get existing device.")
		}

		// Do all updates

		if in, err := cmd.Flags().GetString("app-eui"); err == nil && in != "" {
			appEUI, err := types.ParseAppEUI(in)
			if err != nil {
				ctx.Fatalf("Invalid AppEUI: %s", err)
			}
			dev.GetLorawanDevice().AppEui = &appEUI
		}

		if in, err := cmd.Flags().GetString("dev-eui"); err == nil && in != "" {
			devEUI, err := types.ParseDevEUI(in)
			if err != nil {
				ctx.Fatalf("Invalid DevEUI: %s", err)
			}
			dev.GetLorawanDevice().DevEui = &devEUI
		}

		if in, err := cmd.Flags().GetString("dev-addr"); err == nil && in != "" {
			devAddr, err := types.ParseDevAddr(in)
			if err != nil {
				ctx.Fatalf("Invalid DevAddr: %s", err)
			}
			ctx.Warn("Using a DevAddr that was not issued by the NetworkServer could break connectivity")
			dev.GetLorawanDevice().DevAddr = &devAddr
		}

		if in, err := cmd.Flags().GetString("nwk-s-key"); err == nil && in != "" {
			key, err := types.ParseNwkSKey(in)
			if err != nil {
				ctx.Fatalf("Invalid NwkSKey: %s", err)
			}
			dev.GetLorawanDevice().NwkSKey = &key
		}

		if in, err := cmd.Flags().GetString("app-s-key"); err == nil && in != "" {
			key, err := types.ParseAppSKey(in)
			if err != nil {
				ctx.Fatalf("Invalid AppSKey: %s", err)
			}
			dev.GetLorawanDevice().AppSKey = &key
		}

		if in, err := cmd.Flags().GetString("app-key"); err == nil && in != "" {
			key, err := types.ParseAppKey(in)
			if err != nil {
				ctx.Fatalf("Invalid AppKey: %s", err)
			}
			dev.GetLorawanDevice().AppKey = &key
		}

		if in, err := cmd.Flags().GetInt("fcnt-up"); err == nil && in != -1 {
			dev.GetLorawanDevice().FCntUp = uint32(in)
		}

		if in, err := cmd.Flags().GetInt("fcnt-down"); err == nil && in != -1 {
			dev.GetLorawanDevice().FCntDown = uint32(in)
		}

		if in, err := cmd.Flags().GetBool("disable-fcnt-check"); err == nil {
			dev.GetLorawanDevice().DisableFCntCheck = in
		}

		if in, err := cmd.Flags().GetBool("32-bit-fcnt"); err == nil {
			dev.GetLorawanDevice().Uses32BitFCnt = in
		}

		err = manager.SetDevice(dev)
		if err != nil {
			ctx.WithError(err).Fatal("Could not update Device")
		}

		ctx.WithFields(log.Fields{
			"AppID": appID,
			"DevID": devID,
		}).Info("Updated device")
	},
}

func init() {
	devicesCmd.AddCommand(devicesSetCmd)

	devicesSetCmd.Flags().String("app-eui", "", "Set AppEUI")
	devicesSetCmd.Flags().String("dev-eui", "", "Set DevEUI")
	devicesSetCmd.Flags().String("dev-addr", "", "Set DevAddr")
	devicesSetCmd.Flags().String("nwk-s-key", "", "Set NwkSKey")
	devicesSetCmd.Flags().String("app-s-key", "", "Set AppSKey")
	devicesSetCmd.Flags().String("app-key", "", "Set AppKey")

	devicesSetCmd.Flags().Int("fcnt-up", -1, "Set FCnt Up")
	devicesSetCmd.Flags().Int("fcnt-down", -1, "Set FCnt Down")

	devicesSetCmd.Flags().Bool("disable-fcnt-check", false, "Disable FCnt check")
	devicesSetCmd.Flags().Bool("32-bit-fcnt", false, "Use 32 bit FCnt")
}

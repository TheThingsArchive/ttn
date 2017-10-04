// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/TheThingsNetwork/api"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
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

		// Do all updates

		if in, err := cmd.Flags().GetString("app-eui"); err == nil && in != "" {

			ctx.Warn("Manually changing the AppEUI of a device might break routing for this device")
			if override, _ := cmd.Flags().GetBool("override"); !override {
				ctx.Warnf("Use the --override flag if you're really sure you want to do this")
				os.Exit(0)
			}

			appEUI, err := types.ParseAppEUI(in)
			if err != nil {
				ctx.Fatalf("Invalid AppEUI: %s", err)
			}
			dev.GetLoRaWANDevice().AppEUI = appEUI
		}

		if in, err := cmd.Flags().GetString("dev-eui"); err == nil && in != "" {

			ctx.Warn("Manually changing the DevEUI of a device might break routing for this device")
			if override, _ := cmd.Flags().GetBool("override"); !override {
				ctx.Warnf("Use the --override flag if you're really sure you want to do this")
				os.Exit(0)
			}

			devEUI, err := types.ParseDevEUI(in)
			if err != nil {
				ctx.Fatalf("Invalid DevEUI: %s", err)
			}
			dev.GetLoRaWANDevice().DevEUI = devEUI
		}

		if in, err := cmd.Flags().GetString("dev-addr"); err == nil && in != "" {

			ctx.Warn("Manually changing the DevAddr of a device might break routing for this device")
			if override, _ := cmd.Flags().GetBool("override"); !override {
				ctx.Warnf("Use the --override flag if you're really sure you want to do this")
				os.Exit(0)
			}

			devAddr, err := types.ParseDevAddr(in)
			if err != nil {
				ctx.Fatalf("Invalid DevAddr: %s", err)
			}
			dev.GetLoRaWANDevice().DevAddr = &devAddr
		}

		if in, err := cmd.Flags().GetString("nwk-s-key"); err == nil && in != "" {
			key, err := types.ParseNwkSKey(in)
			if err != nil {
				ctx.Fatalf("Invalid NwkSKey: %s", err)
			}
			dev.GetLoRaWANDevice().NwkSKey = &key
		}

		if in, err := cmd.Flags().GetString("app-s-key"); err == nil && in != "" {
			key, err := types.ParseAppSKey(in)
			if err != nil {
				ctx.Fatalf("Invalid AppSKey: %s", err)
			}
			dev.GetLoRaWANDevice().AppSKey = &key
		}

		if in, err := cmd.Flags().GetString("app-key"); err == nil && in != "" {
			key, err := types.ParseAppKey(in)
			if err != nil {
				ctx.Fatalf("Invalid AppKey: %s", err)
			}
			dev.GetLoRaWANDevice().AppKey = &key
		}

		if in, err := cmd.Flags().GetInt("fcnt-up"); err == nil && in != -1 {
			dev.GetLoRaWANDevice().FCntUp = uint32(in)
		}

		if in, err := cmd.Flags().GetInt("fcnt-down"); err == nil && in != -1 {
			dev.GetLoRaWANDevice().FCntDown = uint32(in)
		}

		if in, err := cmd.Flags().GetBool("enable-fcnt-check"); err == nil && in {
			dev.GetLoRaWANDevice().DisableFCntCheck = false
		}

		if in, err := cmd.Flags().GetBool("disable-fcnt-check"); err == nil && in {
			dev.GetLoRaWANDevice().DisableFCntCheck = true
		}

		if in, err := cmd.Flags().GetBool("32-bit-fcnt"); err == nil && in {
			dev.GetLoRaWANDevice().Uses32BitFCnt = true
		}

		if in, err := cmd.Flags().GetBool("16-bit-fcnt"); err == nil && in {
			dev.GetLoRaWANDevice().Uses32BitFCnt = false
		}

		if in, err := cmd.Flags().GetFloat32("latitude"); err == nil && in != 0 {
			dev.Latitude = in
		}

		if in, err := cmd.Flags().GetFloat32("longitude"); err == nil && in != 0 {
			dev.Longitude = in
		}

		if in, err := cmd.Flags().GetInt32("altitude"); err == nil && in != 0 {
			dev.Altitude = in
		}

		if in, err := cmd.Flags().GetString("description"); err == nil && in != "" {
			dev.Description = in
		}

		if in, err := cmd.Flags().GetStringSlice("attr-set"); err == nil && len(in) > 0 {
			if dev.Attributes == nil {
				dev.Attributes = make(map[string]string, len(in))
			}
			for _, v := range in {
				s := strings.SplitN(v, ":", 2)
				if len(s) == 2 {
					dev.Attributes[s[0]] = s[1]
				} else {
					ctx.Error(fmt.Sprintf("attr-set: cannot parse key:value %s", s))
				}
			}
		}

		if in, err := cmd.Flags().GetStringSlice("attr-remove"); err == nil && len(in) > 0 {
			for _, v := range in {
				delete(dev.Attributes, v)
			}
		}

		err = manager.SetDevice(dev)
		if err != nil {
			ctx.WithError(err).Fatal("Could not update Device")
		}

		ctx.WithFields(ttnlog.Fields{
			"AppID": appID,
			"DevID": devID,
		}).Info("Updated device")
	},
}

func init() {
	devicesCmd.AddCommand(devicesSetCmd)

	devicesSetCmd.Flags().Bool("override", false, "Override protection against breaking changes")

	devicesSetCmd.Flags().String("app-eui", "", "Set AppEUI")
	devicesSetCmd.Flags().String("dev-eui", "", "Set DevEUI")
	devicesSetCmd.Flags().String("dev-addr", "", "Set DevAddr")
	devicesSetCmd.Flags().String("nwk-s-key", "", "Set NwkSKey")
	devicesSetCmd.Flags().String("app-s-key", "", "Set AppSKey")
	devicesSetCmd.Flags().String("app-key", "", "Set AppKey")

	devicesSetCmd.Flags().Int("fcnt-up", -1, "Set FCnt Up")
	devicesSetCmd.Flags().Int("fcnt-down", -1, "Set FCnt Down")

	devicesSetCmd.Flags().Bool("disable-fcnt-check", false, "Disable FCnt check")
	devicesSetCmd.Flags().Bool("enable-fcnt-check", false, "Enable FCnt check (default)")
	devicesSetCmd.Flags().Bool("32-bit-fcnt", false, "Use 32 bit FCnt (default)")
	devicesSetCmd.Flags().Bool("16-bit-fcnt", false, "Use 16 bit FCnt")

	devicesSetCmd.Flags().Float32("latitude", 0, "Set latitude")
	devicesSetCmd.Flags().Float32("longitude", 0, "Set longitude")
	devicesSetCmd.Flags().Int32("altitude", 0, "Set altitude")

	devicesSetCmd.Flags().String("description", "", "Set Description")

	devicesSetCmd.Flags().StringSlice("attr-set", nil, "Add a device attribute (key:value)")
	devicesSetCmd.Flags().StringSlice("attr-remove", nil, "Remove device attribute")
}

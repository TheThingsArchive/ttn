// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"os"
	"strings"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
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
			dev.GetLorawanDevice().AppEui = &appEUI
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
			dev.GetLorawanDevice().DevEui = &devEUI
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

		if in, err := cmd.Flags().GetInt("fcnt-up"); err == nil && in > -1 {
			dev.GetLorawanDevice().FCntUp = uint32(in)
		}

		if in, err := cmd.Flags().GetInt("fcnt-down"); err == nil && in > -1 {
			dev.GetLorawanDevice().FCntDown = uint32(in)
		}

		if in, err := cmd.Flags().GetBool("enable-fcnt-check"); err == nil && in {
			dev.GetLorawanDevice().DisableFCntCheck = false
		}

		if in, err := cmd.Flags().GetBool("disable-fcnt-check"); err == nil && in {
			dev.GetLorawanDevice().DisableFCntCheck = true
		}

		if in, err := cmd.Flags().GetBool("32-bit-fcnt"); err == nil && in {
			dev.GetLorawanDevice().Uses32BitFCnt = true
		}

		if in, err := cmd.Flags().GetBool("16-bit-fcnt"); err == nil && in {
			dev.GetLorawanDevice().Uses32BitFCnt = false
		}

		if in, err := cmd.Flags().GetStringSlice("preferred-gateways"); err == nil {
			if len(in) == 1 && in[0] == "-" {
				dev.GetLorawanDevice().PreferredGateways = []string{}
			} else {
				dev.GetLorawanDevice().PreferredGateways = in
			}
		}

		if in, err := cmd.Flags().GetString("rx2-data-rate"); err == nil && in != "" {
			if in == "-" {
				dev.GetLorawanDevice().Rx2DataRate = ""
			} else {
				dev.GetLorawanDevice().Rx2DataRate = in
			}
		}

		if in, err := cmd.Flags().GetInt64("rx2-frequency"); err == nil && in > -1 {
			dev.GetLorawanDevice().Rx2Frequency = uint64(in)
		}

		if in, err := cmd.Flags().GetString("frequency-plan"); err == nil && in != "" {
			if fp, ok := pb_lorawan.FrequencyPlan_value[in]; ok {
				dev.GetLorawanDevice().FrequencyPlan = pb_lorawan.FrequencyPlan(fp)
			} else {
				var allowedValues []string
				for v := range pb_lorawan.FrequencyPlan_value {
					allowedValues = append(allowedValues, v)
				}
				ctx.Fatalf("Invalid Frequency Plan Name, allowed values are: %s", strings.Join(allowedValues, ", "))
			}
		}

		if in, err := cmd.Flags().GetString("class"); err == nil && in != "" {
			switch strings.ToLower(in) {
			case "a":
				dev.GetLorawanDevice().Class = pb_lorawan.Class_A
			case "b":
				dev.GetLorawanDevice().Class = pb_lorawan.Class_B
			case "c":
				dev.GetLorawanDevice().Class = pb_lorawan.Class_C
			default:
				ctx.Fatalf("Invalid Device Class")
			}
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

	devicesSetCmd.Flags().StringSlice("preferred-gateways", []string{}, "Set device preferred gateways")
	devicesSetCmd.Flags().String("rx2-data-rate", "", "Custom Data Rate to be used in RX2 (reset with \"-\")")
	devicesSetCmd.Flags().Int64("rx2-frequency", -1, "Custom Frequency to be used in RX2 (reset with \"0\")")
	devicesSetCmd.Flags().String("frequency-plan", "", "Set device frequency plan")
	devicesSetCmd.Flags().String("class", "", "Set device class (a/b/c)")

	devicesSetCmd.Flags().Float32("latitude", 0, "Set latitude")
	devicesSetCmd.Flags().Float32("longitude", 0, "Set longitude")
	devicesSetCmd.Flags().Int32("altitude", 0, "Set altitude")

	devicesSetCmd.Flags().String("description", "", "Set Description")
}

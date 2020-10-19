// Copyright © 2020 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/TheThingsNetwork/api"
	"github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	interval = time.Millisecond
)

func exportV3Device(dev *handler.Device, flags *pflag.FlagSet) *v3Device {
	v3Dev := &v3Device{}
	v3Dev.IDs.ApplicationIDs.ApplicationID = dev.AppID
	v3Dev.IDs.DeviceID = dev.DevID

	if dev.Latitude != 0 || dev.Longitude != 0 {
		v3Dev.Locations = map[string]v3DeviceLocation{
			"user": {
				Latitude:  float64(dev.Latitude),
				Longitude: float64(dev.Longitude),
				Altitude:  dev.Altitude,
				Source:    "SOURCE_REGISTRY",
			},
		}
	}

	v3Dev.Name = dev.DevID
	v3Dev.Description = dev.Description
	v3Dev.Attributes = dev.Attributes

	lorawanDevice := dev.GetLoRaWANDevice()

	v3Dev.IDs.JoinEUI = lorawanDevice.GetAppEUI().String()
	v3Dev.IDs.DevEUI = lorawanDevice.GetDevEUI().String()
	if lorawanDevice.AppKey.String() != "" {
		v3Dev.RootKeys = &v3DeviceRootKeys{
			AppKey: v3DeviceKey{Key: lorawanDevice.AppKey.String()},
		}
	}
	if devAddr := lorawanDevice.DevAddr; !devAddr.IsEmpty() && lorawanDevice.NwkSKey.String() != "" {
		v3Dev.DevAddr = devAddr.String()
		v3Dev.Session = &v3DeviceSession{}
		v3Dev.Session.DevAddr = lorawanDevice.DevAddr.String()
		if lorawanDevice.AppSKey.String() != "" {
			v3Dev.Session.Keys.AppSKey = v3DeviceKey{Key: lorawanDevice.AppSKey.String()}
		}
		v3Dev.Session.Keys.FNwkSIntKey = v3DeviceKey{Key: lorawanDevice.NwkSKey.String()}
		v3Dev.Session.LastFCntUp = lorawanDevice.FCntUp
		v3Dev.Session.LastNFCntDown = lorawanDevice.FCntDown
	}

	v3Dev.MACSettings.Supports32BitFCnt = lorawanDevice.Uses32BitFCnt
	v3Dev.MACSettings.ResetsFCnt = lorawanDevice.DisableFCntCheck

	v3Dev.MACVersion = "MAC_V1_0_2"
	v3Dev.PHYVersion = "PHY_V1_0_2_REV_B"

	v3Dev.SupportsJoin = lorawanDevice.AppKey.String() != ""
	v3Dev.MACSettings.Rx1Delay.Value = "RX_DELAY_1"

	frequencyPlanFlag, _ := flags.GetString("frequency-plan-id")
	frequencyPlan, err := getOption(frequencyPlans, frequencyPlanFlag)
	if err != nil {
		ctx.WithError(err).WithFields(log.Fields{
			"frequency_plan_id":            frequencyPlanFlag,
			"available_frequency_plan_ids": frequencyPlans,
		}).Fatal("Invalid --frequency-plan-id argument")
	}
	v3Dev.FrequencyPlanID = frequencyPlan

	return v3Dev
}

var devicesExportCmd = &cobra.Command{
	Use:     "export [Device ID]",
	Short:   "Export a device",
	Long:    `ttnctl devices export exports a device to an external format.`,
	Example: `$ ttnctl devices export test | ttn-lw-cli end-devices create --application-id app-id`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		devID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(devID, "Device ID"); err != nil {
			ctx.Fatal(err.Error())
		}
		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)
		dev, err := manager.GetDevice(appID, devID)
		if err != nil {
			conn.Close()
			ctx.WithError(err).Fatal("Could not get existing device.")
		}

		format, _ := cmd.Flags().GetString("format")
		switch format {
		case "v3":
			v3dev := exportV3Device(dev, cmd.Flags())
			if err := json.NewEncoder(os.Stdout).Encode(v3dev); err != nil {
				conn.Close()
				ctx.WithError(err).Fatal("Could not export device in v3 format")
			}
		}
		conn.Close()
	},
}

var devicesExportAllCmd = &cobra.Command{
	Use:     "export-all",
	Short:   "Export all devices",
	Long:    `ttnctl devices export-all exports all devices to an external format.`,
	Example: `$ ttnctl devices export-all | ttn-lw-cli end-devices create --application-id app-id`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 0, 0)

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)

		devs, err := manager.GetDevicesForApplication(appID, 0, 0)
		if err != nil {
			conn.Close()
			ctx.WithError(err).Fatal("Could not get devices.")
		}

		withFrameCounters, _ := cmd.Flags().GetBool("with-frame-counters")
		for _, dev := range devs {
			if withFrameCounters {
				time.Sleep(interval)
				lwDev, err := manager.GetDevice(dev.AppID, dev.DevID)
				if err != nil {
					ctx.WithError(err).WithField("device_id", dev.DevID).Warn("Could not export frame counters.")
				} else {
					dev.GetLoRaWANDevice().FCntDown = lwDev.GetLoRaWANDevice().FCntDown
					dev.GetLoRaWANDevice().FCntUp = lwDev.GetLoRaWANDevice().FCntUp
					dev.GetLoRaWANDevice().LastSeen = lwDev.GetLoRaWANDevice().LastSeen
				}
			}
			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "v3":
				v3dev := exportV3Device(dev, cmd.Flags())
				if err := json.NewEncoder(os.Stdout).Encode(v3dev); err != nil {
					conn.Close()
					ctx.WithError(err).Fatal("Could not export device in v3 format")
				}
			}
		}
		conn.Close()
	},
}

func v3ExportFlagSet() *pflag.FlagSet {
	set := &pflag.FlagSet{}
	set.String("format", "v3", "Formatting: v3")
	set.String("frequency-plan-id", "", "Specify Frequency Plan ID")
	return set
}

func init() {
	devicesCmd.AddCommand(devicesExportCmd)
	devicesCmd.AddCommand(devicesExportAllCmd)
	devicesExportCmd.Flags().AddFlagSet(v3ExportFlagSet())
	devicesExportAllCmd.Flags().AddFlagSet(v3ExportFlagSet())
	devicesExportAllCmd.Flags().Bool("with-frame-counters", false, "Export frame counters for device. Note that this can take a long time.")
}

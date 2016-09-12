// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"strings"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

// devicesPersonalizeCmd represents the `device personalize` command
var devicesPersonalizeCmd = &cobra.Command{
	Use:   "personalize [Device ID] [NwkSKey] [AppSKey]",
	Short: "Personalize a device",
	Long:  `ttnctl devices personalize can be used to personalize a device (ABP).`,
	Run: func(cmd *cobra.Command, args []string) {

		var err error

		if len(args) == 0 {
			cmd.UsageFunc()(cmd)
			return
		}

		devID := args[0]
		if !api.ValidID(devID) {
			ctx.Fatalf("Invalid Device ID") // TODO: Add link to wiki explaining device IDs
		}

		appID := util.GetAppID(ctx)

		var nwkSKey types.NwkSKey
		if len(args) > 1 {
			nwkSKey, err = types.ParseNwkSKey(args[2])
			if err != nil {
				ctx.Fatalf("Invalid NwkSKey: %s", err)
			}
		} else {
			ctx.Info("Generating random NwkSKey...")
			copy(nwkSKey[:], random.Bytes(16))
		}

		var appSKey types.AppSKey
		if len(args) > 2 {
			appSKey, err = types.ParseAppSKey(args[3])
			if err != nil {
				ctx.Fatalf("Invalid AppSKey: %s", err)
			}
		} else {
			ctx.Info("Generating random AppSKey...")
			copy(appSKey[:], random.Bytes(16))
		}

		conn, manager := util.GetHandlerManager(ctx)
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

		ctx.WithFields(log.Fields{
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

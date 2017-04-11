// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var devicesDeleteCmd = &cobra.Command{
	Use:   "delete [Device ID]",
	Short: "Delete a device",
	Long:  `ttnctl devices delete can be used to delete a device.`,
	Example: `$ ttnctl devices delete test
  INFO Using Application                        AppID=test
Are you sure you want to delete device test from application test?
> yes
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Deleted device                           AppID=test DevID=test
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		devID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(devID, "Device ID"); err != nil {
			ctx.Fatal(err.Error())
		}

		appID := util.GetAppID(ctx)

		if !confirm(fmt.Sprintf("Are you sure you want to delete device %s from application %s?", devID, appID)) {
			ctx.Info("Not doing anything")
			return
		}

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		err := manager.DeleteDevice(appID, devID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not delete device.")
		}

		ctx.WithFields(ttnlog.Fields{
			"AppID": appID,
			"DevID": devID,
		}).Info("Deleted device")
	},
}

func init() {
	devicesCmd.AddCommand(devicesDeleteCmd)
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var devicesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all devices for the current application",
	Long:    `ttnctl devices list can be used to list all devices for the current application.`,
	Example: `$ ttnctl devices list
  INFO Using Application                        AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...

DevID	AppEUI          	DevEUI          	DevAddr
test 	70B3D57EF0000024	0001D544B2936FCE	26001ADA

  INFO Listed 1 devices                         AppID=test
`,
	Run: func(cmd *cobra.Command, args []string) {

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		devices, err := manager.GetDevicesForApplication(appID, 0, 0)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get devices.")
		}

		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("DevID", "AppEUI", "DevEUI", "DevAddr")
		for _, dev := range devices {
			if lorawan := dev.GetLorawanDevice(); lorawan != nil {
				devAddr := lorawan.DevAddr
				if devAddr.IsEmpty() {
					devAddr = nil
				}
				table.AddRow(dev.DevId, lorawan.AppEui, lorawan.DevEui, devAddr)
			} else {
				table.AddRow(dev.DevId)
			}
		}

		fmt.Println()
		fmt.Println(table)
		fmt.Println()

		ctx.WithFields(ttnlog.Fields{
			"AppID": appID,
		}).Infof("Listed %d devices", len(devices))
	},
}

func init() {
	devicesCmd.AddCommand(devicesListCmd)
}

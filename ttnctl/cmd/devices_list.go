// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/apex/log"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

// devicesListCmd represents the `device list` command
var devicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List al devices for the current application",
	Long:  `ttnctl devices list can be used to list all devices for the current application.`,
	Example: `$ ttnctl devices list
  INFO Using Application                        AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...

DevID	AppEUI          	DevEUI          	DevAddr 	Up/Down
test 	70B3D57EF0000024	0001D544B2936FCE	26001ADA	0/0

  INFO Listed 1 devices                         AppID=test
`,
	Run: func(cmd *cobra.Command, args []string) {

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx)
		defer conn.Close()

		devices, err := manager.GetDevicesForApplication(appID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get devices.")
		}

		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("DevID", "AppEUI", "DevEUI", "DevAddr", "Up/Down")
		for _, dev := range devices {
			if lorawan := dev.GetLorawanDevice(); lorawan != nil {
				devAddr := lorawan.DevAddr
				if devAddr.IsEmpty() {
					devAddr = nil
				}
				table.AddRow(dev.DevId, lorawan.AppEui, lorawan.DevEui, devAddr, fmt.Sprintf("%d/%d", lorawan.FCntUp, lorawan.FCntDown))
			} else {
				table.AddRow(dev.DevId)
			}
		}

		fmt.Println()
		fmt.Println(table)
		fmt.Println()

		ctx.WithFields(log.Fields{
			"AppID": appID,
		}).Infof("Listed %d devices", len(devices))
	},
}

func init() {
	devicesCmd.AddCommand(devicesListCmd)
}

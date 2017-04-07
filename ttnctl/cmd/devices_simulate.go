// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"strings"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var devicesSimulateCmd = &cobra.Command{
	Use:   "simulate [Device ID] [Payload]",
	Short: "Simulate uplink for a device",
	Long:  `ttnctl devices simulate can be used to simulate an uplink message for a device.`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)

		devID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(devID, "Device ID"); err != nil {
			ctx.Fatal(err.Error())
		}

		appID := util.GetAppID(ctx)

		port, err := cmd.Flags().GetUint32("port")
		if err != nil {
			ctx.WithError(err).Error("Failed to read port flag")
			return
		}

		payload, err := types.ParseHEX(args[1], len(args[1])/2)
		if err != nil {
			ctx.WithError(err).Error("Invalid Payload")
			return
		}

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		err = manager.SimulateUplink(appID, devID, port, payload)
		if err != nil {
			ctx.WithError(err).Error("Simulate failed")
			return
		}

		ctx.Info("Simulated uplink sent")

	},
}

func init() {
	devicesCmd.AddCommand(devicesSimulateCmd)
	devicesSimulateCmd.Flags().Uint32("port", 1, "Port number")
}

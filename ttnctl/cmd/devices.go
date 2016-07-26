// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// devicesCmd is the entrypoint for handlerctl
var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage devices",
	Long:  `ttnctl devices can be used to manage devices.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)
		ctx.WithFields(log.Fields{
			"AppID":  viper.GetString("app-id"),
			"AppEUI": viper.GetString("app-eui"),
		}).Info("Using Application")
	},
}

func init() {
	RootCmd.AddCommand(devicesCmd)
	devicesCmd.PersistentFlags().String("app-id", "", "The app ID to use")
	viper.BindPFlag("app-id", devicesCmd.PersistentFlags().Lookup("app-id"))
	devicesCmd.PersistentFlags().String("app-eui", "", "The app EUI to use")
	viper.BindPFlag("app-eui", devicesCmd.PersistentFlags().Lookup("app-eui"))
}

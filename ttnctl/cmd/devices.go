// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var devicesCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"device"},
	Short:   "Manage devices",
	Long:    `ttnctl devices can be used to manage devices.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)
		util.GetAccount(ctx)
		ctx.WithFields(ttnlog.Fields{
			"AppID":  util.GetAppID(ctx),
			"AppEUI": util.GetAppEUI(ctx),
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

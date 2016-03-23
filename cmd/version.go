// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get build and version information",
	Long:  `ttn version gets the build and version information of ttn`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx.WithFields(log.Fields{
			"commit":     viper.GetString("gitCommit"),
			"build date": viper.GetString("buildDate"),
		}).Infof("You are running %s of ttn.", viper.GetString("version"))
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

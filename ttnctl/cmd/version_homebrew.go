// +build homebrew

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get build and version information",
	Long:  `ttnctl version gets the build and version information of ttnctl`,
	Run: func(cmd *cobra.Command, args []string) {
		gitCommit := viper.GetString("gitCommit")
		buildDate := viper.GetString("buildDate")
		ctx.WithFields(ttnlog.Fields{
			"Version":   viper.GetString("version") + "-homebrew",
			"Commit":    gitCommit,
			"BuildDate": buildDate,
		}).Info("Got build information")
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

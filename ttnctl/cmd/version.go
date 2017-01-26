// +build !homebrew

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/utils/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const unknown = "unknown"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get build and version information",
	Long:  `ttnctl version gets the build and version information of ttnctl`,
	Run: func(cmd *cobra.Command, args []string) {
		gitBranch := viper.GetString("gitBranch")
		gitCommit := viper.GetString("gitCommit")
		buildDate := viper.GetString("buildDate")

		ctx.WithFields(ttnlog.Fields{
			"Version":   viper.GetString("version"),
			"Branch":    gitBranch,
			"Commit":    gitCommit,
			"BuildDate": buildDate,
		}).Info("Got build information")

		if gitBranch == unknown || gitCommit == unknown || buildDate == unknown {
			ctx.Warn("This is not an official ttnctl build")
			ctx.Warn("If you're building ttnctl from source, you should use the Makefile")
			return
		}

		latest, err := version.GetLatestInfo()
		if err != nil {
			ctx.WithError(err).Warn("Could not get latest version information")
			return
		}

		if latest.Commit == gitCommit {
			ctx.Info("This is an up-to-date ttnctl build")
			return
		}

		if buildDate, err := time.Parse(time.RFC3339, buildDate); err == nil {
			ctx.Warn("This is not the latest official ttnctl build")
			if buildDate.Before(latest.Date) {
				ctx.Warnf("The newest ttnctl build is %s newer.", latest.Date.Sub(buildDate))
			} else {
				ctx.Warn("Your ttnctl build is newer than the latest official one, which is fine if you're a developer")
			}
		} else {
			ctx.Warn("This ttnctl contains invalid build information")
		}

	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

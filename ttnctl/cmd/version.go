// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"time"

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
		ctx = ctx.WithField("Version", viper.GetString("version"))

		gitBranch := viper.GetString("gitBranch")
		ctx = ctx.WithField("Branch", gitBranch)

		gitCommit := viper.GetString("gitCommit")
		ctx = ctx.WithField("Commit", gitCommit)

		buildDate := viper.GetString("buildDate")
		ctx = ctx.WithField("BuildDate", buildDate)

		if gitBranch == unknown || gitCommit == unknown || buildDate == unknown {
			ctx.Warn("This is not an official ttnctl build")
		}

		if gitBranch != unknown {
			if version, err := version.GetLatestInfo(); err == nil {
				if version.Commit == gitCommit {
					ctx.Info("This is an up-to-date ttnctl build")
				} else {
					if buildDate, err := time.Parse(time.RFC3339, buildDate); err == nil {
						if buildDate.Before(version.Date) {
							ctx.Warn("This is not an up-to-date ttnctl build")
							ctx.Warnf("The newest build is %s newer.", version.Date.Sub(buildDate))
						} else {
							ctx.Warn("This is not an official ttnctl build")
						}
					}
				}
			} else {
				ctx.Warn("Could not get latest version information")
			}
		}

		ctx.Info("")
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

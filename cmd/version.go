// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/utils/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const unknown = "unknown"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get build and version information",
	Long:  `ttn version gets the build and version information of ttn`,
	Run: func(cmd *cobra.Command, args []string) {
		gitBranch := viper.GetString("gitBranch")
		gitCommit := viper.GetString("gitCommit")
		buildDate := viper.GetString("buildDate")

		ctx.WithFields(log.Fields{
			"Version":   viper.GetString("version"),
			"Branch":    gitBranch,
			"Commit":    gitCommit,
			"BuildDate": buildDate,
		}).Info("Got build information")

		if gitBranch == unknown || gitCommit == unknown || buildDate == unknown {
			ctx.Warn("This is not an official ttn build")
			ctx.Warn("If you're building ttn from source, you should use the Makefile")
			return
		}

		latest, err := version.GetLatestInfo()
		if err != nil {
			ctx.WithError(err).Warn("Could not get latest version information")
			return
		}

		if latest.Commit == gitCommit {
			ctx.Info("This is an up-to-date ttn build")
			return
		}

		if buildDate, err := time.Parse(time.RFC3339, buildDate); err == nil {
			ctx.Warn("This is not the latest official ttn build")
			if buildDate.Before(latest.Date) {
				ctx.Warnf("The newest ttn build is %s newer.", latest.Date.Sub(buildDate))
			} else {
				ctx.Warn("Your ttn build is newer than the latest official one, which is fine if you're a developer")
			}
		} else {
			ctx.Warn("This ttn contains invalid build information")
		}

	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

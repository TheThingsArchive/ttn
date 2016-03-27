// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationsCmd represents the applications command
var applicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "Show applications",
	Long:  `ttnctl applications retrieves the applications of the logged on user.`,
	Run: func(cmd *cobra.Command, args []string) {
		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found. Please login")
		}

	},
}

func init() {
	RootCmd.AddCommand(applicationsCmd)
}

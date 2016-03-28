// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

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
		server := viper.GetString("ttn-account-server")
		_, err := util.NewRequestWithAuth(server, "GET", fmt.Sprintf("%s/applications", server), nil)
		if err != nil {
			ctx.WithError(err).Fatal("Create request failed")
		}
	},
}

func init() {
	RootCmd.AddCommand(applicationsCmd)
}

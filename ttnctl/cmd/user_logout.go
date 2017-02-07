// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"os"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var userLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout the current user",
	Long:  `ttnctl user logout logs out the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 0, 0)

		err := util.Logout()
		if err != nil {
			if os.IsNotExist(err) {
				ctx.Info("You were not logged in")
				return
			}
			ctx.WithError(err).Fatal("Could not delete credentials")
		}
	},
}

func init() {
	userCmd.AddCommand(userLogoutCmd)
}

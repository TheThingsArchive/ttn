// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an application",
	Long:  `ttnctl devices delete can be used to delete an application.`,
	Run: func(cmd *cobra.Command, args []string) {
		account := util.GetAccount(ctx)
		appID := util.GetAppID(ctx)

		if !confirm(fmt.Sprintf("Are you sure you want to delete application %s?", appID)) {
			ctx.Info("Not doing anything")
			return
		}

		err := account.DeleteApplication(appID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not delete application")
		}
		util.ForceRefreshToken(ctx)

		ctx.Info("Deleted Application")

	},
}

func init() {
	applicationsCmd.AddCommand(applicationsDeleteCmd)
}

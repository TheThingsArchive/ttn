// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var clientDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an OAuth client",
	Long:  "ttnctl clients delete removes an OAuth client.",
	Example: `$ ttnctl clients delete oauthclient1
  INFO OAuth client removed successfully OAuthClientName=oauthclient1
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		ctx = ctx.WithField("OAuthClientName", args[0])

		account := util.GetAccount(ctx)

		err := account.RemoveOAuthClient(args[0])
		if err != nil {
			ctx.WithError(err).Fatal("Failed to delete OAuth client")
		}

		ctx.Info("OAuth client removed successfully")
	},
}

func init() {
	clientsCmd.AddCommand(clientDeleteCmd)
}

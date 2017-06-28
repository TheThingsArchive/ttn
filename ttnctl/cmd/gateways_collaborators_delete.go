// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var gatewaysCollaboratorsDeleteCmd = &cobra.Command{
	Use:     "delete [AppID] [Username]",
	Aliases: []string{"remove"},
	Short:   "Delete a collaborator from a gateway.",
	Long:    `gateways collaborators delete can be used to delete a collaborator from a gateway.`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)
		account := util.GetAccount(ctx)
		ctx.Infof("Removing %s's rights on gateway %s...", args[1], args[0])
		if err := account.RetractGatewayRights(args[0], args[1]); err != nil {
			ctx.WithError(err).Fatal("Could not delete user's gateway rights")
		}
		ctx.Info("Successfully deleted user's gateway rights")
	},
}

func init() {
	gatewaysCollaboratorsCmd.AddCommand(gatewaysCollaboratorsDeleteCmd)
}

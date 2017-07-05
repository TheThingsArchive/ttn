// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsCollaboratorsDeleteCmd = &cobra.Command{
	Use:     "delete [AppID] [Username]",
	Aliases: []string{"remove"},
	Short:   "Delete a collaborator from an application.",
	Long:    `applications collaborators delete can be used to delete a collaborator from an application.`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)
		account := util.GetAccount(ctx)
		ctx.Infof("Removing %s's rights on application %s...", args[1], args[0])
		if err := account.Retract(args[0], args[1]); err != nil {
			ctx.WithError(err).Fatal("Could not delete user's application rights")
		}
		ctx.Info("Successfully deleted user's application rights")
	},
}

func init() {
	applicationsCollaboratorsCmd.AddCommand(applicationsCollaboratorsDeleteCmd)
}

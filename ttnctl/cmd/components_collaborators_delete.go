// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var componentsCollaboratorsDeleteCmd = &cobra.Command{
	Use:     "delete [Type] [ComponentID] [Username]",
	Aliases: []string{"remove"},
	Short:   "Delete a collaborator from a component.",
	Long:    `components collaborators delete can be used to delete a collaborator from a component.`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 3, 3)
		account := util.GetAccount(ctx)
		ctx.Infof("Removing %s's rights on %s %s...", args[2], args[0], args[1])
		if err := account.RetractComponentRights(args[0], args[1], args[2]); err != nil {
			ctx.WithError(err).Fatal("Could not delete user's component rights")
		}
		ctx.Info("Successfully deleted user's component rights")
	},
}

func init() {
	componentsCollaboratorsCmd.AddCommand(componentsCollaboratorsDeleteCmd)
}

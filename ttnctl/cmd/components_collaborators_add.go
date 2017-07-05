// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var componentsCollaboratorsAddCmd = &cobra.Command{
	Use:   "add [Type] [ComponentID] [Username] [Rights...]",
	Short: "Add a collaborator to a component.",
	Long: `components collaborators add can be used to add a collaborator to a component.
Available rights are: ` + joinRights(componentsRights, ", "),
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 3, 0)
		account := util.GetAccount(ctx)
		var rights []types.Right
		if len(args) > 3 {
			for _, right := range args[3:] {
				right := types.Right(right)
				if validRight(componentsRights, right) {
					rights = append(rights, right)
				} else {
					ctx.Warnf(`Right "%s" is invalid and will be ignored`, right)
				}
			}
		} else {
			ctx.Info("No rights supplied, will grant same rights as current user")
			user, err := account.Profile()
			if err != nil {
				ctx.WithError(err).Fatal("Could not get current user")
			}
			component, err := account.FindComponent(args[0], args[1])
			if err != nil {
				ctx.WithError(err).Fatal("Could not get component")
			}
			for _, collaborator := range component.Collaborators {
				if collaborator.Username == user.Username {
					rights = collaborator.Rights
					break
				}
			}
			rights = componentsRights
		}
		if len(rights) == 0 {
			ctx.Fatal("No list of rights supplied. Available rights are: " + joinRights(componentsRights, ", "))
		}
		ctx.Infof("Adding %d rights to user %s on %s %s...", len(rights), args[2], args[0], args[1])
		if err := account.GrantComponentRights(args[0], args[1], args[2], rights); err != nil {
			ctx.WithError(err).Fatal("Could not add component rights to user")
		}
		ctx.Info("Successfully added component rights to user")
	},
}

func init() {
	componentsCollaboratorsCmd.AddCommand(componentsCollaboratorsAddCmd)
}

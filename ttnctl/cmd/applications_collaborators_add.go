// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsCollaboratorsAddCmd = &cobra.Command{
	Use:   "add [AppID] [Username] [Rights...]",
	Short: "Add a collaborator to an application.",
	Long: `applications collaborators add can be used to add a collaborator to an application.
Available rights are: ` + joinRights(applicationRights, ", "),
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 0)
		account := util.GetAccount(ctx)
		var rights []types.Right
		if len(args) > 2 {
			for _, right := range args[2:] {
				right := types.Right(right)
				if validRight(applicationRights, right) {
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
			app, err := account.FindApplication(args[0])
			if err != nil {
				ctx.WithError(err).Fatal("Could not get application")
			}
			for _, collaborator := range app.Collaborators {
				if collaborator.Username == user.Username {
					rights = collaborator.Rights
					break
				}
			}
		}
		if len(rights) == 0 {
			ctx.Fatal("No list of rights supplied. Available rights are: " + joinRights(applicationRights, ", "))
		}
		ctx.Infof("Adding %d rights to user %s on application %s...", len(rights), args[1], args[0])
		if err := account.Grant(args[0], args[1], rights); err != nil {
			ctx.WithError(err).Fatal("Could not add application rights to user")
		}
		ctx.Info("Successfully added application rights to user")
	},
}

func init() {
	applicationsCollaboratorsCmd.AddCommand(applicationsCollaboratorsAddCmd)
}

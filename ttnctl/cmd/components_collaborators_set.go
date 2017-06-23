// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var componentsCollaboratorsSetCmd = &cobra.Command{
	Use:   "set [Type] [ComponentID] [Username]",
	Short: "Grant an user with rights over a network component.",
	Long:  `components collaborators set can be used to add an username as collaborator or edit his rights.`,
	Example: `$ ttnctl components collaborators set handler ttn-handler-dev gomezjdaniel --rights='component:settings,component:delete'
  INFO Successfully user granted with collaborator rights ComponentID=ttn-dev-handler Type=handler Username=gomezjdaniel
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 3, 3)

		var rights []types.Right
		rightsStrings, err := cmd.Flags().GetStringSlice("rights")
		if err != nil {
			ctx.WithError(err).Fatal("Could not read rights options")
		}
		if len(rightsStrings) == 0 {
			ctx.Fatal("Rights has not been provided through --rights flag")
		}

		for _, rightString := range rightsStrings {
			rights = append(rights, types.Right(rightString))
		}

		account := util.GetAccount(ctx)

		if err := account.GrantComponentRights(args[0], args[1], args[2], rights); err != nil {
			ctx.WithError(err).Fatal("Could not grant user with collaborator rights")
		}

		ctx.WithFields(ttnlog.Fields{
			"Type":        args[0],
			"ComponentID": args[1],
			"Username":    args[2],
		}).Info("Successfully user granted with component rights")
	},
}

func init() {
	componentsCollaboratorsCmd.AddCommand(componentsCollaboratorsSetCmd)
	componentsCollaboratorsSetCmd.Flags().StringSlice("rights", []string{}, "Rights to be granted")
}

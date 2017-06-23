// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var componentsCollaboratorsDeleteCmd = &cobra.Command{
	Use:   "delete [Type] [ComponentID] [Username]",
	Short: "Retracts an user as component collaborator.",
	Long:  `components collaborators delete can be used to retract users as collaborators.`,
	Example: `$ ttnctl components collaborators delete handler ttn-handler-dev gomezjdaniel
  INFO Successfully user retracted as component collaborator ComponentID=dev-handler Type=handler Username=gomezjdaniel
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 3, 3)

		account := util.GetAccount(ctx)

		if err := account.RetractComponentRights(args[0], args[1], args[2]); err != nil {
			ctx.WithError(err).Fatal("Could not retract user as component collaborator")
		}

		ctx.WithFields(ttnlog.Fields{
			"Type":        args[0],
			"ComponentID": args[1],
			"Username":    args[2],
		}).Info("Successfully user retracted as component collaborator")
	},
}

func init() {
	componentsCollaboratorsCmd.AddCommand(componentsCollaboratorsDeleteCmd)
}

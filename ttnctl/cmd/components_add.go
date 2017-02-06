// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var componentsAddCmd = &cobra.Command{
	Use:   "add [Type] [ComponentID]",
	Short: "Add a new network component",
	Long:  `ttnctl components add can be used to add a new network component.`,
	Example: `$ ttnctld components add handler test                                                                                                                                             146 !
  INFO Added network component                  id=test type=handler
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)

		account := util.GetAccount(ctx)

		err := account.CreateComponent(args[0], args[1])
		if err != nil {
			ctx.WithError(err).WithField("type", args[0]).WithField("id", args[1]).Fatal("Could not add component")
		}

		util.ForceRefreshToken(ctx)

		ctx.WithField("type", args[0]).WithField("id", args[1]).Info("Added network component")
	},
}

func init() {
	componentsCmd.AddCommand(componentsAddCmd)
}

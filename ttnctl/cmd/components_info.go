// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var componentsInfoCmd = &cobra.Command{
	Use:   "info [Type] [ComponentID]",
	Short: "Get information about a network component.",
	Long:  `components info can be used to retrieve information about a network component.`,
	Example: `$ ttnctl components info handler test
  INFO Found network component

Component ID:   test
Type:           handler
Created:        2016-10-06 09:52:28.766 +0000 UTC
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)

		account := util.GetAccount(ctx)

		component, err := account.FindComponent(args[0], args[1])
		if err != nil {
			ctx.WithError(err).WithField("type", args[0]).WithField("id", args[1]).Fatal("Could not find component")
		}

		ctx.Info("Found network component")

		fmt.Println()
		fmt.Printf("Component ID:   %s\n", component.ID)
		fmt.Printf("Type:           %s\n", component.Type)
		fmt.Printf("Created:        %s\n", component.Created)
		fmt.Println()
	},
}

func init() {
	componentsCmd.AddCommand(componentsInfoCmd)
}

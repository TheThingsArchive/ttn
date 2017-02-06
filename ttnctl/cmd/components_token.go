// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var componentsTokenCmd = &cobra.Command{
	Use:   "token [Type] [ComponentID]",
	Short: "Get the token for a network component.",
	Long:  `components token gets a signed token for the component.`,
	Example: `$ ttnctld components token handler test                                                                                                                                           146 !
  INFO Got component token                      id=test type=handler

eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJ0dG4tYWNjb3VudCIsInN1YiI6InRlc3QxyzJ0eXBlIjoiaGFuZGxlciIsImlhdCI6MTQ3NTc0NzY3MywiZXhwIjoxNDgzNzgyODczfQ.Bf6Gy6xTE2m7fkYSd4WHs3UgRaAEXkox2jjJeaBahVNU365n_wI4_oWX_B3mkMOa1ZL3IB2JagAybo50mTApPtnGiRjDczGjqkkbBiXPcwA8SvmyKTKNkPkrpzGIioq9itjpYDuMJixgLh4gYlK0B_1jkH23ZFoslzn7WfYYe3AKC0JZAhePgQygJ2Zn3w6cGZOqgRvblIIcGynSEqqP3aKyKRhtnwofao-w-jzWqINGvAcMt1iW7JN3hX9yW4IXRicB4_-L0Aaq1sqvRpoh8z9SmpkkE8oBmWqPsUAXTECuoYc4kezjGcDg4YnBfBQtT-itPTfdb8-vq2izxyztsw
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)

		account := util.GetAccount(ctx)

		token, err := account.ComponentToken(args[0], args[1])
		if err != nil {
			ctx.WithError(err).WithField("type", args[0]).WithField("id", args[1]).Fatal("Could not get component token")
		}

		ctx.WithField("type", args[0]).WithField("id", args[1]).Info("Got component token")

		fmt.Println()
		fmt.Println(token)
		fmt.Println()
	},
}

func init() {
	componentsCmd.AddCommand(componentsTokenCmd)
}

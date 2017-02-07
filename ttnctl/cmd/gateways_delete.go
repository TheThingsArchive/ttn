// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var gatewaysDeleteCmd = &cobra.Command{
	Use:   "delete [GatewayID]",
	Short: "Delete a gateway",
	Long:  `ttnctl gateways delete can be used to delete a gateway`,
	Example: `$ ttnctl gateways delete test
  INFO Deleted gateway                          Gateway ID=test
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		gatewayID := args[0]
		if !api.ValidID(gatewayID) {
			ctx.Fatal("Invalid Gateway ID")
		}

		account := util.GetAccount(ctx)
		err := account.DeleteGateway(gatewayID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not list gateways")
		}

		ctx.WithField("Gateway ID", gatewayID).Info("Deleted gateway")
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysDeleteCmd)
}

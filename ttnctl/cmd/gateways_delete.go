// Copyright © 2016 The Things Network
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
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return
		}

		gatewayID := args[0]
		if !api.ValidID(gatewayID) {
			ctx.Fatal("Invalid Gateway ID")
		}

		act := util.GetAccount(ctx)
		err := act.DeleteGateway(gatewayID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not list gateways")
		}

		ctx.WithField("Gateway ID", gatewayID).Info("Deleted gateway")
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysDeleteCmd)
}

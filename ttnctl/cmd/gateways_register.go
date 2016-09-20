// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var gatewaysRegisterCmd = &cobra.Command{
	Use:   "register [GatewayID] [FrequencyPlan]",
	Short: "Register a gateway",
	Long:  `ttnctl gateways register can be used to register a gateway`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.UsageFunc()(cmd)
			return
		}

		gatewayID := args[0]
		frequencyPlan := args[1]

		act := util.GetAccount(ctx)
		gateway, err := act.RegisterGateway(gatewayID, frequencyPlan, nil)
		if err != nil {
			ctx.WithError(err).Fatal("Could not register gateway")
		}

		util.ForceRefreshToken(ctx)

		ctx.WithField("Gateway ID", gateway.ID).Info("Registered Gateway")
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysRegisterCmd)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/account"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var gatewaysRegisterCmd = &cobra.Command{
	Use:   "register [GatewayID] [FrequencyPlan] [Location]",
	Short: "Register a gateway",
	Long:  `ttnctl gateways register can be used to register a gateway`,
	Example: `$ ttnctl gateways register test US 52.37403,4.88968
  INFO Registered gateway                          Gateway ID=test
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 && len(args) != 3 {
			cmd.UsageFunc()(cmd)
			return
		}

		gatewayID := args[0]
		if !api.ValidID(gatewayID) {
			ctx.Fatal("Invalid Gateway ID")
		}

		frequencyPlan := args[1]

		var err error
		var location *account.Location
		if len(args) == 3 {
			location, err = util.ParseLocation(args[2])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid location")
			}
		}

		act := util.GetAccount(ctx)
		gateway, err := act.RegisterGateway(gatewayID, frequencyPlan, location)
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

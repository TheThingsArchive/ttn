// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"strings"

	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/ttn/api"
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
		assertArgsLength(cmd, args, 2, 3)

		gatewayID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(gatewayID, "Gateway ID"); err != nil {
			ctx.Fatal(err.Error())
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

		settings := account.GatewaySettings{
			AntennaLocation: location,
		}

		act := util.GetAccount(ctx)
		gateway, err := act.RegisterGateway(gatewayID, frequencyPlan, settings)
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

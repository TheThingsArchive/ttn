// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var gatewaysEditCmd = &cobra.Command{
	Use:   "edit [GatewayID]",
	Short: "Edit a gateway",
	Long:  `ttnctl gateways edit can be used to edit settings of a gateway`,
	Example: `$ ttnctl gateways edit test --location 52.37403,4.88968 --frequency-plan EU
  INFO Edited gateway                          Gateway ID=test
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		gatewayID := args[0]
		if !api.ValidID(gatewayID) {
			ctx.Fatal("Invalid Gateway ID")
		}

		var edits account.GatewayEdits

		frequencyPlan, err := cmd.Flags().GetString("frequency-plan")
		if err != nil {
			ctx.WithError(err).Fatal("Invalid frequency-plan")
		}

		if frequencyPlan != "" {
			edits.FrequencyPlan = frequencyPlan
		}

		locationStr, err := cmd.Flags().GetString("location")
		if err != nil {
			ctx.WithError(err).Fatal("Invalid location")
		}

		if locationStr != "" {
			location, err := util.ParseLocation(locationStr)
			if err != nil {
				ctx.WithError(err).Fatal("Invalid location")
			}
			edits.Location = location
		}

		act := util.GetAccount(ctx)
		err = act.EditGateway(gatewayID, edits)
		if err != nil {
			ctx.WithError(err).Fatal("Failure editing gateway")
		}

		ctx.WithField("Gateway ID", gatewayID).Info("Edited gateway")
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysEditCmd)
	gatewaysEditCmd.Flags().String("frequency-plan", "", "The frequency plan to use on the gateway")
	gatewaysEditCmd.Flags().String("location", "", "The location of the gateway")
}

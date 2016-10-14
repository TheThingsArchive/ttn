// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var gatewaysInfoCmd = &cobra.Command{
	Use:   "info [GatewayID]",
	Short: "get info about a gateway",
	Long:  `ttnctl gateways info can be used to get information about a gateway`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return
		}

		gatewayID := args[0]
		if !api.ValidID(gatewayID) {
			ctx.Fatal("Invalid Gateway ID")
		}

		account := util.GetAccount(ctx)

		gateway, err := account.FindGateway(gatewayID)
		if err != nil {
			ctx.WithError(err).WithField("id", gatewayID).Fatal("Could not find gateway")
		}

		ctx.Info("Found gateway")

		fmt.Println()
		fmt.Printf("Gateway ID:     %s\n", gateway.ID)
		fmt.Printf("Activated:      %v\n", gateway.Activated)
		fmt.Printf("Frequency Plan: %s\n", gateway.FrequencyPlan)
		locationAccess := "private"
		if gateway.LocationPublic {
			locationAccess = "public"
		}
		if gateway.Location != nil {
			fmt.Printf("Location Info  : (%f, %f) (%s) \n", gateway.Location.Latitude, gateway.Location.Longitude, locationAccess)
		}
		if gateway.StatusPublic {
			fmt.Printf("Status Info:    public (see ttnctl gateways status %s)\n", gatewayID)
		} else {
			fmt.Print("Status Info:    private\n")
		}

		fmt.Println()
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysInfoCmd)
}

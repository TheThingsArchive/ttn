// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var gatewaysInfoCmd = &cobra.Command{
	Use:   "info [GatewayID]",
	Short: "Get info about a gateway",
	Long:  `ttnctl gateways info can be used to get information about a gateway`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		gatewayID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(gatewayID, "Gateway ID"); err != nil {
			ctx.Fatal(err.Error())
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

		if gateway.AntennaLocation != nil {
			fmt.Printf("Location Info  : (%f, %f, %f) (%s) \n", gateway.AntennaLocation.Latitude, gateway.AntennaLocation.Longitude, gateway.AntennaLocation.Altitude, locationAccess)
		}

		if gateway.StatusPublic {
			fmt.Printf("Status Info:    public (see ttnctl gateways status %s)\n", gatewayID)
		} else {
			fmt.Print("Status Info:    private\n")
		}
		if gateway.Key != "" {
			fmt.Printf("Access Key    : %s\n", gateway.Key)
		}

		fmt.Println()
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysInfoCmd)
}

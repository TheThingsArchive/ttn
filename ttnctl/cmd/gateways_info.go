// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/api"
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
		printKV("Gateway ID", gateway.ID)
		printKV("Frequency Plan", gateway.FrequencyPlan)
		if gateway.Router != nil {
			printKV("Router", gateway.Router.ID)
		}
		printBool("Auto Update", gateway.AutoUpdate, "on", "off")
		printKV("Owner", gateway.Owner.Username)
		printBool("Owner Public", gateway.OwnerPublic, "yes", "no")
		if gateway.AntennaLocation != nil {
			printKV("Location", fmt.Sprintf("(%f, %f, %d)", gateway.AntennaLocation.Latitude, gateway.AntennaLocation.Longitude, gateway.AntennaLocation.Altitude))
		}
		printBool("Location Public", gateway.LocationPublic, "yes", "no")
		printBool("Status Public", gateway.StatusPublic, "yes", "no")

		fmt.Println()

		printKV("Brand", gateway.Attributes.Brand)
		printKV("Model", gateway.Attributes.Model)
		printKV("Placement", gateway.Attributes.Placement)
		printKV("AntennaType", gateway.Attributes.AntennaType)
		printKV("AntennaModel", gateway.Attributes.AntennaModel)
		printKV("Description", gateway.Attributes.Description)

		if gateway.Key != "" {
			printKV("Access Key", gateway.Key)
		}

		if len(gateway.Collaborators) > 0 {
			fmt.Println()
			fmt.Println("Collaborators:")
			printCollaborators(gateway.Collaborators)
		}
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysInfoCmd)
}

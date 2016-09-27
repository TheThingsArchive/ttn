// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var gatewaysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your gateways",
	Long:  `ttnctl gateways list can be used to list the gateways you have access to`,
	Example: `$ ttnctl gateways list
 	ID  	Activated	Frequency Plan	Coordinates
1	test	true		US				(52.3740, 4.8896)
`,
	Run: func(cmd *cobra.Command, args []string) {
		account := util.GetAccount(ctx)
		gateways, err := account.ListGateways()
		if err != nil {
			ctx.WithError(err).Fatal("Could not list gateways")
		}

		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("", "ID", "Activated", "Frequency Plan", "Coordinates")
		for i, gateway := range gateways {
			var lat float64
			var lng float64
			if gateway.Location != nil {
				lat = gateway.Location.Latitude
				lng = gateway.Location.Longitude
			}
			table.AddRow(i+1, gateway.ID, gateway.Activated, gateway.FrequencyPlan, fmt.Sprintf("(%f, %f)", lat, lng))
		}

		fmt.Println()
		fmt.Println(table)
		fmt.Println()
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysListCmd)
}

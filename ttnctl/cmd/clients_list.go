// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var clientListCmd = &cobra.Command{
	Use:   "list",
	Short: "List OAuth clients",
	Long:  "ttnctl clients list fetches and shows all the owned OAuth clients",
	Example: `$ ttnctl clients list
  INFO Found one OAuth client:

        Name                    Description             URI                                                          Secret                                                          Scope                        Grants
0       ttn-console             TTN Console			    https://console.thethingsnetwork.org/oauth/callback          156a7c34fbd156d333f53b08ab6015b20b558818148a901927bb3f5b19fe    [profile apps gateways]      [authorization_code refresh_token]
`,
	Run: func(cmd *cobra.Command, args []string) {
		account := util.GetAccount(ctx)

		clients, err := account.ListOAuthClients()
		if err != nil {
			ctx.WithError(err).Fatal("Failed to fetch the OAuth clients")
		}

		switch len(clients) {
		case 0:
			ctx.Info("You don't have any OAuth client")
			return
		case 1:
			ctx.Info("Found one OAuth client:")
		default:
			ctx.Infof("Found %d applications:", len(clients))
		}

		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("", "Name", "Description", "URI", "Secret", "Scope", "Grants")
		for i, client := range clients {
			table.AddRow(i, client.Name, client.Description, client.URI, client.Secret, client.Scope, client.Grants)
		}

		fmt.Println()
		fmt.Println(table)
		fmt.Println()
	},
}

func init() {
	clientsCmd.AddCommand(clientListCmd)
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"sort"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var clientListCmd = &cobra.Command{
	Use:   "list",
	Short: "List OAuth clients",
	Long:  "ttnctl clients list fetches and shows all the owned OAuth clients.",
	Example: `$ ttnctl clients list
  INFO Found 2 OAuth clients:

Name                    URI                                                          Secret                                                          Scope                        Grants							    Status
my-apps-client          https://myttnclient.org/apps/oauth/callback                  b08ab6015b20b558818148a90015b20b558818148a901927bb3f5b19fefd    [profile apps]               [authorization_code refresh_token]    Pending
my-gateway-client       https://myttnclient.org/gateways/oauth/callback              156a7c34fbd156d333f53b08ab6015b20b558818148a901927bb3f5b19fe    [profile gateways]           [authorization_code refresh_token]    Accepted
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
			ctx.Infof("Found %d OAuth clients:", len(clients))
			sort.Slice(clients, func(i, j int) bool { return clients[i].Name < clients[j].Name })
		}

		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("Name", "URI", "Secret", "Scope", "Grants", "Status")
		for _, client := range clients {
			status := "Pending"
			if client.Rejected {
				status = "Rejected"
			} else if client.Accepted {
				status = "Accepted"
			}
			table.AddRow(client.Name, client.URI, client.Secret, client.Scope, client.Grants, status)
		}

		fmt.Println()
		fmt.Println(table)
		fmt.Println()
	},
}

func init() {
	clientsCmd.AddCommand(clientListCmd)
}

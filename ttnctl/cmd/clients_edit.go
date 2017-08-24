// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var clientEditCmd = &cobra.Command{
	Use:   "edit [Name]",
	Short: "Edit the OAuth client",
	Long:  "ttnctl clients edit can be used to edit the OAuth client.",
	Example: `$ ttnctl clients edit my-gateway-client --description "OAuth client for my personal gateway client"
  INFO Retrieving OAuth client...               OAuthClientName=my-gateway-client
  INFO Retrieved OAuth client                   OAuthClientName=my-gateway-client
  INFO OAuth client edited                      OAuthClientName=my-gateway-client
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		var name = args[0]

		account := util.GetAccount(ctx)

		ctx = ctx.WithField("OAuthClientName", name)

		ctx.Info("Retrieving OAuth client...")
		client, err := account.FindOAuthClient(name)
		if err != nil {
			ctx.WithError(err).Fatal("Could not retrieve OAuth client")
		}
		if client == nil {
			ctx.Fatal("No OAuth client returned by the server")
		}
		ctx.Debug("Retrieved OAuth client")

		var edited bool
		if newDescription, err := cmd.Flags().GetString("description"); err != nil {
			ctx.WithError(err).Fatal("Flag error")
		} else if newDescription != "" {
			edited = true
			client.Description = newDescription
		}

		if newURI, err := cmd.Flags().GetString("uri"); err != nil {
			ctx.WithError(err).Fatal("Flag error")
		} else if newURI != "" {
			edited = true
			client.URI = newURI
		}

		if !edited {
			ctx.Info("No property to edit")
			return
		}

		err = account.EditOAuthClient(name, client)
		if err != nil {
			ctx.WithError(err).Fatal("Could not edit OAuth client")
		}

		ctx.Info("OAuth client edited")
	},
}

func init() {
	clientsCmd.AddCommand(clientEditCmd)
	clientEditCmd.Flags().String("description", "", "Edit the description of the OAuth client")
	clientEditCmd.Flags().String("uri", "", "Edit the URI of the OAuth client")
}

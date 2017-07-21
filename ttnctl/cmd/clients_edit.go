// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var clientEditCmd = &cobra.Command{
	Use:   "edit [Name]",
	Short: "Edit the general information of an existing client",
	Long:  "ttnctl clients edit can be used to edit the general information of an existing OAuth client",
	Example: `$ ttnctl clients edit oauthClient1 --name oauthclient-new-name --description "OAuth client used by the console"
  INFO Retrieving OAuth client...               OAuthClientID=oauthClient1
  INFO Retrieved OAuth client                   OAuthClientID=oauthClient1
  INFO OAuth client updated                     OAuthClientID=oauthclient-new-name
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		var name = args[0]

		account := util.GetAccount(ctx)

		clientCtx := ctx.WithField("OAuthClientID", name)

		// Verifying that there is information to edit
		var flagsSet bool
		cmd.Flags().Visit(func(*pflag.Flag) { flagsSet = true })
		if !flagsSet {
			clientCtx.Fatal("No information to edit")
		}

		clientCtx.Info("Retrieving OAuth client...")
		client, err := account.FindOAuthClient(name)
		if err != nil {
			clientCtx.WithError(err).Fatal("Could not find OAuth client")
		}
		if client == nil {
			clientCtx.Fatal("No OAuth client returned by the server")
		}
		clientCtx.Info("Retrieved OAuth client")

		if newName, err := cmd.Flags().GetString("name"); err != nil {
			clientCtx.WithError(err).Fatal("Couldn't parse new name")
		} else if newName != "" {
			client.Name = newName
			clientCtx = ctx.WithField("OAuthClientID", newName)
		}

		if newDescription, err := cmd.Flags().GetString("description"); err != nil {
			clientCtx.WithError(err).Fatal("Couldn't parse description")
		} else if newDescription != "" {
			client.Description = newDescription
		}

		if newURI, err := cmd.Flags().GetString("uri"); err != nil {
			clientCtx.WithError(err).Fatal("Couldn't parse URI")
		} else if newURI != "" {
			client.URI = newURI
		}

		err = account.EditOAuthClient(name, client)
		if err != nil {
			ctx.WithField("OAuthClientID", name).WithError(err).Fatal("Couldn't update OAuth client")
		}

		clientCtx.Info("OAuth client information updated")
	},
}

func init() {
	clientsCmd.AddCommand(clientEditCmd)
	clientEditCmd.Flags().String("name", "", "Edit the name of the OAuth client")
	clientEditCmd.Flags().String("description", "", "Edit the description of the OAuth client")
	clientEditCmd.Flags().String("uri", "", "Edit the URI of the OAuth client")
}

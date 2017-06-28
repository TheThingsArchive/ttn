// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsInfoCmd = &cobra.Command{
	Use:   "info [AppID]",
	Short: "Get information about an application",
	Long:  `ttnctl applications info can be used to info applications.`,
	Example: `$ ttnctl applications info
  INFO Found application

AppID:   test
Name:    Test application
EUIs:
       - 0000000000000000

Access Keys:
       - Name: default key
         Key:  FZYr01cUhdhY1KBiMghUl+/gXyqXhrF6y+1ww7+DzHg=
         Rights: messages:up:r, messages:down:w

Collaborators:
       - Name: yourname
         Rights: settings, delete, collaborators
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 0, 1)

		account := util.GetAccount(ctx)

		var appID string
		if len(args) == 1 {
			appID = args[0]
		} else {
			appID = util.GetAppID(ctx)
		}

		app, err := account.FindApplication(appID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not find application")
		}

		ctx.Info("Found application")

		fmt.Println()

		fmt.Printf("AppID:   %s\n", app.ID)
		fmt.Printf("Name:    %s\n", app.Name)

		fmt.Println()
		fmt.Println("EUIs:")
		for _, eui := range app.EUIs {
			fmt.Printf("       - %s\n", eui)
		}

		fmt.Println()
		fmt.Println("Access Keys:")
		for _, key := range app.AccessKeys {
			fmt.Printf("       - Name: %s\n", key.Name)
			fmt.Printf("         Key:  %s\n", key.Key)
			fmt.Print("         Rights: ")
			for i, right := range key.Rights {
				if i != 0 {
					fmt.Print(", ")
				}
				fmt.Print(right)
			}
			fmt.Println()
		}

		if len(app.Collaborators) > 0 {
			fmt.Println()
			fmt.Println("Collaborators:")
			printCollaborators(app.Collaborators)
		}
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsInfoCmd)
}

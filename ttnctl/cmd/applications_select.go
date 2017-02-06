// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var applicationsSelectCmd = &cobra.Command{
	Use:   "select [AppID [AppEUI]]",
	Short: "select the application to use",
	Long:  `ttnctl applications select can be used to select the application to use in next commands.`,
	Example: `$ ttnctl applications select
  INFO Found one application "test", selecting that one.
  INFO Found one EUI "0000000000000000", selecting that one.
  INFO Updated configuration
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 0, 2)

		account := util.GetAccount(ctx)

		apps, err := account.ListApplications()
		if err != nil {
			ctx.WithError(err).Fatal("Could not select applications")
		}

		var appIDArg string
		if len(args) > 0 {
			appIDArg = args[0]
		}

		var appEUIArg types.AppEUI
		if len(args) > 1 {
			appEUIArg, _ = types.ParseAppEUI(args[1])
		}

		var appIdx int
		var euiIdx int

	selectAppID:
		switch len(apps) {
		case 0:
			ctx.Info("You don't have any applications, not doing anything.")
			return
		case 1:
			ctx.Infof("Found one application \"%s\", selecting that one.", apps[0].ID)
		default:
			table := uitable.New()
			table.MaxColWidth = 70
			table.AddRow("", "ID", "Description")
			for i, app := range apps {
				table.AddRow(i+1, app.ID, app.Name)
				if appIDArg == app.ID {
					appIdx = i
					break selectAppID
				}
			}

			ctx.Infof("Found %d applications:", len(apps))

			fmt.Println()
			fmt.Println(table)
			fmt.Println()

			fmt.Println("Which one do you want to use?")
			fmt.Printf("Enter the number (1 - %d) > ", len(apps))
			fmt.Scanf("%d", &appIdx)
			appIdx--
		}

		if appIdx < 0 || appIdx >= len(apps) {
			ctx.Fatal("Invalid choice for application")
		}
		app := apps[appIdx]

	selectAppEUI:
		switch len(app.EUIs) {
		case 0:
			ctx.Info("You don't have any EUIs in your application")
		case 1:
			ctx.Infof("Found one EUI \"%s\", selecting that one.", app.EUIs[0])
		default:
			table := uitable.New()
			table.MaxColWidth = 70
			table.AddRow("", "EUI")
			for i, eui := range app.EUIs {
				table.AddRow(i+1, eui)
				if appEUIArg == eui {
					appIdx = i
					break selectAppEUI
				}
			}

			ctx.Infof("Found %d EUIs", len(app.EUIs))

			fmt.Println()
			fmt.Println(table)
			fmt.Println()

			fmt.Println("Which one do you want to use?")
			fmt.Printf("Enter the number (1 - %d) > ", len(app.EUIs))
			fmt.Scanf("%d", &euiIdx)
			euiIdx--
		}

		if euiIdx < 0 || euiIdx >= len(apps) {
			ctx.Fatal("Invalid choice for EUI")
		}
		eui := app.EUIs[euiIdx]

		util.SetApp(ctx, app.ID, eui)

		ctx.WithFields(ttnlog.Fields{
			"AppID":  app.ID,
			"AppEUI": eui,
		}).Info("Updated configuration")
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsSelectCmd)
}

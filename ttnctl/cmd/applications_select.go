// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

// applicationsSelectCmd is the entrypoint for handlerctl
var applicationsSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "select the application to use",
	Long:  `ttnctl applications select can be used to select the application to use in next commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		account := util.GetAccount(ctx)

		apps, err := account.ListApplications()
		if err != nil {
			ctx.WithError(err).Fatal("Could not select applications")
		}

		var appIdx int
		var euiIdx int

		switch len(apps) {
		case 0:
			ctx.Info("You don't have any applications, not doing anything.")
			return
		case 1:
			ctx.Infof("Found one application \"%s\", selecting that one.", apps[0].ID)
		default:
			ctx.Infof("Found %d applications:", len(apps))

			table := uitable.New()
			table.MaxColWidth = 70
			table.AddRow("", "ID", "Description")
			for i, app := range apps {
				table.AddRow(i+1, app.ID, app.Name)
			}

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

		switch len(app.EUIs) {
		case 0:
			ctx.Info("You don't have any EUIs in your application")
		case 1:
			ctx.Infof("Found one EUI \"%s\", selecting that one.", app.EUIs[0])
		default:
			ctx.Infof("Found %d EUIs", len(app.EUIs))

			table := uitable.New()
			table.MaxColWidth = 70
			table.AddRow("", "EUI")
			for i, eui := range app.EUIs {
				table.AddRow(i+1, eui)
			}

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

		err = util.SetConfig(map[string]interface{}{
			"app-id":  app.ID,
			"app-eui": eui.String(),
		})
		if err != nil {
			ctx.WithError(err).Fatal("Could not update configuration")
		}

		ctx.Info("Updated configuration")

	},
}

func init() {
	applicationsCmd.AddCommand(applicationsSelectCmd)
}

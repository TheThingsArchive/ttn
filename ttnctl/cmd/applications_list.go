// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var applicationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List applications",
	Long:    `ttnctl applications list can be used to list applications.`,
	Example: `$ ttnctl applications list
  INFO Found one application:

 	ID  	Description     	EUIs	Access Keys	Collaborators
1	test	Test application	1   	1          	1
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 0, 0)

		account := util.GetAccount(ctx)

		apps, err := account.ListApplications()
		if err != nil {
			ctx.WithError(err).Fatal("Could not list applications")
		}

		switch len(apps) {
		case 0:
			ctx.Info("You don't have any applications")
			return
		case 1:
			ctx.Info("Found one application:")
		default:
			ctx.Infof("Found %d applications:", len(apps))
		}

		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("", "ID", "Description", "EUIs", "Access Keys", "Collaborators")
		for i, app := range apps {
			table.AddRow(i+1, app.ID, app.Name, len(app.EUIs), len(app.AccessKeys), len(app.Collaborators))
		}

		fmt.Println()
		fmt.Println(table)
		fmt.Println()
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsListCmd)
}

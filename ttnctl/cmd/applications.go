// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type app struct {
	EUI   string `json:"eui"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// applicationsCmd represents the applications command
var applicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "Show applications",
	Long:  `ttnctl applications retrieves the applications of the logged on user.`,
	Run: func(cmd *cobra.Command, args []string) {
		server := viper.GetString("ttn-account-server")
		req, err := util.NewRequestWithAuth(server, "GET", fmt.Sprintf("%s/applications", server), nil)
		if err != nil {
			ctx.WithError(err).Fatal("Failed to created authenticated request")
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			ctx.WithError(err).Fatal("Failed to get applications")
		}
		if resp.StatusCode != http.StatusOK {
			ctx.Fatalf("Failed to get applications: %s", resp.Status)
		}

		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		var apps []*app
		err = decoder.Decode(&apps)
		if err != nil {
			ctx.WithError(err).Fatal("Failed to read applications")
		}

		ctx.Infof("Found %d application(s)", len(apps))
		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("EUI", "Name", "Owner")
		for _, app := range apps {
			table.AddRow(app.EUI, app.Name, app.Owner)
		}
		fmt.Println(table)
	},
}

func init() {
	RootCmd.AddCommand(applicationsCmd)
}

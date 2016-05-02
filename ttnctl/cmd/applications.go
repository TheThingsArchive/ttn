// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/apex/log"
	"github.com/gosuri/uitable"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationsCmd represents the applications command
var applicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "Show applications",
	Long:  `ttnctl applications retrieves the applications of the logged on user.`,
	Run: func(cmd *cobra.Command, args []string) {
		apps, err := util.GetApplications(ctx)
		if err != nil {
			ctx.WithError(err).Fatal("Failed to get applications")
		}

		ctx.Infof("Found %d application(s)", len(apps))
		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("EUI", "Name", "Owner", "Access Keys", "Valid")
		for _, app := range apps {
			table.AddRow(app.EUI, app.Name, app.Owner, strings.Join(app.AccessKeys, ", "), app.Valid)
		}

		fmt.Println(table)
	},
}

var applicationsCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new application",
	Long:  `ttnctl applications create creates a new application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		server := viper.GetString("ttn-account-server")
		uri := fmt.Sprintf("%s/applications", server)
		values := url.Values{
			"name": {args[0]},
		}
		req, err := util.NewRequestWithAuth(server, "POST", uri, strings.NewReader(values.Encode()))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to create authenticated request")
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			ctx.WithError(err).Fatal("Failed to create application")
		}
		if resp.StatusCode != http.StatusCreated {
			ctx.Fatalf("Failed to create application: %s", resp.Status)
		}

		ctx.Info("Application created successfully")

		// We need to refresh the token to add the new application to the set of
		// claims
		_, err = util.RefreshToken(server)
		if err != nil {
			log.WithError(err).Warn("Failed to refresh token. Please login")
		}
	},
}

var applicationsDeleteCmd = &cobra.Command{
	Use:   "delete [eui]",
	Short: "Delete an application",
	Long:  `ttnctl application delete deletes an existing application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		appEUI, err := util.Parse64(args[0])
		if err != nil {
			ctx.Fatalf("Invalid AppEUI: %s", err)
		}

		server := viper.GetString("ttn-account-server")
		uri := fmt.Sprintf("%s/applications/%s", server, fmt.Sprintf("%X", appEUI))
		req, err := util.NewRequestWithAuth(server, "DELETE", uri, nil)
		if err != nil {
			ctx.WithError(err).Fatal("Failed to create authenticated request")
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			ctx.WithError(err).Fatal("Failed to delete application")
		}
		if resp.StatusCode != http.StatusOK {
			ctx.Fatalf("Failed to delete application: %s", resp.Status)
		}

		ctx.Info("Application deleted successfully")

		// We need to refresh the token to remove the application from the set of
		// claims
		_, err = util.RefreshToken(server)
		if err != nil {
			log.WithError(err).Warn("Failed to refresh token. Please login")
		}
	},
}

var applicationsAuthorizeCmd = &cobra.Command{
	Use:   "authorize [eui] [e-mail]",
	Short: "Authorize a user for the application",
	Long:  `ttnctl applications authorize lets you authorize a user for an application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.Help()
			return
		}

		appEUI, err := util.Parse64(args[0])
		if err != nil {
			ctx.Fatalf("Invalid AppEUI: %s", err)
		}

		server := viper.GetString("ttn-account-server")
		uri := fmt.Sprintf("%s/applications/%s/authorize", server, fmt.Sprintf("%X", appEUI))
		values := url.Values{
			"email": {args[1]},
		}
		req, err := util.NewRequestWithAuth(server, "PUT", uri, strings.NewReader(values.Encode()))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to create authenticated request")
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			ctx.WithError(err).Fatal("Failed to authorize user")
		}
		if resp.StatusCode != http.StatusOK {
			ctx.Fatalf("Failed to authorize user: %s", resp.Status)
		}

		ctx.Info("User authorized successfully")
	},
}

var applicationsUseCmd = &cobra.Command{
	Use:   "use [eui]",
	Short: "Set an application as active",
	Long:  `ttnctl applications use marks an application as the currently active application in ttnctl.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		appEUI, err := util.Parse64(args[0])
		if err != nil {
			ctx.Fatalf("Invalid AppEUI: %s", err)
		}

		// check AppEUI provided is owned by user
		apps, err := util.GetApplications(ctx)
		if err != nil {
			ctx.WithError(err).Fatal("Failed to get applications")
		}

		var found bool
		newEUI := fmt.Sprintf("%X", appEUI)
		for _, app := range apps {
			if app.EUI == newEUI {
				found = true
				break
			}
		}

		if !found {
			ctx.Fatalf("%X not found in registered applications", appEUI)
		}

		// Determine config file
		cFile := viper.ConfigFileUsed()
		if cFile == "" {
			dir, err := homedir.Dir()
			if err != nil {
				ctx.WithError(err).Fatal("Could not get homedir")
			}
			expanded, err := homedir.Expand(dir)
			if err != nil {
				ctx.WithError(err).Fatal("Could not get homedir")
			}
			cFile = path.Join(expanded, ".ttnctl.yaml")
		}

		c := make(map[string]interface{})

		// Read config file
		bytes, err := ioutil.ReadFile(cFile)
		if err == nil {
			err = yaml.Unmarshal(bytes, &c)
		}
		if err != nil {
			ctx.Warnf("Could not read configuration file, will just create a new one")
		}

		// Update app
		c["app-eui"] = newEUI

		// Write config file
		d, err := yaml.Marshal(&c)
		if err != nil {
			ctx.Fatal("Could not generate configiguration file contents")
		}
		err = ioutil.WriteFile(cFile, d, 0644)
		if err != nil {
			ctx.WithError(err).Fatal("Could not write configiguration file")
		}

		ctx.Infof("You are now using application %X.", appEUI)

	},
}

func init() {
	RootCmd.AddCommand(applicationsCmd)
	applicationsCmd.AddCommand(applicationsCreateCmd)
	applicationsCmd.AddCommand(applicationsDeleteCmd)
	applicationsCmd.AddCommand(applicationsAuthorizeCmd)
	applicationsCmd.AddCommand(applicationsUseCmd)
}

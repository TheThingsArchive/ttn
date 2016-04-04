// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/core/components/broker"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/TheThingsNetwork/ttn/utils/tokenkey"
	"github.com/apex/log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// brokerCmd represents the router command
var brokerCmd = &cobra.Command{
	Use:   "broker",
	Short: "The Things Network broker",
	Long: `ttn broker starts the Broker component of The Things Network.

The Broker is responsible for finding the right handler for uplink packets it
receives from Routers. Handlers have register Applications and personalized
devices (with their network session keys) with the Broker.
	`,
	PreRun: func(cmd *cobra.Command, args []string) {
		var statusServer string
		if viper.GetInt("broker.status-port") > 0 {
			statusServer = fmt.Sprintf("%s:%d", viper.GetString("broker.status-address"), viper.GetInt("broker.status-port"))
			stats.Initialize()
		} else {
			statusServer = "disabled"
			stats.Enabled = false
		}
		ctx.WithFields(log.Fields{
			"devices-database":      viper.GetString("broker.db-devices"),
			"applications-database": viper.GetString("broker.db-apps"),
			"status-server":         statusServer,
			"main-server":           fmt.Sprintf("%s:%d", viper.GetString("broker.server-address"), viper.GetInt("broker.server-port")),
		}).Info("Using Configuration")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Status & Health
		statusAddr := fmt.Sprintf("%s:%d", viper.GetString("broker.status-address"), viper.GetInt("broker.status-port"))
		statusAdapter := http.New(
			http.Components{Ctx: ctx.WithField("adapter", "handler-status")},
			http.Options{NetAddr: statusAddr, Timeout: time.Second * 5},
		)
		statusAdapter.Bind(http.Healthz{})
		statusAdapter.Bind(http.StatusPage{})

		// Storage
		var dbDev broker.NetworkController
		devDBString := viper.GetString("broker.db-devices")
		switch {
		case strings.HasPrefix(devDBString, "boltdb:"):

			dbPath, err := filepath.Abs(devDBString[7:])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid devices database path")
			}

			dbDev, err = broker.NewNetworkController(dbPath)
			if err != nil {
				ctx.WithError(err).Fatal("Could not create local storage")
			}

			ctx.WithField("devices database", dbPath).Info("Using local storage")
		default:
			ctx.WithError(fmt.Errorf("Invalid devices database string. Format: \"boltdb:/path/to.db\".")).Fatal("Could not instantiate local storage")
		}

		var dbApp broker.AppStorage
		appDBString := viper.GetString("broker.db-apps")
		switch {
		case strings.HasPrefix(appDBString, "boltdb:"):

			dbPath, err := filepath.Abs(appDBString[7:])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid applications database path")
			}

			dbApp, err = broker.NewAppStorage(dbPath)
			if err != nil {
				ctx.WithError(err).Fatal("Could not create local storage")
			}

			ctx.WithField("applications database", dbPath).Info("Using local storage")
		default:
			ctx.WithError(fmt.Errorf("Invalid applications database string. Format: \"boltdb:/path/to.db\".")).Fatal("Could not instantiate local storage")
		}

		// Broker
		broker := broker.New(
			broker.Components{
				Ctx:               ctx,
				NetworkController: dbDev,
				AppStorage:        dbApp,
			},
			broker.Options{
				NetAddrUp:        fmt.Sprintf("%s:%d", viper.GetString("broker.uplink-address"), viper.GetInt("broker.uplink-port")),
				NetAddrDown:      fmt.Sprintf("%s:%d", viper.GetString("broker.downlink-address"), viper.GetInt("broker.downlink-port")),
				TokenKeyProvider: tokenkey.NewHTTPProvider(fmt.Sprintf("%s/key", viper.GetString("broker.account-server")), viper.GetString("broker.oauth2-keyfile")),
			},
		)

		// Go
		if err := broker.Start(); err != nil {
			ctx.WithError(err).Fatal("Broker has fallen...")
		}
	},
}

func init() {
	RootCmd.AddCommand(brokerCmd)

	brokerCmd.Flags().String("db-apps", "boltdb:/tmp/ttn_broker_apps.db", "Applications Database connection")
	viper.BindPFlag("broker.db-apps", brokerCmd.Flags().Lookup("db-apps"))

	brokerCmd.Flags().String("db-devices", "boltdb:/tmp/ttn_broker_devices.db", "Devices Database connection")
	viper.BindPFlag("broker.db-devices", brokerCmd.Flags().Lookup("db-devices"))

	brokerCmd.Flags().String("status-address", "0.0.0.0", "The IP address to listen for serving status information")
	brokerCmd.Flags().Int("status-port", 10701, "The port of the status server, use 0 to disable")
	viper.BindPFlag("broker.status-address", brokerCmd.Flags().Lookup("status-address"))
	viper.BindPFlag("broker.status-port", brokerCmd.Flags().Lookup("status-port"))

	brokerCmd.Flags().String("uplink-address", "0.0.0.0", "The IP address to listen for uplink communication")
	brokerCmd.Flags().Int("uplink-port", 1881, "The port for uplink communication")
	viper.BindPFlag("broker.uplink-address", brokerCmd.Flags().Lookup("uplink-address"))
	viper.BindPFlag("broker.uplink-port", brokerCmd.Flags().Lookup("uplink-port"))

	brokerCmd.Flags().String("downlink-address", "0.0.0.0", "The IP address to listen for downlink communication")
	brokerCmd.Flags().Int("downlink-port", 1781, "The port for downlink communication")
	viper.BindPFlag("broker.downlink-address", brokerCmd.Flags().Lookup("downlink-address"))
	viper.BindPFlag("broker.downlink-port", brokerCmd.Flags().Lookup("downlink-port"))

	brokerCmd.Flags().String("account-server", "https://account.thethingsnetwork.org", "The address of the OAuth 2.0 server")
	viper.BindPFlag("broker.account-server", brokerCmd.Flags().Lookup("account-server"))

	defaultOAuth2KeyFile := ""
	dir, err := homedir.Dir()
	if err == nil {
		expanded, err := homedir.Expand(dir)
		if err == nil {
			defaultOAuth2KeyFile = path.Join(expanded, ".ttn/oauth2-token.pub")
		}
	}

	brokerCmd.Flags().String("oauth2-keyfile", defaultOAuth2KeyFile, "The OAuth 2.0 public key")
	viper.BindPFlag("broker.oauth2-keyfile", brokerCmd.Flags().Lookup("oauth2-keyfile"))
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/core/components/broker"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// brokerCmd represents the router command
var brokerCmd = &cobra.Command{
	Use:   "broker",
	Short: "The Things Network broker",
	Long: `
The broker is responsible for finding the right handler for uplink packets it
receives from routers. This means that handlers have to register applications
and personalized devices (with their network session keys) with the router.
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
			"database":      viper.GetString("broker.database"),
			"status-server": statusServer,
			"main-server":   fmt.Sprintf("%s:%d", viper.GetString("broker.server-address"), viper.GetInt("broker.server-port")),
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
		var db broker.NetworkController

		dbString := viper.GetString("broker.database")
		switch {
		case strings.HasPrefix(dbString, "boltdb:"):

			dbPath, err := filepath.Abs(dbString[7:])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid database path")
			}

			db, err = broker.NewNetworkController(dbPath)
			if err != nil {
				ctx.WithError(err).Fatal("Could not create local storage")
			}

			ctx.WithField("database", dbPath).Info("Using local storage")
		default:
			ctx.WithError(fmt.Errorf("Invalid database string. Format: \"boltdb:/path/to.db\".")).Fatal("Could not instantiate local storage")
		}

		// Broker
		netAddr := fmt.Sprintf("%s:%d", viper.GetString("broker.server-address"), viper.GetInt("broker.server-port"))
		broker := broker.New(
			broker.Components{Ctx: ctx, NetworkController: db},
			broker.Options{NetAddr: netAddr},
		)

		// Go
		if err := broker.Start(); err != nil {
			ctx.WithError(err).Fatal("Broker has fallen...")
		}
	},
}

func init() {
	RootCmd.AddCommand(brokerCmd)

	brokerCmd.Flags().String("database", "boltdb:/tmp/ttn_broker.db", "Database connection")
	viper.BindPFlag("broker.database", brokerCmd.Flags().Lookup("database"))

	brokerCmd.Flags().String("status-address", "localhost", "The IP address to listen for serving status information")
	brokerCmd.Flags().Int("status-port", 10701, "The port of the status server, use 0 to disable")
	viper.BindPFlag("broker.status-address", brokerCmd.Flags().Lookup("status-address"))
	viper.BindPFlag("broker.status-port", brokerCmd.Flags().Lookup("status-port"))

	brokerCmd.Flags().String("server-address", "", "The IP address to listen for uplink and downlink messages")
	brokerCmd.Flags().Int("server-port", 1881, "The main communication port")
	viper.BindPFlag("broker.server-address", brokerCmd.Flags().Lookup("server-address"))
	viper.BindPFlag("broker.server-port", brokerCmd.Flags().Lookup("server-port"))
}

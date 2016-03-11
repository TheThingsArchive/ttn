// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/handlers"
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
			statusServer = fmt.Sprintf("%s:%d", viper.GetString("broker.status-bind-address"), viper.GetInt("broker.status-port"))
			stats.Initialize()
		} else {
			statusServer = "disabled"
			stats.Enabled = false
		}
		ctx.WithFields(log.Fields{
			"database":      viper.GetString("broker.database"),
			"status-server": statusServer,
			"uplink":        fmt.Sprintf("%s:%d", viper.GetString("broker.uplink-bind-address"), viper.GetInt("broker.uplink-port")),
			"downlink":      fmt.Sprintf("%s:%d", viper.GetString("broker.downlink-bind-address"), viper.GetInt("broker.downlink-port")),
		}).Info("Using Configuration")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Instantiate all components
		rtrNet := fmt.Sprintf("%s:%d", viper.GetString("broker.uplink-bind-address"), viper.GetInt("broker.uplink-port"))
		rtrAdapter, err := http.NewAdapter(rtrNet, nil, ctx.WithField("adapter", "router-http"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Routers Adapter")
		}
		rtrAdapter.Bind(handlers.Collect{})

		hdlNet := fmt.Sprintf("%s:%d", viper.GetString("broker.downlink-bind-address"), viper.GetInt("broker.downlink-port"))
		hdlAdapter, err := http.NewAdapter(hdlNet, nil, ctx.WithField("adapter", "handler-http"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Handlers Adapter")
		}
		hdlAdapter.Bind(handlers.Collect{})
		hdlAdapter.Bind(handlers.PubSub{})
		hdlAdapter.Bind(handlers.Applications{})

		if viper.GetInt("broker.status-port") > 0 {
			statusNet := fmt.Sprintf("%s:%d", viper.GetString("broker.status-bind-address"), viper.GetInt("broker.status-port"))
			statusAdapter, err := http.NewAdapter(statusNet, nil, ctx.WithField("adapter", "status-http"))
			if err != nil {
				ctx.WithError(err).Fatal("Could not start Status Adapter")
			}
			statusAdapter.Bind(handlers.StatusPage{})
			statusAdapter.Bind(handlers.Healthz{})
		}
		// Instantiate Storage

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

		broker := broker.New(db, ctx)

		// Bring the service to life

		// Listen to uplink
		go func() {
			for {
				packet, an, err := rtrAdapter.Next()
				if err != nil {
					ctx.WithError(err).Error("Could not retrieve uplink")
					continue
				}
				go func(packet []byte, an core.AckNacker) {
					if err := broker.HandleUp(packet, an, hdlAdapter); err != nil {
						// We can't do anything with this packet, so we're ignoring it.
						ctx.WithError(err).Debug("Could not process uplink")
					}
				}(packet, an)
			}
		}()

		// List to handler registrations
		go func() {
			for {
				reg, an, err := hdlAdapter.NextRegistration()
				if err != nil {
					ctx.WithError(err).Error("Could not retrieve registration")
					continue
				}
				go func(reg core.Registration, an core.AckNacker) {
					if err := broker.Register(reg, an); err != nil {
						// We can't do anything with this registration, so we're ignoring it.
						ctx.WithError(err).Debug("Could not process registration")
					}
				}(reg, an)
			}
		}()

		<-make(chan bool)
	},
}

func init() {
	RootCmd.AddCommand(brokerCmd)

	brokerCmd.Flags().String("database", "boltdb:/tmp/ttn_broker.db", "Database connection")
	viper.BindPFlag("broker.database", brokerCmd.Flags().Lookup("database"))

	brokerCmd.Flags().String("status-bind-address", "localhost", "The IP address to listen for serving status information")
	brokerCmd.Flags().Int("status-port", 10701, "The port of the status server, use 0 to disable")
	viper.BindPFlag("broker.status-bind-address", brokerCmd.Flags().Lookup("status-bind-address"))
	viper.BindPFlag("broker.status-port", brokerCmd.Flags().Lookup("status-port"))

	brokerCmd.Flags().String("uplink-bind-address", "", "The IP address to listen for uplink messages from routers")
	brokerCmd.Flags().Int("uplink-port", 1881, "The port for the uplink")
	viper.BindPFlag("broker.uplink-bind-address", brokerCmd.Flags().Lookup("uplink-bind-address"))
	viper.BindPFlag("broker.uplink-port", brokerCmd.Flags().Lookup("uplink-port"))

	brokerCmd.Flags().String("downlink-bind-address", "", "The IP address to listen for downlink messages from brokers")
	brokerCmd.Flags().Int("downlink-port", 1781, "The port for the downlink")
	viper.BindPFlag("broker.downlink-bind-address", brokerCmd.Flags().Lookup("downlink-bind-address"))
	viper.BindPFlag("broker.downlink-port", brokerCmd.Flags().Lookup("downlink-port"))
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/parser"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/pubsub"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/statuspage"
	"github.com/TheThingsNetwork/ttn/core/components"
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
		ctx = ctx.WithField("cmd", "broker")
		ctx.WithFields(log.Fields{
			"database":      viper.GetString("broker.database"),
			"routers-port":  viper.GetInt("broker.routers-port"),
			"handlers-port": viper.GetInt("broker.handlers-port"),
		}).Info("Using Configuration")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Instantiate all components
		rtrAdapter, err := http.NewAdapter(uint(viper.GetInt("broker.routers-port")), parser.JSON{}, ctx.WithField("tag", "Routers Adapter"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Routers Adapter")
		}

		hdlHTTPAdapter, err := http.NewAdapter(uint(viper.GetInt("broker.handlers-port")), parser.JSON{}, ctx.WithField("tag", "Handlers Adapter"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Handlers Adapter")
		}

		_, err = statuspage.NewAdapter(hdlHTTPAdapter, ctx.WithField("tag", "Broker Adapter"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Broker Adapter")
		}

		hdlAdapter, err := pubsub.NewAdapter(hdlHTTPAdapter, parser.PubSub{}, ctx.WithField("tag", "Handlers Adapter"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Handlers Adapter")
		}

		db, err := components.NewBrokerStorage()
		if err != nil {
			ctx.WithError(err).Fatal("Could not create a local storage")
		}

		broker := components.NewBroker(db, ctx.WithField("tag", "Broker"))

		// Bring the service to life

		// Listen to uplink
		go func() {
			for {
				packet, an, err := rtrAdapter.Next()
				if err != nil {
					ctx.WithError(err).Error("Could not retrieve uplink")
					continue
				}
				go func(packet core.Packet, an core.AckNacker) {
					if err := broker.HandleUp(packet, an, hdlAdapter); err != nil {
						ctx.WithError(err).Error("Could not process uplink")
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
						ctx.WithError(err).Error("Could not process registration")
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
	brokerCmd.Flags().Int("routers-port", 1690, "TCP port for connections from routers")
	brokerCmd.Flags().Int("handlers-port", 1790, "TCP port for connections from handlers")

	viper.BindPFlag("broker.database", brokerCmd.Flags().Lookup("database"))
	viper.BindPFlag("broker.routers-port", brokerCmd.Flags().Lookup("routers-port"))
	viper.BindPFlag("broker.handlers-port", brokerCmd.Flags().Lookup("handlers-port"))
}

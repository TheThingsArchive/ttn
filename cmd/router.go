// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/broadcast"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/parser"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/statuspage"
	"github.com/TheThingsNetwork/ttn/core/adapters/semtech"
	"github.com/TheThingsNetwork/ttn/core/components"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// routerCmd represents the router command
var routerCmd = &cobra.Command{
	Use:   "router",
	Short: "The Things Network router",
	Long: `The router accepts connections from gateways and forwards uplink packets to one
or more brokers. The router is also responsible for monitoring gateways,
collecting statistics from gateways and for enforcing TTN's fair use policy when
the gateway's duty cycle is (almost) full.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		ctx.WithFields(log.Fields{
			"database":      viper.GetString("router.database"),
			"gateways-port": viper.GetInt("router.gateways-port"),
			"brokers":       viper.GetString("router.brokers"),
			"brokers-port":  viper.GetInt("router.brokers-port"),
		}).Info("Using Configuration")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		gtwAdapter, err := semtech.NewAdapter(uint(viper.GetInt("router.gateways-port")), ctx.WithField("adapter", "gateway-semtech"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Gateway Adapter")
		}

		pktAdapter, err := http.NewAdapter(uint(viper.GetInt("router.brokers-port")), parser.JSON{}, ctx.WithField("adapter", "broker-http"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Broker Adapter")
		}

		_, err = statuspage.NewAdapter(pktAdapter, ctx.WithField("adapter", "statuspage-http"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Broker Adapter")
		}

		var brokers []core.Recipient
		brokersStr := strings.Split(viper.GetString("router.brokers"), ",")
		for i := range brokersStr {
			brokers = append(brokers, core.Recipient{
				Address: strings.Trim(brokersStr[i], " "),
				Id:      i,
			})
		}

		brkAdapter, err := broadcast.NewAdapter(pktAdapter, brokers, ctx.WithField("adapter", "broker-broadcast"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Broker Adapter")
		}

		db, err := components.NewRouterStorage(time.Hour * 8)
		if err != nil {
			ctx.WithError(err).Fatal("Could not create a local storage")
		}

		router := components.NewRouter(db, ctx)

		// Bring the service to life

		// Listen uplink
		go func() {
			for {
				packet, an, err := gtwAdapter.Next()
				if err != nil {
					ctx.WithError(err).Warn("Could not get next packet from gateway")
					continue
				}
				go func(packet core.Packet, an core.AckNacker) {
					if err := router.HandleUp(packet, an, brkAdapter); err != nil {
						ctx.WithError(err).Warn("Could not process packet from gateway")
					}
				}(packet, an)
			}
		}()

		// Listen broker registrations
		go func() {
			for {
				reg, an, err := brkAdapter.NextRegistration()
				if err != nil {
					ctx.WithError(err).Warn("Could not get next registration from broker")
					continue
				}
				go func(reg core.Registration, an core.AckNacker) {
					if err := router.Register(reg, an); err != nil {
						ctx.WithError(err).Warn("Could not process registration from broker")
					}
				}(reg, an)
			}
		}()

		<-make(chan bool)
	},
}

func init() {
	RootCmd.AddCommand(routerCmd)

	routerCmd.Flags().String("database", "boltdb:/tmp/ttn_router.db", "Database connection")
	routerCmd.Flags().Int("gateways-port", 1700, "UDP port for connections from gateways")
	routerCmd.Flags().String("brokers", "localhost:1690", "Comma-separated list of brokers")
	routerCmd.Flags().Int("brokers-port", 1780, "TCP port for connections from brokers")

	viper.BindPFlag("router.database", routerCmd.Flags().Lookup("database"))
	viper.BindPFlag("router.gateways-port", routerCmd.Flags().Lookup("gateways-port"))
	viper.BindPFlag("router.brokers", routerCmd.Flags().Lookup("brokers"))
	viper.BindPFlag("router.brokers-port", routerCmd.Flags().Lookup("brokers-port"))
}

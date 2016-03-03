// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	httpHandlers "github.com/TheThingsNetwork/ttn/core/adapters/http/handlers"
	"github.com/TheThingsNetwork/ttn/core/adapters/udp"
	udpHandlers "github.com/TheThingsNetwork/ttn/core/adapters/udp/handlers"
	"github.com/TheThingsNetwork/ttn/core/components/router"
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
			"uplink-port":   viper.GetInt("router.uplink-port"),
			"downlink-port": viper.GetInt("router.downlink-port"),
			"brokers":       viper.GetString("router.brokers"),
		}).Info("Using Configuration")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		gtwAdapter, err := udp.NewAdapter(uint(viper.GetInt("router.uplink-port")), ctx.WithField("adapter", "gateway-semtech"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Gateway Adapter")
		}
		gtwAdapter.Bind(udpHandlers.Semtech{})

		var brokers []core.Recipient
		brokersStr := strings.Split(viper.GetString("router.brokers"), ",")
		for i := range brokersStr {
			url := fmt.Sprintf("%s/packets/", strings.Trim(brokersStr[i], " "))
			brokers = append(brokers, http.NewRecipient(url, "POST"))
		}

		brkAdapter, err := http.NewAdapter(uint(viper.GetInt("router.downlink-port")), brokers, ctx.WithField("adapter", "broker-http"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Broker Adapter")
		}
		brkAdapter.Bind(httpHandlers.StatusPage{})
		brkAdapter.Bind(httpHandlers.Healthz{})

		var db router.Storage

		dbString := viper.GetString("router.database")
		switch {
		case strings.HasPrefix(dbString, "boltdb:"):

			dbPath, err := filepath.Abs(dbString[7:])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid database path")
			}

			db, err = router.NewStorage(dbPath, time.Hour*8)
			if err != nil {
				ctx.WithError(err).Fatal("Could not create a local storage")
			}

			ctx.WithField("database", dbPath).Info("Using local storage")
		default:
			ctx.WithError(fmt.Errorf("Invalid database string. Format: \"boltdb:/path/to.db\".")).Fatal("Could not instantiate local storage")
		}

		router := router.New(db, ctx)

		// Bring the service to life

		// Listen uplink
		go func() {
			for {
				packet, an, err := gtwAdapter.Next()
				if err != nil {
					ctx.WithError(err).Warn("Could not get next packet from gateway")
					continue
				}
				go func(packet []byte, an core.AckNacker) {
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
	viper.BindPFlag("router.database", routerCmd.Flags().Lookup("database"))

	routerCmd.Flags().Int("uplink-port", 1700, "The UDP port for the uplink")
	viper.BindPFlag("router.uplink-port", routerCmd.Flags().Lookup("uplink-port"))

	routerCmd.Flags().Int("downlink-port", 1780, "The port for the downlink")
	viper.BindPFlag("router.downlink-port", routerCmd.Flags().Lookup("downlink-port"))
	routerCmd.Flags().String("brokers", ":1881", "Comma-separated list of brokers")
	viper.BindPFlag("router.brokers", routerCmd.Flags().Lookup("brokers"))
}

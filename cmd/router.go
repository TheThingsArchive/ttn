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
	"github.com/TheThingsNetwork/ttn/core/dutycycle"
	"github.com/TheThingsNetwork/ttn/utils/stats"
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
		var statusServer string
		if viper.GetInt("router.status-port") > 0 {
			statusServer = fmt.Sprintf("%s:%d", viper.GetString("router.status-bind-address"), viper.GetInt("router.status-port"))
			initStats()
		} else {
			statusServer = "disabled"
			stats.Enabled = false
		}

		ctx.WithFields(log.Fields{
			"db-brokers":    viper.GetString("router.db_brokers"),
			"db-gateways":   viper.GetString("router.db_gateways"),
			"status-server": statusServer,
			"uplink":        fmt.Sprintf("%s:%d", viper.GetString("router.uplink-bind-address"), viper.GetInt("router.uplink-port")),
			"downlink":      fmt.Sprintf("%s:%d", viper.GetString("router.downlink-bind-address"), viper.GetInt("router.downlink-port")),
			"brokers":       viper.GetString("router.brokers"),
		}).Info("Using Configuration")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		gtwNet := fmt.Sprintf("%s:%d", viper.GetString("router.uplink-bind-address"), viper.GetInt("router.uplink-port"))
		gtwAdapter, err := udp.NewAdapter(gtwNet, ctx.WithField("adapter", "gateway-semtech"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Gateway Adapter")
		}
		gtwAdapter.Bind(udpHandlers.Semtech{})

		var brokers []core.Recipient
		brokersStr := strings.Split(viper.GetString("router.brokers"), ",")
		for i := range brokersStr {
			url := strings.Trim(brokersStr[i], " ")
			brokers = append(brokers, http.NewRecipient(url, "POST"))
		}

		brkNet := fmt.Sprintf("%s:%d", viper.GetString("router.downlink-bind-address"), viper.GetInt("router.downlink-port"))
		brkAdapter, err := http.NewAdapter(brkNet, brokers, ctx.WithField("adapter", "broker-http"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Broker Adapter")
		}

		if viper.GetInt("router.status-port") > 0 {
			statusNet := fmt.Sprintf("%s:%d", viper.GetString("router.status-bind-address"), viper.GetInt("router.status-port"))
			statusAdapter, err := http.NewAdapter(statusNet, nil, ctx.WithField("adapter", "status-http"))
			if err != nil {
				ctx.WithError(err).Fatal("Could not start Status Adapter")
			}
			statusAdapter.Bind(httpHandlers.StatusPage{})
			statusAdapter.Bind(httpHandlers.Healthz{})
		}

		var db router.Storage

		dbString := viper.GetString("router.db_brokers")
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

		var dm dutycycle.DutyManager

		dmString := viper.GetString("router.db_gateways")
		switch {
		case strings.HasPrefix(dmString, "boltdb:"):

			dmPath, err := filepath.Abs(dmString[7:])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid database path")
			}

			dm, err = dutycycle.NewManager(dmPath, time.Hour, dutycycle.Europe)
			if err != nil {
				ctx.WithError(err).Fatal("Could not create a local storage")
			}

			ctx.WithField("database", dmPath).Info("Using local storage")
		default:
			ctx.WithError(fmt.Errorf("Invalid database string. Format: \"boltdb:/path/to.db\".")).Fatal("Could not instantiate local storage")
		}

		router := router.New(db, dm, ctx)

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
						// We can't do anything with this packet, so we're ignoring it.
						ctx.WithError(err).Debug("Could not process packet from gateway")
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
						// We can't do anything with this registration, so we're ignoring it.
						ctx.WithError(err).Debug("Could not process registration from broker")
					}
				}(reg, an)
			}
		}()

		<-make(chan bool)
	},
}

func init() {
	RootCmd.AddCommand(routerCmd)

	routerCmd.Flags().String("db_brokers", "boltdb:/tmp/ttn_router_brokers.db", "Database connection of known brokers")
	viper.BindPFlag("router.db_brokers", routerCmd.Flags().Lookup("db_brokers"))

	routerCmd.Flags().String("db_gateways", "boltdb:/tmp/ttn_router_gateways.db", "Database connection of managed gateways")
	viper.BindPFlag("router.db_gateways", routerCmd.Flags().Lookup("db_gateways"))

	routerCmd.Flags().String("status-bind-address", "localhost", "The IP address to listen for serving status information")
	routerCmd.Flags().Int("status-port", 10700, "The port of the status server, use 0 to disable")
	viper.BindPFlag("router.status-bind-address", routerCmd.Flags().Lookup("status-bind-address"))
	viper.BindPFlag("router.status-port", routerCmd.Flags().Lookup("status-port"))

	routerCmd.Flags().String("uplink-bind-address", "", "The IP address to listen for uplink messages from gateways")
	routerCmd.Flags().Int("uplink-port", 1700, "The UDP port for the uplink")
	viper.BindPFlag("router.uplink-bind-address", routerCmd.Flags().Lookup("uplink-bind-address"))
	viper.BindPFlag("router.uplink-port", routerCmd.Flags().Lookup("uplink-port"))

	routerCmd.Flags().String("downlink-bind-address", "", "The IP address to listen for downlink messages from routers")
	routerCmd.Flags().Int("downlink-port", 1780, "The port for the downlink")
	viper.BindPFlag("router.downlink-bind-address", routerCmd.Flags().Lookup("downlink-bind-address"))
	viper.BindPFlag("router.downlink-port", routerCmd.Flags().Lookup("downlink-port"))

	routerCmd.Flags().String("brokers", ":1881", "Comma-separated list of brokers")
	viper.BindPFlag("router.brokers", routerCmd.Flags().Lookup("brokers"))
}

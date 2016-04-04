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
	"github.com/TheThingsNetwork/ttn/core/adapters/udp"
	"github.com/TheThingsNetwork/ttn/core/components/router"
	"github.com/TheThingsNetwork/ttn/core/dutycycle"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// routerCmd represents the router command
var routerCmd = &cobra.Command{
	Use:   "router",
	Short: "The Things Network router",
	Long: `ttn router starts the Router component of The Things Network.

The Router accepts connections from gateways and forwards uplink packets to one
or more brokers. The router is also responsible for monitoring gateways,
collecting statistics from gateways and for enforcing TTN's fair use policy when
the gateway's duty cycle is (almost) full.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		var statusServer string
		if viper.GetInt("router.status-port") > 0 {
			statusServer = fmt.Sprintf("%s:%d", viper.GetString("router.status-address"), viper.GetInt("router.status-port"))
			stats.Initialize()
		} else {
			statusServer = "disabled"
			stats.Enabled = false
		}

		ctx.WithFields(log.Fields{
			"db-brokers":    viper.GetString("router.db-brokers"),
			"db-gateways":   viper.GetString("router.db-gateways"),
			"db-duty":       viper.GetString("router.db-duty"),
			"status-server": statusServer,
			"uplink":        fmt.Sprintf("%s:%d", viper.GetString("router.uplink-address"), viper.GetInt("router.uplink-port")),
			"downlink":      fmt.Sprintf("%s:%d", viper.GetString("router.downlink-address"), viper.GetInt("router.downlink-port")),
			"brokers":       viper.GetString("router.brokers"),
		}).Info("Using Configuration")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Status & Health
		statusAddr := fmt.Sprintf("%s:%d", viper.GetString("router.status-address"), viper.GetInt("router.status-port"))
		statusAdapter := http.New(
			http.Components{Ctx: ctx.WithField("adapter", "router-status")},
			http.Options{NetAddr: statusAddr, Timeout: time.Second * 5},
		)
		statusAdapter.Bind(http.Healthz{})
		statusAdapter.Bind(http.StatusPage{})

		// In-memory packet storage
		var db router.BrkStorage
		dbString := viper.GetString("router.db-brokers")
		switch {
		case strings.HasPrefix(dbString, "boltdb:"):

			dbPath, err := filepath.Abs(dbString[7:])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid database path")
			}

			db, err = router.NewBrkStorage(dbPath, time.Hour*8)
			if err != nil {
				ctx.WithError(err).Fatal("Could not create a local storage")
			}

			ctx.WithField("database", dbPath).Info("Using local storage")
		default:
			ctx.WithError(fmt.Errorf("Invalid database string. Format: \"boltdb:/path/to.db\".")).Fatal("Could not instantiate local storage")
		}

		// Duty Manager
		var dm dutycycle.DutyManager
		dmString := viper.GetString("router.db-duty")
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

		// Gateways
		var dg router.GtwStorage
		dgString := viper.GetString("router.db-gateways")
		switch {
		case strings.HasPrefix(dmString, "boltdb:"):

			dgPath, err := filepath.Abs(dgString[7:])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid database path")
			}

			dg, err = router.NewGtwStorage(dgPath)
			if err != nil {
				ctx.WithError(err).Fatal("Could not create a local storage")
			}

			ctx.WithField("database", dgPath).Info("Using local storage")
		default:
			ctx.WithError(fmt.Errorf("Invalid database string. Format: \"boltdb:/path/to.db\".")).Fatal("Could not instantiate local storage")
		}

		// Broker clients
		var brokers []core.BrokerClient
		brokersStr := strings.Split(viper.GetString("router.brokers"), ",")
		for i := range brokersStr {
			url := strings.Trim(brokersStr[i], " ")
			brokerConn, err := grpc.Dial(url, grpc.WithInsecure(), grpc.WithTimeout(time.Second*15))
			if err != nil {
				ctx.WithError(err).Fatal("Could not dial broker")
			}
			defer brokerConn.Close()
			broker := core.NewBrokerClient(brokerConn)
			brokers = append(brokers, broker)
		}

		// Router
		router := router.New(
			router.Components{
				Ctx:         ctx,
				DutyManager: dm,
				Brokers:     brokers,
				BrkStorage:  db,
				GtwStorage:  dg,
			},
			router.Options{
				NetAddr: fmt.Sprintf("%s:%d", viper.GetString("router.downlink-address"), viper.GetInt("router.downlink-port")),
			},
		)

		// Gateway Adapter
		gtwNet := fmt.Sprintf("%s:%d", viper.GetString("router.uplink-address"), viper.GetInt("router.uplink-port"))
		err := udp.Start(
			udp.Components{
				Ctx:    ctx.WithField("adapter", "gateway-semtech"),
				Router: router,
			},
			udp.Options{
				NetAddr:              gtwNet,
				MaxReconnectionDelay: 25 * 10000 * time.Millisecond,
			},
		)
		if err != nil {
			ctx.WithError(err).Fatal("Could not start Gateway Adapter")
		}

		// Go
		if err := router.Start(); err != nil {
			ctx.WithError(err).Fatal("Router has fallen...")
		}
	},
}

func init() {
	RootCmd.AddCommand(routerCmd)

	routerCmd.Flags().String("db-brokers", "boltdb:/tmp/ttn_router_brokers.db", "Database connection of known brokers")
	viper.BindPFlag("router.db-brokers", routerCmd.Flags().Lookup("db-brokers"))

	routerCmd.Flags().String("db-gateways", "boltdb:/tmp/ttn_router_gateways.db", "Database connection of managed gateways")
	viper.BindPFlag("router.db-gateways", routerCmd.Flags().Lookup("db-gateways"))

	routerCmd.Flags().String("db-duty", "boltdb:/tmp/ttn_router_duty.db", "Database connection of managed dutycycles")
	viper.BindPFlag("router.db-duty", routerCmd.Flags().Lookup("db-duty"))

	routerCmd.Flags().String("status-address", "0.0.0.0", "The IP address to listen for serving status information")
	routerCmd.Flags().Int("status-port", 10700, "The port of the status server, use 0 to disable")
	viper.BindPFlag("router.status-address", routerCmd.Flags().Lookup("status-address"))
	viper.BindPFlag("router.status-port", routerCmd.Flags().Lookup("status-port"))

	routerCmd.Flags().String("uplink-address", "0.0.0.0", "The IP address to listen for uplink communication from gateways")
	routerCmd.Flags().Int("uplink-port", 1700, "The UDP port for uplink communication from gateways")
	viper.BindPFlag("router.uplink-address", routerCmd.Flags().Lookup("uplink-address"))
	viper.BindPFlag("router.uplink-port", routerCmd.Flags().Lookup("uplink-port"))

	routerCmd.Flags().String("downlink-address", "0.0.0.0", "The IP address to listen for downlink communication")
	routerCmd.Flags().Int("downlink-port", 1780, "The port for downlink communication")
	viper.BindPFlag("router.downlink-address", routerCmd.Flags().Lookup("downlink-address"))
	viper.BindPFlag("router.downlink-port", routerCmd.Flags().Lookup("downlink-port"))

	routerCmd.Flags().String("brokers", "localhost:1881", "Comma-separated list of brokers")
	viper.BindPFlag("router.brokers", routerCmd.Flags().Lookup("brokers"))
}

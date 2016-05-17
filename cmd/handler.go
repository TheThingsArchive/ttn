// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core/adapters/fields"
	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	handlerMQTT "github.com/TheThingsNetwork/ttn/core/adapters/mqtt"
	"github.com/TheThingsNetwork/ttn/core/components/broker"
	"github.com/TheThingsNetwork/ttn/core/components/handler"
	ttnMQTT "github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// handlerCmd represents the handler command
var handlerCmd = &cobra.Command{
	Use:   "handler",
	Short: "The Things Network handler",
	Long: `ttn handler starts a default Handler for The Things Network

The Handler is the bridge between The Things Network and applications.
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		var statusServer string
		if viper.GetInt("handler.status-port") > 0 {
			statusServer = fmt.Sprintf("%s:%d", viper.GetString("handler.status-address"), viper.GetInt("handler.status-port"))
			stats.Initialize()
		} else {
			statusServer = "disabled"
			stats.Enabled = false
		}
		ctx.WithFields(log.Fields{
			"devices-database": viper.GetString("handler.db-devices"),
			"packets-database": viper.GetString("handler.db-packets"),
			"status-server":    statusServer,
			"internal-server":  fmt.Sprintf("%s:%d", viper.GetString("handler.internal-address"), viper.GetInt("handler.internal-port")),
			"public-server":    fmt.Sprintf("%s:%d", viper.GetString("handler.public-address"), viper.GetInt("handler.public-port")),
			"ttn-broker":       viper.GetString("handler.ttn-broker"),
			"mqtt-broker":      viper.GetString("handler.mqtt-broker"),
		}).Info("Using Configuration")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Status & Health
		statusAddr := fmt.Sprintf("%s:%d", viper.GetString("handler.status-address"), viper.GetInt("handler.status-port"))
		statusAdapter := http.New(
			http.Components{Ctx: ctx.WithField("adapter", "handler-status")},
			http.Options{NetAddr: statusAddr, Timeout: time.Second * 5},
		)
		statusAdapter.Bind(http.Healthz{})
		statusAdapter.Bind(http.StatusPage{})

		// In-memory devices storage
		var devicesDB handler.DevStorage

		devDBString := viper.GetString("handler.db-devices")
		switch {
		case strings.HasPrefix(devDBString, "boltdb:"):

			devDBPath, err := filepath.Abs(devDBString[7:])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid devices database path")
			}

			devicesDB, err = handler.NewDevStorage(devDBPath)
			if err != nil {
				ctx.WithError(err).Fatal("Could not create local devices storage")
			}

			ctx.WithField("database", devDBPath).Info("Using local storage")
		default:
			ctx.WithError(fmt.Errorf("Invalid database string. Format: \"boltdb:/path/to.db\".")).Fatal("Could not instantiate local devices storage")
		}

		// In-memory packets storage
		var packetsDB handler.PktStorage

		pktDBString := viper.GetString("handler.db-packets")
		switch {
		case strings.HasPrefix(pktDBString, "boltdb:"):

			pktDBPath, err := filepath.Abs(pktDBString[7:])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid packets database path")
			}

			packetsDB, err = handler.NewPktStorage(pktDBPath, 1)
			if err != nil {
				ctx.WithError(err).Fatal("Could not create local packets storage")
			}

			ctx.WithField("database", pktDBPath).Info("Using local storage")
		default:
			ctx.WithError(fmt.Errorf("Invalid database string. Format: \"boltdb:/path/to.db\".")).Fatal("Could not instantiate local packets storage")
		}

		// BrokerClient
		brokerClient, err := broker.NewClient(viper.GetString("handler.ttn-broker"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not dial broker")
		}

		// MQTT Client & adapter
		mqttClient := ttnMQTT.NewClient(
			ctx.WithField("adapter", "handler-mqtt"),
			"ttnhdl",
			viper.GetString("handler.mqtt-username"),
			viper.GetString("handler.mqtt-password"),
			fmt.Sprintf("tcp://%s", viper.GetString("handler.mqtt-broker")),
		)
		err = mqttClient.Connect()
		if err != nil {
			ctx.WithError(err).Fatal("Could not connect to MQTT")
		}

		storage, err := fields.ConnectRedis(viper.GetString("handler.redis-addr"), int64(viper.GetInt("handler.redis-db")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not connect to Redis")
		}

		fieldsAdapter := fields.NewAdapter(
			ctx.WithField("adapter", "fields-adapter"),
			storage,
			handlerMQTT.NewAdapter(
				ctx.WithField("adapter", "mqtt-adapter"),
				mqttClient,
			),
		)

		// Handler
		handler := handler.New(
			handler.Components{
				Ctx:        ctx,
				DevStorage: devicesDB,
				PktStorage: packetsDB,
				Broker:     brokerClient,
				AppAdapter: fieldsAdapter,
			},
			handler.Options{
				PublicNetAddr:          fmt.Sprintf("%s:%d", viper.GetString("handler.public-address"), viper.GetInt("handler.public-port")),
				PrivateNetAddr:         fmt.Sprintf("%s:%d", viper.GetString("handler.internal-address"), viper.GetInt("handler.internal-port")),
				PrivateNetAddrAnnounce: fmt.Sprintf("%s:%d", viper.GetString("handler.internal-address-announce"), viper.GetInt("handler.internal-port")),
			},
		)

		fieldsAdapter.SubscribeDownlink(handler)

		if err := handler.Start(); err != nil {
			ctx.WithError(err).Fatal("Handler has fallen...")
		}
	},
}

func init() {
	RootCmd.AddCommand(handlerCmd)

	handlerCmd.Flags().String("db-devices", "boltdb:/tmp/ttn_handler_devices.db", "Devices Database connection")
	handlerCmd.Flags().String("db-packets", "boltdb:/tmp/ttn_handler_packets.db", "Packets Database connection")
	viper.BindPFlag("handler.db-devices", handlerCmd.Flags().Lookup("db-devices"))
	viper.BindPFlag("handler.db-packets", handlerCmd.Flags().Lookup("db-packets"))

	handlerCmd.Flags().String("status-address", "0.0.0.0", "The IP address to listen for serving status information")
	handlerCmd.Flags().Int("status-port", 10702, "The port of the status server, use 0 to disable")
	viper.BindPFlag("handler.status-address", handlerCmd.Flags().Lookup("status-address"))
	viper.BindPFlag("handler.status-port", handlerCmd.Flags().Lookup("status-port"))

	handlerCmd.Flags().String("internal-address", "0.0.0.0", "The IP address to listen for communication from other components")
	handlerCmd.Flags().String("internal-address-announce", "localhost", "The hostname to announce for communication from other components")
	handlerCmd.Flags().Int("internal-port", 1882, "The port for communication from other components")
	viper.BindPFlag("handler.internal-address", handlerCmd.Flags().Lookup("internal-address"))
	viper.BindPFlag("handler.internal-address-announce", handlerCmd.Flags().Lookup("internal-address-announce"))
	viper.BindPFlag("handler.internal-port", handlerCmd.Flags().Lookup("internal-port"))

	handlerCmd.Flags().String("public-address", "0.0.0.0", "The IP address to listen for communication with the wild open")
	handlerCmd.Flags().Int("public-port", 1782, "The port for communication with the wild open")
	viper.BindPFlag("handler.public-address", handlerCmd.Flags().Lookup("public-address"))
	viper.BindPFlag("handler.public-port", handlerCmd.Flags().Lookup("public-port"))

	handlerCmd.Flags().String("mqtt-broker", "localhost:1883", "The address of the MQTT broker (uplink)")
	handlerCmd.Flags().String("mqtt-username", "handler", "The username for the MQTT broker (uplink)")
	handlerCmd.Flags().String("mqtt-password", "", "The password for the MQTT broker (uplink)")
	viper.BindPFlag("handler.mqtt-broker", handlerCmd.Flags().Lookup("mqtt-broker"))
	viper.BindPFlag("handler.mqtt-username", handlerCmd.Flags().Lookup("mqtt-username"))
	viper.BindPFlag("handler.mqtt-password", handlerCmd.Flags().Lookup("mqtt-password"))

	handlerCmd.Flags().String("redis-addr", "localhost:6379", "The address of Redis")
	handlerCmd.Flags().Int64("redis-db", 0, "The database of Redis")
	viper.BindPFlag("handler.redis-addr", handlerCmd.Flags().Lookup("redis-addr"))
	viper.BindPFlag("handler.redis-db", handlerCmd.Flags().Lookup("redis-db"))

	handlerCmd.Flags().String("ttn-broker", "localhost:1781", "The address of the TTN broker (downlink)")
	viper.BindPFlag("handler.ttn-broker", handlerCmd.Flags().Lookup("ttn-broker"))
}

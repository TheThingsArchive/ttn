// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	httpHandlers "github.com/TheThingsNetwork/ttn/core/adapters/http/handlers"
	"github.com/TheThingsNetwork/ttn/core/adapters/mqtt"
	mqttHandlers "github.com/TheThingsNetwork/ttn/core/adapters/mqtt/handlers"
	"github.com/TheThingsNetwork/ttn/core/components/handler"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// handlerCmd represents the handler command
var handlerCmd = &cobra.Command{
	Use:   "handler",
	Short: "The Things Network handler",
	Long: `
The default handler is the bridge between The Things Network and applications.
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		ctx.WithFields(log.Fields{
			"devicesDatabase": viper.GetString("handler.dev-database"),
			"packetsDatabase": viper.GetString("handler.pkt-database"),
			"uplink-port":     viper.GetInt("handler.uplink-port"),
			"mqtt-broker":     viper.GetString("handler.mqtt-broker"),
		}).Info("Using Configuration")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// ----- Start Adapters
		brkAdapter, err := http.NewAdapter(uint(viper.GetInt("handler.uplink-port")), nil, ctx.WithField("adapter", "broker-adapter"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start broker adapter")
		}
		brkAdapter.Bind(httpHandlers.Collect{})
		brkAdapter.Bind(httpHandlers.StatusPage{})
		brkAdapter.Bind(httpHandlers.Healthz{})

		mqttClient, err := mqtt.NewClient("handler-client", viper.GetString("handler.mqtt-broker"), mqtt.TCP)
		if err != nil {
			ctx.WithError(err).Fatal("Could not start MQTT client")
		}
		appAdapter := mqtt.NewAdapter(mqttClient, ctx.WithField("adapter", "app-adapter"))
		appAdapter.Bind(mqttHandlers.Activation{})

		// Instantiate in-memory devices storage

		var devicesDB handler.DevStorage

		devDBString := viper.GetString("handler.dev-database")
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

		// Instantiate in-memory packets storage

		var packetsDB handler.PktStorage

		pktDBString := viper.GetString("handler.pkt-database")
		switch {
		case strings.HasPrefix(pktDBString, "boltdb:"):

			pktDBPath, err := filepath.Abs(pktDBString[7:])
			if err != nil {
				ctx.WithError(err).Fatal("Invalid packets database path")
			}

			packetsDB, err = handler.NewPktStorage(pktDBPath)
			if err != nil {
				ctx.WithError(err).Fatal("Could not create local packets storage")
			}

			ctx.WithField("database", pktDBPath).Info("Using local storage")
		default:
			ctx.WithError(fmt.Errorf("Invalid database string. Format: \"boltdb:/path/to.db\".")).Fatal("Could not instantiate local packets storage")
		}

		// Instantiate the actual handler
		handler := handler.New(devicesDB, packetsDB, ctx)

		// Bring the service to life

		// Listen uplink
		go func() {
			for {
				packet, an, err := brkAdapter.Next()
				if err != nil {
					ctx.WithError(err).Warn("Could not get next packet fom brokers")
					continue
				}

				go func(packet []byte, an core.AckNacker) {
					if err := handler.HandleUp(packet, an, appAdapter); err != nil {
						ctx.WithError(err).Warn("Could not process packet from brokers")
					}
				}(packet, an)
			}
		}()

		// Listen downlink
		go func() {
			for {
				packet, an, err := appAdapter.Next()
				if err != nil {
					ctx.WithError(err).Warn("Could not get next packet fom applications")
					continue
				}

				go func(packet []byte, an core.AckNacker) {
					if err := handler.HandleDown(packet, an, brkAdapter); err != nil {
						ctx.WithError(err).Warn("Could not process packet from applications")
					}
				}(packet, an)
			}
		}()

		// Listen registrations
		go func() {
			for {
				reg, an, err := appAdapter.NextRegistration()
				if err != nil {
					ctx.WithError(err).Warn("Could not get next registration from applications")
					continue
				}

				go func(reg core.Registration, an core.AckNacker) {
					if err := handler.Register(reg, an); err != nil {
						ctx.WithError(err).Warn("Could not process registration from applications")
					}
				}(reg, an)
			}
		}()

		// Wait
		<-make(chan bool)
	},
}

func init() {
	RootCmd.AddCommand(handlerCmd)

	handlerCmd.Flags().String("dev-database", "boltdb:/tmp/ttn_handler_devices.db", "Devices Database connection")
	handlerCmd.Flags().String("pkt-database", "boltdb:/tmp/ttn_handler_packets.db", "Packets Database connection")
	viper.BindPFlag("handler.dev-database", handlerCmd.Flags().Lookup("dev-database"))
	viper.BindPFlag("handler.pkt-database", handlerCmd.Flags().Lookup("pkt-database"))

	handlerCmd.Flags().Int("uplink-port", 1882, "The UDP port for the uplink")
	viper.BindPFlag("handler.uplink-port", handlerCmd.Flags().Lookup("uplink-port"))

	handlerCmd.Flags().String("mqtt-broker", "localhost:1883", "The address of the MQTT broker")
	viper.BindPFlag("handler.mqtt-broker", handlerCmd.Flags().Lookup("mqtt-broker"))
}

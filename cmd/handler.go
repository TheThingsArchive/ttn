// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/handler"
	"github.com/TheThingsNetwork/ttn/core/proxy"
	"github.com/TheThingsNetwork/ttn/core/proxy/jsonpb"
	"github.com/TheThingsNetwork/ttn/utils/parse"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"gopkg.in/redis.v5"
)

// handlerCmd represents the handler command
var handlerCmd = &cobra.Command{
	Use:   "handler",
	Short: "The Things Network handler",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		ctx.WithFields(ttnlog.Fields{
			"Server":        fmt.Sprintf("%s:%d", viper.GetString("handler.server-address"), viper.GetInt("handler.server-port")),
			"HTTP Proxy":    fmt.Sprintf("%s:%d", viper.GetString("handler.http-address"), viper.GetInt("handler.http-port")),
			"Announce":      fmt.Sprintf("%s:%d", viper.GetString("handler.server-address-announce"), viper.GetInt("handler.server-port")),
			"Database":      fmt.Sprintf("%s/%d", viper.GetString("handler.redis-address"), viper.GetInt("handler.redis-db")),
			"TTN Broker ID": viper.GetString("handler.broker-id"),
			"MQTT":          viper.GetString("handler.mqtt-address"),
			"AMQP":          viper.GetString("handler.amqp-address"),
		}).Info("Initializing Handler")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Redis Client
		client := redis.NewClient(&redis.Options{
			Addr:     viper.GetString("handler.redis-address"),
			Password: "", // no password set
			DB:       viper.GetInt("handler.redis-db"),
		})

		connectRedis(client)

		// Component
		component, err := component.New(ttnlog.Get(), "handler", fmt.Sprintf("%s:%d", viper.GetString("handler.server-address-announce"), viper.GetInt("handler.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize component")
		}

		httpActive := viper.GetString("handler.http-address") != "" && viper.GetInt("handler.http-port") != 0
		if httpActive && component.Identity.ApiAddress == "" {
			component.Identity.ApiAddress = fmt.Sprintf("http://%s:%d", viper.GetString("handler.server-address-announce"), viper.GetInt("handler.http-port"))
		}

		// Handler
		handler := handler.NewRedisHandler(
			client,
			viper.GetString("handler.broker-id"),
		)
		if viper.GetString("handler.mqtt-address") != "" {
			handler = handler.WithMQTT(
				viper.GetString("handler.mqtt-username"),
				viper.GetString("handler.mqtt-password"),
				viper.GetString("handler.mqtt-address"),
			)

			mqttPort, err := parse.Port(viper.GetString("handler.mqtt-address"))
			if err != nil {
				ctx.WithError(err).Error("Could not announce the handler")
			}
			if announceAddr := viper.GetString("handler.mqtt-address-announce"); announceAddr != "" {
				component.Identity.MqttAddress = fmt.Sprintf("%s:%d", announceAddr, mqttPort)
			} else {
				component.Identity.MqttAddress = fmt.Sprintf("%s:%d", viper.GetString("handler.server-address-announce"), mqttPort)
			}
		} else {
			ctx.Warn("MQTT is not enabled in your configuration")
		}
		if viper.GetString("handler.amqp-address") != "" {
			handler = handler.WithAMQP(
				viper.GetString("handler.amqp-username"),
				viper.GetString("handler.amqp-password"),
				viper.GetString("handler.amqp-address"),
				viper.GetString("handler.amqp-exchange"),
			)

			amqpPort, err := parse.Port(viper.GetString("handler.amqp-address"))
			if err != nil {
				ctx.WithError(err).Error("Could not announce the handler")
			}
			if announceAddr := viper.GetString("handler.amqp-address-announce"); announceAddr != "" {
				component.Identity.AmqpAddress = fmt.Sprintf("%s:%d", announceAddr, amqpPort)
			} else {
				component.Identity.AmqpAddress = fmt.Sprintf("%s:%d", viper.GetString("handler.server-address-announce"), amqpPort)
			}
		} else {
			ctx.Warn("AMQP is not enabled in your configuration")
		}
		err = handler.Init(component)
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize handler")
		}
		defer handler.Shutdown()

		// gRPC Server
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viper.GetString("handler.server-address"), viper.GetInt("handler.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start gRPC server")
		}
		grpc := grpc.NewServer(component.ServerOptions()...)

		// Register and Listen
		component.RegisterHealthServer(grpc)
		handler.RegisterRPC(grpc)
		handler.RegisterManager(grpc)
		go grpc.Serve(lis)
		defer grpc.Stop()

		if httpActive {
			proxyConn, err := component.Identity.Dial()
			if err != nil {
				ctx.WithError(err).Fatal("Could not start client for gRPC proxy")
			}
			mux := runtime.NewServeMux(runtime.WithMarshalerOption("*", &jsonpb.GoGoJSONPb{
				OrigName: true,
			}))
			netCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pb.RegisterApplicationManagerHandler(netCtx, mux, proxyConn)

			prxy := proxy.WithToken(mux)
			prxy = proxy.WithPagination(prxy)
			prxy = proxy.WithLogger(prxy, ctx)

			go func() {
				err := http.ListenAndServe(
					fmt.Sprintf("%s:%d", viper.GetString("handler.http-address"), viper.GetInt("handler.http-port")),
					prxy,
				)
				if err != nil {
					ctx.WithError(err).Fatal("Error in gRPC proxy")
				}
			}()
		}

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		ctx.WithField("signal", <-sigChan).Info("signal received")

	},
}

func init() {
	RootCmd.AddCommand(handlerCmd)

	handlerCmd.Flags().String("redis-address", "localhost:6379", "Redis host and port")
	viper.BindPFlag("handler.redis-address", handlerCmd.Flags().Lookup("redis-address"))
	handlerCmd.Flags().Int("redis-db", 0, "Redis database")
	viper.BindPFlag("handler.redis-db", handlerCmd.Flags().Lookup("redis-db"))

	handlerCmd.Flags().String("broker-id", "dev", "The ID of the TTN Broker as announced in the Discovery server")
	viper.BindPFlag("handler.broker-id", handlerCmd.Flags().Lookup("broker-id"))

	handlerCmd.Flags().String("mqtt-address", "", "MQTT host and port. Leave empty to disable MQTT")
	handlerCmd.Flags().String("mqtt-address-announce", "", "MQTT address to announce (takes value of server-address-announce if empty while enabled)")
	handlerCmd.Flags().String("mqtt-username", "", "MQTT username")
	handlerCmd.Flags().String("mqtt-password", "", "MQTT password")
	viper.BindPFlag("handler.mqtt-address", handlerCmd.Flags().Lookup("mqtt-address"))
	viper.BindPFlag("handler.mqtt-address-announce", handlerCmd.Flags().Lookup("mqtt-address-announce"))
	viper.BindPFlag("handler.mqtt-username", handlerCmd.Flags().Lookup("mqtt-username"))
	viper.BindPFlag("handler.mqtt-password", handlerCmd.Flags().Lookup("mqtt-password"))

	handlerCmd.Flags().String("amqp-address", "", "AMQP host and port. Leave empty to disable AMQP")
	handlerCmd.Flags().String("amqp-address-announce", "", "AMQP address to announce (takes value of server-address-announce if empty while enabled)")
	handlerCmd.Flags().String("amqp-username", "guest", "AMQP username")
	handlerCmd.Flags().String("amqp-password", "guest", "AMQP password")
	handlerCmd.Flags().String("amqp-exchange", "ttn.handler", "AMQP exchange")
	viper.BindPFlag("handler.amqp-address", handlerCmd.Flags().Lookup("amqp-address"))
	viper.BindPFlag("handler.amqp-address-announce", handlerCmd.Flags().Lookup("amqp-address-announce"))
	viper.BindPFlag("handler.amqp-username", handlerCmd.Flags().Lookup("amqp-username"))
	viper.BindPFlag("handler.amqp-password", handlerCmd.Flags().Lookup("amqp-password"))
	viper.BindPFlag("handler.amqp-exchange", handlerCmd.Flags().Lookup("amqp-exchange"))

	handlerCmd.Flags().String("server-address", "0.0.0.0", "The IP address to listen for communication")
	handlerCmd.Flags().String("server-address-announce", "localhost", "The public IP address to announce")
	handlerCmd.Flags().Int("server-port", 1904, "The port for communication")
	viper.BindPFlag("handler.server-address", handlerCmd.Flags().Lookup("server-address"))
	viper.BindPFlag("handler.server-address-announce", handlerCmd.Flags().Lookup("server-address-announce"))
	viper.BindPFlag("handler.server-port", handlerCmd.Flags().Lookup("server-port"))

	handlerCmd.Flags().String("http-address", "0.0.0.0", "The IP address where the gRPC proxy should listen")
	handlerCmd.Flags().Int("http-port", 8084, "The port where the gRPC proxy should listen")
	viper.BindPFlag("handler.http-address", handlerCmd.Flags().Lookup("http-address"))
	viper.BindPFlag("handler.http-port", handlerCmd.Flags().Lookup("http-port"))
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/handler"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// handlerCmd represents the handler command
var handlerCmd = &cobra.Command{
	Use:   "handler",
	Short: "The Things Network handler",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		ctx.WithFields(log.Fields{
			"Server":     fmt.Sprintf("%s:%d", viper.GetString("handler.server-address"), viper.GetInt("handler.server-port")),
			"Announce":   fmt.Sprintf("%s:%d", viper.GetString("handler.server-address-announce"), viper.GetInt("handler.server-port")),
			"Database":   fmt.Sprintf("%s/%d", viper.GetString("handler.redis-address"), viper.GetInt("handler.redis-db")),
			"TTNbroker":  viper.GetString("handler.ttn-broker"),
			"MQTTbroker": viper.GetString("handler.mqtt-broker"),
		}).Info("Initializing Handler")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Redis Client
		client := redis.NewClient(&redis.Options{
			Addr:     viper.GetString("handler.redis-address"),
			Password: "", // no password set
			DB:       int64(viper.GetInt("handler.redis-db")),
		})

		connectRedis(client)

		// Component
		component := core.NewComponent(ctx, "handler", fmt.Sprintf("%s:%d", viper.GetString("handler.server-address-announce"), viper.GetInt("handler.server-port")))

		// Broker
		handler := handler.NewRedisHandler(
			client,
			viper.GetString("handler.ttn-broker"),
			viper.GetString("handler.mqtt-username"),
			viper.GetString("handler.mqtt-password"),
			viper.GetString("handler.mqtt-broker"),
		)
		err := handler.Init(component)
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize handler")
		}

		// gRPC Server
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viper.GetString("handler.server-address"), viper.GetInt("handler.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start gRPC server")
		}
		grpc := grpc.NewServer(component.ServerOptions()...)

		// Register and Listen
		handler.RegisterRPC(grpc)
		handler.RegisterManager(grpc)
		go grpc.Serve(lis)

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		ctx.WithField("signal", <-sigChan).Info("signal received")

		grpc.Stop()

	},
}

func init() {
	RootCmd.AddCommand(handlerCmd)

	handlerCmd.Flags().String("redis-address", "localhost:6379", "Redis host and port")
	viper.BindPFlag("handler.redis-address", handlerCmd.Flags().Lookup("redis-address"))
	handlerCmd.Flags().Int("redis-db", 0, "Redis database")
	viper.BindPFlag("handler.redis-db", handlerCmd.Flags().Lookup("redis-db"))

	handlerCmd.Flags().String("ttn-broker", "localhost:1902", "TTN broker host and port")
	viper.BindPFlag("handler.ttn-broker", handlerCmd.Flags().Lookup("ttn-broker"))

	handlerCmd.Flags().String("mqtt-broker", "localhost:1883", "MQTT broker host and port")
	viper.BindPFlag("handler.mqtt-broker", handlerCmd.Flags().Lookup("mqtt-broker"))

	handlerCmd.Flags().String("mqtt-username", "handler", "MQTT username")
	viper.BindPFlag("handler.mqtt-username", handlerCmd.Flags().Lookup("mqtt-username"))

	handlerCmd.Flags().String("mqtt-password", "", "MQTT password")
	viper.BindPFlag("handler.mqtt-password", handlerCmd.Flags().Lookup("mqtt-password"))

	handlerCmd.Flags().String("server-address", "0.0.0.0", "The IP address to listen for communication")
	handlerCmd.Flags().String("server-address-announce", "localhost", "The public IP address to announce")
	handlerCmd.Flags().Int("server-port", 1904, "The port for communication")
	viper.BindPFlag("handler.server-address", handlerCmd.Flags().Lookup("server-address"))
	viper.BindPFlag("handler.server-address-announce", handlerCmd.Flags().Lookup("server-address-announce"))
	viper.BindPFlag("handler.server-port", handlerCmd.Flags().Lookup("server-port"))
}

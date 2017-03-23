// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/networkserver"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gopkg.in/redis.v5"
)

// networkserverCmd represents the networkserver command
var networkserverCmd = &cobra.Command{
	Use:   "networkserver",
	Short: "The Things Network networkserver",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		ctx.WithFields(ttnlog.Fields{
			"Server":   fmt.Sprintf("%s:%d", viper.GetString("networkserver.server-address"), viper.GetInt("networkserver.server-port")),
			"Database": fmt.Sprintf("%s/%d", viper.GetString("networkserver.redis-address"), viper.GetInt("networkserver.redis-db")),
			"NetID":    viper.GetString("networkserver.net-id"),
		}).Info("Initializing Network Server")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Redis Client
		client := redis.NewClient(&redis.Options{
			Addr:     viper.GetString("networkserver.redis-address"),
			Password: viper.GetString("networkserver.redis-password"),
			DB:       viper.GetInt("networkserver.redis-db"),
		})

		if err := connectRedis(client); err != nil {
			ctx.WithError(err).Fatal("Could not initialize database connection")
		}

		// Component
		component, err := component.New(ttnlog.Get(), "networkserver", fmt.Sprintf("%s:%d", viper.GetString("networkserver.server-address-announce"), viper.GetInt("networkserver.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize component")
		}

		// networkserver Server
		networkserver := networkserver.NewRedisNetworkServer(client, viper.GetInt("networkserver.net-id"))

		// Register Prefixes
		for prefix, usage := range viper.GetStringMapString("networkserver.prefixes") {
			prefix, err := types.ParseDevAddrPrefix(prefix)
			if err != nil {
				ctx.WithError(err).Warn("Could not use DevAddr Prefix. Skipping.")
				continue
			}
			err = networkserver.UsePrefix(prefix, strings.Split(usage, ","))
			if err != nil {
				ctx.WithError(err).Fatal("Could not initialize networkserver")
				continue
			}
			ctx.Infof("Using DevAddr prefix %s (%v)", prefix, usage)
		}

		err = networkserver.Init(component)
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize networkserver")
		}

		// gRPC Server
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viper.GetString("networkserver.server-address"), viper.GetInt("networkserver.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start gRPC server")
		}
		grpc := grpc.NewServer(component.ServerOptions()...)

		// Register and Listen
		component.RegisterHealthServer(grpc)
		networkserver.RegisterRPC(grpc)
		networkserver.RegisterManager(grpc)
		go grpc.Serve(lis)

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		ctx.WithField("signal", <-sigChan).Info("signal received")

		grpc.Stop()
		networkserver.Shutdown()
	},
}

func init() {
	RootCmd.AddCommand(networkserverCmd)

	networkserverCmd.Flags().String("redis-address", "localhost:6379", "Redis server and port")
	viper.BindPFlag("networkserver.redis-address", networkserverCmd.Flags().Lookup("redis-address"))
	networkserverCmd.Flags().String("redis-password", "", "Redis password")
	viper.BindPFlag("networkserver.redis-password", networkserverCmd.Flags().Lookup("redis-password"))
	networkserverCmd.Flags().Int("redis-db", 0, "Redis database")
	viper.BindPFlag("networkserver.redis-db", networkserverCmd.Flags().Lookup("redis-db"))

	networkserverCmd.Flags().Int("net-id", 19, "LoRaWAN NetID")
	viper.BindPFlag("networkserver.net-id", networkserverCmd.Flags().Lookup("net-id"))

	viper.SetDefault("networkserver.prefixes", map[string]string{
		"26000000/20": "otaa,abp,world,local,private,testing",
	})

	networkserverCmd.Flags().String("server-address", "0.0.0.0", "The IP address to listen for communication")
	networkserverCmd.Flags().String("server-address-announce", "localhost", "The public IP address to announce")
	networkserverCmd.Flags().Int("server-port", 1903, "The port for communication")
	viper.BindPFlag("networkserver.server-address", networkserverCmd.Flags().Lookup("server-address"))
	viper.BindPFlag("networkserver.server-address-announce", networkserverCmd.Flags().Lookup("server-address-announce"))
	viper.BindPFlag("networkserver.server-port", networkserverCmd.Flags().Lookup("server-port"))
}

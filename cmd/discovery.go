// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/discovery"
	"github.com/TheThingsNetwork/ttn/core/discovery/announcement"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gopkg.in/redis.v5"
)

// discoveryCmd represents the discovery command
var discoveryCmd = &cobra.Command{
	Use:   "discovery",
	Short: "The Things Network discovery",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		ctx.WithFields(log.Fields{
			"Server":   fmt.Sprintf("%s:%d", viper.GetString("discovery.server-address"), viper.GetInt("discovery.server-port")),
			"Database": fmt.Sprintf("%s/%d", viper.GetString("discovery.redis-address"), viper.GetInt("discovery.redis-db")),
		}).Info("Initializing Discovery")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Redis Client
		client := redis.NewClient(&redis.Options{
			Addr:     viper.GetString("discovery.redis-address"),
			Password: "", // no password set
			DB:       viper.GetInt("discovery.redis-db"),
		})

		connectRedis(client)

		// Component
		component, err := component.New(ctx, "discovery", fmt.Sprintf("%s:%d", viper.GetString("discovery.server-address-announce"), viper.GetInt("discovery.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize component")
		}

		// Discovery Server
		discovery := discovery.NewRedisDiscovery(client)
		if viper.GetBool("discovery.cache") {
			discovery.WithCache(announcement.DefaultCacheOptions)
		}
		discovery.WithMasterAuthServers(viper.GetStringSlice("discovery.master-auth-servers")...)
		err = discovery.Init(component)
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize discovery")
		}

		// gRPC Server
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viper.GetString("discovery.server-address"), viper.GetInt("discovery.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start gRPC server")
		}
		grpc := grpc.NewServer(component.ServerOptions()...)

		// Register and Listen
		component.RegisterHealthServer(grpc)
		discovery.RegisterRPC(grpc)
		go grpc.Serve(lis)

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		ctx.WithField("signal", <-sigChan).Info("signal received")

		grpc.Stop()
	},
}

func init() {
	RootCmd.AddCommand(discoveryCmd)

	discoveryCmd.Flags().String("redis-address", "localhost:6379", "Redis server and port")
	viper.BindPFlag("discovery.redis-address", discoveryCmd.Flags().Lookup("redis-address"))
	discoveryCmd.Flags().Int("redis-db", 0, "Redis database")
	viper.BindPFlag("discovery.redis-db", discoveryCmd.Flags().Lookup("redis-db"))

	discoveryCmd.Flags().String("server-address", "0.0.0.0", "The IP address to listen for communication")
	discoveryCmd.Flags().Int("server-port", 1900, "The port for communication")
	viper.BindPFlag("discovery.server-address", discoveryCmd.Flags().Lookup("server-address"))
	viper.BindPFlag("discovery.server-port", discoveryCmd.Flags().Lookup("server-port"))

	discoveryCmd.Flags().Bool("cache", false, "Add a cache in front of the database")
	viper.BindPFlag("discovery.cache", discoveryCmd.Flags().Lookup("cache"))

	discoveryCmd.Flags().StringSlice("master-auth-servers", []string{"ttn-account"}, "Auth servers that are allowed to manage this network")
	viper.BindPFlag("discovery.master-auth-servers", discoveryCmd.Flags().Lookup("master-auth-servers"))
}

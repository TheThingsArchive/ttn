// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/discovery"
	"github.com/TheThingsNetwork/ttn/core/discovery/announcement"
	"github.com/TheThingsNetwork/ttn/core/proxy"
	"github.com/TheThingsNetwork/ttn/core/proxy/jsonpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
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
		ctx.WithFields(ttnlog.Fields{
			"Server":     fmt.Sprintf("%s:%d", viper.GetString("discovery.server-address"), viper.GetInt("discovery.server-port")),
			"HTTP Proxy": fmt.Sprintf("%s:%d", viper.GetString("discovery.http-address"), viper.GetInt("discovery.http-port")),
			"Database":   fmt.Sprintf("%s/%d", viper.GetString("discovery.redis-address"), viper.GetInt("discovery.redis-db")),
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

		if err := connectRedis(client); err != nil {
			ctx.WithError(err).Fatal("Could not initialize database connection")
		}

		// Component
		component, err := component.New(ttnlog.Get(), "discovery", fmt.Sprintf("%s:%d", "localhost", viper.GetInt("discovery.server-port")))
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

		if viper.GetString("discovery.http-address") != "" && viper.GetInt("discovery.http-port") != 0 {
			proxyConn, err := component.Identity.Dial()
			if err != nil {
				ctx.WithError(err).Fatal("Could not start client for gRPC proxy")
			}
			mux := runtime.NewServeMux(runtime.WithMarshalerOption("*", &jsonpb.GoGoJSONPb{
				OrigName: true,
			}))
			netCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pb.RegisterDiscoveryHandler(netCtx, mux, proxyConn)

			prxy := proxy.WithLogger(mux, ctx)
			prxy = proxy.WithPagination(prxy)

			go func() {
				err := http.ListenAndServe(
					fmt.Sprintf("%s:%d", viper.GetString("discovery.http-address"), viper.GetInt("discovery.http-port")),
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

	discoveryCmd.Flags().StringSlice("master-auth-servers", []string{"ttn-account-v2"}, "Auth servers that are allowed to manage this network")
	viper.BindPFlag("discovery.master-auth-servers", discoveryCmd.Flags().Lookup("master-auth-servers"))

	discoveryCmd.Flags().String("http-address", "0.0.0.0", "The IP address where the gRPC proxy should listen")
	discoveryCmd.Flags().Int("http-port", 8080, "The port where the gRPC proxy should listen")
	viper.BindPFlag("discovery.http-address", discoveryCmd.Flags().Lookup("http-address"))
	viper.BindPFlag("discovery.http-port", discoveryCmd.Flags().Lookup("http-port"))
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/router"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// routerCmd represents the router command
var routerCmd = &cobra.Command{
	Use:   "router",
	Short: "The Things Network router",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		ctx.WithFields(ttnlog.Fields{
			"Server":   fmt.Sprintf("%s:%d", viper.GetString("router.server-address"), viper.GetInt("router.server-port")),
			"Announce": fmt.Sprintf("%s:%d", viper.GetString("router.server-address-announce"), viper.GetInt("router.server-port")),
		}).Info("Initializing Router")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Component
		component, err := component.New(ttnlog.Get(), "router", fmt.Sprintf("%s:%d", viper.GetString("router.server-address-announce"), viper.GetInt("router.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize component")
		}

		// Router
		router := router.NewRouter()
		err = router.Init(component)
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize router")
		}

		// gRPC Server
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viper.GetString("router.server-address"), viper.GetInt("router.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start gRPC server")
		}
		grpc := grpc.NewServer(component.ServerOptions()...)

		// Register and Listen
		component.RegisterHealthServer(grpc)
		router.RegisterRPC(grpc)
		router.RegisterManager(grpc)
		go grpc.Serve(lis)

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		ctx.WithField("signal", <-sigChan).Info("signal received")

		grpc.Stop()
		router.Shutdown()
	},
}

func init() {
	RootCmd.AddCommand(routerCmd)
	routerCmd.Flags().String("server-address", "0.0.0.0", "The IP address to listen for communication")
	routerCmd.Flags().String("server-address-announce", "localhost", "The public IP address to announce")
	routerCmd.Flags().Int("server-port", 1901, "The port for communication")
	routerCmd.Flags().Bool("skip-verify-gateway-token", false, "Skip verification of the gateway token")
	viper.BindPFlag("router.server-address", routerCmd.Flags().Lookup("server-address"))
	viper.BindPFlag("router.server-address-announce", routerCmd.Flags().Lookup("server-address-announce"))
	viper.BindPFlag("router.server-port", routerCmd.Flags().Lookup("server-port"))
	viper.BindPFlag("router.skip-verify-gateway-token", routerCmd.Flags().Lookup("skip-verify-gateway-token"))
}

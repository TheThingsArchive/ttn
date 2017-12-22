// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/broker"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// brokerCmd represents the broker command
var brokerCmd = &cobra.Command{
	Use:   "broker",
	Short: "The Things Network broker",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		ctx.WithFields(ttnlog.Fields{
			"Server":             fmt.Sprintf("%s:%d", viper.GetString("broker.server-address"), viper.GetInt("broker.server-port")),
			"Announce":           fmt.Sprintf("%s:%d", viper.GetString("broker.server-address-announce"), viper.GetInt("broker.server-port")),
			"NetworkServer":      viper.GetString("broker.networkserver-address"),
			"DeduplicationDelay": viper.GetString("broker.deduplication-delay"),
		}).Info("Initializing Broker")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		// Component
		component, err := component.New(ttnlog.Get(), "broker", fmt.Sprintf("%s:%d", viper.GetString("broker.server-address-announce"), viper.GetInt("broker.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize component")
		}

		var nsCert string
		if nsCertFile := viper.GetString("broker.networkserver-cert"); nsCertFile != "" {
			contents, err := ioutil.ReadFile(nsCertFile)
			if err != nil {
				ctx.WithError(err).Fatal("Could not get Networkserver certificate")
			}
			nsCert = string(contents)
		}

		// Broker
		broker := broker.NewBroker(
			time.Duration(viper.GetInt("broker.deduplication-delay")) * time.Millisecond,
		)
		broker.SetNetworkServer(viper.GetString("broker.networkserver-address"), nsCert, viper.GetString("broker.networkserver-token"))
		err = broker.Init(component)
		if err != nil {
			ctx.WithError(err).Fatal("Could not initialize broker")
		}

		// gRPC Server
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viper.GetString("broker.server-address"), viper.GetInt("broker.server-port")))
		if err != nil {
			ctx.WithError(err).Fatal("Could not start gRPC server")
		}
		grpc := grpc.NewServer(component.ServerOptions()...)

		// Register and Listen
		broker.RegisterRPC(grpc)
		broker.RegisterManager(grpc)
		component.RegisterHealthServer(grpc) // must be last one

		go grpc.Serve(lis)

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		ctx.WithField("signal", <-sigChan).Info("signal received")

		grpc.Stop()
		broker.Shutdown()
	},
}

func init() {
	RootCmd.AddCommand(brokerCmd)

	brokerCmd.Flags().String("networkserver-address", "localhost:1903", "Networkserver host and port")
	viper.BindPFlag("broker.networkserver-address", brokerCmd.Flags().Lookup("networkserver-address"))
	brokerCmd.Flags().String("networkserver-cert", "", "Networkserver certificate to use")
	viper.BindPFlag("broker.networkserver-cert", brokerCmd.Flags().Lookup("networkserver-cert"))
	brokerCmd.Flags().String("networkserver-token", "", "Networkserver token to use")
	viper.BindPFlag("broker.networkserver-token", brokerCmd.Flags().Lookup("networkserver-token"))

	brokerCmd.Flags().Int("deduplication-delay", 200, "Deduplication delay (in ms)")
	viper.BindPFlag("broker.deduplication-delay", brokerCmd.Flags().Lookup("deduplication-delay"))

	brokerCmd.Flags().String("server-address", "0.0.0.0", "The IP address to listen for communication")
	brokerCmd.Flags().String("server-address-announce", "localhost", "The public IP address to announce")
	brokerCmd.Flags().Int("server-port", 1902, "The port for communication")
	viper.BindPFlag("broker.server-address", brokerCmd.Flags().Lookup("server-address"))
	viper.BindPFlag("broker.server-address-announce", brokerCmd.Flags().Lookup("server-address-announce"))
	viper.BindPFlag("broker.server-port", brokerCmd.Flags().Lookup("server-port"))
}

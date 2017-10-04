// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/TheThingsNetwork/api/discovery"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/pool"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc/metadata"
)

// brokerRegisterPrefixCmd represents the secure command
var brokerRegisterPrefixCmd = &cobra.Command{
	Use:   "register-prefix [prefix ...]",
	Short: "Register a prefix to this Broker",
	Long:  `ttn broker register prefix registers a prefix to this Broker`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.UsageFunc()(cmd)
			return
		}

		path := filepath.Clean(viper.GetString("key-dir") + "/ca.cert")
		cert, err := ioutil.ReadFile(path)
		if err == nil && !pool.RootCAs.AppendCertsFromPEM(cert) {
			ctx.Warnf("Could not add root certificates from %s", path)
		}

		conn, err := api.Dial(viper.GetString("discovery-address"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not connect to Discovery server")
		}
		client := discovery.NewDiscoveryClient(conn)

		client.GetAll(context.Background(), &discovery.GetServiceRequest{})

		md := metadata.Pairs(
			"service-name", "broker",
			"id", viper.GetString("id"),
			"token", viper.GetString("auth-token"),
		)
		dscContext := metadata.NewOutgoingContext(context.Background(), md)

		success := true
		for _, prefixString := range args {
			ctx := ctx.WithField("Prefix", prefixString)
			prefix, err := types.ParseDevAddrPrefix(prefixString)
			if err != nil {
				ctx.WithError(err).Error("Could not register prefix")
				success = false
				continue
			}
			_, err = client.AddMetadata(dscContext, &discovery.MetadataRequest{
				ServiceName: "broker",
				ID:          viper.GetString("id"),
				Metadata: discovery.Metadata{Metadata: &discovery.Metadata_DevAddrPrefix{
					DevAddrPrefix: prefix.Bytes(),
				}},
			})
			if err != nil {
				ctx.WithError(err).Error("Could not register prefix")
				success = false
				continue
			}
			ctx.Info("Registered prefix")
		}

		if !success {
			os.Exit(1)
		}
	},
}

func init() {
	brokerCmd.AddCommand(brokerRegisterPrefixCmd)
}

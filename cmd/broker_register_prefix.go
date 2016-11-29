// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
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
		}

		conn, err := grpc.Dial(viper.GetString("discovery-address"), append(api.DialOptions, grpc.WithInsecure(), grpc.WithBlock())...)
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
		dscContext := metadata.NewContext(context.Background(), md)

		for _, prefixString := range args {
			ctx := ctx.WithField("Prefix", prefixString)
			prefix, err := types.ParseDevAddrPrefix(prefixString)
			if err != nil {
				ctx.WithError(err).Error("Could not register prefix")
				continue
			}
			_, err = client.AddMetadata(dscContext, &discovery.MetadataRequest{
				ServiceName: "broker",
				Id:          viper.GetString("id"),
				Metadata: &discovery.Metadata{Metadata: &discovery.Metadata_DevAddrPrefix{
					DevAddrPrefix: prefix.Bytes(),
				}},
			})
			if err != nil {
				ctx.WithError(err).Error("Could not register prefix")
			}
			ctx.Info("Registered prefix")
		}
	},
}

func init() {
	brokerCmd.AddCommand(brokerRegisterPrefixCmd)
}

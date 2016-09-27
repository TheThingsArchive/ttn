// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
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

		conn, err := grpc.Dial(viper.GetString("discovery-server"), append(api.DialOptions, grpc.WithInsecure())...)
		if err != nil {
			ctx.WithError(err).Fatal("Could not connect to Discovery server")
		}
		client := discovery.NewDiscoveryClient(conn)

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
				Metadata: &discovery.Metadata{
					Key:   discovery.Metadata_PREFIX,
					Value: []byte{byte(prefix.Length), prefix.DevAddr[0], prefix.DevAddr[1], prefix.DevAddr[2], prefix.DevAddr[3]},
				},
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

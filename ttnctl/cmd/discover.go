// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/spf13/cobra"
)

const serviceFmt = "%-36s %-36s %-20s %-6s\n"

var discoverCmd = &cobra.Command{
	Use:    "discover [ServiceType]",
	Short:  "Discover routing services",
	Long:   `ttnctl discover is used to discover routing services`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		serviceType := strings.TrimRight(args[0], "s") // Allow both singular and plural

		switch serviceType {
		case "router":
			ctx.Info("Discovering routers...")
		case "broker":
			ctx.Info("Discovering brokers and their prefixes...")
		case "handler":
			ctx.Info("Discovering handlers and their apps...")
		default:
			ctx.Fatalf("Service type %s unknown", serviceType)
		}

		conn, client := util.GetDiscovery(ctx)
		defer conn.Close()

		res, err := client.GetAll(util.GetContext(ctx), &discovery.GetServiceRequest{
			ServiceName: serviceType,
		})
		if err != nil {
			ctx.WithError(errors.FromGRPCError(err)).Fatalf("Could not get %ss", serviceType)
		}

		ctx.Infof("Discovered %d %ss", len(res.Services), serviceType)

		fmt.Printf(serviceFmt, "ID", "ADDRESS", "VERSION", "PUBLIC")
		fmt.Printf(serviceFmt, "==", "=======", "=======", "======")
		fmt.Println()
		for _, service := range res.Services {
			fmt.Printf(serviceFmt, service.Id, crop(service.NetAddress, 36), crop(service.ServiceVersion, 20), fmt.Sprintf("%v", service.Public))
			if showMetadata, _ := cmd.Flags().GetBool("metadata"); showMetadata {
				switch serviceType {
				case "broker":
					fmt.Println("  DevAddr Prefixes:")
					for _, prefix := range service.DevAddrPrefixes() {
						min := types.DevAddr{0x00, 0x00, 0x00, 0x00}.WithPrefix(prefix)
						max := types.DevAddr{0xff, 0xff, 0xff, 0xff}.WithPrefix(prefix)
						fmt.Printf("   %s (%s-%s)\n", prefix, min, max)
					}
				case "handler":
					fmt.Println("  AppIDs:")
					for _, appID := range service.AppIDs() {
						fmt.Println("   ", appID)
					}
				}
				fmt.Println()
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(discoverCmd)
	discoverCmd.Flags().Bool("metadata", false, "Show additional metadata")
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"strings"
	"time"

	"github.com/TheThingsNetwork/api"
	"github.com/TheThingsNetwork/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/health"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [ServiceType] [ServiceID]",
	Short: "Check routing services",
	Long:  `ttnctl components check is used to check the status of routing services`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)

		serviceType := strings.ToLower(args[0])
		switch serviceType {
		case "router", "broker", "handler":
		default:
			ctx.Fatalf("Service type %s unknown", serviceType)
		}

		serviceID := strings.ToLower(args[1])
		if err := api.NotEmptyAndValidID(serviceID, "Service ID"); err != nil {
			ctx.Fatal(err.Error())
		}

		dscConn, client := util.GetDiscovery(ctx)
		defer dscConn.Close()

		res, err := client.Get(util.GetContext(ctx), &discovery.GetRequest{
			ServiceName: serviceType,
			ID:          serviceID,
		})
		if err != nil {
			ctx.WithError(errors.FromGRPCError(err)).Fatalf("Could not get %s %s", serviceType, serviceID)
		}

		conn, err := res.Dial(nil)
		if err != nil {
			ctx.WithError(errors.FromGRPCError(err)).Fatalf("Could not dial %s %s", serviceType, serviceID)
		}
		defer conn.Close()

		start := time.Now()
		ok, err := health.Check(conn)
		if err != nil {
			ctx.WithError(errors.FromGRPCError(err)).Fatalf("Could not check %s %s", serviceType, serviceID)
		}
		ctx = ctx.WithField("Duration", time.Now().Sub(start))

		if ok {
			ctx.Infof("%s %s is up and running", serviceType, serviceID)
		} else {
			ctx.Warnf("%s %s is not feeling well", serviceType, serviceID)
		}

	},
}

func init() {
	componentsCmd.AddCommand(checkCmd)
}

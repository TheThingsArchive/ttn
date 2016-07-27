// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/apex/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// GetRouter starts a connection with the router
func GetRouter(ctx log.Interface) (*grpc.ClientConn, router.RouterClient) {
	dscConn, client := GetDiscovery(ctx)
	defer dscConn.Close()
	routerAnnouncement, err := client.Get(GetContext(ctx), &discovery.GetRequest{
		ServiceName: "router",
		Id:          viper.GetString("ttn-router"),
	})
	if err != nil {
		ctx.WithError(err).Fatal("Could not get Router from Discovery")
	}
	rtrConn, err := routerAnnouncement.Dial()
	if err != nil {
		ctx.WithError(err).Fatal("Could not connect to Router")
	}
	return rtrConn, router.NewRouterClient(rtrConn)
}

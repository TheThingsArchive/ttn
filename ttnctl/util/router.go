// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// GetRouter starts a connection with the router
func GetRouter(ctx ttnlog.Interface) (*grpc.ClientConn, *router.Client) {
	ctx.Info("Discovering Router...")
	dscConn, client := GetDiscovery(ctx)
	defer dscConn.Close()
	routerAnnouncement, err := client.Get(GetContext(ctx), &discovery.GetRequest{
		ServiceName: "router",
		Id:          viper.GetString("router-id"),
	})
	if err != nil {
		ctx.WithError(errors.FromGRPCError(err)).Fatal("Could not get Router from Discovery")
	}
	ctx.Info("Connecting with Router...")
	rtrConn, err := routerAnnouncement.Dial(nil)
	ctx.Info("Connected to Router")
	rtrClient := router.NewClient(router.DefaultClientConfig)
	rtrClient.AddServer(viper.GetString("router-id"), rtrConn)
	return rtrConn, rtrClient
}

// GetRouterManager starts a management connection with the router
func GetRouterManager(ctx ttnlog.Interface) (*grpc.ClientConn, router.RouterManagerClient) {
	ctx.Info("Discovering Router...")
	dscConn, client := GetDiscovery(ctx)
	defer dscConn.Close()
	routerAnnouncement, err := client.Get(GetContext(ctx), &discovery.GetRequest{
		ServiceName: "router",
		Id:          viper.GetString("router-id"),
	})
	if err != nil {
		ctx.WithError(errors.FromGRPCError(err)).Fatal("Could not get Router from Discovery")
	}
	ctx.Info("Connecting with Router...")
	rtrConn, err := routerAnnouncement.Dial(nil)
	if err != nil {
		ctx.WithError(err).Fatal("Could not connect to Router")
	}
	ctx.Info("Connected to Router")
	return rtrConn, router.NewRouterManagerClient(rtrConn)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/apex/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// GetHandlerManager gets a new HandlerManager for ttnctl
func GetHandlerManager(ctx log.Interface) (*grpc.ClientConn, *handler.ManagerClient) {
	dscConn, client := GetDiscovery(ctx)
	defer dscConn.Close()
	handlerAnnouncement, err := client.Get(GetContext(ctx), &discovery.GetRequest{
		ServiceName: "handler",
		Id:          viper.GetString("ttn-handler"),
	})
	if err != nil {
		ctx.WithError(err).Fatal("Could not find Handler")
	}
	token, err := GetTokenSource(ctx).Token()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get token")
	}
	hdlConn, err := handlerAnnouncement.Dial()
	if err != nil {
		ctx.WithError(err).Fatal("Could not connect to Handler")
	}
	managerClient, err := handler.NewManagerClient(hdlConn, token.AccessToken)
	if err != nil {
		ctx.WithError(err).Fatal("Could not create Handler Manager")
	}
	return hdlConn, managerClient
}

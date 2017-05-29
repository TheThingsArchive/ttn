// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"github.com/TheThingsNetwork/go-account-lib/scope"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// GetHandlerManager gets a new HandlerManager for ttnctl
func GetHandlerManager(ctx ttnlog.Interface, appID string) (*grpc.ClientConn, *handler.ManagerClient) {
	ctx.WithField("Handler", viper.GetString("handler-id")).Info("Discovering Handler...")
	dscConn, client := GetDiscovery(ctx)
	defer dscConn.Close()
	handlerAnnouncement, err := client.Get(GetContext(ctx), &discovery.GetRequest{
		ServiceName: "handler",
		Id:          viper.GetString("handler-id"),
	})
	if err != nil {
		ctx.WithError(errors.FromGRPCError(err)).Fatal("Could not find Handler")
	}

	token := TokenForScope(ctx, scope.App(appID))

	ctx.WithField("Handler", handlerAnnouncement.NetAddress).Info("Connecting with Handler...")
	hdlConn, err := handlerAnnouncement.Dial(nil)
	if err != nil {
		ctx.WithError(err).Fatal("Could not connect to Handler")
	}
	managerClient, err := handler.NewManagerClientWithToken(hdlConn, token)
	if err != nil {
		ctx.WithError(err).Fatal("Could not create Handler Manager")
	}
	return hdlConn, managerClient
}

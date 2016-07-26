// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/apex/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// GetDiscovery gets the Discovery client for ttnctl
func GetDiscovery(ctx log.Interface) (*grpc.ClientConn, discovery.DiscoveryClient) {
	conn, err := grpc.Dial(viper.GetString("discovery-server"), append(api.DialOptions, grpc.WithInsecure())...)
	if err != nil {
		ctx.WithError(err).Fatal("Could not connect to Discovery server")
	}
	return conn, discovery.NewDiscoveryClient(conn)
}

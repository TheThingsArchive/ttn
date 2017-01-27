// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"io/ioutil"
	"path"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// GetDiscovery gets the Discovery client for ttnctl
func GetDiscovery(ctx ttnlog.Interface) (*grpc.ClientConn, discovery.DiscoveryClient) {
	path := path.Join(GetDataDir(), "/ca.cert")
	cert, err := ioutil.ReadFile(path)
	if err == nil && !api.RootCAs.AppendCertsFromPEM(cert) {
		ctx.Warnf("Could not add root certificates from %s", path)
	}

	conn, err := api.Dial(viper.GetString("discovery-address"))
	if err != nil {
		ctx.WithError(err).Fatal("Could not connect to Discovery server")
	}
	return conn, discovery.NewDiscoveryClient(conn)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"os"
	"os/user"

	"github.com/apex/log"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// GetContext returns a new context
func GetContext(ctx log.Interface) context.Context {
	id := "client"
	if user, err := user.Current(); err == nil {
		id += "-" + user.Username
	}
	if hostname, err := os.Hostname(); err == nil {
		id += "@" + hostname
	}
	token, err := GetTokenSource(ctx).Token()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get token")
	}
	md := metadata.Pairs(
		"id", id,
		"service-name", "ttnctl",
		"token", token.AccessToken,
	)
	return metadata.NewContext(context.Background(), md)
}

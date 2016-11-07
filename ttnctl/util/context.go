// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"os"
	"os/user"

	"github.com/apex/log"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc/metadata"
)

// GetID retrns the ID of this ttnctl
func GetID() string {
	id := "ttnctl"
	if user, err := user.Current(); err == nil {
		id += "-" + user.Username
	}
	if hostname, err := os.Hostname(); err == nil {
		id += "@" + hostname
	}
	return id
}

// GetContext returns a new context
func GetContext(ctx log.Interface, extraPairs ...string) context.Context {
	token, err := GetTokenSource(ctx).Token()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get token")
	}
	md := metadata.Pairs(
		"id", GetID(),
		"service-name", "ttnctl",
		"token", token.AccessToken,
	)
	return metadata.NewContext(context.Background(), md)
}

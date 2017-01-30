// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"os"
	"os/user"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/spf13/viper"
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
func GetContext(ctx ttnlog.Interface, extraPairs ...string) context.Context {
	token, err := GetTokenSource(ctx).Token()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get token")
	}
	md := metadata.Pairs(
		"id", GetID(),
		"service-name", "ttnctl",
		"service-version", fmt.Sprintf("%s-%s (%s)", viper.GetString("version"), viper.GetString("gitCommit"), viper.GetString("buildDate")),
		"token", token.AccessToken,
	)
	return metadata.NewContext(context.Background(), md)
}

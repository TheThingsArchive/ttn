// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"os"
	"os/user"

	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/spf13/viper"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
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
func GetContext(log ttnlog.Interface, extraPairs ...string) context.Context {
	token, err := GetTokenSource(log).Token()
	if err != nil {
		log.WithError(err).Fatal("Could not get token")
	}
	ctx := context.Background()
	ctx = ttnctx.OutgoingContextWithID(ctx, GetID())
	ctx = ttnctx.OutgoingContextWithServiceInfo(ctx, "ttnctl", fmt.Sprintf("%s-%s (%s)", viper.GetString("version"), viper.GetString("gitCommit"), viper.GetString("buildDate")), "")
	ctx = ttnctx.OutgoingContextWithToken(ctx, token.AccessToken)
	return ctx
}

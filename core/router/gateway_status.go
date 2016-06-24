// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"time"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
)

func (r *router) HandleGatewayStatus(gatewayEUI types.GatewayEUI, status *pb_gateway.Status) error {
	ctx := r.Ctx.WithField("GatewayEUI", gatewayEUI)
	var err error
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle gateway status")
		}
	}()

	gateway := r.getGateway(gatewayEUI)
	gateway.LastSeen = time.Now()
	err = gateway.Status.Update(status)
	if err != nil {
		return err
	}
	return nil
}

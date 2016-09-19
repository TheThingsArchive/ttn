// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"time"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
)

func (r *router) HandleGatewayStatus(gatewayID string, status *pb_gateway.Status) error {
	ctx := r.Ctx.WithField("GatewayID", gatewayID)
	var err error
	start := time.Now()
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle gateway status")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled gateway status")
		}
	}()

	gateway := r.getGateway(gatewayID)
	gateway.LastSeen = time.Now()
	err = gateway.Status.Update(status)
	if err != nil {
		return err
	}
	return nil
}

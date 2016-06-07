package router

import (
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
	err = gateway.Status.Update(status)
	if err != nil {
		return err
	}
	return nil
}

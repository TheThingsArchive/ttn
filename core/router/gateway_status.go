package router

import (
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
)

func (r *router) HandleGatewayStatus(gatewayEUI types.GatewayEUI, status *pb_gateway.StatusMessage) error {
	gateway := r.getGateway(gatewayEUI)
	return gateway.Status.Update(status)
}

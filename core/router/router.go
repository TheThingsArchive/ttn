package router

import (
	"github.com/TheThingsNetwork/ttn/api/gateway"
	pb "github.com/TheThingsNetwork/ttn/api/router"
)

type Router interface {
	HandleGatewayStatus(status *gateway.StatusMessage) error
	HandleUplink(uplink *pb.UplinkMessage) error
	HandleDownlink() error
	HandleActivation(activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error)
}

type router struct {
	gatewayStatusStore GatewayStatusStore
}

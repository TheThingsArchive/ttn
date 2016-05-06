package gateway

import "github.com/TheThingsNetwork/ttn/core/types"

func NewGateway(eui types.GatewayEUI) *Gateway {
	return &Gateway{
		EUI:         eui,
		Status:      NewStatusStore(),
		Utilization: NewUtilization(),
		Schedule:    NewSchedule(),
	}
}

type Gateway struct {
	EUI         types.GatewayEUI
	Status      StatusStore
	Utilization Utilization
	Schedule    Schedule
}

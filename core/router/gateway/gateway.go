package gateway

import "github.com/TheThingsNetwork/ttn/core/types"

// NewGateway creates a new in-memory Gateway structure
func NewGateway(eui types.GatewayEUI) *Gateway {
	return &Gateway{
		EUI:         eui,
		Status:      NewStatusStore(),
		Utilization: NewUtilization(),
		Schedule:    NewSchedule(),
	}
}

// Gateway contains the state of a gateway
type Gateway struct {
	EUI         types.GatewayEUI
	Status      StatusStore
	Utilization Utilization
	Schedule    Schedule
}

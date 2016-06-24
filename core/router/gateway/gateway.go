// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/apex/log"
)

// NewGateway creates a new in-memory Gateway structure
func NewGateway(ctx log.Interface, eui types.GatewayEUI) *Gateway {
	return &Gateway{
		EUI:         eui,
		Status:      NewStatusStore(),
		Utilization: NewUtilization(),
		Schedule:    NewSchedule(ctx.WithField("GatewayEUI", eui)),
	}
}

// Gateway contains the state of a gateway
type Gateway struct {
	EUI         types.GatewayEUI
	Status      StatusStore
	Utilization Utilization
	Schedule    Schedule
	LastSeen    time.Time
}

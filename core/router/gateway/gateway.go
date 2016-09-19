// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"time"

	"github.com/apex/log"
)

// NewGateway creates a new in-memory Gateway structure
func NewGateway(ctx log.Interface, id string) *Gateway {
	return &Gateway{
		ID:          id,
		Status:      NewStatusStore(),
		Utilization: NewUtilization(),
		Schedule:    NewSchedule(ctx.WithField("GatewayID", id)),
	}
}

// Gateway contains the state of a gateway
type Gateway struct {
	ID          string
	Status      StatusStore
	Utilization Utilization
	Schedule    Schedule
	LastSeen    time.Time
}

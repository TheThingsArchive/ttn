// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"time"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
)

func (g *Gateway) HandleStatus(status *pb_gateway.Status) (err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	status.GatewayTrusted = g.status.GatewayTrusted // we take this from the existing status
	status.Messages = append(g.status.Messages, status.Messages...)
	if len(status.Messages) > 10 {
		status.Messages = status.Messages[len(status.Messages)-10:] // we keep the last 10 messages
	}
	g.status = *status
	g.lastSeen = time.Now()
	if g.frequencyPlan == nil && status.FrequencyPlan != "" {
		g.setFrequencyPlan(status.FrequencyPlan)
	}
	g.sendToMonitor(status)
	return nil
}

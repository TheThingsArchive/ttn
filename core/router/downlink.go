// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"strings"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/fields"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

func (r *router) SubscribeDownlink(gatewayID string) (<-chan *pb.DownlinkMessage, error) {
	ctx := r.Ctx.WithField("GatewayID", gatewayID)
	gateway := r.getGateway(gatewayID)
	if fromSchedule := gateway.SubscribeDownlink(); fromSchedule != nil {
		toGateway := make(chan *pb.DownlinkMessage)
		go func() {
			ctx.Debug("Activate downlink")
			for message := range fromSchedule {
				gateway.HandleDownlink(message)
				ctx.WithFields(fields.Get(message)).Debug("Send downlink")
				toGateway <- message
			}
			ctx.Debug("Deactivate downlink")
			close(toGateway)
		}()
		return toGateway, nil
	}
	return nil, errors.NewErrInternal(fmt.Sprintf("Could not subscribe to downlink for %s", gatewayID))
}

func (r *router) UnsubscribeDownlink(gatewayID string) error {
	r.getGateway(gatewayID).StopDownlink()
	return nil
}

func (r *router) HandleDownlink(downlink *pb_broker.DownlinkMessage) (err error) {
	r.status.downlink.Mark(1)
	downlink.Trace = downlink.Trace.WithEvent(trace.ReceiveEvent)
	option := downlink.DownlinkOption
	downlinkMessage := &pb.DownlinkMessage{
		Payload:               downlink.Payload,
		ProtocolConfiguration: option.ProtocolConfig,
		GatewayConfiguration:  option.GatewayConfig,
		Trace:                 downlink.Trace,
	}
	identifier := option.Identifier
	if r.Component != nil && r.Component.Identity != nil {
		identifier = strings.TrimPrefix(option.Identifier, fmt.Sprintf("%s:", r.Component.Identity.Id))
	}
	if identifier == "" {
		return errors.NewErrInvalidArgument("Downlink", "identifier missing")
	}
	gateway := r.getGateway(downlink.DownlinkOption.GatewayId)
	return gateway.ScheduleDownlink(identifier, downlinkMessage)
}

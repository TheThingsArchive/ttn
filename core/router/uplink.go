// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/fields"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

func (r *router) HandleUplink(gatewayID string, uplink *pb.UplinkMessage) (err error) {
	ctx := r.Ctx.WithField("GatewayID", gatewayID).WithFields(fields.Get(uplink))
	start := time.Now()
	gateway := r.getGateway(gatewayID)
	defer func() {
		if err != nil {
			uplink.Trace = uplink.Trace.WithEvent(trace.DropEvent, "reason", err)
			ctx.WithError(err).Warn("Could not handle uplink")
		}
		gateway.SendToMonitor(uplink)
	}()
	r.status.uplink.Mark(1)
	uplink.Trace = uplink.Trace.WithEvent(trace.ReceiveEvent, "gateway", gatewayID)

	if err := uplink.UnmarshalPayload(); err != nil {
		return err
	}

	if uplink.GetMessage().GetLorawan().GetMType() == pb_lorawan.MType_JOIN_REQUEST {
		req := uplink.GetMessage().GetLorawan().GetJoinRequestPayload()
		if req == nil {
			return errors.NewErrInvalidArgument("Join Request", "does not contain a JoinRequest payload")
		}
		ctx.WithFields(ttnlog.Fields{
			"DevEUI": req.DevEui,
			"AppEUI": req.AppEui,
		}).Debug("Handle Uplink as Activation")
		r.HandleActivation(gatewayID, &pb.DeviceActivationRequest{
			Payload:          uplink.Payload,
			DevEui:           &req.DevEui,
			AppEui:           &req.AppEui,
			ProtocolMetadata: uplink.ProtocolMetadata,
			GatewayMetadata:  uplink.GatewayMetadata,
			Trace:            uplink.Trace.WithEvent("handle uplink as activation"),
		})
		return nil
	}

	if err := gateway.HandleUplink(uplink); err != nil {
		return err
	}

	if mType := uplink.GetMessage().GetLorawan().GetMType(); mType != pb_lorawan.MType_UNCONFIRMED_UP && mType != pb_lorawan.MType_CONFIRMED_UP {
		ctx.Info("Accidentally received non-uplink message")
		return nil
	}

	mac := uplink.GetMessage().GetLorawan().GetMacPayload()
	if mac == nil {
		return errors.NewErrInvalidArgument("Uplink", "does not contain a MAC payload")
	}

	downlinkOptions, err := gateway.GetDownlinkOptions(uplink)
	if err != nil {
		uplink.Trace = uplink.Trace.WithEvent(trace.WarnEvent, trace.ErrorField, fmt.Sprintf("could not build downlink options: %s", err))
	} else {
		uplink.Trace = uplink.Trace.WithEvent(trace.BuildDownlinkEvent, "options", len(downlinkOptions))
		ctx = ctx.WithField("DownlinkOptions", len(downlinkOptions))
	}
	if r.Component != nil && r.Component.Identity != nil {
		for _, opt := range downlinkOptions {
			opt.Identifier = fmt.Sprintf("%s:%s", r.Component.Identity.Id, opt.Identifier)
		}
	}

	// Find Broker
	brokers, err := r.Discovery.GetAllBrokersForDevAddr(mac.DevAddr)
	if err != nil {
		return err
	}

	if len(brokers) == 0 {
		ctx.Debug("No brokers to forward message to")
		uplink.Trace = uplink.Trace.WithEvent(trace.DropEvent, "reason", "no brokers")
		return nil
	}
	ctx = ctx.WithField("NumBrokers", len(brokers))
	uplink.Trace = uplink.Trace.WithEvent(trace.ForwardEvent, "brokers", len(brokers))

	// Forward to all brokers
	for _, broker := range brokers {
		broker, err := r.getBroker(broker)
		if err != nil {
			continue
		}
		broker.uplink <- &pb_broker.UplinkMessage{
			Payload:          uplink.Payload,
			ProtocolMetadata: uplink.ProtocolMetadata,
			GatewayMetadata:  uplink.GatewayMetadata,
			DownlinkOptions:  downlinkOptions,
			Trace:            uplink.Trace,
		}
	}

	ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled uplink")

	return nil
}

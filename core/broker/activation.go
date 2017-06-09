// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/fields"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

type challengeResponseWithHandler struct {
	handler  *pb_discovery.Announcement
	client   pb_handler.HandlerClient
	response *pb.ActivationChallengeResponse
}

var errDuplicateActivation = errors.New("Not handling duplicate activation on this gateway")

func (b *broker) HandleActivation(activation *pb.DeviceActivationRequest) (res *pb.DeviceActivationResponse, err error) {
	ctx := b.Ctx.WithFields(fields.Get(activation))
	start := time.Now()
	deduplicatedActivationRequest := new(pb.DeduplicatedDeviceActivationRequest)
	deduplicatedActivationRequest.ServerTime = start.UnixNano()

	defer func() {
		if err != nil {
			if deduplicatedActivationRequest != nil {
				deduplicatedActivationRequest.Trace = deduplicatedActivationRequest.Trace.WithEvent(trace.DropEvent, "reason", err)
			}
			ctx.WithError(err).Warn("Could not handle activation")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled activation")
		}
		if deduplicatedActivationRequest != nil && b.monitorStream != nil {
			b.monitorStream.Send(deduplicatedActivationRequest)
		}
	}()

	b.status.activations.Mark(1)

	activation.Trace = activation.Trace.WithEvent(trace.ReceiveEvent)

	// De-duplicate uplink messages
	duplicates := b.deduplicateActivation(activation)
	if len(duplicates) == 0 {
		return nil, errDuplicateActivation
	}
	ctx = ctx.WithField("Duplicates", len(duplicates))

	b.status.activationsUnique.Mark(1)

	deduplicatedActivationRequest.Payload = duplicates[0].Payload
	deduplicatedActivationRequest.DevEui = duplicates[0].DevEui
	deduplicatedActivationRequest.AppEui = duplicates[0].AppEui
	deduplicatedActivationRequest.ProtocolMetadata = duplicates[0].ProtocolMetadata
	deduplicatedActivationRequest.ActivationMetadata = duplicates[0].ActivationMetadata
	deduplicatedActivationRequest.Trace = deduplicatedActivationRequest.Trace.WithEvent(trace.DeduplicateEvent,
		"duplicates", len(duplicates),
	)
	for _, duplicate := range duplicates {
		if duplicate.Trace != nil {
			deduplicatedActivationRequest.Trace.Parents = append(deduplicatedActivationRequest.Trace.Parents, duplicate.Trace)
		}
	}

	// Collect GatewayMetadata and DownlinkOptions
	var downlinkOptions []downlinkOption
	for _, duplicate := range duplicates {
		deduplicatedActivationRequest.GatewayMetadata = append(deduplicatedActivationRequest.GatewayMetadata, duplicate.GatewayMetadata)
		for _, option := range duplicate.DownlinkOptions {
			if option.RxDelay != 5 {
				continue // The Join Accept needs to have an RX delay of 5 seconds
			}
			downlinkOptions = append(downlinkOptions, downlinkOption{
				uplinkMetadata: duplicate.GatewayMetadata,
				option:         option,
			})
		}
	}

	// Select best DownlinkOption
	if len(downlinkOptions) > 0 {
		deduplicatedActivationRequest.ResponseTemplate = &pb.DeviceActivationResponse{
			DownlinkOption: selectBestDownlink(downlinkOptions),
		}
	}

	// Send Activate to NS
	deduplicatedActivationRequest, err = b.ns.PrepareActivation(b.Component.GetContext(b.nsToken), deduplicatedActivationRequest)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "NetworkServer refused to prepare activation")
	}

	ctx = ctx.WithFields(ttnlog.Fields{
		"AppID": deduplicatedActivationRequest.AppId,
		"DevID": deduplicatedActivationRequest.DevId,
	})

	// Find Handler (based on AppEUI)
	var announcements []*pb_discovery.Announcement
	announcements, err = b.Discovery.GetAllHandlersForAppID(deduplicatedActivationRequest.AppId)
	if err != nil {
		return nil, err
	}
	if len(announcements) == 0 {
		return nil, errors.NewErrNotFound(fmt.Sprintf("Handler for AppID %s", deduplicatedActivationRequest.AppId))
	}

	ctx = ctx.WithField("NumHandlers", len(announcements))

	// LoRaWAN: Unmarshal and prepare version without MIC
	var phyPayload lorawan.PHYPayload
	err = phyPayload.UnmarshalBinary(deduplicatedActivationRequest.Payload)
	if err != nil {
		return nil, err
	}
	correctMIC := phyPayload.MIC
	phyPayload.MIC = [4]byte{0, 0, 0, 0}
	phyPayloadWithoutMIC, err := phyPayload.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Build Challenge
	challenge := &pb.ActivationChallengeRequest{
		Payload: phyPayloadWithoutMIC,
		AppId:   deduplicatedActivationRequest.AppId,
		DevId:   deduplicatedActivationRequest.DevId,
		AppEui:  deduplicatedActivationRequest.AppEui,
		DevEui:  deduplicatedActivationRequest.DevEui,
	}

	// Send Challenge to all handlers and collect responses
	var wg sync.WaitGroup
	responses := make(chan *challengeResponseWithHandler, len(announcements))
	for _, announcement := range announcements {
		conn, err := b.getHandlerConn(announcement.Id)
		if err != nil {
			ctx.WithError(err).Warn("Could not dial handler for Activation")
			continue
		}
		client := pb_handler.NewHandlerClient(conn)

		// Do async request
		wg.Add(1)
		go func(announcement *pb_discovery.Announcement) {
			res, err := client.ActivationChallenge(b.Component.GetContext(""), challenge)
			if err == nil && res != nil {
				responses <- &challengeResponseWithHandler{
					handler:  announcement,
					client:   client,
					response: res,
				}
			}
			wg.Done()
		}(announcement)
	}

	// Make sure to close channel when all requests are done
	go func() {
		wg.Wait()
		close(responses)
	}()

	var gotFirst bool
	var joinHandler *pb_discovery.Announcement
	var joinHandlerClient pb_handler.HandlerClient
	for res := range responses {
		var phyPayload lorawan.PHYPayload
		err = phyPayload.UnmarshalBinary(res.response.Payload)
		if err != nil {
			continue
		}
		if phyPayload.MIC != correctMIC {
			continue
		}

		if gotFirst {
			ctx.Warn("Duplicate Activation Response")
		} else {
			gotFirst = true
			joinHandler = res.handler
			joinHandlerClient = res.client
		}
	}

	// Activation not accepted by any broker
	if !gotFirst {
		ctx.Debug("Activation not accepted by any Handler")
		return nil, errors.New("Activation not accepted by any Handler")
	}

	ctx.WithField("HandlerID", joinHandler.Id).Debug("Forward Activation")
	deduplicatedActivationRequest.Trace = deduplicatedActivationRequest.Trace.WithEvent(trace.ForwardEvent,
		"handler", joinHandler.Id,
	)

	handlerResponse, err := joinHandlerClient.Activate(b.Component.GetContext(""), deduplicatedActivationRequest)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Handler refused activation")
	}

	handlerResponse.Trace = handlerResponse.Trace.WithEvent(trace.ReceiveEvent)

	handlerResponse, err = b.ns.Activate(b.Component.GetContext(b.nsToken), handlerResponse)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "NetworkServer refused activation")
	}

	handlerResponse.Trace = handlerResponse.Trace.WithEvent(trace.ForwardEvent)

	res = &pb.DeviceActivationResponse{
		Payload:        handlerResponse.Payload,
		Message:        handlerResponse.Message,
		DownlinkOption: handlerResponse.DownlinkOption,
		Trace:          handlerResponse.Trace,
	}

	return res, nil
}

func (b *broker) deduplicateActivation(duplicate *pb.DeviceActivationRequest) (activations []*pb.DeviceActivationRequest) {
	sum := md5.Sum(duplicate.Payload)
	key := hex.EncodeToString(sum[:])
	list := b.activationDeduplicator.Deduplicate(key, duplicate)
	if len(list) == 0 {
		return
	}
	for _, duplicate := range list {
		activations = append(activations, duplicate.(*pb.DeviceActivationRequest))
	}
	return
}

package broker

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"time"

	"google.golang.org/grpc"

	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/broker/application"
	"github.com/apex/log"
)

func (b *broker) HandleActivation(activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	ctx := b.Ctx.WithFields(log.Fields{
		"GatewayEUI": *activation.GatewayMetadata.GatewayEui,
		"AppEUI":     *activation.AppEui,
		"DevEUI":     *activation.DevEui,
	})
	var err error
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle activation")
		}
	}()

	time := time.Now()

	// De-duplicate uplink messages
	duplicates := b.deduplicateActivation(activation)
	if len(duplicates) == 0 {
		err = errors.New("ttn/broker: No duplicates")
		return nil, err
	}

	base := duplicates[0]

	// Collect GatewayMetadata and DownlinkOptions
	var gatewayMetadata []*gateway.RxMetadata
	var downlinkOptions []*pb.DownlinkOption
	var deviceActivationResponse *pb.DeviceActivationResponse
	for _, duplicate := range duplicates {
		gatewayMetadata = append(gatewayMetadata, duplicate.GatewayMetadata)
		downlinkOptions = append(downlinkOptions, duplicate.DownlinkOptions...)
	}

	// Select best DownlinkOption
	if len(downlinkOptions) > 0 {
		deviceActivationResponse = &pb.DeviceActivationResponse{
			DownlinkOption: selectBestDownlink(downlinkOptions),
		}
	}

	// Build Uplink
	deduplicatedActivationRequest := &pb.DeduplicatedDeviceActivationRequest{
		Payload:            base.Payload,
		DevEui:             base.DevEui,
		AppEui:             base.AppEui,
		ProtocolMetadata:   base.ProtocolMetadata,
		GatewayMetadata:    gatewayMetadata,
		ActivationMetadata: base.ActivationMetadata,
		ServerTime:         time.UnixNano(),
		ResponseTemplate:   deviceActivationResponse,
	}

	// Send Activate to NS
	deduplicatedActivationRequest, err = b.ns.PrepareActivation(b.Component.GetContext(), deduplicatedActivationRequest)
	if err != nil {
		return nil, err
	}

	// Find Handler (based on AppEUI)
	var application *application.Application
	application, err = b.applications.Get(*base.AppEui)
	if err != nil {
		return nil, err
	}

	ctx = ctx.WithField("HandlerID", application.HandlerID)
	ctx.Debug("Forward Activation")

	var conn *grpc.ClientConn
	conn, err = grpc.Dial(application.HandlerNetAddress, api.DialOptions...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := pb_handler.NewHandlerClient(conn)
	var handlerResponse *pb_handler.DeviceActivationResponse
	handlerResponse, err = client.Activate(b.Component.GetContext(), deduplicatedActivationRequest)
	if err != nil {
		return nil, err
	}

	handlerResponse, err = b.ns.Activate(b.Component.GetContext(), handlerResponse)
	if err != nil {
		return nil, err
	}

	deviceActivationResponse = &pb.DeviceActivationResponse{
		Payload:        handlerResponse.Payload,
		DownlinkOption: handlerResponse.DownlinkOption,
	}

	ctx.Debug("Successful Activation")

	return deviceActivationResponse, nil
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

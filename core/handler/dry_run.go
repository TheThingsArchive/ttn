// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"encoding/json"

	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/handler/functions"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
)

// DryUplink converts the uplink message payload by running the payload
// functions that are provided in the DryUplinkMessage, without actually going to the network.
// This is helpful for testing the payload functions without having to save them.
func (h *handlerManager) DryUplink(ctx context.Context, in *pb.DryUplinkMessage) (*pb.DryUplinkResult, error) {
	app := in.App

	logger := functions.NewEntryLogger()

	flds := ""
	valid := true
	if app != nil && app.Decoder != "" {
		functions := &UplinkFunctions{
			Decoder:   app.Decoder,
			Converter: app.Converter,
			Validator: app.Validator,
			Logger:    logger,
		}

		fields, val, err := functions.Process(in.Payload, uint8(in.Port))
		if err != nil {
			return nil, err
		}

		valid = val

		marshalled, err := json.Marshal(fields)
		if err != nil {
			return nil, err
		}

		flds = string(marshalled)
	}

	return &pb.DryUplinkResult{
		Payload: in.Payload,
		Fields:  flds,
		Valid:   valid,
		Logs:    logger.Logs,
	}, nil
}

// DryDownlink converts the downlink message payload by running the payload
// functions that are provided in the DryDownlinkMessage, without actually going to the network.
// This is helpful for testing the payload functions without having to save them.
func (h *handlerManager) DryDownlink(ctx context.Context, in *pb.DryDownlinkMessage) (*pb.DryDownlinkResult, error) {
	app := in.App

	if in.Payload != nil {
		if in.Fields != "" {
			return nil, errors.NewErrInvalidArgument("Downlink", "Both Fields and Payload provided")
		}
		return &pb.DryDownlinkResult{
			Payload: in.Payload,
		}, nil
	}

	if in.Fields == "" {
		return nil, errors.NewErrInvalidArgument("Downlink", "Neither Fields nor Payload provided")
	}

	if app == nil || app.Encoder == "" {
		return nil, errors.NewErrInvalidArgument("Encoder", "Not specified")
	}

	logger := functions.NewEntryLogger()

	functions := &DownlinkFunctions{
		Encoder: app.Encoder,
		Logger:  logger,
	}

	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(in.Fields), &parsed)
	if err != nil {
		return nil, errors.NewErrInvalidArgument("Fields", err.Error())
	}

	payload, _, err := functions.Process(parsed, uint8(in.Port))
	if err != nil {
		return nil, err
	}

	return &pb.DryDownlinkResult{
		Payload: payload,
		Logs:    logger.Logs,
	}, nil
}

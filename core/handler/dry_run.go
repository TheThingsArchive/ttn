// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"golang.org/x/net/context"
)

// DryUplink converts the uplink message payload by running the payload
// functions that are provided in the DryUplinkMessage, without actually going to the network.
// This is helpful for testing the payload functions without having to save them.
func (h *handler) DryUplink(ctx context.Context, in *pb.DryUplinkMessage) (*pb.DryUplinkResult, error) {
	app := in.App

	flds := ""
	valid := true
	if app != nil && app.Decoder != "" {
		functions := &UplinkFunctions{
			Decoder:   app.Decoder,
			Converter: app.Converter,
			Validator: app.Validator,
		}

		fields, val, err := functions.Process(in.Payload)
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
	}, nil
}

// DryDownlink converts the downlink message payload by running the payload
// functions that are provided in the DryDownlinkMessage, without actually going to the network.
// This is helpful for testing the payload functions without having to save them.
func (h *handler) DryDownlink(ctx context.Context, in *pb.DryDownlinkMessage) (*pb.DryDownlinkResult, error) {
	app := in.App

	if in.Payload != nil {
		if in.Fields != "" {
			return nil, errors.New("Both fields and payload provided on downlink message")
		}
		return &pb.DryDownlinkResult{
			Payload: in.Payload,
		}, nil
	}

	if in.Fields == "" {
		return nil, errors.New("Neither fields or payload provided on downlink message")
	}

	functions := &DownlinkFunctions{
		Encoder: app.Encoder,
	}

	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(in.Fields), &parsed)
	if err != nil {
		return nil, fmt.Errorf("Could not parse fields: %s", err)
	}

	payload, _, err := functions.Process(parsed)
	if err != nil {
		return nil, err
	}

	return &pb.DryDownlinkResult{
		Payload: payload,
	}, nil
}

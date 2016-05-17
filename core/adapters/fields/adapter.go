// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package fields

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/mqtt"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Adapter represents the interface of an application
type Adapter interface {
	core.AppClient
	Storage() AppStorage
	SubscribeDownlink(handler core.HandlerServer) error
}

type adapter struct {
	ctx     log.Interface
	storage AppStorage
	mqtt    mqtt.Adapter
}

// NewAdapter returns a new adapter that processes binary payload to fields
func NewAdapter(ctx log.Interface, storage AppStorage, next mqtt.Adapter) Adapter {
	return &adapter{ctx, storage, next}
}

func (s *adapter) HandleData(context context.Context, in *core.DataAppReq, opt ...grpc.CallOption) (*core.DataAppRes, error) {
	var appEUI types.AppEUI
	appEUI.Unmarshal(in.AppEUI)
	var devEUI types.DevEUI
	devEUI.Unmarshal(in.DevEUI)
	ctx := s.ctx.WithFields(&log.Fields{"AppEUI": appEUI, "DevEUI": devEUI})

	req := core.DataUpAppReq{
		Payload:  in.Payload,
		Metadata: core.ProtoMetaToAppMeta(in.Metadata...),
		FPort:    uint8(in.FPort),
		FCnt:     in.FCnt,
		DevEUI:   devEUI.String(),
	}

	functions, err := s.storage.GetFunctions(appEUI)
	if err != nil {
		ctx.WithError(err).Warn("Failed to get functions")
		// If we can't get the functions here, just publish it anyway without fields
		return new(core.DataAppRes), s.mqtt.PublishUplink(appEUI, devEUI, req)
	}

	if functions == nil {
		// Publish when there are no payload functions set
		return new(core.DataAppRes), s.mqtt.PublishUplink(appEUI, devEUI, req)
	}

	fields, valid, err := functions.Process(in.Payload)
	if err != nil {
		// If there were errors processing the payload, just publish it anyway
		// without fields
		ctx.WithError(err).Warn("Failed to process payload")
		return new(core.DataAppRes), s.mqtt.PublishUplink(appEUI, devEUI, req)
	}

	if !valid {
		// If the payload has been processed successfully but is not valid, it should
		// not be published
		ctx.Info("The processed payload is not valid")
		return new(core.DataAppRes), errors.New(errors.Operational, "The processed payload is not valid")
	}

	req.Fields = fields
	if err := s.mqtt.PublishUplink(appEUI, devEUI, req); err != nil {
		return new(core.DataAppRes), errors.New(errors.Operational, "Failed to publish data")
	}

	return new(core.DataAppRes), nil
}

func (s *adapter) HandleJoin(context context.Context, in *core.JoinAppReq, opt ...grpc.CallOption) (*core.JoinAppRes, error) {
	var appEUI types.AppEUI
	appEUI.Unmarshal(in.AppEUI)
	var devEUI types.DevEUI
	devEUI.Unmarshal(in.DevEUI)

	req := core.OTAAAppReq{
		Metadata: core.ProtoMetaToAppMeta(in.Metadata...),
	}

	return new(core.JoinAppRes), s.mqtt.PublishActivation(appEUI, devEUI, req)
}

func (s *adapter) SubscribeDownlink(handler core.HandlerServer) error {
	return s.mqtt.SubscribeDownlink(handler)
}

func (s *adapter) Storage() AppStorage {
	return s.storage
}

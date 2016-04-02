// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"github.com/TheThingsNetwork/ttn/core"
	ttnMQTT "github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Adapter defines a public interface for the mqtt adapter
type Adapter interface {
	core.AppClient
	SubscribeDownlink(handler core.HandlerServer) error
}

type defaultAdapter struct {
	ctx    log.Interface
	client ttnMQTT.Client
}

// HandleData implements the core.AppClient interface
func (a *defaultAdapter) HandleData(_ context.Context, req *core.DataAppReq, _ ...grpc.CallOption) (*core.DataAppRes, error) {
	if err := validateData(req); err != nil {
		return new(core.DataAppRes), errors.New(errors.Structural, err)
	}

	dataUp := core.DataUpAppReq{
		Payload:  req.Payload,
		Metadata: core.ProtoMetaToAppMeta(req.Metadata...),
	}

	if a.ctx != nil {
		a.ctx.WithFields(log.Fields{
			"AppEUI": req.AppEUI,
			"DevEUI": req.DevEUI,
		}).Debug("Publishing Uplink")
	}

	token := a.client.PublishUplink(req.AppEUI, req.DevEUI, dataUp)
	if token.Wait(); token.Error() != nil {
		return new(core.DataAppRes), errors.New(errors.Structural, token.Error())
	}
	return new(core.DataAppRes), nil
}

func validateData(req *core.DataAppReq) error {
	var err error
	switch {
	case req == nil:
		err = errors.New(errors.Structural, "Received Nil Application Request")
	case len(req.Payload) == 0:
		err = errors.New(errors.Structural, "Invalid Packet Payload")
	case len(req.DevEUI) != 8:
		err = errors.New(errors.Structural, "Invalid Device EUI")
	case len(req.AppEUI) != 8:
		err = errors.New(errors.Structural, "Invalid Application EUI")
	case req.Metadata == nil:
		err = errors.New(errors.Structural, "Missing Mandatory Metadata")
	}

	if err != nil {
		stats.MarkMeter("mqtt_adapter.uplink.invalid")
		return err
	}

	return nil
}

// HandleJoin implements the core.AppClient interface
func (a *defaultAdapter) HandleJoin(_ context.Context, req *core.JoinAppReq, _ ...grpc.CallOption) (*core.JoinAppRes, error) {
	if err := validateJoin(req); err != nil {
		return new(core.JoinAppRes), errors.New(errors.Structural, err)
	}

	otaa := core.OTAAAppReq{
		Metadata: core.ProtoMetaToAppMeta(req.Metadata...),
	}

	if a.ctx != nil {
		a.ctx.WithFields(log.Fields{
			"AppEUI": req.AppEUI,
			"DevEUI": req.DevEUI,
		}).Debug("Publishing Activation")
	}

	token := a.client.PublishActivation(req.AppEUI, req.DevEUI, otaa)
	if token.Wait(); token.Error() != nil {
		return new(core.JoinAppRes), errors.New(errors.Structural, token.Error())
	}
	return new(core.JoinAppRes), nil
}

func validateJoin(req *core.JoinAppReq) error {
	var err error
	switch {
	case req == nil:
		err = errors.New(errors.Structural, "Received Nil Application Request")
	case len(req.DevEUI) != 8:
		err = errors.New(errors.Structural, "Invalid Device EUI")
	case len(req.AppEUI) != 8:
		err = errors.New(errors.Structural, "Invalid Application EUI")
	case req.Metadata == nil:
		err = errors.New(errors.Structural, "Missing Mandatory Metadata")
	}

	if err != nil {
		stats.MarkMeter("mqtt_adapter.join.invalid")
		return err
	}

	return nil
}

func (a *defaultAdapter) SubscribeDownlink(handler core.HandlerServer) error {
	token := a.client.SubscribeDownlink(func(client ttnMQTT.Client, appEUI []byte, devEUI []byte, req core.DataDownAppReq) {
		if len(req.Payload) == 0 {
			if a.ctx != nil {
				a.ctx.Debug("Skipping empty downlink")
			}
			return
		}

		if a.ctx != nil {
			a.ctx.WithFields(log.Fields{
				"AppEUI": appEUI,
				"DevEUI": devEUI,
			}).Debug("Receiving Downlink")
		}

		// Convert it to an handler downlink
		handler.HandleDataDown(context.Background(), &core.DataDownHandlerReq{
			Payload: req.Payload,
			TTL:     req.TTL,
			AppEUI:  appEUI,
			DevEUI:  devEUI,
		})
	})

	token.Wait()
	return token.Error()
}

// NewAdapter returns a new MQTT handler adapter
func NewAdapter(ctx log.Interface, client ttnMQTT.Client) Adapter {
	return &defaultAdapter{ctx, client}
}

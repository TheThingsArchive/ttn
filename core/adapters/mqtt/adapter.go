// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	ttnMQTT "github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"golang.org/x/net/context"
)

const publishTimeout = 20 * time.Millisecond

// Adapter represents the interface of an application
type Adapter interface {
	PublishUplink(appEUI types.AppEUI, devEUI types.DevEUI, req core.DataUpAppReq) error
	PublishActivation(appEUI types.AppEUI, devEUI types.DevEUI, req core.OTAAAppReq) error
	SubscribeDownlink(handler core.HandlerServer) error
}

type defaultAdapter struct {
	ctx    log.Interface
	client ttnMQTT.Client
}

// NewAdapter returns a new MQTT handler adapter
func NewAdapter(ctx log.Interface, client ttnMQTT.Client) Adapter {
	return &defaultAdapter{ctx, client}
}

// HandleData implements the core.AppClient interface
func (a *defaultAdapter) PublishUplink(appEUI types.AppEUI, devEUI types.DevEUI, req core.DataUpAppReq) error {
	ctx := a.ctx.WithFields(log.Fields{
		"AppEUI": appEUI,
		"DevEUI": devEUI,
	})
	ctx.Debug("Publishing Uplink")

	token := a.client.PublishUplink(appEUI, devEUI, req)
	if token.WaitTimeout(publishTimeout) {
		// token did not timeout: just return
		if token.Error() != nil {
			return errors.New(errors.Structural, token.Error())
		}
	} else {
		// token did timeout: wait for it in background and just return
		go func() {
			token.Wait()
			if token.Error() != nil {
				ctx.WithError(token.Error()).Warn("Could not publish uplink")
			}
		}()
	}
	return nil
}

func (a *defaultAdapter) PublishActivation(appEUI types.AppEUI, devEUI types.DevEUI, req core.OTAAAppReq) error {
	ctx := a.ctx.WithFields(log.Fields{
		"AppEUI": appEUI,
		"DevEUI": devEUI,
	})
	ctx.Debug("Publishing Activation")

	token := a.client.PublishActivation(appEUI, devEUI, req)
	if token.WaitTimeout(publishTimeout) {
		// token did not timeout: just return
		if token.Error() != nil {
			return errors.New(errors.Structural, token.Error())
		}
	} else {
		// token did timeout: wait for it in background and just return
		go func() {
			token.Wait()
			if token.Error() != nil {
				ctx.WithError(token.Error()).Warn("Could not publish activation")
			}
		}()
	}
	return nil
}

func (a *defaultAdapter) SubscribeDownlink(handler core.HandlerServer) error {
	token := a.client.SubscribeDownlink(func(client ttnMQTT.Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataDownAppReq) {
		if len(req.Payload) == 0 {
			a.ctx.Debug("Skipping empty downlink")
			return
		}

		a.ctx.WithFields(log.Fields{
			"AppEUI": appEUI,
			"DevEUI": devEUI,
		}).Debug("Receiving Downlink")

		// Convert it to an handler downlink
		handler.HandleDataDown(context.Background(), &core.DataDownHandlerReq{
			Payload: req.Payload,
			TTL:     req.TTL,
			AppEUI:  appEUI.Bytes(),
			DevEUI:  devEUI.Bytes(),
		})
	})

	token.Wait()
	return token.Error()
}

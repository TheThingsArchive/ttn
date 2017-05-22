// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/json"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// ActivationHandler is called for activations
type ActivationHandler func(client Client, appID string, devID string, req types.Activation)

// PublishActivation publishes an activation
func (c *DefaultClient) PublishActivation(activation types.Activation) Token {
	appID := activation.AppID
	devID := activation.DevID
	return c.PublishDeviceEvent(appID, devID, types.ActivationEvent, activation)
}

// SubscribeDeviceActivations subscribes to all activations for the given application and device
func (c *DefaultClient) SubscribeDeviceActivations(appID string, devID string, handler ActivationHandler) Token {
	return c.SubscribeDeviceEvents(appID, devID, types.ActivationEvent, func(_ Client, appID string, devID string, _ types.EventType, payload []byte) {
		activation := types.Activation{}
		if err := json.Unmarshal(payload, &activation); err != nil {
			c.ctx.Warnf("mqtt: could not unmarshal activation: %s", err)
			return
		}
		activation.AppID = appID
		activation.DevID = devID
		// Call the Activation handler
		handler(c, appID, devID, activation)
	})
}

// SubscribeAppActivations subscribes to all activations for the given application
func (c *DefaultClient) SubscribeAppActivations(appID string, handler ActivationHandler) Token {
	return c.SubscribeDeviceActivations(appID, "", handler)
}

// SubscribeActivations subscribes to all activations that the current user has access to
func (c *DefaultClient) SubscribeActivations(handler ActivationHandler) Token {
	return c.SubscribeDeviceActivations("", "", handler)
}

// UnsubscribeDeviceActivations unsubscribes from the activations for the given application and device
func (c *DefaultClient) UnsubscribeDeviceActivations(appID string, devID string) Token {
	return c.UnsubscribeDeviceEvents(appID, devID, types.ActivationEvent)
}

// UnsubscribeAppActivations unsubscribes from the activations for the given application
func (c *DefaultClient) UnsubscribeAppActivations(appID string) Token {
	return c.UnsubscribeDeviceEvents(appID, "", types.ActivationEvent)
}

// UnsubscribeActivations unsubscribes from the activations that the current user has access to
func (c *DefaultClient) UnsubscribeActivations() Token {
	return c.UnsubscribeDeviceEvents("", "", types.ActivationEvent)
}

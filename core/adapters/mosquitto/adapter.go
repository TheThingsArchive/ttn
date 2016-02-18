// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mosquitto

import (
	"fmt"
	"strings"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

type Adapter struct {
	ctx           log.Interface
	registrations chan core.Registration
}

type PersonnalizedActivation struct {
	DevAddr lorawan.DevAddr
	NwkSKey lorawan.AES128Key
	AppSKey lorawan.AES128Key
}

const (
	TOPIC_ACTIVATIONS string = "activations"
	TOPIC_UPLINK             = "up"
	TOPIC_DOWNLINK           = "down"
	RESOURCE                 = "devices"
	PERSONNALIZED            = "personnalized"
)

// NewAdapter constructs a new mqtt adapter
func NewAdapter(client *MQTT.Client, ctx log.Interface) (*Adapter, error) {
	a := &Adapter{
		ctx:           ctx,
		registrations: make(chan core.Registration),
	}

	token := client.Subscribe(fmt.Sprintf("+/devices/+/activations"), 2, a.handleActivation)
	if token.Wait() && token.Error() != nil {
		ctx.WithError(token.Error()).Error("Unable to instantiate the adapter")
		return nil, errors.New(ErrFailedOperation, token.Error())
	}

	return a, nil
}

func (a *Adapter) handleActivation(client *MQTT.Client, message MQTT.Message) {
	topicInfos := strings.Split(message.Topic(), "/")
	appEUI := topicInfos[0]
	devEUI := topicInfos[2]

	if devEUI != PERSONNALIZED {
		a.ctx.WithField("Device Address", devEUI).Warn("OTAA not yet supported. Unable to register device")
		return
	}

	payload := message.Payload()
	if len(payload) != 36 {
		a.ctx.WithField("Payload", payload).Error("Invalid registration payload")
		return
	}

	var devAddr lorawan.DevAddr
	var nwkSKey lorawan.AES128Key
	var appSKey lorawan.AES128Key
	copy(devAddr[:], message.Payload()[:4])
	copy(nwkSKey[:], message.Payload()[4:20])
	copy(appSKey[:], message.Payload()[20:])

	a.registrations <- core.Registration{
		DevAddr: devAddr,
		Recipient: core.Recipient{
			Id:      appEUI,
			Address: "DoestNotMatterWillBeRefactored",
		},
		Options: struct {
			NwkSKey lorawan.AES128Key
			AppSKey lorawan.AES128Key
		}{
			NwkSKey: nwkSKey,
			AppSKey: appSKey,
		},
	}
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, r ...core.Recipient) (core.Packet, error) {
	return core.Packet{}, nil
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() (core.Packet, core.AckNacker, error) {
	return core.Packet{}, nil, nil
}

// NextRegistration implements the core.Adapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	r := <-a.registrations
	return r, nil, nil
}

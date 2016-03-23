// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Interface defines a public interface for the mqtt adapter
type Interface interface {
	core.AppClient
	Start(inMsg <-chan Msg, handler core.HandlerServer)
}

type adapter struct {
	Components
}

// Components defines a structure to make the instantiation easier to read
type Components struct {
	Client Client
	Ctx    log.Interface
}

// Options defines a structure to make the instantiation easier to read
type Options struct {
}

// Msg are emitted by an MQTT subscriber towards the adapter
type Msg struct {
	Topic   string
	Payload []byte
	Type    msgType
}

// msgType constants are used in MQTTMsg to characterise the kind of message processed
const (
	Down msgType = iota + 1
)

type msgType byte

// New constructs an mqtt adapter responsible for making the bridge between the handler and
// application.
func New(c Components, o Options) Interface {
	return adapter{Components: c}
}

// Start eventually launches the mqtt message internal consumer
func (a adapter) Start(inMsg <-chan Msg, handler core.HandlerServer) {
	go a.consumeMQTTMsg(inMsg, handler)
}

// HandleData implements the core.AppClient interface
func (a adapter) HandleData(bctx context.Context, req *core.DataAppReq, _ ...grpc.CallOption) (*core.DataAppRes, error) {
	stats.MarkMeter("mqtt_adapter.send")

	// Verify the packet integrity
	// TODO Move this elsewhere, make it a function call validate() ...
	if req == nil {
		stats.MarkMeter("mqtt_adapter.uplink.invalid")
		return nil, errors.New(errors.Structural, "Received Nil Application Request")
	}
	if len(req.Payload) == 0 {
		stats.MarkMeter("mqtt_adapter.uplink.invalid")
		return nil, errors.New(errors.Structural, "Invalid Packet Payload")
	}
	if len(req.DevEUI) != 8 {
		stats.MarkMeter("mqtt_adapter.uplink.invalid")
		return nil, errors.New(errors.Structural, "Invalid Device EUI")
	}
	if len(req.AppEUI) != 8 {
		stats.MarkMeter("mqtt_adapter.uplink.invalid")
		return nil, errors.New(errors.Structural, "Invalid Application EUI")
	}
	if req.Metadata == nil {
		stats.MarkMeter("mqtt_adapter.uplink.invalid")
		return nil, errors.New(errors.Structural, "Missing Mandatory Metadata")
	}
	ctx := a.Ctx.WithField("appEUI", req.AppEUI).WithField("devEUI", req.DevEUI)

	// Marshal the packet
	dataUp := core.DataUpAppReq{
		Payload:  req.Payload,
		Metadata: core.ProtoMetaToAppMeta(req.Metadata...),
	}
	msg, err := dataUp.MarshalMsg(nil)
	if err != nil {
		return nil, errors.New(errors.Structural, "Unable to marshal the application payload")
	}

	// Actually send it
	ctx.Debug("Sending Packet")
	deui, aeui := hex.EncodeToString(req.DevEUI), hex.EncodeToString(req.AppEUI)
	err = a.Client.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS2,
		Retain:    false,
		TopicName: []byte(fmt.Sprintf("%s/devices/%s/up", aeui, deui)),
		Message:   msg,
	})

	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return nil, nil
}

// HandleJoin implements the core.AppClient interface
func (a adapter) HandleJoin(bctx context.Context, req *core.JoinAppReq, _ ...grpc.CallOption) (*core.JoinAppRes, error) {
	if len(req.AppEUI) != 8 || len(req.DevEUI) != 8 || len(req.Metadata) == 0 {
		a.Ctx.Debug("Received invalid JoinAppReq")
		return nil, errors.New(errors.Structural, "Invalid request parameters")
	}
	otaa := core.OTAAAppReq{
		Metadata: core.ProtoMetaToAppMeta(req.Metadata...),
	}
	data, err := otaa.MarshalMsg(nil)
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	deui, aeui := hex.EncodeToString(req.DevEUI), hex.EncodeToString(req.AppEUI)
	err = a.Client.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS2,
		Retain:    false,
		TopicName: []byte(fmt.Sprintf("%s/devices/%s/activations", aeui, deui)),
		Message:   data,
	})
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return nil, nil
}

// consumeMQTTMsg processes incoming messages from MQTT broker.
//
// It runs in its own goroutine
func (a adapter) consumeMQTTMsg(chmsg <-chan Msg, handler core.HandlerServer) {
	a.Ctx.Debug("Start consuming MQTT messages")
	for msg := range chmsg {
		switch msg.Type {
		case Down:
			req, err := handleDataDown(msg)
			if err == nil {
				_, err = handler.HandleDataDown(context.Background(), req)
			}
			if err != nil {
				a.Ctx.WithError(err).Debug("Unable to consume data down")
			}
		default:
			a.Ctx.Debug("Unsupported MQTT message's type")
		}
	}
	a.Ctx.Debug("Stop consuming MQTT messages")
}

// handleDataDown parses and handles Downlink message coming through MQTT
func handleDataDown(msg Msg) (*core.DataDownHandlerReq, error) {
	// Ensure the query / topic parameters are valid
	topicInfos := strings.Split(msg.Topic, "/")
	if len(topicInfos) != 4 {
		return nil, errors.New(errors.Structural, "Unexpect (and invalid) mqtt topic")
	}
	appEUI, erra := hex.DecodeString(topicInfos[0])
	devEUI, errd := hex.DecodeString(topicInfos[2])
	if erra != nil || errd != nil || len(appEUI) != 8 || len(devEUI) != 8 {
		return nil, errors.New(errors.Structural, "Topic constituted of invalid AppEUI or DevEUI")
	}

	// Retrieve the message payload
	var req core.DataDownAppReq
	if _, err := req.UnmarshalMsg(msg.Payload); err != nil {
		if err = json.Unmarshal(msg.Payload, &req); err != nil {
			return nil, errors.New(errors.Structural, err)
		}
	}
	if len(req.Payload) == 0 {
		return nil, errors.New(errors.Structural, "There's now data to handle")
	}

	// Convert it to an handler downlink
	return &core.DataDownHandlerReq{
		Payload: req.Payload,
		TTL:     req.TTL,
		AppEUI:  appEUI,
		DevEUI:  devEUI,
	}, nil
}

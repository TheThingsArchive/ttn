// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type adapter struct {
	*client.Client
	handler core.HandlerClient
	ctx     log.Interface
}

// Msg are emitted by an MQTT subscriber towards the adapter
type Msg struct {
	Topic   string
	Payload []byte
	Type    msgType
}

type msgType byte

// msgType constants are used in MQTTMsg to characterise the kind of message processed
const (
	Down msgType = iota
	ABP
	OTAA
)

// New constructs an mqtt adapter responsible for making the bridge between the handler and
// application.
func New(handler core.HandlerClient, client *client.Client, chmsg <-chan Msg, ctx log.Interface) core.AppClient {
	a := adapter{
		Client:  client,
		handler: handler,
		ctx:     ctx,
	}

	go a.consumeMQTTMsg(chmsg)

	return a
}

// NewClient creates and connects a mqtt client with predefined options.
func NewClient(id string, netAddr string, ctx log.Interface) (*client.Client, chan Msg, error) {
	var cli *client.Client
	delay := 25 * time.Millisecond
	chmsg := make(chan Msg)

	tryConnect := func() error {
		ctx.WithField("id", id).WithField("address", netAddr).Debug("(Re)Connecting MQTT Client")
		err := cli.Connect(&client.ConnectOptions{
			Network:  "tcp",
			Address:  netAddr,
			ClientID: []byte(id),
		})

		if err != nil {
			return err
		}

		return cli.Subscribe(&client.SubscribeOptions{
			SubReqs: []*client.SubReq{
				&client.SubReq{
					TopicFilter: []byte("+/devices/+/down"),
					QoS:         mqtt.QoS2,
					Handler: func(topic, msg []byte) {
						chmsg <- Msg{
							Topic:   string(topic),
							Payload: msg,
							Type:    Down,
						}
					},
				},
				&client.SubReq{
					TopicFilter: []byte("+/devices/personalized/activations"),
					QoS:         mqtt.QoS2,
					Handler: func(topic, msg []byte) {
						chmsg <- Msg{
							Topic:   string(topic),
							Payload: msg,
							Type:    ABP,
						}
					},
				},
			},
		})
	}

	var reconnect func(fault error)
	reconnect = func(fault error) {
		if cli == nil {
			ctx.Fatal("Attempt reconnection on non-existing client")
		}
		if delay > 10000*delay {
			cli.Terminate()
			ctx.WithError(fault).Fatal("Unable to reconnect the mqtt client")
		}
		<-time.After(delay)
		if err := tryConnect(); err != nil {
			delay *= 10
			ctx.WithError(err).Debugf("Failed to reconnect MQTT client. Trying again in %s", delay)
			reconnect(fault)
			return
		}
		delay = 25 * time.Millisecond
	}

	cli = client.New(&client.Options{ErrorHandler: reconnect})

	if err := tryConnect(); err != nil {
		cli.Terminate()
		return nil, nil, errors.New(errors.Operational, err)
	}
	return cli, chmsg, nil
}

// HandleData implements the core.AppClient interface
func (a adapter) HandleData(bctx context.Context, req *core.DataAppReq, _ ...grpc.CallOption) (*core.DataAppRes, error) {
	stats.MarkMeter("mqtt_adapter.send")

	// Verify the packet integrity
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
	ctx := a.ctx.WithField("appEUI", req.AppEUI).WithField("devEUI", req.DevEUI)

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
	err = a.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS2,
		Retain:    true,
		TopicName: []byte(fmt.Sprintf("%s/devices/%s/up", aeui, deui)),
		Message:   msg,
	})

	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return nil, nil
}

// consumeMQTTMsg processes incoming messages from MQTT broker.
//
// It runs in its own goroutine
func (a adapter) consumeMQTTMsg(chmsg <-chan Msg) {
	a.ctx.Debug("Start consuming MQTT message")
	for msg := range chmsg {
		switch msg.Type {
		case Down:
			req, err := handleDataDown(msg)
			if err == nil {
				_, err = a.handler.HandleDataDown(context.Background(), req)
			}
			if err != nil {
				a.ctx.WithError(err).Debug("Unable to consume data down")
			}
		case ABP:
			req, err := handleABP(msg)
			if err == nil {
				_, err = a.handler.SubscribePersonalized(context.Background(), req)
			}
			if err != nil {
				a.ctx.WithError(err).Debug("Unable to consume ABP")
			}
		default:
			a.ctx.Debug("Unsupported MQTT message's type")
		}
	}
}

// handleABP parses and handles Application By Personalization request coming through MQTT
func handleABP(msg Msg) (*core.ABPSubHandlerReq, error) {
	// Ensure the query / topic parameters are valid
	topicInfos := strings.Split(msg.Topic, "/")
	if len(topicInfos) != 4 {
		return nil, errors.New(errors.Structural, "Unexpect (and invalid) mqtt topic")
	}

	// Get the actual message, try messagePack then JSON
	var req core.ABPSubAppReq
	if _, err := req.UnmarshalMsg(msg.Payload); err != nil {
		if err = json.Unmarshal(msg.Payload, &req); err != nil {
			return nil, errors.New(errors.Structural, err)
		}
	}

	// Verify each parameter
	appEUI, err := hex.DecodeString(topicInfos[0])
	if err != nil || len(appEUI) != 8 {
		return nil, errors.New(errors.Structural, "Invalid Application EUI")
	}

	devAddr, err := hex.DecodeString(req.DevAddr)
	if err != nil || len(devAddr) != 4 {
		return nil, errors.New(errors.Structural, "Invalid Device Address")
	}

	nwkSKey, err := hex.DecodeString(req.NwkSKey)
	if err != nil || len(nwkSKey) != 16 {
		return nil, errors.New(errors.Structural, "Invalid Network Session Key")
	}

	appSKey, err := hex.DecodeString(req.AppSKey)
	if err != nil || len(appSKey) != 16 {
		return nil, errors.New(errors.Structural, "Invalid Application Session Key")
	}

	// Convert it to an handler subscription
	return &core.ABPSubHandlerReq{
		AppEUI:  appEUI,
		DevAddr: devAddr,
		NwkSKey: nwkSKey,
		AppSKey: appSKey,
	}, nil
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
		AppEUI:  appEUI,
		DevEUI:  devEUI,
	}, nil
}

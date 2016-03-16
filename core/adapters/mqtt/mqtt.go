// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/hex"
	"fmt"
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
	Handler core.HandlerClient
	ctx     log.Interface
}

// New constructs an mqtt adapter responsible for making the bridge between the handler and
// application.
func New(handler core.HandlerClient, client *client.Client, ctx log.Interface) core.AppClient {
	return adapter{
		Client:  client,
		Handler: handler,
		ctx:     ctx,
	}
}

// NewClient creates and connects a mqtt client with predefined options.
func NewClient(id string, netAddr string, ctx log.Interface) (*client.Client, error) {
	var cli *client.Client
	var delay time.Duration = 25 * time.Millisecond

	tryConnect := func() error {
		ctx.WithField("id", id).WithField("address", netAddr).Debug("(Re)Connecting MQTT Client")
		return cli.Connect(&client.ConnectOptions{
			Network:  "tcp",
			Address:  netAddr,
			ClientID: []byte(id),
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
		return nil, errors.New(errors.Operational, err)
	}
	return cli, nil
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
	appPayload := core.AppPayload{
		Payload:  req.Payload,
		Metadata: core.ProtoMetaToAppMeta(req.Metadata...),
	}
	msg, err := appPayload.MarshalMsg(nil)
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

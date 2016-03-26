// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

// InitialReconnectDelay represents the initial delay of reconnection in case of lose. The client
// will attempt to reconnect several times after a given delay being 10x the previous one.
const InitialReconnectDelay = 25 * time.Millisecond

// Client provides an interface for an MQTT client
type Client interface {
	// Publish pushes a message on a given topic
	Publish(*client.PublishOptions) error
	// Terminate kills internal client goroutine and processes
	Terminate()
}

// connecter is an alias used by methods belows
type connecter func() error

// NewClient creates and connects a mqtt client with predefined options.
func NewClient(id string, netAddr string, ctx log.Interface) (Client, chan Msg, error) {
	ctx = ctx.WithField("id", id).WithField("address", netAddr)
	chcmd := make(chan interface{})
	chmsg := make(chan Msg)

	go monitorClient(id, netAddr, chcmd, ctx)

	tryConnect := createConnecter(id, netAddr, chmsg, chcmd, ctx)
	if err := tryConnect(); err != nil {
		close(chcmd)
		return nil, nil, errors.New(errors.Operational, err)
	}

	return safeClient{
		chcmd: chcmd,
	}, chmsg, nil
}

// monitorClient is used to keep all accesses to the client completely concurrent-safe. It also
// allows to replace the current client by a new one in case of error. /
//
// When the client loses its connection and isn't able to re-establish it, we need to create a new
// client. However, because that client is likely to be accessed by several goroutines at "the same
// time", we cannot just swap two variables somewhere. The hereby monitor enables a safe client
// swapping and managing. (See safeClient struct as well)
func monitorClient(id string, netAddr string, chcmd <-chan interface{}, ctx log.Interface) {
	var cli *client.Client
	ctx = ctx.WithField("process", "monitorClient")
	ctx.Debug("Start monitoring MQTT client")

	for cmd := range chcmd {
		if cli == nil {
			init, ok := cmd.(cmdClient)
			if !ok {
				ctx.Warn("Received cmd whereas client is nil. Ignored")
				continue
			}
			ctx.Debug("Setup initial MQTT client")
			cli = init.options
			init.cherr <- nil
			continue
		}
		switch cmd.(type) {
		case cmdPublish:
			cmd := cmd.(cmdPublish)
			cmd.cherr <- cli.Publish(cmd.options)
		case cmdTerminate:
			cmd := cmd.(cmdTerminate)
			cli.Terminate()
			cli = nil
			cmd.cherr <- nil
		case cmdClient:
			cmd := cmd.(cmdClient)
			cli.Terminate()
			cli = cmd.options
			cmd.cherr <- nil
		default:
			ctx.WithField("cmd", cmd).Warn("Received unreckognized command")
		}
	}

	if cli != nil {
		_ = cli.Disconnect()
		cli.Terminate()
	}
	ctx.Debug("Stop monitoring MQTT client")
}

// createConnecter is used to start and subscribe a new client. It also make sure that if the
// created client goes down, another one is automatically created such that the client recover
// itself.
func createConnecter(id string, netAddr string, chmsg chan<- Msg, chcmd chan<- interface{}, ctx log.Interface) connecter {
	ctx.Debug("Create new connecter for MQTT client")
	var cli *client.Client
	cli = client.New(&client.Options{
		ErrorHandler: createErrorHandler(
			func() error { return createConnecter(id, netAddr, chmsg, chcmd, ctx)() },
			10000*InitialReconnectDelay,
			ctx,
		),
	})

	return func() error {
		ctx.Debug("(Re)Connecting MQTT client")
		err := cli.Connect(&client.ConnectOptions{
			Network:  "tcp",
			Address:  netAddr,
			ClientID: []byte(id),
		})

		if err != nil {
			return err
		}

		err = cli.Subscribe(&client.SubscribeOptions{
			SubReqs: []*client.SubReq{
				&client.SubReq{
					TopicFilter: []byte("+/devices/+/down"),
					QoS:         mqtt.QoS2,
					Handler: func(topic, msg []byte) {
						if len(msg) == 0 {
							return
						}
						chmsg <- Msg{
							Topic:   string(topic),
							Payload: msg,
							Type:    Down,
						}
					},
				},
			},
		})

		if err != nil {
			return err
		}

		cherr := make(chan error)
		select {
		case chcmd <- cmdClient{options: cli, cherr: cherr}:
			return <-cherr
		case <-time.After(time.Second):
			return errors.New(errors.Operational, "Timeout. Unable to set new client")
		}
	}
}

// createErrorHandler use the client reference to create an error handler function which will
// attempt to reconnect the client after a failure.
func createErrorHandler(tryReconnect connecter, maxDelay time.Duration, ctx log.Interface) client.ErrorHandler {
	delay := InitialReconnectDelay
	var reconnect func(fault error)
	reconnect = func(fault error) {
		if delay > maxDelay {
			ctx.WithError(fault).Error("Unable to reconnect the mqtt client")
			return
		}
		ctx.Debugf("Connection lost with MQTT broker. Trying to reconnect in %s", delay)
		<-time.After(delay)
		if err := tryReconnect(); err != nil {
			delay *= 10
			ctx.WithError(err).Debug("Failed to reconnect MQTT client.")
			reconnect(fault)
		}
	}

	return reconnect
}

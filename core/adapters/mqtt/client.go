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

const InitialReconnectDelay = 25 * time.Millisecond

type Client interface {
	Subscribe(*client.SubscribeOptions) error
	Unsubscribe(*client.UnsubscribeOptions) error
	Publish(*client.PublishOptions) error
	Connect(*client.ConnectOptions) error
	Disconnect() error
	Terminate()
}

type connecter func() error

type recoverableClient struct {
	chcmd chan<- interface{}
	cherr <-chan error
}

type disconnectOptions struct{}
type terminateOptions struct{}

func (c recoverableClient) Subscribe(o *client.SubscribeOptions) error {
	c.chcmd <- o
	return <-c.cherr
}

func (c recoverableClient) Unsubscribe(o *client.UnsubscribeOptions) error {
	c.chcmd <- o
	return <-c.cherr
}

func (c recoverableClient) Publish(o *client.PublishOptions) error {
	c.chcmd <- o
	return <-c.cherr
}

func (c recoverableClient) Connect(o *client.ConnectOptions) error {
	c.chcmd <- o
	return <-c.cherr
}

func (c recoverableClient) Disconnect() error {
	c.chcmd <- disconnectOptions{}
	return <-c.cherr
}

func (c recoverableClient) Terminate() {
	c.chcmd <- terminateOptions{}
}

// NewClient creates and connects a mqtt client with predefined options.
func NewClient(id string, netAddr string, ctx log.Interface) (Client, chan Msg, error) {
	ctx = ctx.WithField("id", id).WithField("address", netAddr)
	chcmd := make(chan interface{})
	cherr := make(chan error)
	chmsg := make(chan Msg)

	go monitorClient(id, netAddr, chcmd, cherr, ctx)

	tryConnect := createConnecter(id, netAddr, chmsg, chcmd, ctx)
	if err := tryConnect(); err != nil {
		close(chcmd)
		return nil, nil, errors.New(errors.Operational, err)
	}

	return recoverableClient{
		chcmd: chcmd,
		cherr: cherr,
	}, chmsg, nil
}

func monitorClient(id string, netAddr string, chcmd <-chan interface{}, cherr chan<- error, ctx log.Interface) {
	var cli *client.Client
	ctx = ctx.WithField("process", "monitorClient")
	ctx.Debug("Start monitoring MQTT client")

	for cmd := range chcmd {
		if cli == nil {
			init, ok := cmd.(*client.Client)
			if !ok {
				ctx.Warn("Received cmd whereas client is nil. Ignored")
				continue
			}
			ctx.Debug("Setup initial MQTT client")
			cli = init
			continue
		}
		switch cmd.(type) {
		case *client.SubscribeOptions:
			ctx.Debug("Client received new subscription order")
			cherr <- cli.Subscribe(cmd.(*client.SubscribeOptions))
		case *client.UnsubscribeOptions:
			ctx.Debug("Client received new unsubscription order")
			cherr <- cli.Unsubscribe(cmd.(*client.UnsubscribeOptions))
		case *client.PublishOptions:
			ctx.Debug("Client received new publication order")
			cherr <- cli.Publish(cmd.(*client.PublishOptions))
		case *client.ConnectOptions:
			ctx.Debug("Client received new connection order")
			cherr <- cli.Connect(cmd.(*client.ConnectOptions))
		case disconnectOptions:
			ctx.Debug("Client received disconnection order")
			cherr <- cli.Disconnect()
		case terminateOptions:
			ctx.Debug("Client received termination order")
			cli.Terminate()
			cli = nil
		case *client.Client:
			ctx.Debug("Replacing client with another one")
			_ = cli.Disconnect()
			cli.Terminate()
			cli = cmd.(*client.Client)
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

func createConnecter(id string, netAddr string, chmsg chan<- Msg, chcmd chan<- interface{}, ctx log.Interface) connecter {
	ctx.Debug("Create new connecter for MQTT client")
	var cli *client.Client
	cli = client.New(&client.Options{
		ErrorHandler: func(err error) {
			createErrorHandler(
				createConnecter(id, netAddr, chmsg, chcmd, ctx),
				10000*InitialReconnectDelay,
				ctx,
			)(err)
		},
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

		if err != nil {
			return err
		}

		select {
		case chcmd <- cli:
			return nil
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

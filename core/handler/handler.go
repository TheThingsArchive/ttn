// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/TheThingsNetwork/ttn/api"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	pb "github.com/TheThingsNetwork/ttn/api/handler"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"gopkg.in/redis.v3"
)

// Handler component
type Handler interface {
	core.ComponentInterface
	core.ManagementInterface

	HandleUplink(uplink *pb_broker.DeduplicatedUplinkMessage) error
	HandleActivation(activation *pb_broker.DeduplicatedDeviceActivationRequest) (*pb.DeviceActivationResponse, error)
	EnqueueDownlink(appDownlink *mqtt.DownlinkMessage) error
}

// NewRedisHandler creates a new Redis-backed Handler
func NewRedisHandler(client *redis.Client, ttnBrokerAddr string, mqttUsername string, mqttPassword string, mqttBrokers ...string) Handler {
	return &handler{
		devices:       device.NewRedisDeviceStore(client),
		applications:  application.NewRedisApplicationStore(client),
		ttnBrokerAddr: ttnBrokerAddr,
		mqttUsername:  mqttUsername,
		mqttPassword:  mqttPassword,
		mqttBrokers:   mqttBrokers,
	}
}

type handler struct {
	*core.Component

	devices            device.Store
	applications       application.Store
	applicationIds     []string // this is a cache
	applicationIdsLock sync.RWMutex

	ttnBrokerAddr    string
	ttnBrokerConn    *grpc.ClientConn
	ttnBroker        pb_broker.BrokerClient
	ttnBrokerManager pb_broker.BrokerManagerClient
	ttnDeviceManager pb_lorawan.DeviceManagerClient

	downlink chan *pb_broker.DownlinkMessage

	mqttClient   mqtt.Client
	mqttUsername string
	mqttPassword string
	mqttBrokers  []string

	mqttUp         chan *mqtt.UplinkMessage
	mqttActivation chan *mqtt.Activation
}

func (h *handler) announce() error {
	h.applicationIdsLock.RLock()
	defer h.applicationIdsLock.RUnlock()
	h.Component.Identity.Metadata = []*pb_discovery.Metadata{}
	for _, id := range h.applicationIds {
		h.Identity.Metadata = append(h.Component.Identity.Metadata, &pb_discovery.Metadata{
			Key: pb_discovery.Metadata_APP_ID, Value: []byte(id),
		})
	}
	return h.Component.Announce()
}

func (h *handler) Init(c *core.Component) error {
	h.Component = c
	err := h.Component.UpdateTokenKey()
	if err != nil {
		return err
	}

	h.applications.Set(&application.Application{
		AppID: "htdvisser",
	})

	apps, err := h.applications.List()
	if err != nil {
		return err
	}
	h.applicationIdsLock.Lock()
	for _, app := range apps {
		h.applicationIds = append(h.applicationIds, app.AppID)
	}
	h.applicationIdsLock.Unlock()
	err = h.announce()
	if err != nil {
		return err
	}

	var brokers []string
	for _, broker := range h.mqttBrokers {
		brokers = append(brokers, fmt.Sprintf("tcp://%s", broker))
	}
	err = h.HandleMQTT(h.mqttUsername, h.mqttPassword, brokers...)
	if err != nil {
		return err
	}

	err = h.associateBroker()
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) associateBroker() error {
	conn, err := grpc.Dial(h.ttnBrokerAddr, api.DialOptions...)
	if err != nil {
		return err
	}
	h.ttnBrokerConn = conn
	h.ttnBroker = pb_broker.NewBrokerClient(conn)
	h.ttnBrokerManager = pb_broker.NewBrokerManagerClient(conn)
	h.ttnDeviceManager = pb_lorawan.NewDeviceManagerClient(conn)

	ctx := h.GetContext()

	h.downlink = make(chan *pb_broker.DownlinkMessage)

	go func() {
		for {
			upStream, err := h.ttnBroker.Subscribe(ctx, &pb_broker.SubscribeRequest{})
			if err != nil {
				<-time.After(api.Backoff)
				continue
			}
			for {
				in, err := upStream.Recv()
				if err != nil {
					h.Ctx.Errorf("ttn/handler: Error in Broker subscribe: %s", err)
					break
				}
				go h.HandleUplink(in)
			}
		}
	}()

	go func() {
		for {
			downStream, err := h.ttnBroker.Publish(ctx)
			if err != nil {
				<-time.After(api.Backoff)
				continue
			}
			for downlink := range h.downlink {
				err := downStream.Send(downlink)
				if err != nil {
					h.Ctx.Errorf("ttn/handler: Error in Broker publish: %s", err)
					break
				}
			}
		}
	}()

	return nil
}

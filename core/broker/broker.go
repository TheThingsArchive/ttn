// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc"

	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/discovery"
)

type Broker interface {
	core.ComponentInterface
	core.ManagementInterface

	HandleUplink(uplink *pb.UplinkMessage) error
	HandleDownlink(downlink *pb.DownlinkMessage) error
	HandleActivation(activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error)

	ActivateRouter(id string) (<-chan *pb.DownlinkMessage, error)
	DeactivateRouter(id string) error
	ActivateHandler(id string) (<-chan *pb.DeduplicatedUplinkMessage, error)
	DeactivateHandler(id string) error
}

func NewRedisBroker(client *redis.Client, networkserver string, timeout time.Duration) Broker {
	return &broker{
		routers:                make(map[string]chan *pb.DownlinkMessage),
		handlers:               make(map[string]chan *pb.DeduplicatedUplinkMessage),
		uplinkDeduplicator:     NewDeduplicator(timeout),
		activationDeduplicator: NewDeduplicator(timeout),
		nsAddr:                 networkserver,
	}
}

type broker struct {
	*core.Component
	routers                map[string]chan *pb.DownlinkMessage
	routersLock            sync.RWMutex
	handlerDiscovery       discovery.HandlerDiscovery
	handlers               map[string]chan *pb.DeduplicatedUplinkMessage
	handlersLock           sync.RWMutex
	nsAddr                 string
	ns                     networkserver.NetworkServerClient
	nsManager              pb_lorawan.DeviceManagerClient
	uplinkDeduplicator     Deduplicator
	activationDeduplicator Deduplicator
}

func (b *broker) Init(c *core.Component) error {
	b.Component = c
	err := b.Component.UpdateTokenKey()
	if err != nil {
		return err
	}
	err = b.Component.Announce()
	if err != nil {
		return err
	}
	b.handlerDiscovery = discovery.NewHandlerDiscovery(b.Component)
	b.handlerDiscovery.All() // Update cache
	conn, err := grpc.Dial(b.nsAddr, api.DialOptions...)
	if err != nil {
		return err
	}
	b.ns = networkserver.NewNetworkServerClient(conn)
	b.nsManager = pb_lorawan.NewDeviceManagerClient(conn)
	return nil
}

func (b *broker) ActivateRouter(id string) (<-chan *pb.DownlinkMessage, error) {
	b.routersLock.Lock()
	defer b.routersLock.Unlock()
	if existing, ok := b.routers[id]; ok {
		return existing, errors.New("Router already active")
	}
	b.routers[id] = make(chan *pb.DownlinkMessage)
	return b.routers[id], nil
}

func (b *broker) DeactivateRouter(id string) error {
	b.routersLock.Lock()
	defer b.routersLock.Unlock()
	if channel, ok := b.routers[id]; ok {
		close(channel)
		delete(b.routers, id)
		return nil
	}
	return errors.New("Router not active")
}

func (b *broker) getRouter(id string) (chan<- *pb.DownlinkMessage, error) {
	b.routersLock.RLock()
	defer b.routersLock.RUnlock()
	if router, ok := b.routers[id]; ok {
		return router, nil
	}
	return nil, errors.New("Router not active")
}

func (b *broker) ActivateHandler(id string) (<-chan *pb.DeduplicatedUplinkMessage, error) {
	b.handlersLock.Lock()
	defer b.handlersLock.Unlock()
	if existing, ok := b.handlers[id]; ok {
		return existing, errors.New("Handler already active")
	}
	b.handlers[id] = make(chan *pb.DeduplicatedUplinkMessage)
	return b.handlers[id], nil
}

func (b *broker) DeactivateHandler(id string) error {
	b.handlersLock.Lock()
	defer b.handlersLock.Unlock()
	if channel, ok := b.handlers[id]; ok {
		close(channel)
		delete(b.handlers, id)
		return nil
	}
	return errors.New("Handler not active")
}

func (b *broker) getHandler(id string) (chan<- *pb.DeduplicatedUplinkMessage, error) {
	b.handlersLock.RLock()
	defer b.handlersLock.RUnlock()
	if handler, ok := b.handlers[id]; ok {
		return handler, nil
	}
	return nil, errors.New("Handler not active")
}

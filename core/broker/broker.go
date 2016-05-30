package broker

import (
	"errors"
	"sync"

	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/core/broker/application"
)

type Broker interface {
	HandleUplink(uplink *pb.UplinkMessage) error
	HandleDownlink(downlink *pb.DownlinkMessage) error
	HandleActivation(activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error)

	ActivateRouter(id string) (<-chan *pb.DownlinkMessage, error)
	DeactivateRouter(id string) error
	ActivateHandler(id string) (<-chan *pb.DeduplicatedUplinkMessage, error)
	DeactivateHandler(id string) error
}

type broker struct {
	identity               *pb_discovery.Announcement
	routers                map[string]chan *pb.DownlinkMessage
	routersLock            sync.RWMutex
	handlers               map[string]chan *pb.DeduplicatedUplinkMessage
	handlersLock           sync.RWMutex
	applications           application.Store
	ns                     networkserver.NetworkServerClient
	uplinkDeduplicator     Deduplicator
	activationDeduplicator Deduplicator
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

func (b *broker) getContext() context.Context {
	var id, token string
	if b.identity != nil {
		id = b.identity.Id
		token = b.identity.Token
	}
	md := metadata.Pairs(
		"token", token,
		"id", id,
	)
	ctx := metadata.NewContext(context.Background(), md)
	return ctx
}

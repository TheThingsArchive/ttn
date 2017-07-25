// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"fmt"
	"strings"
	"sync"
	"time"

	pb "github.com/TheThingsNetwork/api/broker"
	"github.com/TheThingsNetwork/api/monitor/monitorclient"
	"github.com/TheThingsNetwork/api/networkserver"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/go-utils/grpc/auth"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"google.golang.org/grpc"
)

type Broker interface {
	component.Interface
	component.ManagementInterface

	SetNetworkServer(addr, cert, token string)

	HandleUplink(uplink *pb.UplinkMessage) error
	HandleDownlink(downlink *pb.DownlinkMessage) error
	HandleActivation(activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error)

	ActivateRouter(id string) (<-chan *pb.DownlinkMessage, error)
	DeactivateRouter(id string) error
	ActivateHandlerUplink(id string) (<-chan *pb.DeduplicatedUplinkMessage, error)
	DeactivateHandlerUplink(id string) error
}

func NewBroker(timeout time.Duration) Broker {
	return &broker{
		routers:                make(map[string]chan *pb.DownlinkMessage),
		handlers:               make(map[string]*handler),
		uplinkDeduplicator:     NewDeduplicator(timeout),
		activationDeduplicator: NewDeduplicator(timeout),
	}
}

func (b *broker) SetNetworkServer(addr, cert, token string) {
	b.nsAddr = addr
	b.nsCert = cert
	b.nsToken = token
}

type broker struct {
	*component.Component
	routers                map[string]chan *pb.DownlinkMessage
	routersLock            sync.RWMutex
	handlers               map[string]*handler
	handlersLock           sync.RWMutex
	nsAddr                 string
	nsCert                 string
	nsToken                string
	nsConn                 *grpc.ClientConn
	ns                     networkserver.NetworkServerClient
	uplinkDeduplicator     Deduplicator
	activationDeduplicator Deduplicator
	status                 *status
	monitorStream          monitorclient.Stream
}

func (b *broker) checkPrefixAnnouncements() error {
	// Get prefixes from NS
	nsPrefixes := map[types.DevAddrPrefix]string{}
	devAddrClient := pb_lorawan.NewDevAddrManagerClient(b.nsConn)
	resp, err := devAddrClient.GetPrefixes(b.GetContext(""), &pb_lorawan.PrefixesRequest{})
	if err != nil {
		return errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not return prefixes")
	}
	for _, mapping := range resp.Prefixes {
		prefix, err := types.ParseDevAddrPrefix(mapping.Prefix)
		if err != nil {
			continue
		}
		nsPrefixes[prefix] = strings.Join(mapping.Usage, ",")
	}

	// Get self from Discovery
	self, err := b.Component.Discover("broker", b.Component.Identity.ID)
	if err != nil {
		return err
	}
	announcedPrefixes := self.DevAddrPrefixes()

nextPrefix:
	for nsPrefix, usage := range nsPrefixes {
		if !strings.Contains(usage, "world") && !strings.Contains(usage, "local") {
			continue
		}
		for _, announcedPrefix := range announcedPrefixes {
			if nsPrefix.DevAddr == announcedPrefix.DevAddr && nsPrefix.Length == announcedPrefix.Length {
				b.Ctx.WithField("NSPrefix", nsPrefix).WithField("DPrefix", announcedPrefix).Info("Prefix found in Discovery")
				continue nextPrefix
			}
		}
		b.Ctx.WithField("Prefix", nsPrefix).Warn("Prefix not announced in Discovery")
	}

	return nil
}

func (b *broker) Init(c *component.Component) error {
	b.Component = c
	b.InitStatus()
	err := b.Component.UpdateTokenKey()
	if err != nil {
		return err
	}
	err = b.Component.Announce()
	if err != nil {
		return err
	}
	b.Discovery.GetAll("handler") // Update cache
	var conn *grpc.ClientConn
	if b.nsCert == "" {
		conn, err = api.Dial(b.nsAddr)
	} else {
		conn, err = api.DialWithCert(b.nsAddr, b.nsCert)
	}
	if err != nil {
		return err
	}
	b.nsConn = conn
	b.ns = networkserver.NewNetworkServerClient(conn)
	b.checkPrefixAnnouncements()
	b.Component.SetStatus(component.StatusHealthy)
	if b.Component.Monitor != nil {
		b.monitorStream = b.Component.Monitor.BrokerClient(b.Context, grpc.PerRPCCredentials(auth.WithStaticToken(b.AccessToken)))
		go func() {
			for range time.Tick(b.Component.Config.StatusInterval) {
				b.monitorStream.Send(b.GetStatus())
			}
		}()
	}
	return nil
}

func (b *broker) Shutdown() {}

func (b *broker) ActivateRouter(id string) (<-chan *pb.DownlinkMessage, error) {
	b.routersLock.Lock()
	defer b.routersLock.Unlock()
	if existing, ok := b.routers[id]; ok {
		return existing, errors.NewErrInternal(fmt.Sprintf("Router %s already active", id))
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
	return errors.NewErrInternal(fmt.Sprintf("Router %s not active", id))
}

func (b *broker) getRouter(id string) (chan<- *pb.DownlinkMessage, error) {
	b.routersLock.RLock()
	defer b.routersLock.RUnlock()
	if router, ok := b.routers[id]; ok {
		return router, nil
	}
	return nil, errors.NewErrInternal(fmt.Sprintf("Router %s not active", id))
}

type handler struct {
	conn   *grpc.ClientConn
	uplink chan *pb.DeduplicatedUplinkMessage
	sync.Mutex
}

func (b *broker) getHandler(id string) *handler {
	b.handlersLock.Lock()
	defer b.handlersLock.Unlock()
	if existing, ok := b.handlers[id]; ok {
		return existing
	}
	b.handlers[id] = new(handler)
	return b.handlers[id]
}

func (b *broker) ActivateHandlerUplink(id string) (<-chan *pb.DeduplicatedUplinkMessage, error) {
	hdl := b.getHandler(id)
	hdl.Lock()
	defer hdl.Unlock()
	if hdl.uplink != nil {
		return hdl.uplink, errors.NewErrInternal(fmt.Sprintf("Handler %s already active", id))
	}
	hdl.uplink = make(chan *pb.DeduplicatedUplinkMessage)
	return hdl.uplink, nil
}

func (b *broker) DeactivateHandlerUplink(id string) error {
	hdl := b.getHandler(id)
	hdl.Lock()
	defer hdl.Unlock()
	if hdl.uplink == nil {
		return errors.NewErrInternal(fmt.Sprintf("Handler %s not active", id))
	}
	close(hdl.uplink)
	hdl.uplink = nil
	return nil
}

func (b *broker) getHandlerUplink(id string) (chan<- *pb.DeduplicatedUplinkMessage, error) {
	hdl := b.getHandler(id)
	hdl.Lock()
	defer hdl.Unlock()
	if hdl.uplink == nil {
		return nil, errors.NewErrInternal(fmt.Sprintf("Handler %s not active", id))
	}
	return hdl.uplink, nil
}

func (b *broker) getHandlerConn(id string) (*grpc.ClientConn, error) {
	hdl := b.getHandler(id)
	hdl.Lock()
	defer hdl.Unlock()
	if hdl.conn != nil {
		return hdl.conn, nil
	}
	announcement, err := b.Discover("handler", id)
	if err != nil {
		return nil, err
	}
	conn, err := announcement.Dial(b.Pool)
	if err != nil {
		return nil, err
	}
	hdl.conn = conn
	return hdl.conn, nil
}

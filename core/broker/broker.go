// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"google.golang.org/grpc"
)

type Broker interface {
	core.ComponentInterface
	core.ManagementInterface

	SetNetworkServer(addr, cert, token string)

	HandleUplink(uplink *pb.UplinkMessage) error
	HandleDownlink(downlink *pb.DownlinkMessage) error
	HandleActivation(activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error)

	ActivateRouter(id string) (<-chan *pb.DownlinkMessage, error)
	DeactivateRouter(id string) error
	ActivateHandler(id string) (<-chan *pb.DeduplicatedUplinkMessage, error)
	DeactivateHandler(id string) error
}

func NewBroker(timeout time.Duration) Broker {
	return &broker{
		routers:                make(map[string]chan *pb.DownlinkMessage),
		handlers:               make(map[string]chan *pb.DeduplicatedUplinkMessage),
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
	*core.Component
	routers                map[string]chan *pb.DownlinkMessage
	routersLock            sync.RWMutex
	handlers               map[string]chan *pb.DeduplicatedUplinkMessage
	handlersLock           sync.RWMutex
	nsAddr                 string
	nsCert                 string
	nsToken                string
	nsConn                 *grpc.ClientConn
	ns                     networkserver.NetworkServerClient
	uplinkDeduplicator     Deduplicator
	activationDeduplicator Deduplicator
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
	var announcedPrefixes []types.DevAddrPrefix
	self, err := b.Component.Discover("broker", b.Component.Identity.Id)
	if err != nil {
		return err
	}
	for _, meta := range self.Metadata {
		if meta.Key == pb_discovery.Metadata_PREFIX && len(meta.Value) == 5 {
			var prefix types.DevAddrPrefix
			copy(prefix.DevAddr[:], meta.Value[1:])
			prefix.Length = int(meta.Value[0])
			announcedPrefixes = append(announcedPrefixes, prefix)
		}
	}

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
	b.Discovery.GetAll("handler") // Update cache
	conn, err := api.DialWithCert(b.nsAddr, b.nsCert)
	if err != nil {
		return err
	}
	b.nsConn = conn
	b.ns = networkserver.NewNetworkServerClient(conn)
	b.checkPrefixAnnouncements()
	b.Component.SetStatus(core.StatusHealthy)
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

func (b *broker) ActivateHandler(id string) (<-chan *pb.DeduplicatedUplinkMessage, error) {
	b.handlersLock.Lock()
	defer b.handlersLock.Unlock()
	if existing, ok := b.handlers[id]; ok {
		return existing, errors.NewErrInternal(fmt.Sprintf("Handler %s already active", id))
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
	return errors.NewErrInternal(fmt.Sprintf("Handler %s not active", id))
}

func (b *broker) getHandler(id string) (chan<- *pb.DeduplicatedUplinkMessage, error) {
	b.handlersLock.RLock()
	defer b.handlersLock.RUnlock()
	if handler, ok := b.handlers[id]; ok {
		return handler, nil
	}
	return nil, errors.NewErrInternal(fmt.Sprintf("Handler %s not active", id))
}

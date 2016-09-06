// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"sync"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// BrokerCacheTime indicates how long the BrokerDiscovery should cache the services
var BrokerCacheTime = 30 * time.Minute

// BrokerDiscovery is used as a client to discover Brokers
type BrokerDiscovery interface {
	Discover(devAddr types.DevAddr) ([]*pb.Announcement, error)
	All() ([]*pb.Announcement, error)
}

type brokerDiscovery struct {
	component       *core.Component
	cache           []*pb.Announcement
	cacheLock       sync.RWMutex
	cacheValidUntil time.Time
}

// NewBrokerDiscovery returns a new BrokerDiscovery on top of the given gRPC connection
func NewBrokerDiscovery(component *core.Component) BrokerDiscovery {
	return &brokerDiscovery{component: component}
}

func (d *brokerDiscovery) refreshCache() error {
	res, err := d.component.Discovery.GetAll(d.component.GetContext(""), &pb.GetAllRequest{ServiceName: "broker"})
	if err != nil {
		return err
	}
	// TODO: validate response
	d.cacheLock.Lock()
	defer d.cacheLock.Unlock()
	d.cacheValidUntil = time.Now().Add(BrokerCacheTime)
	d.cache = res.Services
	return nil
}

func (d *brokerDiscovery) All() (announcements []*pb.Announcement, err error) {
	d.cacheLock.Lock()
	defer d.cacheLock.Unlock()
	if time.Now().After(d.cacheValidUntil) {
		d.cacheValidUntil = time.Now().Add(10 * time.Second)
		go d.refreshCache()
	}
	announcements = d.cache
	return
}

func (d *brokerDiscovery) Discover(devAddr types.DevAddr) ([]*pb.Announcement, error) {
	d.cacheLock.Lock()
	defer d.cacheLock.Unlock()

	if time.Now().After(d.cacheValidUntil) {
		d.cacheValidUntil = time.Now().Add(10 * time.Second)
		go d.refreshCache()
	}

	matches := []*pb.Announcement{}
	for _, service := range d.cache {
		for _, meta := range service.Metadata {
			if meta.Key == pb.Metadata_PREFIX && len(meta.Value) == 5 {
				var prefix types.DevAddrPrefix
				copy(prefix.DevAddr[:], meta.Value[1:])
				prefix.Length = int(meta.Value[0])

				if devAddr.HasPrefix(prefix) {
					matches = append(matches, service)
					break
				}
			}
		}
	}

	return matches, nil
}

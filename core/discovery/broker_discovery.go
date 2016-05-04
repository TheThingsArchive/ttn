package discovery

import (
	"bytes"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core/types"
	"golang.org/x/net/context"
)

// BrokerCacheTime indicates how long the BrokerDiscovery should cache the services
var BrokerCacheTime = 30 * time.Minute

// BrokerDiscovery is used as a client to discover Brokers
type BrokerDiscovery interface {
	Discover(devAddr types.DevAddr) ([]*pb.Announcement, error)
}

type brokerDiscovery struct {
	serverAddress   string
	cache           []*pb.Announcement
	cacheLock       sync.RWMutex
	cacheValidUntil time.Time
}

// NewBrokerDiscovery returns a new BrokerDiscovery on top of the given gRPC connection
func NewBrokerDiscovery(serverAddress string) BrokerDiscovery {
	return &brokerDiscovery{serverAddress: serverAddress}
}

func (d *brokerDiscovery) refreshCache() error {
	// Connect to the server
	conn, err := grpc.Dial(d.serverAddress, DialOptions...)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewDiscoveryClient(conn)
	res, err := client.Discover(context.Background(), &pb.DiscoverRequest{ServiceName: "broker"})
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

func (d *brokerDiscovery) Discover(devAddr types.DevAddr) ([]*pb.Announcement, error) {
	d.cacheLock.Lock()
	if time.Now().After(d.cacheValidUntil) {
		d.cacheValidUntil = time.Now().Add(10 * time.Second)
		go d.refreshCache()
	}
	d.cacheLock.Unlock()
	d.cacheLock.RLock()
	defer d.cacheLock.RUnlock()
	matches := []*pb.Announcement{}
	for _, service := range d.cache {
		for _, meta := range service.Metadata {
			if meta.Key == pb.Metadata_PREFIX && bytes.HasPrefix(devAddr.Bytes(), meta.Value) {
				matches = append(matches, service)
				break
			}
		}
	}
	return matches, nil
}

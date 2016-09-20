// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

// Client is used to manage applications and devices on a handler
import (
	"fmt"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/bluele/gcache"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// CacheSize indicates the number of components that are cached
var CacheSize = 1000

// CacheExpiration indicates the time a cached item is valid
var CacheExpiration = 5 * time.Minute

// Client is used as the main client to the Discovery server
type Client interface {
	Announce(token string) error
	GetAll(serviceName string) ([]*Announcement, error)
	Get(serviceName, id string) (*Announcement, error)
	AddMetadata(key Metadata_Key, value []byte, token string) error
	DeleteMetadata(key Metadata_Key, value []byte, token string) error
	GetAllForMetadata(serviceName string, key Metadata_Key, matchFunc func(value []byte) bool) ([]*Announcement, error)
	GetAllBrokersForDevAddr(devAddr types.DevAddr) ([]*Announcement, error)
	GetAllHandlersForAppID(appID string) ([]*Announcement, error)
	Close() error
}

// NewClient returns a new Client
func NewClient(server string, announcement *Announcement, tokenFunc func() string) (Client, error) {
	conn, err := grpc.Dial(server, append(api.DialOptions, grpc.WithBlock(), grpc.WithInsecure())...)
	if err != nil {
		return nil, err
	}
	client := &DefaultClient{
		lists:        make(map[string][]*Announcement),
		listsUpdated: make(map[string]time.Time),
		self:         announcement,
		tokenFunc:    tokenFunc,
		conn:         conn,
		client:       NewDiscoveryClient(conn),
	}
	client.cache = gcache.
		New(CacheSize).
		Expiration(CacheExpiration).
		ARC().
		LoaderFunc(func(k interface{}) (interface{}, error) {
			key, ok := k.(cacheKey)
			if !ok {
				return nil, fmt.Errorf("wrong type for cacheKey: %T", k)
			}
			return client.get(key.serviceName, key.id)
		}).
		Build()
	return client, nil
}

// DefaultClient is a wrapper around DiscoveryClient
type DefaultClient struct {
	sync.Mutex
	cache        gcache.Cache
	listsUpdated map[string]time.Time
	lists        map[string][]*Announcement
	self         *Announcement
	tokenFunc    func() string
	conn         *grpc.ClientConn
	client       DiscoveryClient
}

type cacheKey struct {
	serviceName string
	id          string
}

func (c *DefaultClient) getContext(token string) context.Context {
	if token == "" {
		token = c.tokenFunc()
	}
	md := metadata.Pairs(
		"service-name", c.self.ServiceName,
		"id", c.self.Id,
		"token", token,
		"net-address", c.self.NetAddress,
	)
	ctx := metadata.NewContext(context.Background(), md)
	return ctx
}

func (c *DefaultClient) get(serviceName, id string) (*Announcement, error) {
	res, err := c.client.Get(c.getContext(""), &GetRequest{
		ServiceName: serviceName,
		Id:          id,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *DefaultClient) getAll(serviceName string) ([]*Announcement, error) {
	res, err := c.client.GetAll(c.getContext(""), &GetAllRequest{ServiceName: serviceName})
	if err != nil {
		return nil, err
	}
	c.lists[serviceName] = res.Services
	c.listsUpdated[serviceName] = time.Now()
	for _, announcement := range res.Services {
		c.cache.Set(&cacheKey{serviceName: announcement.ServiceName, id: announcement.Id}, announcement)
	}
	return res.Services, nil
}

// Announce announces the configured announcement to the discovery server
func (c *DefaultClient) Announce(token string) error {
	_, err := c.client.Announce(c.getContext(token), c.self)
	return err
}

// GetAll returns all services of the given service type
func (c *DefaultClient) GetAll(serviceName string) ([]*Announcement, error) {
	c.Lock()
	defer c.Unlock()

	// If list initialized, return cached version
	if list, ok := c.lists[serviceName]; ok && len(list) > 0 {
		// And update if expired
		if c.listsUpdated[serviceName].Add(CacheExpiration).Before(time.Now()) {
			go func() {
				c.Lock()
				defer c.Unlock()
				c.getAll(serviceName)
			}()
		}
		return list, nil
	}

	// If list not initialized, do request
	return c.getAll(serviceName)
}

// Get returns the (cached) service annoucement for the given service type and id
func (c *DefaultClient) Get(serviceName, id string) (*Announcement, error) {
	res, err := c.cache.Get(cacheKey{serviceName, id})
	if err != nil {
		return nil, err
	}
	return res.(*Announcement), nil
}

// AddMetadata publishes metadata for the current component to the Discovery server
func (c *DefaultClient) AddMetadata(key Metadata_Key, value []byte, token string) error {
	_, err := c.client.AddMetadata(c.getContext(token), &MetadataRequest{
		ServiceName: c.self.ServiceName,
		Id:          c.self.Id,
		Metadata: &Metadata{
			Key:   key,
			Value: value,
		},
	})
	return err
}

// DeleteMetadata deletes metadata for the current component from the Discovery server
func (c *DefaultClient) DeleteMetadata(key Metadata_Key, value []byte, token string) error {
	_, err := c.client.DeleteMetadata(c.getContext(token), &MetadataRequest{
		ServiceName: c.self.ServiceName,
		Id:          c.self.Id,
		Metadata: &Metadata{
			Key:   key,
			Value: value,
		},
	})
	return err
}

// GetAllForMetadata returns all annoucements of given type that contain given metadata and match the given function
func (c *DefaultClient) GetAllForMetadata(serviceName string, key Metadata_Key, matchFunc func(value []byte) bool) ([]*Announcement, error) {
	announcements, err := c.GetAll(serviceName)
	if err != nil {
		return nil, err
	}
	res := make([]*Announcement, 0, len(announcements))
nextAnnouncement:
	for _, announcement := range announcements {
		for _, meta := range announcement.Metadata {
			if meta.Key == key && matchFunc(meta.Value) {
				res = append(res, announcement)
				continue nextAnnouncement
			}
		}
	}
	return res, nil
}

// GetAllBrokersForDevAddr returns all brokers that can handle the given DevAddr
func (c *DefaultClient) GetAllBrokersForDevAddr(devAddr types.DevAddr) ([]*Announcement, error) {
	return c.GetAllForMetadata("broker", Metadata_PREFIX, func(value []byte) bool {
		if len(value) != 5 {
			return false
		}
		var prefix types.DevAddrPrefix
		copy(prefix.DevAddr[:], value[1:])
		prefix.Length = int(value[0])
		return devAddr.HasPrefix(prefix)
	})
}

// GetAllHandlersForAppID returns all handlers that can handle the given AppID
func (c *DefaultClient) GetAllHandlersForAppID(appID string) ([]*Announcement, error) {
	return c.GetAllForMetadata("handler", Metadata_APP_ID, func(value []byte) bool {
		return string(value) == appID
	})
}

// Close purges the cache and closes the connection with the Discovery server
func (c *DefaultClient) Close() error {
	c.cache.Purge()
	return c.conn.Close()
}

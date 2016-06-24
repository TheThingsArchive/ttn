// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package discovery implements TTN Service Discovery.
package discovery

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/redis.v3"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core"
)

// Discovery specifies the interface for the TTN Service Discovery component
type Discovery interface {
	core.ComponentInterface
	Announce(announcement *pb.Announcement) error
	Discover(serviceName string, ids ...string) ([]*pb.Announcement, error)
}

// discovery is a reference implementation for a TTN Service Discovery component.
type discovery struct {
	*core.Component
	services map[string]map[string]*pb.Announcement
	sync.RWMutex
}

func (d *discovery) Init(c *core.Component) error {
	d.Component = c
	err := d.Component.UpdateTokenKey()
	if err != nil {
		return err
	}
	return nil
}

func (d *discovery) Announce(announcement *pb.Announcement) error {
	d.Lock()
	defer d.Unlock()

	// Get the list
	services, ok := d.services[announcement.ServiceName]
	if !ok {
		services = map[string]*pb.Announcement{}
		d.services[announcement.ServiceName] = services
	}

	// Find an existing announcement
	service, ok := services[announcement.Id]
	if ok {
		if announcement.Token == service.Token {
			*service = *announcement
		} else {
			return errors.New("ttn/core: Invalid token")
		}
	} else {
		services[announcement.Id] = announcement
	}

	return nil
}

func (d *discovery) Discover(serviceName string, ids ...string) ([]*pb.Announcement, error) {
	d.RLock()
	defer d.RUnlock()

	// Get the list
	services, ok := d.services[serviceName]
	if !ok {
		return []*pb.Announcement{}, nil
	}

	// Traverse the list
	announcements := make([]*pb.Announcement, 0, len(services))
	for _, service := range services {
		serviceCopy := *service
		serviceCopy.Token = ""
		if len(ids) == 0 {
			announcements = append(announcements, &serviceCopy)
		} else {
			for _, id := range ids {
				if service.Id == id {
					announcements = append(announcements, &serviceCopy)
					break
				}
			}
		}
	}
	return announcements, nil
}

// NewRedisDiscovery creates a new Redis-based discovery service
func NewRedisDiscovery(client *redis.Client) Discovery {
	return &redisDiscovery{
		client: client,
	}
}

const redisAnnouncementPrefix = "service"

type redisDiscovery struct {
	*core.Component
	client *redis.Client
}

func (d *redisDiscovery) Init(c *core.Component) error {
	d.Component = c
	err := d.Component.UpdateTokenKey()
	if err != nil {
		return err
	}
	return nil
}

func (d *redisDiscovery) Announce(announcement *pb.Announcement) error {
	key := fmt.Sprintf("%s:%s:%s", redisAnnouncementPrefix, announcement.ServiceName, announcement.Id)

	if token, err := d.client.HGet(key, "token").Result(); err == nil && token != announcement.Token {
		return errors.New("ttn/core: Invalid token")
	}

	dmap, err := announcement.ToStringStringMap(pb.AnnouncementProperties...)
	if err != nil {
		return err
	}

	return d.client.HMSetMap(key, dmap).Err()
}

func (d *redisDiscovery) Discover(serviceName string, ids ...string) ([]*pb.Announcement, error) {
	announcements := []*pb.Announcement{}
	if len(ids) == 0 {
		keys, err := d.client.Keys(fmt.Sprintf("%s:%s:*", redisAnnouncementPrefix, serviceName)).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			if parts := strings.Split(key, ":"); len(parts) == 3 {
				ids = append(ids, parts[2])
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(ids))
	results := make(chan *pb.Announcement)
	go func() {
		wg.Wait()
		close(results)
	}()
	for _, id := range ids {
		go func(id string) {
			res, err := d.Get(serviceName, id)
			if err == nil && res != nil {
				results <- res
			}
			wg.Done()
		}(id)
	}

	for res := range results {
		announcements = append(announcements, res)
	}

	return announcements, nil
}

func (d *redisDiscovery) Get(serviceName string, id string) (*pb.Announcement, error) {
	key := fmt.Sprintf("%s:%s:%s", redisAnnouncementPrefix, serviceName, id)
	res, err := d.client.HGetAllMap(key).Result()
	if err != nil {
		return nil, err
	} else if len(res) == 0 {
		return nil, redis.Nil // This might be a bug in redis package
	}
	announcement := &pb.Announcement{}
	err = announcement.FromStringStringMap(res)
	if err != nil {
		return nil, err
	}
	announcement.Token = ""
	return announcement, nil
}

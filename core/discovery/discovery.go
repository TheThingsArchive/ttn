// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package discovery implements TTN Service Discovery.
package discovery

import (
	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/discovery/announcement"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"gopkg.in/redis.v4"
)

// Discovery specifies the interface for the TTN Service Discovery component
type Discovery interface {
	core.ComponentInterface
	WithCache(options announcement.CacheOptions)
	Announce(announcement *pb.Announcement) error
	GetAll(serviceName string) ([]*pb.Announcement, error)
	Get(serviceName string, id string) (*pb.Announcement, error)
	AddMetadata(serviceName string, id string, metadata *pb.Metadata) error
	DeleteMetadata(serviceName string, id string, metadata *pb.Metadata) error
}

// discovery is a reference implementation for a TTN Service Discovery component.
type discovery struct {
	*core.Component
	services announcement.Store
}

func (d *discovery) WithCache(options announcement.CacheOptions) {
	d.services = announcement.NewCachedAnnouncementStore(d.services, options)
}

func (d *discovery) Init(c *core.Component) error {
	d.Component = c
	err := d.Component.UpdateTokenKey()
	if err != nil {
		return err
	}
	d.Component.SetStatus(core.StatusHealthy)
	return nil
}

func (d *discovery) Shutdown() {}

func (d *discovery) Announce(in *pb.Announcement) error {
	service, err := d.services.Get(in.ServiceName, in.Id)
	if err != nil && errors.GetErrType(err) != errors.NotFound {
		return err
	}
	if service == nil {
		service = new(announcement.Announcement)
	}

	service.StartUpdate()

	service.ID = in.Id
	service.ServiceName = in.ServiceName
	service.ServiceVersion = in.ServiceVersion
	service.Description = in.Description
	service.NetAddress = in.NetAddress
	service.PublicKey = in.PublicKey
	service.Certificate = in.Certificate

	return d.services.Set(service)
}

func (d *discovery) Get(serviceName string, id string) (*pb.Announcement, error) {
	service, err := d.services.Get(serviceName, id)
	if err != nil {
		return nil, err
	}
	return service.ToProto(), nil
}

func (d *discovery) GetAll(serviceName string) ([]*pb.Announcement, error) {
	services, err := d.services.ListService(serviceName)
	if err != nil {
		return nil, err
	}
	serviceCopies := make([]*pb.Announcement, 0, len(services))
	for _, service := range services {
		serviceCopies = append(serviceCopies, service.ToProto())
	}
	return serviceCopies, nil
}

func (d *discovery) AddMetadata(serviceName string, id string, in *pb.Metadata) error {
	meta := announcement.MetadataFromProto(in)
	return d.services.AddMetadata(serviceName, id, meta)
}

func (d *discovery) DeleteMetadata(serviceName string, id string, in *pb.Metadata) error {
	meta := announcement.MetadataFromProto(in)
	return d.services.RemoveMetadata(serviceName, id, meta)
}

// NewRedisDiscovery creates a new Redis-based discovery service
func NewRedisDiscovery(client *redis.Client) Discovery {
	return &discovery{
		services: announcement.NewRedisAnnouncementStore(client, "discovery"),
	}
}

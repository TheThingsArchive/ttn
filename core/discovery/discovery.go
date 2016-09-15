// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package discovery implements TTN Service Discovery.
package discovery

import (
	"bytes"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/discovery/announcement"
	"github.com/TheThingsNetwork/ttn/core/discovery/kv"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"gopkg.in/redis.v3"
)

// Discovery specifies the interface for the TTN Service Discovery component
type Discovery interface {
	core.ComponentInterface
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
	appIDs   kv.Store
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

func (d *discovery) Announce(in *pb.Announcement) error {
	existing, err := d.services.Get(in.ServiceName, in.Id)
	if errors.GetErrType(err) == errors.NotFound {
		// Not found; create new
		existing = &pb.Announcement{}
	} else if err != nil {
		return err
	}
	in.Metadata = existing.Metadata
	return d.services.Set(in)
}

func (d *discovery) Get(serviceName string, id string) (*pb.Announcement, error) {
	service, err := d.services.Get(serviceName, id)
	if err != nil {
		return nil, err
	}
	serviceCopy := *service
	return &serviceCopy, nil
}

func (d *discovery) GetAll(serviceName string) ([]*pb.Announcement, error) {
	services, err := d.services.ListService(serviceName)
	if err != nil {
		return nil, err
	}
	serviceCopies := make([]*pb.Announcement, 0, len(services))
	for _, service := range services {
		serviceCopy := *service
		serviceCopies = append(serviceCopies, &serviceCopy)
	}
	return serviceCopies, nil
}

func (d *discovery) AddMetadata(serviceName string, id string, in *pb.Metadata) error {
	existing, err := d.services.Get(serviceName, id)
	if err != nil {
		return err
	}
	// Skip if already existing
	for _, md := range existing.Metadata {
		if md.Key == in.Key && bytes.Equal(md.Value, in.Value) {
			return nil
		}
	}

	// Pre-update
	switch in.Key {
	case pb.Metadata_APP_ID:
		existingHandler, err := d.appIDs.Get(string(in.Value))
		if err == nil {
			d.DeleteMetadata("handler", existingHandler, in)
		}
	}

	// Update
	existing.Metadata = append(existing.Metadata, in)
	err = d.services.Set(existing)
	if err != nil {
		return err
	}

	// Post-update
	switch in.Key {
	case pb.Metadata_APP_ID:
		err := d.appIDs.Set(string(in.Value), id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *discovery) DeleteMetadata(serviceName string, id string, in *pb.Metadata) error {
	existing, err := d.services.Get(serviceName, id)
	if err != nil {
		return err
	}
	newMeta := make([]*pb.Metadata, 0, len(existing.Metadata))
	for _, md := range existing.Metadata {
		if md.Key == in.Key && bytes.Equal(md.Value, in.Value) {
			continue
		}
		newMeta = append(newMeta, md)
	}
	existing.Metadata = newMeta
	return d.services.Set(existing)
}

// NewDiscovery creates a new memory-based discovery service
func NewDiscovery(client *redis.Client) Discovery {
	return &discovery{
		services: announcement.NewAnnouncementStore(),
		appIDs:   kv.NewKVStore(),
	}
}

// NewRedisDiscovery creates a new Redis-based discovery service
func NewRedisDiscovery(client *redis.Client) Discovery {
	return &discovery{
		services: announcement.NewRedisAnnouncementStore(client),
		appIDs:   kv.NewRedisKVStore(client, "app-id"),
	}
}

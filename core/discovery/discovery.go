// Package discovery implements TTN Service Discovery.
package discovery

import (
	"errors"
	"fmt"
	"sync"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
)

// Discovery specifies the interface for the TTN Service Discovery component
type Discovery interface {
	Announce(announcement *pb.Announcement) error
	Discover(serviceName string) ([]*pb.Announcement, error)
}

// discovery is a reference implementation for a TTN Service Discovery component.
// TODO: Implement one with a real database
type discovery struct {
	services map[string]map[string]*pb.Announcement
	sync.RWMutex
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

func (d *discovery) Discover(serviceName string) ([]*pb.Announcement, error) {
	d.RLock()
	defer d.RUnlock()

	// Get the list
	services, ok := d.services[serviceName]
	if !ok {
		return nil, fmt.Errorf("Service %s does not exist", serviceName)
	}

	// Traverse the list
	announcements := make([]*pb.Announcement, 0, len(services))
	for _, service := range services {
		serviceCopy := *service
		serviceCopy.Token = ""
		announcements = append(announcements, &serviceCopy)
	}
	return announcements, nil
}

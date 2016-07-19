// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package announcement

import (
	"errors"
	"fmt"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"gopkg.in/redis.v3"
)

var (
	// ErrNotFound is returned when a announcement was not found
	ErrNotFound = errors.New("ttn/discovery: Announcement not found")
)

// Store is used to store announcement configurations
type Store interface {
	// List all announcements
	List() ([]*pb.Announcement, error)
	// List all announcements for a service
	ListService(serviceName string) ([]*pb.Announcement, error)
	// Get the full announcement
	Get(serviceName, serviceID string) (*pb.Announcement, error)
	// Set the announcement.
	Set(announcement *pb.Announcement) error
	// Delete a announcement
	Delete(serviceName, serviceID string) error
}

// NewAnnouncementStore creates a new in-memory Announcement store
func NewAnnouncementStore() Store {
	return &announcementStore{
		announcements: make(map[string]map[string]*pb.Announcement),
	}
}

// announcementStore is an in-memory Announcement store. It should only be used for testing
// purposes. Use the redisAnnouncementStore for actual deployments.
type announcementStore struct {
	announcements map[string]map[string]*pb.Announcement
}

func (s *announcementStore) List() ([]*pb.Announcement, error) {
	announcements := make([]*pb.Announcement, 0, len(s.announcements))
	for _, service := range s.announcements {
		for _, announcement := range service {
			announcements = append(announcements, announcement)
		}
	}
	return announcements, nil
}

func (s *announcementStore) ListService(serviceName string) ([]*pb.Announcement, error) {
	if service, ok := s.announcements[serviceName]; ok {
		announcements := make([]*pb.Announcement, 0, len(s.announcements))
		for _, announcement := range service {
			announcements = append(announcements, announcement)
		}
		return announcements, nil
	}
	return []*pb.Announcement{}, nil
}

func (s *announcementStore) Get(serviceName, serviceID string) (*pb.Announcement, error) {
	if service, ok := s.announcements[serviceName]; ok {
		if announcement, ok := service[serviceID]; ok {
			return announcement, nil
		}
	}
	return nil, ErrNotFound
}

func (s *announcementStore) Set(new *pb.Announcement) error {
	if _, ok := s.announcements[new.ServiceName]; !ok {
		s.announcements[new.ServiceName] = map[string]*pb.Announcement{}
	}
	s.announcements[new.ServiceName][new.Id] = new
	return nil
}

func (s *announcementStore) Delete(serviceName, serviceID string) error {
	if service, ok := s.announcements[serviceName]; ok {
		delete(service, serviceID)
	}
	return nil
}

// NewRedisAnnouncementStore creates a new Redis-based status store
func NewRedisAnnouncementStore(client *redis.Client) Store {
	return &redisAnnouncementStore{
		client: client,
	}
}

const redisAnnouncementPrefix = "discovery:announcement"

type redisAnnouncementStore struct {
	client *redis.Client
}

func (s *redisAnnouncementStore) getForKeys(keys []string) ([]*pb.Announcement, error) {
	pipe := s.client.Pipeline()
	defer pipe.Close()

	// Add all commands to pipeline
	cmds := make(map[string]*redis.StringStringMapCmd)
	for _, key := range keys {
		cmds[key] = s.client.HGetAllMap(key)
	}

	// Execute pipeline
	_, err := pipe.Exec()
	if err != nil {
		return nil, err
	}

	// Get all results from pipeline
	announcements := make([]*pb.Announcement, 0, len(keys))
	for _, cmd := range cmds {
		dmap, err := cmd.Result()
		if err == nil {
			announcement := &pb.Announcement{}
			err := announcement.FromStringStringMap(dmap)
			if err == nil {
				announcements = append(announcements, announcement)
			}
		}
	}

	return announcements, nil
}

func (s *redisAnnouncementStore) List() ([]*pb.Announcement, error) {
	keys, err := s.client.Keys(fmt.Sprintf("%s:*", redisAnnouncementPrefix)).Result()
	if err != nil {
		return nil, err
	}
	return s.getForKeys(keys)
}

func (s *redisAnnouncementStore) ListService(serviceName string) ([]*pb.Announcement, error) {
	keys, err := s.client.Keys(fmt.Sprintf("%s:%s:*", redisAnnouncementPrefix, serviceName)).Result()
	if err != nil {
		return nil, err
	}
	return s.getForKeys(keys)
}

func (s *redisAnnouncementStore) Get(serviceName, serviceID string) (*pb.Announcement, error) {
	res, err := s.client.HGetAllMap(fmt.Sprintf("%s:%s:%s", redisAnnouncementPrefix, serviceName, serviceID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	} else if len(res) == 0 {
		return nil, ErrNotFound
	}
	announcement := &pb.Announcement{}
	err = announcement.FromStringStringMap(res)
	if err != nil {
		return nil, err
	}
	return announcement, nil
}

func (s *redisAnnouncementStore) Set(new *pb.Announcement) error {
	key := fmt.Sprintf("%s:%s:%s", redisAnnouncementPrefix, new.ServiceName, new.Id)
	dmap, err := new.ToStringStringMap(pb.AnnouncementProperties...)
	if err != nil {
		return err
	}
	err = s.client.HMSetMap(key, dmap).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *redisAnnouncementStore) Delete(serviceName, serviceID string) error {
	key := fmt.Sprintf("%s:%s:%s", redisAnnouncementPrefix, serviceName, serviceID)
	err := s.client.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}

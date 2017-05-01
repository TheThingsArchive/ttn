// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core/handler/device/migrate"
	"github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"gopkg.in/redis.v5"
)

// Store interface for Devices
type Store interface {
	Count() (int, error)
	CountForApp(appID string) (int, error)
	List(opts *storage.ListOptions) ([]*Device, error)
	ListForApp(appID string, opts *storage.ListOptions) ([]*Device, error)
	Get(appID, devID string) (*Device, error)
	DownlinkQueue(appID, devID string) (DownlinkQueue, error)
	Set(new *Device, properties ...string) (err error)
	Delete(appID, devID string) error
	AttrWhitelist() map[string]bool
}

const defaultRedisPrefix = "handler"
const redisDevicePrefix = "device"
const redisDownlinkQueuePrefix = "downlink"
const redisUpStreamMetadataKeyWhitelist = ":attr-whitelist"

// NewRedisDeviceStore creates a new Redis-based Device store
func NewRedisDeviceStore(client *redis.Client, prefix string) *RedisDeviceStore {
	if prefix == "" {
		prefix = defaultRedisPrefix
	}
	store := storage.NewRedisMapStore(client, prefix+":"+redisDevicePrefix)
	store.SetBase(Device{}, "")
	for v, f := range migrate.DeviceMigrations(prefix) {
		store.AddMigration(v, f)
	}
	s, _ := client.Get(prefix + ":" + redisDevicePrefix + redisUpStreamMetadataKeyWhitelist).Result()
	m := map[string]bool{}
	for _, v := range strings.Split(s, ":") {
		m[v] = true
	}
	queues := storage.NewRedisQueueStore(client, prefix+":"+redisDownlinkQueuePrefix)
	return &RedisDeviceStore{
		store:         store,
		queues:        queues,
		attrWhitelist: m,
	}
}

// RedisDeviceStore stores Devices in Redis.
// - Devices are stored as a Hash
type RedisDeviceStore struct {
	store         *storage.RedisMapStore
	queues        *storage.RedisQueueStore
	attrWhitelist map[string]bool
}

// Count all devices in the store
func (s *RedisDeviceStore) Count() (int, error) {
	return s.store.Count("")
}

// CountForApp counts all devices for an Application
func (s *RedisDeviceStore) CountForApp(appID string) (int, error) {
	return s.store.Count(fmt.Sprintf("%s:*", appID))
}

// List all Devices
func (s *RedisDeviceStore) List(opts *storage.ListOptions) ([]*Device, error) {
	devicesI, err := s.store.List("", opts)
	if err != nil {
		return nil, err
	}
	devices := make([]*Device, len(devicesI))
	for i, deviceI := range devicesI {
		if device, ok := deviceI.(Device); ok {
			devices[i] = &device
		}
	}
	return devices, nil
}

// ListForApp lists all devices for a specific Application
func (s *RedisDeviceStore) ListForApp(appID string, opts *storage.ListOptions) ([]*Device, error) {
	devicesI, err := s.store.List(fmt.Sprintf("%s:*", appID), opts)
	if err != nil {
		return nil, err
	}
	devices := make([]*Device, len(devicesI))
	for i, deviceI := range devicesI {
		if device, ok := deviceI.(Device); ok {
			devices[i] = &device
		}
	}
	return devices, nil
}

// Get a specific Device
func (s *RedisDeviceStore) Get(appID, devID string) (*Device, error) {
	deviceI, err := s.store.Get(fmt.Sprintf("%s:%s", appID, devID))
	if err != nil {
		return nil, err
	}
	if device, ok := deviceI.(Device); ok {
		return &device, nil
	}
	return nil, errors.New("Database did not return a Device")
}

// DownlinkQueue for a specific Device
func (s *RedisDeviceStore) DownlinkQueue(appID, devID string) (DownlinkQueue, error) {
	return &RedisDownlinkQueue{
		appID:  appID,
		devID:  devID,
		queues: s.queues,
	}, nil
}

// Set a new Device or update an existing one
func (s *RedisDeviceStore) Set(new *Device, properties ...string) (err error) {
	now := time.Now()
	new.UpdatedAt = now
	key := fmt.Sprintf("%s:%s", new.AppID, new.DevID)
	if new.old == nil {
		new.CreatedAt = now
	}
	err = s.store.Set(key, *new, properties...)
	if err != nil {
		return
	}
	return nil
}

// Delete a Device
func (s *RedisDeviceStore) Delete(appID, devID string) error {
	key := fmt.Sprintf("%s:%s", appID, devID)
	if err := s.queues.Delete(key); err != nil {
		return err
	}
	return s.store.Delete(key)
}

// attrWhitelist return a map of the whitelisted
func (s *RedisDeviceStore) AttrWhitelist() map[string]bool {
	return s.attrWhitelist
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"gopkg.in/redis.v5"
)

// Store interface for Devices
type Store interface {
	List(opts *storage.ListOptions) ([]*Device, error)
	ListForApp(appID string, opts *storage.ListOptions) ([]*Device, error)
	Get(appID, devID string) (*Device, error)
	Set(new *Device, properties ...string) (err error)
	Delete(appID, devID string) error
}

const defaultRedisPrefix = "handler"
const redisDevicePrefix = "device"

// NewRedisDeviceStore creates a new Redis-based Device store
func NewRedisDeviceStore(client *redis.Client, prefix string) *RedisDeviceStore {
	if prefix == "" {
		prefix = defaultRedisPrefix
	}
	store := storage.NewRedisMapStore(client, prefix+":"+redisDevicePrefix)
	store.SetBase(Device{}, "")
	return &RedisDeviceStore{
		store: store,
	}
}

// RedisDeviceStore stores Devices in Redis.
// - Devices are stored as a Hash
type RedisDeviceStore struct {
	store *storage.RedisMapStore
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

// Set a new Device or update an existing one
func (s *RedisDeviceStore) Set(new *Device, properties ...string) (err error) {

	now := time.Now()
	new.UpdatedAt = now

	key := fmt.Sprintf("%s:%s", new.AppID, new.DevID)
	if new.old != nil {
		err = s.store.Update(key, *new, properties...)
	} else {
		new.CreatedAt = now
		err = s.store.Create(key, *new, properties...)
	}
	if err != nil {
		return
	}

	return nil
}

// Delete a Device
func (s *RedisDeviceStore) Delete(appID, devID string) error {
	key := fmt.Sprintf("%s:%s", appID, devID)
	return s.store.Delete(key)
}

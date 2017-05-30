// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/handler/device/migrate"
	"github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"gopkg.in/redis.v5"
)

const maxAttr uint8 = 5
const maxAttrLength = 200

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
	SetAttributesList(string)
}

const defaultRedisPrefix = "handler"
const redisDevicePrefix = "device"
const redisDownlinkQueuePrefix = "downlink"

var defaultDeviceAttributeList = map[string]bool{
	"ttn-model":   true,
	"ttn-type":    true,
	"ttn-version": true,
}

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
	queues := storage.NewRedisQueueStore(client, prefix+":"+redisDownlinkQueuePrefix)
	return &RedisDeviceStore{
		store:          store,
		queues:         queues,
		attributesKeys: defaultDeviceAttributeList,
	}
}

// RedisDeviceStore stores Devices in Redis.
// - Devices are stored as a Hash
type RedisDeviceStore struct {
	store          *storage.RedisMapStore
	queues         *storage.RedisQueueStore
	attributesKeys map[string]bool
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
	s.sortAttributes(new)
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

// SetAttributesList set the key that will always be added to the Attribute map.
func (s *RedisDeviceStore) SetAttributesList(a string) {
	l := strings.Split(a, ":")
	for _, key := range l {
		if api.ValidID(key) {
			s.attributesKeys[key] = true
		}
	}
}

// sortAttributes allow restriction on the builtin attributesKeys.
// The attribute in list will always be allowed.
// The attributesKeys already present will not be replaced/deleted unless stated otherwise.
// Then the new attributesKeys wil be added if they key is valid and if there is enough key slot available
func (s *RedisDeviceStore) sortAttributes(new *Device) {
	m, max := new.MapOldAttributes(s.attributesKeys)
	max = maxAttr - max
	if m == nil {
		m = make(map[string]string, max)
	}
	for _, v := range new.Attributes {
		_, ok := m[v.Key]
		_, okw := s.attributesKeys[v.Key]
		if v.Val == "" {
			if ok {
				delete(m, v.Key)
				max++
			}
		} else if (ok || okw || (max > 0 && api.ValidID(v.Key))) &&
			len(v.Val) < maxAttrLength {
			m[v.Key] = v.Val
			if !ok && !okw {
				max--
			}
		}
	}
	new.AttributesFromMap(m)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"gopkg.in/redis.v3"
)

// Store is used to store device configurations
type Store interface {
	// List all devices
	List() ([]*Device, error)
	// ListForApp lists all devices for an app
	ListForApp(appID string) ([]*Device, error)
	// Get the full information about a device
	Get(appID, devID string) (*Device, error)
	// Set the given fields of a device. If fields empty, it sets all fields.
	Set(device *Device, fields ...string) error
	// Delete a device
	Delete(appID, devID string) error
}

// NewDeviceStore creates a new in-memory Device store
func NewDeviceStore() Store {
	return &deviceStore{
		devices: make(map[string]map[string]*Device),
	}
}

// deviceStore is an in-memory Device store. It should only be used for testing
// purposes. Use the redisDeviceStore for actual deployments.
type deviceStore struct {
	devices map[string]map[string]*Device
}

func (s *deviceStore) List() ([]*Device, error) {
	devices := make([]*Device, 0, len(s.devices))
	for _, app := range s.devices {
		for _, device := range app {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

func (s *deviceStore) ListForApp(appID string) ([]*Device, error) {
	devices := make([]*Device, 0, len(s.devices))
	if app, ok := s.devices[appID]; ok {
		for _, device := range app {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

func (s *deviceStore) Get(appID, devID string) (*Device, error) {
	if app, ok := s.devices[appID]; ok {
		if dev, ok := app[devID]; ok {
			return dev, nil
		}
	}
	return nil, core.NewErrNotFound(fmt.Sprintf("%s/%s", appID, devID))
}

func (s *deviceStore) Set(new *Device, fields ...string) error {
	// NOTE: We don't care about fields for testing
	if app, ok := s.devices[new.AppID]; ok {
		app[new.DevID] = new
	} else {
		s.devices[new.AppID] = map[string]*Device{new.DevID: new}
	}
	return nil
}

func (s *deviceStore) Delete(appID, devID string) error {
	if app, ok := s.devices[appID]; ok {
		delete(app, devID)
	}
	return nil
}

// NewRedisDeviceStore creates a new Redis-based status store
func NewRedisDeviceStore(client *redis.Client) Store {
	return &redisDeviceStore{
		client: client,
	}
}

const redisDevicePrefix = "handler:device"

type redisDeviceStore struct {
	client *redis.Client
}

func (s *redisDeviceStore) getForKeys(keys []string) ([]*Device, error) {
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
	devices := make([]*Device, 0, len(keys))
	for _, cmd := range cmds {
		dmap, err := cmd.Result()
		if err == nil {
			device := &Device{}
			err := device.FromStringStringMap(dmap)
			if err == nil {
				devices = append(devices, device)
			}
		}
	}

	return devices, nil
}

func (s *redisDeviceStore) List() ([]*Device, error) {
	keys, err := s.client.Keys(fmt.Sprintf("%s:*", redisDevicePrefix)).Result()
	if err != nil {
		return nil, err
	}
	return s.getForKeys(keys)
}

func (s *redisDeviceStore) ListForApp(appID string) ([]*Device, error) {
	keys, err := s.client.Keys(fmt.Sprintf("%s:%s:*", redisDevicePrefix, appID)).Result()
	if err != nil {
		return nil, err
	}
	return s.getForKeys(keys)
}

func (s *redisDeviceStore) Get(appID, devID string) (*Device, error) {
	res, err := s.client.HGetAllMap(fmt.Sprintf("%s:%s:%s", redisDevicePrefix, appID, devID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, core.NewErrNotFound(fmt.Sprintf("%s/%s", appID, devID))
		}
		return nil, err
	} else if len(res) == 0 {
		return nil, core.NewErrNotFound(fmt.Sprintf("%s/%s", appID, devID))
	}
	device := &Device{}
	err = device.FromStringStringMap(res)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (s *redisDeviceStore) Set(new *Device, fields ...string) error {
	if len(fields) == 0 {
		fields = DeviceProperties
	}

	key := fmt.Sprintf("%s:%s:%s", redisDevicePrefix, new.AppID, new.DevID)
	new.UpdatedAt = time.Now()
	dmap, err := new.ToStringStringMap(fields...)
	if err != nil {
		return err
	}
	if len(dmap) == 0 {
		return nil
	}
	err = s.client.HMSetMap(key, dmap).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *redisDeviceStore) Delete(appID, devID string) error {
	key := fmt.Sprintf("%s:%s:%s", redisDevicePrefix, appID, devID)
	err := s.client.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}

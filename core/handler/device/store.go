// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"errors"
	"fmt"

	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core/types"
)

var (
	// ErrNotFound is returned when a device was not found
	ErrNotFound = errors.New("ttn/handler: Device not found")
)

// Store is used to store device configurations
type Store interface {
	// List all devices
	List() ([]*Device, error)
	// Get the full information about a device
	Get(appEUI types.AppEUI, devEUI types.DevEUI) (*Device, error)
	// Set the given fields of a device. If fields empty, it sets all fields.
	Set(device *Device, fields ...string) error
	// Delete a device
	Delete(types.AppEUI, types.DevEUI) error
}

// NewDeviceStore creates a new in-memory Device store
func NewDeviceStore() Store {
	return &deviceStore{
		devices: make(map[types.AppEUI]map[types.DevEUI]*Device),
	}
}

// deviceStore is an in-memory Device store. It should only be used for testing
// purposes. Use the redisDeviceStore for actual deployments.
type deviceStore struct {
	devices map[types.AppEUI]map[types.DevEUI]*Device
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

func (s *deviceStore) Get(appEUI types.AppEUI, devEUI types.DevEUI) (*Device, error) {
	if app, ok := s.devices[appEUI]; ok {
		if dev, ok := app[devEUI]; ok {
			return dev, nil
		}
	}
	return nil, ErrNotFound
}

func (s *deviceStore) Set(new *Device, fields ...string) error {
	// NOTE: We don't care about fields for testing
	if app, ok := s.devices[new.AppEUI]; ok {
		app[new.DevEUI] = new
	} else {
		s.devices[new.AppEUI] = map[types.DevEUI]*Device{new.DevEUI: new}
	}
	return nil
}

func (s *deviceStore) Delete(appEUI types.AppEUI, devEUI types.DevEUI) error {
	if app, ok := s.devices[appEUI]; ok {
		delete(app, devEUI)
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

func (s *redisDeviceStore) List() ([]*Device, error) {
	var devices []*Device
	keys, err := s.client.Keys(fmt.Sprintf("%s:*", redisDevicePrefix)).Result()
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		res, err := s.client.HGetAllMap(key).Result()
		if err != nil {
			return nil, err
		}
		device := &Device{}
		err = device.FromStringStringMap(res)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func (s *redisDeviceStore) Get(appEUI types.AppEUI, devEUI types.DevEUI) (*Device, error) {
	res, err := s.client.HGetAllMap(fmt.Sprintf("%s:%s:%s", redisDevicePrefix, appEUI, devEUI)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	} else if len(res) == 0 {
		return nil, ErrNotFound
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

	key := fmt.Sprintf("%s:%s:%s", redisDevicePrefix, new.AppEUI, new.DevEUI)
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

func (s *redisDeviceStore) Delete(appEUI types.AppEUI, devEUI types.DevEUI) error {
	key := fmt.Sprintf("%s:%s:%s", redisDevicePrefix, appEUI, devEUI)
	err := s.client.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}

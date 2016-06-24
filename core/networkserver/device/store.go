// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"errors"
	"fmt"
	"sync"

	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core/types"
)

var (
	// ErrNotFound is returned when a device was not found
	ErrNotFound = errors.New("ttn/networkserver: Device not found")
)

// Store is used to store device configurations
type Store interface {
	// List all devices
	List() ([]*Device, error)
	// Get the full information about a device
	Get(appEUI types.AppEUI, devEUI types.DevEUI) (*Device, error)
	// Get a list of devices matching the DevAddr
	GetWithAddress(devAddr types.DevAddr) ([]*Device, error)
	// Set the given fields of a device. If fields empty, it sets all fields.
	Set(device *Device, fields ...string) error
	// Activate a device
	Activate(types.AppEUI, types.DevEUI, types.DevAddr, types.NwkSKey) error
	// Delete a device
	Delete(types.AppEUI, types.DevEUI) error
}

// NewDeviceStore creates a new in-memory Device store
func NewDeviceStore() Store {
	return &deviceStore{
		devices:   make(map[types.AppEUI]map[types.DevEUI]*Device),
		byAddress: make(map[types.DevAddr][]*Device),
	}
}

// deviceStore is an in-memory Device store. It should only be used for testing
// purposes. Use the redisDeviceStore for actual deployments.
type deviceStore struct {
	devices   map[types.AppEUI]map[types.DevEUI]*Device
	byAddress map[types.DevAddr][]*Device
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

func (s *deviceStore) GetWithAddress(devAddr types.DevAddr) ([]*Device, error) {
	if devices, ok := s.byAddress[devAddr]; ok {
		return devices, nil
	}
	return []*Device{}, nil
}

func (s *deviceStore) Set(new *Device, fields ...string) error {
	// NOTE: We don't care about fields for testing
	if app, ok := s.devices[new.AppEUI]; ok {
		if old, ok := app[new.DevEUI]; ok {
			// DevAddr Updated
			if new.DevAddr != old.DevAddr && !old.DevAddr.IsEmpty() {
				// Remove the old DevAddr
				newList := make([]*Device, 0, len(s.byAddress[old.DevAddr]))
				for _, candidate := range s.byAddress[old.DevAddr] {
					if candidate.DevEUI != old.DevEUI || candidate.AppEUI != old.AppEUI {
						newList = append(newList, candidate)
					}
				}
				s.byAddress[old.DevAddr] = newList
			}
		}
		app[new.DevEUI] = new
	} else {
		s.devices[new.AppEUI] = map[types.DevEUI]*Device{new.DevEUI: new}
	}

	if !new.DevAddr.IsEmpty() && !new.NwkSKey.IsEmpty() {
		if devices, ok := s.byAddress[new.DevAddr]; ok {
			var exists bool
			for _, candidate := range devices {
				if candidate.AppEUI == new.AppEUI && candidate.DevEUI == new.DevEUI {
					exists = true
					break
				}
			}
			if !exists {
				s.byAddress[new.DevAddr] = append(devices, new)
			}
		} else {
			s.byAddress[new.DevAddr] = []*Device{new}
		}
	}

	return nil
}

func (s *deviceStore) Activate(appEUI types.AppEUI, devEUI types.DevEUI, devAddr types.DevAddr, nwkSKey types.NwkSKey) error {
	dev, err := s.Get(appEUI, devEUI)
	if err != nil {
		return err
	}
	dev.DevAddr = devAddr
	dev.NwkSKey = nwkSKey
	dev.FCntUp = 0
	dev.FCntDown = 0
	return s.Set(dev)
}

func (s *deviceStore) Delete(appEUI types.AppEUI, devEUI types.DevEUI) error {
	if app, ok := s.devices[appEUI]; ok {
		if old, ok := app[devEUI]; ok {
			delete(app, devEUI)
			newList := make([]*Device, 0, len(s.byAddress[old.DevAddr]))
			for _, candidate := range s.byAddress[old.DevAddr] {
				if candidate.DevEUI != old.DevEUI || candidate.AppEUI != old.AppEUI {
					newList = append(newList, candidate)
				}
			}
			s.byAddress[old.DevAddr] = newList
		}
	}

	return nil
}

// NewRedisDeviceStore creates a new Redis-based status store
func NewRedisDeviceStore(client *redis.Client) Store {
	return &redisDeviceStore{
		client: client,
	}
}

const redisDevicePrefix = "ns:device"
const redisDevAddrPrefix = "ns:dev_addr"

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

func (s *redisDeviceStore) GetWithAddress(devAddr types.DevAddr) ([]*Device, error) {
	keys, err := s.client.SMembers(fmt.Sprintf("%s:%s", redisDevAddrPrefix, devAddr)).Result()
	if err != nil {
		return nil, err
	}

	// TODO: If someone finds a nice way to do this more efficiently, please submit a PR!

	var wg sync.WaitGroup
	responses := make(chan *Device)
	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			dmap, err := s.client.HGetAllMap(key).Result()
			if err == nil {
				device := &Device{}
				err := device.FromStringStringMap(dmap)
				if err == nil {
					responses <- device
				}
			}
			wg.Done()
		}(key)
	}

	go func() {
		wg.Wait()
		close(responses)
	}()

	devices := make([]*Device, 0, len(keys))
	for res := range responses {
		devices = append(devices, res)
	}

	return devices, nil
}

func (s *redisDeviceStore) Set(new *Device, fields ...string) error {
	if len(fields) == 0 {
		fields = DeviceProperties
	}

	key := fmt.Sprintf("%s:%s:%s", redisDevicePrefix, new.AppEUI, new.DevEUI)

	// Check for old DevAddr
	if devAddr, err := s.client.HGet(key, "dev_addr").Result(); err == nil {
		// Delete old DevAddr
		if devAddr != "" {
			err := s.client.SRem(fmt.Sprintf("%s:%s", redisDevAddrPrefix, devAddr), key).Err()
			if err != nil {
				return err
			}
		}
	}

	dmap, err := new.ToStringStringMap(fields...)
	if err != nil {
		return err
	}
	s.client.HMSetMap(key, dmap)

	if !new.DevAddr.IsEmpty() && !new.NwkSKey.IsEmpty() {
		err := s.client.SAdd(fmt.Sprintf("%s:%s", redisDevAddrPrefix, new.DevAddr), key).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *redisDeviceStore) Activate(appEUI types.AppEUI, devEUI types.DevEUI, devAddr types.DevAddr, nwkSKey types.NwkSKey) error {
	key := fmt.Sprintf("%s:%s:%s", redisDevicePrefix, appEUI, devEUI)

	// Find existing device
	exists, err := s.client.Exists(key).Result()
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	// Check for old DevAddr
	if devAddr, err := s.client.HGet(key, "dev_addr").Result(); err == nil {
		// Delete old DevAddr
		if devAddr != "" {
			err := s.client.SRem(fmt.Sprintf("%s:%s", redisDevAddrPrefix, devAddr), key).Err()
			if err != nil {
				return err
			}
		}
	}

	// Update Device
	dev := &Device{
		DevAddr:  devAddr,
		NwkSKey:  nwkSKey,
		FCntUp:   0,
		FCntDown: 0,
	}

	// Don't touch Utilization and Options
	dmap, err := dev.ToStringStringMap("dev_addr", "nwk_s_key", "f_cnt_up", "f_cnt_down")

	// Register Device
	err = s.client.HMSetMap(key, dmap).Err()
	if err != nil {
		return err
	}

	// Register DevAddr
	err = s.client.SAdd(fmt.Sprintf("%s:%s", redisDevAddrPrefix, devAddr), key).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *redisDeviceStore) Delete(appEUI types.AppEUI, devEUI types.DevEUI) error {
	key := fmt.Sprintf("%s:%s:%s", redisDevicePrefix, appEUI, devEUI)
	if devAddr, err := s.client.HGet(key, "dev_addr").Result(); err == nil {
		// Delete old DevAddr
		if devAddr != "" {
			err := s.client.SRem(fmt.Sprintf("%s:%s", redisDevAddrPrefix, devAddr), key).Err()
			if err != nil {
				return err
			}
		}
	}
	err := s.client.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}

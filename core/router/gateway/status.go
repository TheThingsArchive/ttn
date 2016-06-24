// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"sync"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	"gopkg.in/redis.v3"
)

// StatusStore is a database for setting and retrieving the latest gateway status
type StatusStore interface {
	// Insert or Update the status
	Update(status *pb_gateway.Status) error
	// Get the last status
	Get() (*pb_gateway.Status, error)
}

// NewStatusStore creates a new in-memory status store
func NewStatusStore() StatusStore {
	return &statusStore{}
}

type statusStore struct {
	sync.RWMutex
	lastStatus *pb_gateway.Status
}

func (s *statusStore) Update(status *pb_gateway.Status) error {
	s.Lock()
	defer s.Unlock()
	s.lastStatus = status
	return nil
}

func (s *statusStore) Get() (*pb_gateway.Status, error) {
	s.RLock()
	defer s.RUnlock()
	if s.lastStatus != nil {
		return s.lastStatus, nil
	}
	return &pb_gateway.Status{}, nil
}

// NewRedisStatusStore creates a new Redis-based status store
func NewRedisStatusStore(client *redis.Client, eui types.GatewayEUI) StatusStore {
	return &redisStatusStore{
		client: client,
		key:    fmt.Sprintf("router:gateway:%s", eui),
	}
}

type redisStatusStore struct {
	client *redis.Client
	key    string
}

func (s *redisStatusStore) Update(status *pb_gateway.Status) error {
	m, err := status.ToStringStringMap(pb_gateway.StatusMessageProperties...)
	if err != nil {
		return err
	}
	return s.client.HMSetMap(s.key, m).Err()
}

func (s *redisStatusStore) Get() (*pb_gateway.Status, error) {
	status := &pb_gateway.Status{}
	res, err := s.client.HGetAllMap(s.key).Result()
	if err != nil {
		return status, nil
	}
	err = status.FromStringStringMap(res)
	if err != nil {
		return nil, err
	}
	return status, nil
}

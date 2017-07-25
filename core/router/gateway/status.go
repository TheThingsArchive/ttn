// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"sync"

	pb_gateway "github.com/TheThingsNetwork/api/gateway"
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

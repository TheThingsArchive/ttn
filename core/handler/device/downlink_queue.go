// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"encoding/json"
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// DownlinkQueue stores the Downlink queue
type DownlinkQueue interface {
	Length() (int, error)
	Next() (*types.DownlinkMessage, error)
	Replace(msg *types.DownlinkMessage) error
	PushFirst(msg *types.DownlinkMessage) error
	PushLast(msg *types.DownlinkMessage) error
}

// RedisDownlinkQueue implements the downlink queue in Redis
type RedisDownlinkQueue struct {
	appID  string
	devID  string
	queues *storage.RedisQueueStore
}

func (s *RedisDownlinkQueue) key() string {
	return fmt.Sprintf("%s:%s", s.appID, s.devID)
}

// Length of the downlink queue
func (s *RedisDownlinkQueue) Length() (int, error) {
	return s.queues.Length(s.key())
}

// Next item in the downlink queue
func (s *RedisDownlinkQueue) Next() (*types.DownlinkMessage, error) {
	qd, err := s.queues.Next(s.key())
	if err != nil {
		return nil, err
	}
	if qd == "" {
		return nil, nil
	}
	msg := new(types.DownlinkMessage)
	if err := json.Unmarshal([]byte(qd), msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// Replace the downlink queue with msg
func (s *RedisDownlinkQueue) Replace(msg *types.DownlinkMessage) error {
	if err := s.queues.Delete(s.key()); err != nil {
		return err
	}
	return s.PushFirst(msg)
}

// PushFirst message to the downlink queue
func (s *RedisDownlinkQueue) PushFirst(msg *types.DownlinkMessage) error {
	qd, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return s.queues.AddFront(s.key(), string(qd))
}

// PushLast message to the downlink queue
func (s *RedisDownlinkQueue) PushLast(msg *types.DownlinkMessage) error {
	qd, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return s.queues.AddEnd(s.key(), string(qd))
}

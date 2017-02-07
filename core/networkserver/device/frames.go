// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"encoding/json"
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// FrameHistory for a device
type FrameHistory interface {
	Push(frame *Frame) error
	Get() ([]*Frame, error)
	Clear() error
}

// RedisFrameHistory implements the frame history in Redis
type RedisFrameHistory struct {
	appEUI types.AppEUI
	devEUI types.DevEUI
	store  *storage.RedisQueueStore
}

// FramesHistorySize for ADR
const FramesHistorySize = 20

// Frame collected for ADR
type Frame struct {
	FCnt         uint32  `json:"f_cnt"`
	SNR          float32 `json:"snr"`
	GatewayCount uint32  `json:"gw_cnt"`
}

func (s *RedisFrameHistory) key() string {
	return fmt.Sprintf("%s:%s", s.appEUI, s.devEUI)
}

// Push a Frame to the device's history
func (s *RedisFrameHistory) Push(frame *Frame) error {
	frameBytes, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	if err := s.store.AddFront(s.key(), string(frameBytes)); err != nil {
		return err
	}
	return s.Trim()
}

// Get the last frames from the device's history
func (s *RedisFrameHistory) Get() (out []*Frame, err error) {
	frames, err := s.store.GetFront(s.key(), FramesHistorySize)
	for _, frameStr := range frames {
		frame := new(Frame)
		if err := json.Unmarshal([]byte(frameStr), frame); err != nil {
			return nil, err
		}
		out = append(out, frame)
	}
	return
}

// Trim frames in the device's history
func (s *RedisFrameHistory) Trim() error {
	return s.store.Trim(s.key(), FramesHistorySize)
}

// Clear frames in the device's history
func (s *RedisFrameHistory) Clear() error {
	return s.store.Delete(s.key())
}

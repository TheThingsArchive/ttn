// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"sync"
	"time"

	"github.com/TheThingsNetwork/go-utils/backoff"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
)

type deviceIdentifier struct {
	appID string
	devID string
}

type activeDownlinks struct {
	mu      sync.Mutex
	devices map[deviceIdentifier]int
}

func (h *handler) activateDownlink(dev *device.Device) {
	identifier := deviceIdentifier{dev.AppID, dev.DevID}
	h.activeDownlinks.mu.Lock()
	pending, ok := h.activeDownlinks.devices[identifier]
	h.activeDownlinks.devices[identifier] = pending + 1
	h.activeDownlinks.mu.Unlock()
	if ok {
		return
	}
	go func() {
		attempts := 0
		for {
			downlink, err := h.ttnBroker.PrepareDownlink(h.GetContext(""), &pb_broker.PrepareDownlinkRequest{
				AppId:  dev.AppID,
				DevId:  dev.DevID,
				AppEui: &dev.AppEUI,
				DevEui: &dev.DevEUI,
			})
			attempts++
			if err == nil {
				err = h.tryDownlink(dev.AppID, dev.DevID, downlink)
				if err == nil {
					attempts = 0
					h.activeDownlinks.mu.Lock()
					pending, _ := h.activeDownlinks.devices[identifier]
					if pending <= 1 {
						delete(h.activeDownlinks.devices, identifier)
						h.activeDownlinks.mu.Unlock()
						return
					}
					h.activeDownlinks.devices[identifier] = pending - 1
					h.activeDownlinks.mu.Unlock()
				}
			}
			time.Sleep(backoff.Backoff(attempts))
		}
	}()
}

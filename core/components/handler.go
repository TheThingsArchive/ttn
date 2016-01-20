// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/apex/log"
	//"github.com/brocaar/lorawan"
)

const BUFFER_DELAY = time.Millisecond * 300

var ErrNotImplemented = fmt.Errorf("Ilegal call on non implemented method")

type Handler struct {
	ctx log.Interface
	db  handlerStorage
	set chan<- uplinkBundle
}

type bundleId [20]byte

type uplinkBundle struct {
	id     bundleId
	an     core.AckNacker
	packet core.Packet
}

func NewHandler(db handlerStorage, ctx log.Interface) (*Handler, error) {
	h := Handler{
		ctx: ctx,
		db:  db,
	}

	bundles := make(chan []uplinkBundle)
	set := make(chan uplinkBundle)

	go h.consumeBundles(bundles)
	go h.manageBuffers(bundles, set)
	h.set = set

	return &h, nil
}

func (h *Handler) Register(reg core.Registration, an core.AckNacker) error {
	return nil
}

func (h *Handler) HandleUp(p core.Packet, an core.AckNacker, upAdapter core.Adapter) error {
	return nil
}

func (h *Handler) HandleDown(p core.Packet, an core.AckNacker, downAdapter core.Adapter) error {
	return ErrNotImplemented
}

func (h *Handler) consumeBundles(bundles <-chan []uplinkBundle) {
	//for bundle := range bundles {
	// Deduplicate
	// DecryptPayload
	// AddMeta
	// AckOrNack each packets
	// Store into mongo
	//}
}

// manageBuffers gather new incoming bundles that possess the same id
// It then flushs them once a given delay has passed since the reception of the first bundle.
func (h *Handler) manageBuffers(bundles chan<- []uplinkBundle, set <-chan uplinkBundle) {
	buffers := make(map[bundleId][]uplinkBundle)
	alarm := make(chan bundleId)

	for {
		select {
		case id := <-alarm:
			b := buffers[id]
			delete(buffers, id)
			go func(b []uplinkBundle) { bundles <- b }(b)
		case bundle := <-set:
			b := append(buffers[bundle.id], bundle)
			if len(b) == 1 {
				go setAlarm(alarm, bundle.id, time.Millisecond*300)
			}
			buffers[bundle.id] = b
		}
	}
}

// setAlarm will trigger a message on the given channel after a given delay.
func setAlarm(alarm chan<- bundleId, id bundleId, delay time.Duration) {
	<-time.After(delay)
	alarm <- id
}

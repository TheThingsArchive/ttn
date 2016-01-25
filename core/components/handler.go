// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/apex/log"
)

const BUFFER_DELAY = time.Millisecond * 300

var ErrNotImplemented = fmt.Errorf("Ilegal call on non implemented method")

type Handler struct {
	ctx log.Interface
	db  handlerStorage
	set chan<- uplinkBundle
}

type bundleId [22]byte // AppEUI | DevAddr | FCnt

type uplinkBundle struct {
	id      bundleId
	entry   handlerEntry
	packet  core.Packet
	adapter core.Adapter
	chresp  chan interface{} // Error or decrypted packet
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
	partition, err := h.db.partition([]core.Packet{p})
	if err != nil {
		an.Nack()
		return err
	}

	fcnt, err := p.Fcnt()
	if err != nil {
		an.Nack()
		return err
	}

	chresp := make(chan interface{})
	var id bundleId
	buf := new(bytes.Buffer)
	buf.Write(partition[0].id[:]) // Partition is necessarily of length 1, associated to 1 packet, the same we gave
	binary.Write(buf, binary.BigEndian, fcnt)
	copy(id[:], buf.Bytes())
	h.set <- uplinkBundle{
		id:      id,
		packet:  p,
		entry:   partition[0].handlerEntry,
		adapter: upAdapter,
		chresp:  chresp,
	}

	resp := <-chresp
	switch resp.(type) {
	case core.Packet:
		an.Ack(resp.(core.Packet))
		return nil
	case error:
		an.Nack()
		return resp.(error)
	default:
		an.Ack()
		return nil
	}
}

func (h *Handler) HandleDown(p core.Packet, an core.AckNacker, downAdapter core.Adapter) error {
	return ErrNotImplemented
}

func (h *Handler) consumeBundles(chbundles <-chan []uplinkBundle) {
	for bundles := range chbundles {
		var packet *core.Packet
		var sendToAdapter func(packet core.Packet) error
		for _, bundle := range bundles {
			if packet == nil {
				*packet = core.Packet{
					Payload: bundle.packet.Payload,
					Metadata: core.Metadata{
						Group: []core.Metadata{bundle.packet.Metadata},
					},
				}
				// The handler assumes payload encrypted with AppSKey only !
				if err := packet.Payload.DecryptMACPayload(bundle.entry.AppSKey); err != nil {
					for _, bundle := range bundles {
						bundle.chresp <- err
					}
					break
				}

				sendToAdapter = func(packet core.Packet) error {
					// NOTE We'll have to look here for the downlink !
					_, err := bundle.adapter.Send(packet, core.Recipient{
						Address: bundle.entry.DevAddr,
						Id:      bundle.entry.AppEUI,
					})
					return err
				}
				continue
			}
			packet.Metadata.Group = append(packet.Metadata.Group, bundle.packet.Metadata)
		}

		err := sendToAdapter(*packet)
		for _, bundle := range bundles {
			bundle.chresp <- err
		}
	}
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

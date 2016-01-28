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
	"github.com/brocaar/lorawan"
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
	h.ctx.WithField("registration", reg).Debug("New registration request")
	options, okOpts := reg.Options.(struct {
		AppSKey lorawan.AES128Key
		NwkSKey lorawan.AES128Key
	})
	appEUI, okId := reg.Recipient.Id.(lorawan.EUI64)

	if !okId || !okOpts {
		an.Nack()
		return ErrBadOptions
	}

	err := h.db.store(reg.DevAddr, handlerEntry{
		AppEUI:  appEUI,
		AppSKey: options.AppSKey,
		NwkSKey: options.NwkSKey,
		DevAddr: reg.DevAddr,
	})

	if err != nil {
		an.Nack()
		return err
	}

	an.Ack()
	return nil
}

func (h *Handler) HandleUp(p core.Packet, an core.AckNacker, upAdapter core.Adapter) error {
	h.ctx.Debug("Handling new uplink packet")
	partition, err := h.db.partition([]core.Packet{p})
	if err != nil {
		h.ctx.WithError(err).Debug("Unable to find entry")
		an.Nack()
		return err
	}

	fcnt, err := p.Fcnt()
	if err != nil {
		h.ctx.WithError(err).Debug("Unable to retrieve fcnt")
		an.Nack()
		return err
	}

	chresp := make(chan interface{})
	var id bundleId
	buf := new(bytes.Buffer)
	buf.Write(partition[0].id[:]) // Partition is necessarily of length 1, associated to 1 packet, the same we gave
	binary.Write(buf, binary.BigEndian, fcnt)
	copy(id[:], buf.Bytes())
	h.ctx.WithField("bundleId", id).Debug("Defining new bundle")
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
		h.ctx.WithField("bundleId", id).Debug("Received response with packet. Sending ack")
		an.Ack(resp.(core.Packet))
		return nil
	case error:
		h.ctx.WithField("bundleId", id).WithError(resp.(error)).Debug("Received response. Sending Nack")
		an.Nack()
		return resp.(error)
	default:
		h.ctx.WithField("bundleId", id).Debug("Received response. Sending ack")
		an.Ack()
		return nil
	}
}

func (h *Handler) HandleDown(p core.Packet, an core.AckNacker, downAdapter core.Adapter) error {
	return ErrNotImplemented
}

func (h *Handler) consumeBundles(chbundles <-chan []uplinkBundle) {
	ctx := h.ctx.WithField("goroutine", "consumer")
	ctx.Debug("Starting bundle consumer")
browseBundles:
	for bundles := range chbundles {
		var packet *core.Packet
		var sendToAdapter func(packet core.Packet) error
		ctx.WithField("nb", len(bundles)).Debug("Consuming new bundles set")
		for _, bundle := range bundles {
			if packet == nil {
				ctx.WithField("entry", bundle.entry).Debug("Preparing ground for given entry")
				packet = new(core.Packet)
				*packet = core.Packet{
					Payload: bundle.packet.Payload,
					Metadata: core.Metadata{
						Group: []core.Metadata{bundle.packet.Metadata},
					},
				}
				// The handler assumes payloads encrypted with AppSKey only !
				payload, ok := packet.Payload.MACPayload.(*lorawan.MACPayload)
				if !ok {
					ctx.WithError(ErrInvalidPacket).Debug("Unable to extract MACPayload")
					for _, bundle := range bundles {
						bundle.chresp <- ErrInvalidPacket
					}
					continue browseBundles
				}

				if err := payload.DecryptFRMPayload(bundle.entry.AppSKey); err != nil {
					ctx.WithError(err).Debug("Unable to decrypt MAC Payload with given AppSKey")
					for _, bundle := range bundles {
						bundle.chresp <- err
					}
					continue browseBundles
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
		ctx.WithField("error", err).Debug("Sending to bundle adapter")
		for _, bundle := range bundles {
			bundle.chresp <- err
		}
	}
}

// manageBuffers gather new incoming bundles that possess the same id
// It then flushs them once a given delay has passed since the reception of the first bundle.
func (h *Handler) manageBuffers(bundles chan<- []uplinkBundle, set <-chan uplinkBundle) {
	ctx := h.ctx.WithField("goroutine", "bufferer")
	ctx.Debug("Starting uplink packets buffering")

	buffers := make(map[bundleId][]uplinkBundle)
	alarm := make(chan bundleId)

	for {
		select {
		case id := <-alarm:
			b := buffers[id]
			delete(buffers, id)
			go func(b []uplinkBundle) { bundles <- b }(b)
			ctx.WithField("bundleId", id).Debug("Alarm done. Consuming collected bundles")
		case bundle := <-set:
			b := append(buffers[bundle.id], bundle)
			if len(b) == 1 {
				go setAlarm(alarm, bundle.id, time.Millisecond*300)
				ctx.WithField("bundleId", bundle.id).Debug("Starting buffering. New alarm set")
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

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"reflect"
	"time"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	"github.com/apex/log"
)

const buffer_delay time.Duration = time.Millisecond * 300
const max_duty_cycle = 90 // 90%

// component implements the core.Component interface
type component struct {
	ctx     log.Interface
	devices DevStorage
	packets PktStorage
	set     chan<- bundle
}

type bundle struct {
	Adapter Adapter
	Chresp  chan interface{}
	Entry   devEntry
	Id      [20]byte
	Packet  HPacket
}

// New construct a new Handler
func New(devDb DevStorage, pktDb PktStorage, ctx log.Interface) Component {
	h := component{
		ctx:     ctx,
		devices: devDb,
		packets: pktDb,
	}

	set := make(chan bundle)
	bundles := make(chan []bundle)

	h.set = set
	go h.consumeBundles(bundles)
	go h.consumeSet(bundles, set)

	return h
}

// Register implements the core.Component interface
func (h component) Register(reg Registration, an AckNacker) (err error) {
	h.ctx.WithField("registration", reg).Debug("New registration request")
	defer ensureAckNack(an, nil, &err)

	hreg, ok := reg.(HRegistration)
	if !ok {
		return errors.New(errors.Structural, "Not a Handler registration")
	}

	if err = h.devices.Store(hreg); err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

// HandleUp implements the core.Component interface
func (h component) HandleUp(data []byte, an AckNacker, up Adapter) (err error) {
	// Make sure we don't forget the AckNacker
	var ack Packet
	defer ensureAckNack(an, &ack, &err)

	itf, err := UnmarshalPacket(data)
	if err != nil {
		return errors.New(errors.Structural, data)
	}

	switch itf.(type) {
	case HPacket:
		// 0. Retrieve the handler packet
		packet := itf.(HPacket)
		appEUI := packet.AppEUI()
		devEUI := packet.DevEUI()

		// 1. Lookup for the associated AppSKey + Recipient
		entry, err := h.devices.Lookup(appEUI, devEUI)
		if err != nil {
			return errors.New(errors.Operational, err)
		}

		// 2. Prepare a channel to receive the response from the consumer
		chresp := make(chan interface{})

		// 3. Create a "bundle" which holds info waiting for other related packets
		var bundleId [20]byte // AppEUI(8) | DevEUI(8)
		rw := readwriter.New(nil)
		rw.Write(appEUI)
		rw.Write(devEUI)
		rw.Write(packet.FCnt())
		data, err := rw.Bytes()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		copy(bundleId[:], data[:])

		// 4. Send the actual bundle to the consumer
		ctx := h.ctx.WithField("BundleID", bundleId)
		ctx.Debug("Define new bundle")
		h.set <- bundle{
			Id:      bundleId,
			Packet:  packet,
			Entry:   entry,
			Adapter: up,
			Chresp:  chresp,
		}

		// 5. Wait for the response. Could be an error, a packet or nothing.
		// We'll respond to a maximum of one node. The handler will use the
		// rssi + gateway's duty cycle to select to best fit.
		// All other channels will get a nil response.
		// If there's an error, all channels get the error.
		resp := <-chresp
		switch resp.(type) {
		case Packet:
			ctx.Debug("Received response with packet. Sending Ack")
			an.Ack(resp.(Packet))
		case error:
			ctx.WithError(resp.(error)).Warn("Received errored response. Sending Ack")
			an.Nack()
			return errors.New(errors.Operational, resp.(error))
		default:
			ctx.Debug("Received empty response. Sending empty Ack")
			an.Ack(nil)
		}

		return nil
	case JPacket:
		return errors.New(errors.Implementation, "Join Request not yet implemented")
	default:
		return errors.New(errors.Implementation, "Unhandled packet type")
	}
}

func computeScore(dutyCycle uint, rssi int) uint {
	if dutyCycle > max_duty_cycle {
		return 0
	}

	if dutyCycle > 2*max_duty_cycle/3 {
		return uint(1000 - rssi)
	}

	return uint(10000 - rssi)
}

// consumeBundles processes list of bundle generated overtime, decrypt the underlying packet,
// deduplicate them, and send a single enhanced packet to the upadapter for further processing.
func (h component) consumeBundles(chbundle <-chan []bundle) {
	ctx := h.ctx.WithField("goroutine", "bundle consumer")
	ctx.Debug("Starting bundle consumer")

browseBundles:
	for bundles := range chbundle {
		var metadata []Metadata
		var payload []byte
		var bestBundle bundle
		var bestScore uint

		for i, bundle := range bundles {
			// We only decrypt the payload of the first bundle's packet.
			// We assume all the other to be equal and we'll merely collect
			// metadata from other bundle.
			if i == 0 {
				var err error
				payload, err = bundle.Packet.Payload(bundle.Entry.AppSKey)
				if err != nil {
					go h.abortConsume(err, bundles)
					continue browseBundles
				}
				bestBundle = bundle
			}

			// Append metadata for each of them
			metadata = append(metadata, bundle.Packet.Metadata())

			// And try to find the best recipient to which answer
			duty := bundle.Packet.Metadata().Duty
			rssi := bundle.Packet.Metadata().Rssi
			if duty == nil || rssi == nil {
				continue
			}
			score := computeScore(*duty, *rssi)
			if score > bestScore {
				bestScore = score
				bestBundle = bundle
			}
		}

		// Then create an application-level packet
		packet, err := NewAPacket(payload, bestBundle.Packet.DevEUI(), metadata)
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		// And send it to the wild open
		// we don't expect a response from the adapter, end of the chain.
		recipient, err := bestBundle.Adapter.GetRecipient(bestBundle.Entry.Recipient)
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		_, err = bestBundle.Adapter.Send(packet, recipient)
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		// Now handle the downlink
		down, err := h.packets.Pull(bestBundle.Packet.AppEUI(), bestBundle.Packet.DevEUI())
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		// Then respond to node -> no response for the moment
		for _, bundle := range bundles {
			if bundle.Id == bestBundle.Id {
				bundle.Chresp <- down
			} else {
				bundle.Chresp <- nil
			}
		}
	}
}

// Abort consume forward the given error to all bundle recipients
func (h component) abortConsume(fault error, bundles []bundle) {
	err := errors.New(errors.Structural, fault)
	h.ctx.WithError(err).Debug("Unable to consume bundle")
	for _, bundle := range bundles {
		bundle.Chresp <- err
	}
}

// consumeSet gathers new incoming bundles which possess the same id (i.e. appEUI & devEUI & Fcnt)
// It then flushes them once a given delay has passed since the reception of the first bundle.
func (h component) consumeSet(chbundles chan<- []bundle, chset <-chan bundle) {
	ctx := h.ctx.WithField("goroutine", "set consumer")
	ctx.Debug("Starting packets buffering")

	// NOTE Processed is likely to grow quickly. One has to define a more efficient data stucture
	// with a ttl for each entry. Processed is merely there to avoid late packets from being
	// processed again. The TTL could be only of several seconds or minutes.
	processed := make(map[[16]byte][]byte) // AppEUI | DevEUI | FCnt -> hasBeenProcessed ?
	buffers := make(map[[20]byte][]bundle) // AppEUI | DevEUI | FCnt ->  buffered bundles
	alarm := make(chan [20]byte)           // Communication channel with subsequent alarms

	for {
		select {
		case id := <-alarm:
			// Get all bundles
			bundles := buffers[id]
			delete(buffers, id)

			// Register the last processed entry
			var pid [16]byte
			copy(pid[:], id[:16])
			processed[pid] = id[16:]

			// Actually send the bundle to the be processed
			go func(bundles []bundle) { chbundles <- bundles }(bundles)
			ctx.WithField("BundleID", id).Debug("Consuming collected bundles")
		case b := <-chset:
			ctx = ctx.WithField("BundleID", b.Id)

			// Check if bundle has already been processed
			var pid [16]byte
			copy(pid[:], b.Id[:16])
			if reflect.DeepEqual(processed[pid], b.Id[16:]) {
				ctx.Debug("Reject already processed bundle")
				go func(b bundle) {
					b.Chresp <- errors.New(errors.Behavioural, "Already processed")
				}(b)
				continue
			}

			// Add the bundle to the stack, and set the alarm if its the first
			bundles := append(buffers[b.Id], b)
			if len(bundles) == 1 {
				go setAlarm(alarm, b.Id, buffer_delay)
				ctx.Debug("Buffering started -> new alarm set")
			}
			buffers[b.Id] = bundles
		}
	}
}

// setAlarm will trigger a message on the given channel after the given delay
func setAlarm(alarm chan<- [20]byte, id [20]byte, delay time.Duration) {
	<-time.After(delay)
	alarm <- id
}

// HandleDown implements the core.Component interface
func (h component) HandleDown(data []byte, an AckNacker, down Adapter) (err error) {
	// Make sure we don't forget the AckNacker
	var ack Packet
	defer ensureAckNack(an, &ack, &err)

	// Unmarshal the given packet and see what gift we get
	itf, err := UnmarshalPacket(data)
	if err != nil {
		return errors.New(errors.Structural, data)
	}
	return nil

	switch itf.(type) {
	case HPacket:
		return h.packets.Push(itf.(HPacket))
	default:
		return errors.New(errors.Implementation, "Unhandled packet type")
	}
}

func ensureAckNack(an AckNacker, ack *Packet, err *error) {
	if err != nil && *err != nil {
		an.Nack()
	} else {
		var p Packet
		if ack != nil {
			p = *ack
		}
		an.Ack(p)
	}
}

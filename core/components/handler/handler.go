// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"time"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/dutycycle"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

const bufferDelay time.Duration = time.Millisecond * 300

// component implements the core.Component interface
type component struct {
	broker  JSONRecipient
	ctx     log.Interface
	devices DevStorage
	packets PktStorage
	set     chan<- bundle
}

type bundle struct {
	Adapter Adapter
	Chresp  chan interface{}
	Entry   devEntry
	ID      [20]byte
	Packet  HPacket
	Time    time.Time
}

// New construct a new Handler
func New(devDb DevStorage, pktDb PktStorage, broker JSONRecipient, ctx log.Interface) Handler {
	h := component{
		ctx:     ctx,
		devices: devDb,
		packets: pktDb,
		broker:  broker,
	}

	set := make(chan bundle)
	bundles := make(chan []bundle)

	h.set = set
	go h.consumeBundles(bundles)
	go h.consumeSet(bundles, set)

	return h
}

// Register implements the core.Component interface
func (h component) Register(reg Registration, an AckNacker, sub Subscriber) (err error) {
	h.ctx.WithField("registration", reg).Debug("New registration request")
	defer ensureAckNack(an, nil, &err)
	stats.MarkMeter("handler.registration.in")

	hreg, ok := reg.(HRegistration)
	if !ok {
		stats.MarkMeter("handler.registration.invalid")
		return errors.New(errors.Structural, "Not a Handler registration")
	}

	if err = h.devices.StorePersonalized(hreg); err != nil {
		return errors.New(errors.Operational, err)
	}

	return sub.Subscribe(brokerRegistration{
		recipient: h.broker,
		appEUI:    hreg.AppEUI(),
		devEUI:    hreg.DevEUI(),
		nwkSKey:   hreg.NwkSKey(),
	})
}

// HandleUp implements the core.Component interface
func (h component) HandleUp(data []byte, an AckNacker, up Adapter) (err error) {
	// Make sure we don't forget the AckNacker
	var ack Packet
	defer ensureAckNack(an, &ack, &err)
	stats.MarkMeter("handler.uplink.in")

	itf, err := UnmarshalPacket(data)
	if err != nil {
		stats.MarkMeter("handler.uplink.invalid")
		return errors.New(errors.Structural, err)
	}

	switch itf.(type) {
	case HPacket:
		stats.MarkMeter("handler.uplink.data")

		// 0. Retrieve the handler packet
		packet := itf.(HPacket)
		appEUI := packet.AppEUI()
		devEUI := packet.DevEUI()

		// 1. Lookup for the associated AppSKey + Recipient
		h.ctx.WithField("appEUI", appEUI).WithField("devEUI", devEUI).Debug("Perform lookup")
		entry, err := h.devices.Lookup(appEUI, devEUI)
		if err != nil {
			return err
		}

		// 2. Prepare a channel to receive the response from the consumer
		chresp := make(chan interface{})

		// 3. Create a "bundle" which holds info waiting for other related packets
		var bundleID [20]byte // AppEUI(8) | DevEUI(8) | FCnt
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, appEUI[:])
		binary.Write(buf, binary.BigEndian, devEUI[:])
		binary.Write(buf, binary.BigEndian, packet.FCnt())
		data := buf.Bytes()
		if len(data) != 20 {
			return errors.New(errors.Structural, "Unable to generate bundleID")
		}
		copy(bundleID[:], data[:])

		// 4. Send the actual bundle to the consumer
		ctx := h.ctx.WithField("BundleID", bundleID)
		ctx.Debug("Define new bundle")
		h.set <- bundle{
			ID:      bundleID,
			Packet:  packet,
			Entry:   entry,
			Adapter: up,
			Chresp:  chresp,
			Time:    time.Now(),
		}

		// 5. Wait for the response. Could be an error, a packet or nothing.
		// We'll respond to a maximum of one node. The handler will use the
		// rssi + gateway's duty cycle to select to best fit.
		// All other channels will get a nil response.
		// If there's an error, all channels get the error.
		resp := <-chresp
		switch resp.(type) {
		case BPacket:
			stats.MarkMeter("handler.uplink.ack.with_response")
			stats.MarkMeter("handler.downlink.out")
			ctx.Debug("Received response with packet. Sending Ack")
			ack = resp.(Packet)
		case error:
			stats.MarkMeter("handler.uplink.error")
			ctx.WithError(resp.(error)).Warn("Received errored response. Sending Ack")
			return resp.(error)
		default:
			stats.MarkMeter("handler.uplink.ack.without_response")
			ctx.Debug("Received empty response. Sending empty Ack")
		}

		return nil
	case JPacket:
		stats.MarkMeter("handler.uplink.join_request")
		return errors.New(errors.Implementation, "Join Request not yet implemented")
	default:
		stats.MarkMeter("handler.uplink.unknown")
		return errors.New(errors.Implementation, "Unhandled packet type")
	}
}

// consumeBundles processes list of bundle generated overtime, decrypt the underlying packet,
// deduplicate them, and send a single enhanced packet to the upadapter for further processing.
func (h component) consumeBundles(chbundle <-chan []bundle) {
	ctx := h.ctx.WithField("goroutine", "bundle consumer")
	ctx.Debug("Starting bundle consumer")

browseBundles:
	for bundles := range chbundle {
		ctx.WithField("BundleID", bundles[0].ID).Debug("Consume new bundle")
		var metadata []Metadata
		var payload []byte
		var firstTime time.Time

		if len(bundles) < 1 {
			continue browseBundles
		}
		b := bundles[0]
		h.ctx.WithField("Metadata", b.Packet.Metadata()).Debug("Considering first packet")

		computer, scores, err := dutycycle.NewScoreComputer(b.Packet.Metadata().Datr)
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		stats.UpdateHistogram("handler.uplink.duplicate.count", int64(len(bundles)))

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
				firstTime = bundle.Time
				stats.MarkMeter("handler.uplink.in.unique")
			} else {
				diff := bundle.Time.Sub(firstTime).Nanoseconds()
				stats.UpdateHistogram("handler.uplink.duplicate.delay", diff/1000)
			}

			// Append metadata for each of them
			metadata = append(metadata, bundle.Packet.Metadata())
			scores = computer.Update(scores, i, bundle.Packet.Metadata())
		}

		// Then create an application-level packet
		packet, err := NewAPacket(b.Packet.AppEUI(), b.Packet.DevEUI(), payload, metadata)
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		// And send it to the wild open
		// we don't expect a response from the adapter, end of the chain.
		recipient, err := b.Adapter.GetRecipient(b.Entry.Recipient)
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		_, err = b.Adapter.Send(packet, recipient)
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}
		stats.MarkMeter("handler.uplink.out")

		// Now handle the downlink and respond to node
		h.ctx.Debug("Looking for downlink response")
		best := computer.Get(scores)
		h.ctx.WithField("Bundle", best).Debug("Determine best gateway")
		var down APacket
		if best != nil { // Avoid pulling when there's no gateway available for an answer
			down, err = h.packets.Pull(b.Packet.AppEUI(), b.Packet.DevEUI())
		}
		if err != nil && err.(errors.Failure).Nature != errors.NotFound {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}
		h.ctx.WithField("Packet", down).Debug("Pull downlink from storage")
		for i, bundle := range bundles {
			if best != nil && best.ID == i && down != nil && err == nil {
				stats.MarkMeter("handler.downlink.pull")

				bpacket, err := h.buildDownlink(down, bundle.Packet, bundle.Entry, best.IsRX2)
				if err != nil {
					go h.abortConsume(errors.New(errors.Structural, err), bundles)
					continue browseBundles
				}
				if err := h.devices.UpdateFCnt(b.Packet.AppEUI(), b.Packet.DevEUI(), bpacket.FCnt()); err != nil {
					go h.abortConsume(errors.New(errors.Structural, err), bundles)
					continue browseBundles
				}
				bundle.Chresp <- bpacket
			} else {
				bundle.Chresp <- nil
			}
		}
	}
}

// Abort consume forward the given error to all bundle recipients
func (h component) abortConsume(err error, bundles []bundle) {
	stats.MarkMeter("handler.uplink.invalid")
	h.ctx.WithError(err).Debug("Unable to consume bundle")
	for _, bundle := range bundles {
		bundle.Chresp <- err
	}
}

// constructs a downlink packet from something we pulled from the gathered downlink, and, the actual
// uplink.
func (h component) buildDownlink(down APacket, up HPacket, entry devEntry, isRX2 bool) (BPacket, error) {
	macPayload := lorawan.NewMACPayload(false)
	macPayload.FHDR = lorawan.FHDR{
		FCnt:    entry.FCntDown + 1,
		DevAddr: entry.DevAddr,
	}
	macPayload.FPort = 1
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{
		Bytes: down.Payload(),
	}}

	if err := macPayload.EncryptFRMPayload(entry.AppSKey); err != nil {
		return nil, err
	}

	payload := lorawan.NewPHYPayload(false)
	payload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataDown, // TODO Handle Confirmed data down
		Major: lorawan.LoRaWANR1,
	}
	payload.MACPayload = macPayload

	data, err := payload.MarshalBinary()
	if err != nil {
		return nil, err
	}
	pmetadata := up.Metadata()

	if pmetadata.Tmst == nil || pmetadata.Freq == nil || pmetadata.Codr == nil || pmetadata.Datr == nil {
		return nil, errors.New(errors.Structural, "Missing mandatory metadata in uplink packet")
	}

	metadata := Metadata{
		Freq: pmetadata.Freq,
		Codr: pmetadata.Codr,
		Datr: pmetadata.Datr,
		Size: pointer.Uint(uint(len(data))),
		Tmst: pointer.Uint(*pmetadata.Tmst + 1000),
	}

	if isRX2 { // Should we reply on RX2, metadata aren't the same
		// TODO Handle different regions with non hard-coded values
		metadata.Freq = pointer.Float64(869.5)
		metadata.Datr = pointer.String("SF9BW125")
		metadata.Tmst = pointer.Uint(*pmetadata.Tmst + 2000)
	}

	return NewBPacket(payload, metadata)
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
			ctx = ctx.WithField("BundleID", b.ID)

			// Check if bundle has already been processed
			var pid [16]byte
			copy(pid[:], b.ID[:16])
			if reflect.DeepEqual(processed[pid], b.ID[16:]) {
				ctx.Debug("Reject already processed bundle")
				go func(b bundle) {
					b.Chresp <- errors.New(errors.Behavioural, "Already processed")
				}(b)
				continue
			}

			// Add the bundle to the stack, and set the alarm if its the first
			bundles := append(buffers[b.ID], b)
			if len(bundles) == 1 {
				go setAlarm(alarm, b.ID, bufferDelay)
				ctx.Debug("Buffering started -> new alarm set")
			}
			buffers[b.ID] = bundles
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
	stats.MarkMeter("handler.downlink.in")

	h.ctx.Debug("Handle downlink message")

	// Unmarshal the given packet and see what gift we get
	itf, err := UnmarshalPacket(data)
	if err != nil {
		stats.MarkMeter("handler.downlink.invalid")
		return errors.New(errors.Structural, err)
	}

	switch itf.(type) {
	case APacket:
		apacket := itf.(APacket)
		h.ctx.WithField("DevEUI", apacket.DevEUI()).WithField("AppEUI", apacket.AppEUI()).Debug("Save downlink for later")
		return h.packets.Push(apacket)
	default:
		stats.MarkMeter("handler.downlink.invalid")
		return errors.New(errors.Implementation, "Unhandled packet type")
	}
}

func ensureAckNack(an AckNacker, ack *Packet, err *error) {
	if err != nil && *err != nil {
		an.Nack(*err)
	} else {
		var p Packet
		if ack != nil {
			p = *ack
		}
		an.Ack(p)
	}
}

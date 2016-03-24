// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"bytes"
	"crypto/aes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/dutycycle"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// bufferDelay defines the timeframe length during which we bufferize packets
const bufferDelay time.Duration = time.Millisecond * 300

// dataRates makes correspondance between string datarate identifier and lorawan uint descriptors
var dataRates = map[string]uint8{
	"SF12BW125": 0,
	"SF11BW125": 1,
	"SF10BW125": 2,
	"SF9BW125":  3,
	"SF8BW125":  4,
	"SF7BW125":  5,
}

// component implements the core.Component interface
type component struct {
	Components
	Set                    chan<- bundle
	PublicNetAddr          string
	PrivateNetAddr         string
	PrivateNetAddrAnnounce string
	Configuration          struct {
		CFList      [5]uint32
		NetID       [3]byte
		RX1DROffset uint8
		RX2DataRate string
		RX2Freq     float32
		RXDelay     uint8
		PowerRX1    uint32
		PowerRX2    uint32
		InvPolarity bool
		JoinDelay   uint8
	}
}

// Interface defines the Handler interface
type Interface interface {
	core.HandlerServer
	core.HandlerManagerServer
	Start() error
}

// Components is used to make handler instantiation easier
type Components struct {
	Broker     core.AuthBrokerClient
	Ctx        log.Interface
	DevStorage DevStorage
	PktStorage PktStorage
	AppAdapter core.AppClient
}

// Options is used to make handler instantiation easier
type Options struct {
	PublicNetAddr          string // Net Address used to communicate with the handler from the outside
	PrivateNetAddr         string // Net Address the handler listens on for internal communications
	PrivateNetAddrAnnounce string // Net Address the handler announces to brokers for internal communications
}

// bundle are used to materialize an incoming request being bufferized, waiting for the others.
type bundle struct {
	Chresp   chan interface{}
	Entry    devEntry
	ID       [20]byte
	Packet   interface{}
	DataRate string
	Time     time.Time
}

// New construct a new Handler
func New(c Components, o Options) Interface {
	h := &component{
		Components:             c,
		PublicNetAddr:          o.PublicNetAddr,
		PrivateNetAddr:         o.PrivateNetAddr,
		PrivateNetAddrAnnounce: o.PrivateNetAddrAnnounce,
	}

	// TODO Make it configurable
	h.Configuration.CFList = [5]uint32{867100000, 867300000, 867500000, 867700000, 867900000}
	h.Configuration.NetID = [3]byte{14, 14, 14}
	h.Configuration.RX1DROffset = 0
	h.Configuration.RX2DataRate = "SF9BW125"
	h.Configuration.RX2Freq = 869.525
	h.Configuration.RXDelay = 1
	h.Configuration.JoinDelay = 5
	h.Configuration.PowerRX1 = 14
	h.Configuration.PowerRX2 = 27
	h.Configuration.InvPolarity = true

	set := make(chan bundle)
	bundles := make(chan []bundle)

	h.Set = set
	go h.consumeBundles(bundles)
	go h.consumeSet(bundles, set)

	return h
}

// Start actually runs the component and starts the rpc server
func (h *component) Start() error {
	pubConn, errPub := net.Listen("tcp", h.PublicNetAddr)
	priConn, errPri := net.Listen("tcp", h.PrivateNetAddr)

	if errPub != nil || errPri != nil {
		return errors.New(errors.Operational, fmt.Sprintf("Unable to open connections: %s | %s", errPub, errPri))
	}

	server := grpc.NewServer()
	core.RegisterHandlerServer(server, h)
	core.RegisterHandlerManagerServer(server, h)

	cherr := make(chan error)

	go func() { cherr <- server.Serve(pubConn) }()
	go func() { cherr <- server.Serve(priConn) }()

	if err := <-cherr; err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

// HandleJoin implements the core.HandlerServer interface
func (h component) HandleJoin(bctx context.Context, req *core.JoinHandlerReq) (*core.JoinHandlerRes, error) {
	stats.MarkMeter("handler.joinrequest.in")

	// 0. Check the packet integrity
	if req == nil || len(req.DevEUI) != 8 || len(req.AppEUI) != 8 || len(req.DevNonce) != 2 || req.Metadata == nil {
		h.Ctx.Debug("Invalid join request packet")
		return new(core.JoinHandlerRes), errors.New(errors.Structural, "Invalid parameters")
	}

	// 1. Lookup for the associated AppKey
	h.Ctx.WithField("appEUI", req.AppEUI).WithField("devEUI", req.DevEUI).Debug("Perform lookup")
	entry, err := h.DevStorage.read(req.AppEUI, req.DevEUI)
	if err != nil {
		return new(core.JoinHandlerRes), err
	}
	if entry.AppKey == nil { // Trying to activate an ABP device
		return new(core.JoinHandlerRes), errors.New(errors.Behavioural, "Trying to activate a personalized device")
	}

	// 1.5. (Yep, indexed comments sucks) Verify MIC
	payload := lorawan.NewPHYPayload(true)
	payload.MHDR = lorawan.MHDR{Major: lorawan.LoRaWANR1, MType: lorawan.JoinRequest}
	joinPayload := lorawan.JoinRequestPayload{}
	copy(payload.MIC[:], req.MIC)
	copy(joinPayload.AppEUI[:], req.AppEUI)
	copy(joinPayload.DevEUI[:], req.DevEUI)
	copy(joinPayload.DevNonce[:], req.DevNonce)
	payload.MACPayload = &joinPayload
	if ok, err := payload.ValidateMIC(lorawan.AES128Key(*entry.AppKey)); err != nil || !ok {
		h.Ctx.WithError(err).Debug("Unable to validate MIC from incoming join-request")
		return new(core.JoinHandlerRes), errors.New(errors.Structural, "Unable to validate MIC")
	}

	// 2. Prepare a channel to receive the response from the consumer
	chresp := make(chan interface{})

	// 3. Create a "bundle" which holds info waiting for other related packets
	var bundleID [20]byte // AppEUI(8) | DevEUI(8) | DevNonce | [ 0 0 ]
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, req.AppEUI)
	_ = binary.Write(buf, binary.BigEndian, req.DevEUI)
	_ = binary.Write(buf, binary.BigEndian, req.DevNonce)
	copy(bundleID[:], buf.Bytes())

	// 4. Send the actual bundle to the consumer
	ctx := h.Ctx.WithField("BundleID", bundleID)
	ctx.Debug("Define new bundle")
	req.Metadata.Time = time.Now().Format(time.RFC3339Nano)
	h.Set <- bundle{
		ID:       bundleID,
		Packet:   req,
		DataRate: req.Metadata.DataRate,
		Entry:    entry,
		Chresp:   chresp,
		Time:     time.Now(),
	}

	// 5. Control the response
	resp := <-chresp
	switch resp.(type) {
	case *core.JoinHandlerRes:
		stats.MarkMeter("handler.join.send_accept")
		ctx.Debug("Sending Join-Accept")
		return resp.(*core.JoinHandlerRes), nil
	case error:
		stats.MarkMeter("handler.join.error")
		ctx.WithError(resp.(error)).Warn("Error while processing join-request.")
		return new(core.JoinHandlerRes), resp.(error)
	default:
		ctx.Debug("No response to send.")
		return new(core.JoinHandlerRes), nil
	}
}

// HandleDataDown implements the core.HandlerServer interface
func (h component) HandleDataDown(bctx context.Context, req *core.DataDownHandlerReq) (*core.DataDownHandlerRes, error) {
	stats.MarkMeter("handler.downlink.in")
	h.Ctx.Debug("Handle downlink message")

	// Unmarshal the given packet and see what gift we get

	if len(req.AppEUI) != 8 {
		stats.MarkMeter("handler.downlink.invalid")
		return new(core.DataDownHandlerRes), errors.New(errors.Structural, "Invalid Application EUI")
	}

	if len(req.DevEUI) != 8 {
		stats.MarkMeter("handler.downlink.invalid")
		return new(core.DataDownHandlerRes), errors.New(errors.Structural, "Invalid Device EUI")
	}

	if len(req.Payload) == 0 {
		stats.MarkMeter("handler.downlink.invalid")
		return new(core.DataDownHandlerRes), errors.New(errors.Structural, "Invalid payload")
	}

	ttl, err := time.ParseDuration(req.TTL)
	if err != nil || ttl == 0 {
		stats.MarkMeter("handler.downlink.invalid")
		return new(core.DataDownHandlerRes), errors.New(errors.Structural, "Invalid TTL")
	}

	h.Ctx.WithField("DevEUI", req.DevEUI).WithField("AppEUI", req.AppEUI).Debug("Save downlink for later")
	return new(core.DataDownHandlerRes), h.PktStorage.enqueue(pktEntry{
		Payload: req.Payload,
		AppEUI:  req.AppEUI,
		DevEUI:  req.DevEUI,
		TTL:     time.Now().Add(ttl),
	})
}

// HandleDataUp implements the core.HandlerServer interface
func (h component) HandleDataUp(bctx context.Context, req *core.DataUpHandlerReq) (*core.DataUpHandlerRes, error) {
	stats.MarkMeter("handler.uplink.in")

	// 0. Check the packet integrity
	if len(req.Payload) == 0 {
		stats.MarkMeter("handler.uplink.invalid")
		return new(core.DataUpHandlerRes), errors.New(errors.Structural, "Invalid Packet Payload")
	}
	if len(req.DevEUI) != 8 {
		stats.MarkMeter("handler.uplink.invalid")
		return new(core.DataUpHandlerRes), errors.New(errors.Structural, "Invalid Device EUI")
	}
	if len(req.AppEUI) != 8 {
		stats.MarkMeter("handler.uplink.invalid")
		return new(core.DataUpHandlerRes), errors.New(errors.Structural, "Invalid Application EUI")
	}
	if req.Metadata == nil {
		stats.MarkMeter("handler.uplink.invalid")
		return new(core.DataUpHandlerRes), errors.New(errors.Structural, "Missing Mandatory Metadata")
	}
	stats.MarkMeter("handler.uplink.data")

	// 1. Lookup for the associated AppSKey + Application
	h.Ctx.WithField("appEUI", req.AppEUI).WithField("devEUI", req.DevEUI).Debug("Perform lookup")
	entry, err := h.DevStorage.read(req.AppEUI, req.DevEUI)
	if err != nil {
		return new(core.DataUpHandlerRes), err
	}
	h.Ctx.Debugf("%+v", entry)
	if len(entry.DevAddr) != 4 { // Not Activated
		return new(core.DataUpHandlerRes), errors.New(errors.Structural, "Tried to send uplink on non-activated device")
	}

	// 2. Prepare a channel to receive the response from the consumer
	chresp := make(chan interface{})

	// 3. Create a "bundle" which holds info waiting for other related packets
	var bundleID [20]byte // AppEUI(8) | DevEUI(8) | FCnt
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, req.AppEUI)
	_ = binary.Write(buf, binary.BigEndian, req.DevEUI)
	_ = binary.Write(buf, binary.BigEndian, req.FCnt)
	copy(bundleID[:], buf.Bytes())

	// 4. Send the actual bundle to the consumer
	ctx := h.Ctx.WithField("BundleID", bundleID)
	ctx.Debug("Define new bundle")
	req.Metadata.Time = time.Now().Format(time.RFC3339Nano)
	h.Set <- bundle{
		ID:       bundleID,
		Packet:   req,
		DataRate: req.Metadata.DataRate,
		Entry:    entry,
		Chresp:   chresp,
		Time:     time.Now(),
	}

	// 5. Wait for the response. Could be an error, a packet or nothing.
	// We'll respond to a maximum of one node. The handler will use the
	// rssi + gateway's duty cycle to select to best fit.
	// All other channels will get a nil response.
	// If there's an error, all channels get the error.
	resp := <-chresp
	switch resp.(type) {
	case *core.DataUpHandlerRes:
		stats.MarkMeter("handler.uplink.ack.with_response")
		stats.MarkMeter("handler.downlink.out")
		ctx.Debug("Sending downlink packet as response.")
		return resp.(*core.DataUpHandlerRes), nil
	case error:
		stats.MarkMeter("handler.uplink.error")
		ctx.WithError(resp.(error)).Warn("Error while processing dowlink.")
		return new(core.DataUpHandlerRes), resp.(error)
	default:
		stats.MarkMeter("handler.uplink.ack.without_response")
		ctx.Debug("No response to send.")
		return new(core.DataUpHandlerRes), nil
	}
}

// consumeSet gathers new incoming bundles which possess the same id (i.e. appEUI & devEUI & Fcnt)
// It then flushes them once a given delay has passed since the reception of the first bundle.
func (h component) consumeSet(chbundles chan<- []bundle, chset <-chan bundle) {
	ctx := h.Ctx.WithField("goroutine", "set consumer")
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

// consumeBundles processes list of bundle generated overtime, decrypt the underlying packet,
// deduplicate them, and send a single enhanced packet to the upadapter for further processing.
func (h component) consumeBundles(chbundle <-chan []bundle) {
	ctx := h.Ctx.WithField("goroutine", "bundle consumer")
	ctx.Debug("Starting bundle consumer")

browseBundles:
	for bundles := range chbundle {
		ctx.WithField("BundleID", bundles[0].ID).Debug("Consume new bundle")
		if len(bundles) < 1 {
			continue browseBundles
		}
		b := bundles[0]
		switch b.Packet.(type) {
		case *core.DataUpHandlerReq:
			pkt := b.Packet.(*core.DataUpHandlerReq)
			go h.consumeDown(pkt.AppEUI, pkt.DevEUI, b.DataRate, bundles)
		case *core.JoinHandlerReq:
			pkt := b.Packet.(*core.JoinHandlerReq)
			// Entry.AppKey not nil, checked before creating any bundles
			go h.consumeJoin(pkt.AppEUI, pkt.DevEUI, *b.Entry.AppKey, b.DataRate, bundles)
		}
	}
}

// consume Join actually consumes a set of join-request packets
func (h component) consumeJoin(appEUI []byte, devEUI []byte, appKey [16]byte, dataRate string, bundles []bundle) {
	ctx := h.Ctx.WithField("AppEUI", appEUI).WithField("DevEUI", devEUI)
	ctx.Debug("Consuming join-request")

	// Compute score while gathering metadata
	var metadata []*core.Metadata
	computer, scores, err := dutycycle.NewScoreComputer(dataRate)
	if err != nil {
		ctx.WithError(err).Debug("Unable to instantiate score computer")
		h.abortConsume(err, bundles)
		return
	}

	ctx.Debug("Compute scores for each packet")
	for i, bundle := range bundles {
		packet := bundle.Packet.(*core.JoinHandlerReq)
		metadata = append(metadata, packet.Metadata)
		scores = computer.Update(scores, i, *packet.Metadata)
	}

	// Check if at least one is available
	best := computer.Get(scores)
	ctx.WithField("Best", best).Debug("Determine best recipient to reply")
	if best == nil {
		h.abortConsume(errors.New(errors.Operational, "No gateway is available for an answer"), bundles)
		return
	}
	packet := bundles[best.ID].Packet.(*core.JoinHandlerReq)

	// Generate a DevAddr + NwkSKey + AppSKey
	ctx.Debug("Generate DevAddr, NwkSKey and AppSKey")
	rdn := rand.New(rand.NewSource(int64(packet.Metadata.Rssi)))
	appNonce, devAddr := make([]byte, 4), [4]byte{}
	binary.BigEndian.PutUint32(appNonce, rdn.Uint32())
	binary.BigEndian.PutUint32(devAddr[:], rdn.Uint32())
	devAddr[0] = (h.Configuration.NetID[2] << 1) | (devAddr[0] & 1) // DevAddr 7 msb are NetID 7 lsb

	buf := make([]byte, 16)
	copy(buf[1:4], appNonce[:3])
	copy(buf[4:7], h.Configuration.NetID[:])
	copy(buf[7:9], packet.DevNonce)

	block, err := aes.NewCipher(appKey[:])
	if err != nil || block.BlockSize() != 16 {
		h.abortConsume(errors.New(errors.Structural, "Unable to create cipher to generate keys"), bundles)
		return
	}

	var nwkSKey, appSKey [16]byte
	buf[0] = 0x1
	block.Encrypt(nwkSKey[:], buf)
	buf[0] = 0x2
	block.Encrypt(appSKey[:], buf)

	// Update the internal storage entry
	err = h.DevStorage.upsert(devEntry{
		AppEUI:   appEUI,
		AppKey:   &appKey,
		AppSKey:  appSKey,
		DevAddr:  devAddr[:],
		DevEUI:   devEUI,
		FCntDown: 0,
		FCntUp:   0,
		NwkSKey:  nwkSKey,
	})
	if err != nil {
		ctx.WithError(err).Debug("Unable to initialize devEntry with activation")
		h.abortConsume(err, bundles)
		return
	}

	// Build join-accept and send it
	joinAccept, err := h.buildJoinAccept(packet, appKey, appNonce[:3], devAddr, best.IsRX2)
	if err != nil {
		ctx.WithError(err).Debug("Unable to build join accept")
		h.abortConsume(err, bundles)
		return
	}
	joinAccept.NwkSKey = nwkSKey[:]
	joinAccept.DevAddr = devAddr[:]

	// Notify the application
	_, err = h.AppAdapter.HandleJoin(context.Background(), &core.JoinAppReq{
		Metadata: metadata,
		AppEUI:   appEUI,
		DevEUI:   devEUI,
	})
	if err != nil {
		ctx.WithError(err).Debug("Fails to notify application")
	}

	for i, bundle := range bundles {
		if i == best.ID {
			bundle.Chresp <- joinAccept
		} else {
			bundle.Chresp <- nil
		}
	}
}

// consume Down actually consumes a set of downlink packets
func (h component) consumeDown(appEUI []byte, devEUI []byte, dataRate string, bundles []bundle) {
	stats.UpdateHistogram("handler.uplink.duplicate.count", int64(len(bundles)))
	var metadata []*core.Metadata
	var payload []byte
	var firstTime time.Time

	computer, scores, err := dutycycle.NewScoreComputer(dataRate)
	if err != nil {
		h.abortConsume(err, bundles)
		return
	}

	for i, bundle := range bundles {
		// We only decrypt the payload of the first bundle's packet.
		// We assume all the other to be equal and we'll merely collect
		// metadata from other bundle.
		packet := bundle.Packet.(*core.DataUpHandlerReq)
		if i == 0 {
			var err error
			var devAddr lorawan.DevAddr
			copy(devAddr[:], bundle.Entry.DevAddr)
			payload, err = lorawan.EncryptFRMPayload(
				bundle.Entry.AppSKey,
				true,
				devAddr,
				packet.FCnt,
				packet.Payload,
			)
			if err != nil {
				h.abortConsume(err, bundles)
				return
			}
			firstTime = bundle.Time
			stats.MarkMeter("handler.uplink.in.unique")
		} else {
			diff := bundle.Time.Sub(firstTime).Nanoseconds()
			stats.UpdateHistogram("handler.uplink.duplicate.delay", diff/1000)
		}

		// Append metadata for each of them
		metadata = append(metadata, packet.Metadata)
		scores = computer.Update(scores, i, *packet.Metadata) // Nil check already done
	}

	// Then create an application-level packet and send it to the wild open
	// we don't expect a response from the adapter, end of the chain.
	_, err = h.AppAdapter.HandleData(context.Background(), &core.DataAppReq{
		AppEUI:   appEUI,
		DevEUI:   devEUI,
		Payload:  payload,
		Metadata: metadata,
	})
	if err != nil {
		h.abortConsume(errors.New(errors.Operational, err), bundles)
		return
	}

	stats.MarkMeter("handler.uplink.out")

	// Now handle the downlink and respond to node
	h.Ctx.Debug("Looking for downlink response")
	best := computer.Get(scores)
	h.Ctx.WithField("Bundle", best).Debug("Determine best gateway")
	var downlink pktEntry
	if best != nil { // Avoid pulling when there's no gateway available for an answer
		downlink, err = h.PktStorage.dequeue(appEUI, devEUI)
	}
	if err != nil && err.(errors.Failure).Nature != errors.NotFound {
		h.abortConsume(err, bundles)
		return
	}

	// One of those bundle might be available for a response
	upType := lorawan.MType(bundles[0].Packet.(*core.DataUpHandlerReq).MType)
	for i, bundle := range bundles {
		if best != nil && best.ID == i && (downlink.Payload != nil || upType == lorawan.ConfirmedDataUp) {
			stats.MarkMeter("handler.downlink.pull")
			var downType lorawan.MType
			switch upType {
			case lorawan.UnconfirmedDataUp:
				downType = lorawan.UnconfirmedDataDown
			case lorawan.ConfirmedDataUp:
				downType = lorawan.ConfirmedDataDown
			default:
				h.abortConsume(errors.New(errors.Implementation, "Unrecognized uplink MType"), bundles)
				return
			}
			downlink, err := h.buildDownlink(downlink.Payload, downType, *bundle.Packet.(*core.DataUpHandlerReq), bundle.Entry, best.IsRX2)
			if err != nil {
				h.abortConsume(errors.New(errors.Structural, err), bundles)
				return
			}

			bundle.Entry.FCntDown = downlink.Payload.MACPayload.FHDR.FCnt
			bundle.Entry.FCntUp = bundle.Packet.(*core.DataUpHandlerReq).FCnt
			err = h.DevStorage.upsert(bundle.Entry)
			if err != nil {
				h.abortConsume(err, bundles)
				return
			}
			bundle.Chresp <- downlink
		} else {
			bundle.Chresp <- nil
		}
	}

	// Then, if there was no downlink, we still update the Frame Counter Up in the storage
	if best == nil || downlink.Payload == nil && upType != lorawan.ConfirmedDataUp {
		bundles[0].Entry.FCntUp = bundles[0].Packet.(*core.DataUpHandlerReq).FCnt
		if err := h.DevStorage.upsert(bundles[0].Entry); err != nil {
			h.Ctx.WithError(err).Debug("Unable to update Frame Counter Up")
		}
	}
}

// Abort consume forward the given error to all bundle recipients
func (h component) abortConsume(err error, bundles []bundle) {
	stats.MarkMeter("handler.uplink.invalid")
	h.Ctx.WithError(err).Debug("Unable to consume bundle")
	for _, bundle := range bundles {
		bundle.Chresp <- err
	}
}

// constructs a downlink packet from something we pulled from the gathered downlink, and, the actual
// uplink.
func (h component) buildDownlink(down []byte, mtype lorawan.MType, up core.DataUpHandlerReq, entry devEntry, isRX2 bool) (*core.DataUpHandlerRes, error) {
	macpayload := lorawan.NewMACPayload(false)
	macpayload.FHDR = lorawan.FHDR{
		FCnt: entry.FCntDown + 1,
	}
	copy(macpayload.FHDR.DevAddr[:], entry.DevAddr)
	macpayload.FPort = new(uint8)
	*macpayload.FPort = 1
	if mtype == lorawan.ConfirmedDataDown {
		macpayload.FHDR.FCtrl.ACK = true
	}
	var frmpayload []byte
	if down != nil {
		macpayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: down}}
		err := macpayload.EncryptFRMPayload(entry.AppSKey)
		if err != nil {
			return nil, errors.New(errors.Structural, err)
		}
		frmpayload, err = macpayload.FRMPayload[0].MarshalBinary()
		if err != nil {
			return nil, errors.New(errors.Structural, err)
		}
	}

	payload := lorawan.NewPHYPayload(false)
	payload.MHDR = lorawan.MHDR{
		MType: mtype,
		Major: lorawan.LoRaWANR1,
	}
	payload.MACPayload = macpayload

	data, err := payload.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	metadata := h.buildMetadata(*up.Metadata, uint32(len(data)), 1000000*uint32(h.Configuration.RXDelay), isRX2)

	return &core.DataUpHandlerRes{
		Payload: &core.LoRaWANData{
			MHDR: &core.LoRaWANMHDR{
				MType: uint32(payload.MHDR.MType),
				Major: uint32(payload.MHDR.Major),
			},
			MACPayload: &core.LoRaWANMACPayload{
				FHDR: &core.LoRaWANFHDR{
					DevAddr: macpayload.FHDR.DevAddr[:],
					FCnt:    macpayload.FHDR.FCnt,
					FCtrl: &core.LoRaWANFCtrl{
						ADR:       macpayload.FHDR.FCtrl.ADR,
						ADRAckReq: macpayload.FHDR.FCtrl.ADRACKReq,
						Ack:       macpayload.FHDR.FCtrl.ACK,
						FPending:  macpayload.FHDR.FCtrl.FPending,
					},
				},
				FPort:      uint32(*macpayload.FPort),
				FRMPayload: frmpayload,
			},
			MIC: payload.MIC[:],
		},
		Metadata: &metadata,
	}, nil
}

func (h component) buildJoinAccept(joinReq *core.JoinHandlerReq, appKey [16]byte, appNonce []byte, devAddr [4]byte, isRX2 bool) (*core.JoinHandlerRes, error) {
	payload := lorawan.NewPHYPayload(false)
	payload.MHDR = lorawan.MHDR{
		MType: lorawan.JoinAccept,
		Major: lorawan.LoRaWANR1,
	}
	joinAcceptPayload := &lorawan.JoinAcceptPayload{
		NetID:   lorawan.NetID(h.Configuration.NetID),
		DevAddr: lorawan.DevAddr(devAddr),
		DLSettings: lorawan.DLsettings{
			RX1DRoffset: h.Configuration.RX1DROffset,
			RX2DataRate: dataRates[h.Configuration.RX2DataRate],
		},
		RXDelay: h.Configuration.RXDelay,
	}
	cflist := lorawan.CFList(h.Configuration.CFList)
	joinAcceptPayload.CFList = &cflist
	copy(joinAcceptPayload.AppNonce[:], appNonce)
	payload.MACPayload = joinAcceptPayload
	if err := payload.SetMIC(lorawan.AES128Key(appKey)); err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	if err := payload.EncryptJoinAcceptPayload(lorawan.AES128Key(appKey)); err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	data, err := payload.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	m := h.buildMetadata(*joinReq.Metadata, uint32(len(data)), 1000000*uint32(h.Configuration.JoinDelay), isRX2)
	return &core.JoinHandlerRes{
		Payload: &core.LoRaWANJoinAccept{
			Payload: data,
		},
		Metadata: &m,
	}, nil
}

// buildMetadata construct a new Metadata
func (h component) buildMetadata(metadata core.Metadata, size uint32, baseDelay uint32, isRX2 bool) core.Metadata {
	m := core.Metadata{
		Frequency:   metadata.Frequency,
		CodingRate:  metadata.CodingRate,
		DataRate:    metadata.DataRate,
		Modulation:  metadata.Modulation,
		RFChain:     metadata.RFChain,
		InvPolarity: h.Configuration.InvPolarity,
		Power:       h.Configuration.PowerRX1,
		PayloadSize: size,
		Timestamp:   metadata.Timestamp + baseDelay,
	}

	if isRX2 { // Should we reply on RX2, metadata aren't the same
		// TODO Handle different regions with non hard-coded values
		m.Frequency = h.Configuration.RX2Freq
		m.DataRate = h.Configuration.RX2DataRate
		m.Power = h.Configuration.PowerRX2
		m.Timestamp = metadata.Timestamp + baseDelay + 1000000
	}
	return m
}

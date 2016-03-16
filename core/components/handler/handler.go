// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"bytes"
	"encoding/binary"
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

const bufferDelay time.Duration = time.Millisecond * 300

// component implements the core.Component interface
type component struct {
	broker  core.BrokerClient
	ctx     log.Interface
	devices DevStorage
	packets PktStorage
	set     chan<- bundle
	adapter core.AppClient
	netAddr string
}

type bundle struct {
	Chresp chan interface{}
	Entry  devEntry
	ID     [20]byte
	Packet core.DataUpHandlerReq
	Time   time.Time
}

// New construct a new Handler
func New(devDb DevStorage, pktDb PktStorage, broker core.BrokerClient, adapter core.AppClient, ctx log.Interface) core.HandlerServer {
	return &component{
		ctx:     ctx,
		devices: devDb,
		packets: pktDb,
		broker:  broker,
		adapter: adapter,
	}
}

// Start actually runs the component and starts the rpc server
func (h *component) Start(netAddr string) error {
	conn, err := net.Listen("tcp", netAddr)
	if err != nil {
		return errors.New(errors.Operational, err)
	}

	server := grpc.NewServer()
	core.RegisterHandlerServer(server, h)

	set := make(chan bundle)
	h.set = set
	bundles := make(chan []bundle)
	go h.consumeBundles(bundles)
	go h.consumeSet(bundles, set)

	defer close(bundles)
	defer close(set)

	if err := server.Serve(conn); err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

// RegisterPersonalized implements the core.HandlerServer interface
func (h component) SubscribePersonalized(bctx context.Context, req *core.SubPersoHandlerReq) (*core.SubPersoHandlerRes, error) {
	h.ctx.Debug("New personalized subscription request")
	stats.MarkMeter("handler.registration.in")

	if len(req.AppEUI) != 8 {
		stats.MarkMeter("handler.registration.invalid")
		return nil, errors.New(errors.Structural, "Invalid Application EUI")
	}

	if len(req.DevAddr) != 4 {
		stats.MarkMeter("handler.registration.invalid")
		return nil, errors.New(errors.Structural, "Invalid Device Address")
	}
	var devAddr [4]byte
	copy(devAddr[:], req.DevAddr)

	if len(req.NwkSKey) != 16 {
		stats.MarkMeter("handler.registration.invalid")
		return nil, errors.New(errors.Structural, "Invalid Network Session Key")
	}
	var nwkSKey [16]byte
	copy(nwkSKey[:], req.NwkSKey)

	if len(req.AppSKey) != 16 {
		stats.MarkMeter("handler.registration.invalid")
		return nil, errors.New(errors.Structural, "Invalid Application Session Key")
	}
	var appSKey [16]byte
	copy(appSKey[:], req.AppSKey)

	if h.netAddr == "" {
		return nil, errors.New(errors.Operational, "Illegal call. Start the server first")
	}

	if err := h.devices.StorePersonalized(req.AppEUI, devAddr, nwkSKey, appSKey); err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	_, err := h.broker.SubscribePersonalized(context.Background(), &core.SubPersoBrokerReq{
		HandlerNet: h.netAddr,
		AppEUI:     req.AppEUI,
		DevAddr:    req.DevAddr,
		NwkSKey:    req.NwkSKey,
	})

	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return nil, nil
}

// HandleDataDown implements the core.HandlerServer interface
func (h component) HandleDataDown(bctx context.Context, req *core.DataDownHandlerReq) (*core.DataDownHandlerRes, error) {
	stats.MarkMeter("handler.downlink.in")
	h.ctx.Debug("Handle downlink message")

	// Unmarshal the given packet and see what gift we get

	if len(req.AppEUI) != 8 {
		stats.MarkMeter("handler.downlink.invalid")
		return nil, errors.New(errors.Structural, "Invalid Application EUI")
	}

	if len(req.DevEUI) != 8 {
		stats.MarkMeter("handler.downlink.invalid")
		return nil, errors.New(errors.Structural, "Invalid Device EUI")
	}

	if len(req.Payload) == 0 {
		stats.MarkMeter("handler.downlink.invalid")
		return nil, errors.New(errors.Structural, "Invalid payload")
	}

	h.ctx.WithField("DevEUI", req.DevEUI).WithField("AppEUI", req.AppEUI).Debug("Save downlink for later")
	return nil, h.packets.Push(req.AppEUI, req.DevEUI, req.Payload)
}

// HandleDataUp implements the core.HandlerServer interface
func (h component) HandleDataUp(bctx context.Context, req *core.DataUpHandlerReq) (*core.DataUpHandlerRes, error) {
	stats.MarkMeter("handler.uplink.in")

	// 0. Check the packet integrity
	if len(req.Payload) == 0 {
		stats.MarkMeter("handler.uplink.invalid")
		return nil, errors.New(errors.Structural, "Invalid Packet Payload")
	}
	if len(req.DevEUI) != 8 {
		stats.MarkMeter("handler.uplink.invalid")
		return nil, errors.New(errors.Structural, "Invalid Device EUI")
	}
	if len(req.AppEUI) != 8 {
		stats.MarkMeter("handler.uplink.invalid")
		return nil, errors.New(errors.Structural, "Invalid Application EUI")
	}
	if req.Metadata == nil {
		stats.MarkMeter("handler.uplink.invalid")
		return nil, errors.New(errors.Structural, "Missing Mandatory Metadata")
	}
	stats.MarkMeter("handler.uplink.data")

	// 1. Lookup for the associated AppSKey + Application
	h.ctx.WithField("appEUI", req.AppEUI).WithField("devEUI", req.DevEUI).Debug("Perform lookup")
	entry, err := h.devices.Lookup(req.AppEUI, req.DevEUI)
	if err != nil {
		return nil, err
	}

	// 2. Prepare a channel to receive the response from the consumer
	chresp := make(chan interface{})

	// 3. Create a "bundle" which holds info waiting for other related packets
	var bundleID [20]byte // AppEUI(8) | DevEUI(8) | FCnt
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, req.AppEUI)
	binary.Write(buf, binary.BigEndian, req.DevEUI)
	binary.Write(buf, binary.BigEndian, req.FCnt)
	data := buf.Bytes()
	if len(data) != 20 {
		return nil, errors.New(errors.Structural, "Unable to generate bundleID")
	}
	copy(bundleID[:], data[:])

	// 4. Send the actual bundle to the consumer
	ctx := h.ctx.WithField("BundleID", bundleID)
	ctx.Debug("Define new bundle")
	h.set <- bundle{
		ID:     bundleID,
		Packet: *req,
		Entry:  entry,
		Chresp: chresp,
		Time:   time.Now(),
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
		return nil, resp.(error)
	default:
		stats.MarkMeter("handler.uplink.ack.without_response")
		ctx.Debug("No response to send.")
		return nil, nil
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
	ctx := h.ctx.WithField("goroutine", "bundle consumer")
	ctx.Debug("Starting bundle consumer")

browseBundles:
	for bundles := range chbundle {
		ctx.WithField("BundleID", bundles[0].ID).Debug("Consume new bundle")
		var metadata []*core.Metadata
		var payload []byte
		var firstTime time.Time

		if len(bundles) < 1 {
			continue browseBundles
		}
		b := bundles[0]

		computer, scores, err := dutycycle.NewScoreComputer(b.Packet.Metadata.DataRate) // Nil check already done
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
				payload, err = lorawan.DecryptFRMPayload(
					bundle.Packet.Payload,
					lorawan.DevAddr(bundle.Entry.DevAddr),
					bundle.Packet.FCnt,
					true,
					bundle.Entry.AppSKey,
				)
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
			metadata = append(metadata, bundle.Packet.Metadata)
			scores = computer.Update(scores, i, *bundle.Packet.Metadata) // Nil check already done
		}

		// Then create an application-level packet and send it to the wild open
		// we don't expect a response from the adapter, end of the chain.
		_, err = h.adapter.HandleData(context.Background(), &core.DataAppReq{
			AppEUI:   b.Packet.AppEUI,
			DevEUI:   b.Packet.DevEUI,
			Payload:  payload,
			Metadata: metadata,
		})
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		stats.MarkMeter("handler.uplink.out")

		// Now handle the downlink and respond to node
		h.ctx.Debug("Looking for downlink response")
		best := computer.Get(scores)
		h.ctx.WithField("Bundle", best).Debug("Determine best gateway")
		var downlink []byte
		if best != nil { // Avoid pulling when there's no gateway available for an answer
			downlink, err = h.packets.Pull(b.Packet.AppEUI, b.Packet.DevEUI)
		}
		if err != nil && err.(errors.Failure).Nature != errors.NotFound {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		// One of those bundle might be available for a response
		for i, bundle := range bundles {
			if best != nil && best.ID == i && downlink != nil && err == nil {
				stats.MarkMeter("handler.downlink.pull")

				downlink, err := h.buildDownlink(downlink, bundle.Packet, bundle.Entry, best.IsRX2)
				if err != nil {
					go h.abortConsume(errors.New(errors.Structural, err), bundles)
					continue browseBundles
				}
				err = h.devices.UpdateFCnt(b.Packet.AppEUI, b.Packet.DevEUI, downlink.Payload.MACPayload.FHDR.FCnt)
				if err != nil {
					go h.abortConsume(errors.New(errors.Structural, err), bundles)
					continue browseBundles
				}
				bundle.Chresp <- downlink
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
func (h component) buildDownlink(down []byte, up core.DataUpHandlerReq, entry devEntry, isRX2 bool) (*core.DataUpHandlerRes, error) {
	macpayload := lorawan.NewMACPayload(false)
	macpayload.FHDR = lorawan.FHDR{
		FCnt:    entry.FCntDown + 1,
		DevAddr: entry.DevAddr,
	}
	macpayload.FPort = 1
	macpayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: down}}

	if err := macpayload.EncryptFRMPayload(entry.AppSKey); err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	frmpayload, err := macpayload.FRMPayload[0].MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	payload := lorawan.NewPHYPayload(false)
	payload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataDown, // TODO Handle Confirmed data down
		Major: lorawan.LoRaWANR1,
	}
	payload.MACPayload = macpayload

	data, err := payload.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	metadata := core.Metadata{
		Frequency:   up.Metadata.Frequency,
		CodingRate:  up.Metadata.CodingRate,
		DataRate:    up.Metadata.DataRate,
		PayloadSize: uint32(len(data)),
		Timestamp:   up.Metadata.Timestamp + 1000,
	}

	if isRX2 { // Should we reply on RX2, metadata aren't the same
		// TODO Handle different regions with non hard-coded values
		metadata.Frequency = 869.50
		metadata.DataRate = "SF9BW125"
		metadata.Timestamp = up.Metadata.Timestamp + 2000
	}

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
				FPort:      uint32(macpayload.FPort),
				FRMPayload: frmpayload,
			},
			MIC: payload.MIC[:],
		},
		Metadata: &metadata,
	}, nil
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/parser"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
)

// Adapter type materializes an http adapter which implements the basic http protocol
type Adapter struct {
	parser.PacketParser                // The adapter's parser contract
	http.Client                        // Adapter is also an http client
	serveMux            *http.ServeMux // Holds a references to the adapter servemux in order to dynamically define endpoints
	packets             chan pktReq    // Channel used to "transforms" incoming request to something we can handle concurrently
	ctx                 log.Interface  // Just a logger, no one really cares about him.
}

// Message sent through the packets channel when an incoming request arrives
type pktReq struct {
	core.Packet             // The actual packet that has been parsed
	response    chan pktRes // A response channel waiting for an success or reject confirmation
}

// Message sent through the response channel of a pktReq
type pktRes struct {
	statusCode int    // The http status code to set as an answer
	content    []byte // The response content.
}

// NewAdapter constructs and allocates a new http adapter
func NewAdapter(port uint, parser parser.PacketParser, ctx log.Interface) (*Adapter, error) {
	a := Adapter{
		PacketParser: parser,
		serveMux:     http.NewServeMux(),
		packets:      make(chan pktReq),
		ctx:          ctx,
		Client:       http.Client{},
	}

	a.RegisterEndpoint("/packets", a.handlePostPacket)
	go a.listenRequests(port)

	return &a, nil
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, r ...core.Recipient) (core.Packet, error) {
	stats.MarkMeter("http_adapter.send")
	stats.UpdateHistogram("http_adapter.send_recipients", int64(len(r)))

	// Generate payload from core packet
	m, err := json.Marshal(p.Metadata)
	if err != nil {
		a.ctx.WithError(err).Warn("Invalid Packet")
		return core.Packet{}, errors.New(ErrInvalidStructure, err)
	}
	pl, err := p.Payload.MarshalBinary()
	if err != nil {
		a.ctx.WithError(err).Warn("Invalid Packet")
		return core.Packet{}, errors.New(ErrInvalidStructure, err)
	}
	payload := fmt.Sprintf(`{"payload":"%s","metadata":%s}`, base64.StdEncoding.EncodeToString(pl), m)

	devAddr, _ := p.DevAddr()
	ctx := a.ctx.WithField("devAddr", devAddr)
	ctx.Debug("Sending Packet")

	// Prepare ground for parrallel http request
	nb := len(r)
	cherr := make(chan error, nb)
	chresp := make(chan core.Packet, nb)
	wg := sync.WaitGroup{}
	wg.Add(nb)

	// Run each request
	for _, recipient := range r {
		go func(recipient core.Recipient) {
			defer wg.Done()

			ctx := ctx.WithField("recipient", recipient.Address)
			ctx.Debug("POST Request")

			buf := new(bytes.Buffer)
			buf.Write([]byte(payload))

			// Send request
			resp, err := a.Post(fmt.Sprintf("http://%s", recipient.Address.(string)), "application/json", buf)
			if err != nil {
				cherr <- err
				return
			}

			defer func() {
				// This is needed because the default HTTP client's Transport does not
				// attempt to reuse HTTP/1.0 or HTTP/1.1 TCP connections unless the Body
				// is read to completion and is closed.
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()
			}()

			// Check response code
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
				ctx.WithField("response", resp.StatusCode).Warn("Unexpected response")
				cherr <- fmt.Errorf("Unexpected response from server: %s (%d)", resp.Status, resp.StatusCode)
				return
			}

			// Process response body
			raw, err := ioutil.ReadAll(resp.Body)
			if err != nil && err != io.EOF {
				cherr <- err
				return
			}

			// Process packet
			var packet core.Packet
			if err := json.Unmarshal(raw, &packet); err != nil {
				cherr <- err
				return
			}

			chresp <- packet
		}(recipient)
	}

	// Wait for each request to be done, and return
	stats.IncCounter("http_adapter.waiting_for_send")
	wg.Wait()
	stats.DecCounter("http_adapter.waiting_for_send")

	// Collect errors
	var errs []error
	for i := 0; i < len(cherr); i += 1 {
		err := <-cherr
		ctx.WithError(err).Error("POST Failed")
		errs = append(errs, err)
	}

	// Check responses
	if len(chresp) > 1 {
		ctx.WithField("response_count", len(chresp)).Error("Received too many positive answers")
		return core.Packet{}, errors.New(ErrWrongBehavior, "Received too many positive answers")
	}

	// Get packet
	select {
	case packet := <-chresp:
		return packet, nil
	default:
		if errs != nil {
			return core.Packet{}, errors.New(ErrFailedOperation, fmt.Sprintf("%+v", errs))
		}
		ctx.Error("No response packet available")
		return core.Packet{}, errors.New(ErrWrongBehavior, "No response packet available")
	}
}

// RegisterEndpoint can be used by an external agent to register a handler to the adapter servemux
func (a *Adapter) RegisterEndpoint(url string, handler func(w http.ResponseWriter, req *http.Request)) {
	a.ctx.WithField("url", url).Info("Register new endpoint")
	a.serveMux.HandleFunc(url, handler)
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() (core.Packet, core.AckNacker, error) {
	pktReq := <-a.packets
	return pktReq.Packet, packetAckNacker{response: pktReq.response}, nil
}

// NextRegistration implements the core.Adapter interface. Not implemented for this adapter.
//
// See broadcast and pubsub adapters for mechanisms to handle registrations.
func (a *Adapter) NextRegistration() (core.Packet, core.AckNacker, error) {
	return core.Packet{}, nil, errors.New(ErrNotSupported, "NextRegistration not supported for http adapter")
}

// listenRequests handles incoming registration request sent through http to the adapter
func (a *Adapter) listenRequests(port uint) {
	server := http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: a.serveMux,
	}
	a.ctx.WithField("port", port).Info("Starting Server")
	err := server.ListenAndServe()
	a.ctx.WithError(err).Warn("HTTP connection lost")
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/parser"
	"github.com/apex/log"
)

var ErrInvalidPort = fmt.Errorf("The given port is invalid")
var ErrInvalidPacket = fmt.Errorf("The given packet is invalid")
var ErrNotImplemented = fmt.Errorf("Illegal call on non-implemented method")

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
	// Generate payload from core packet
	m, err := json.Marshal(p.Metadata)
	if err != nil {
		a.ctx.WithError(err).Warn("Invalid Packet")
		return core.Packet{}, ErrInvalidPacket
	}
	pl, err := p.Payload.MarshalBinary()
	if err != nil {
		a.ctx.WithError(err).Warn("Invalid Packet")
		return core.Packet{}, ErrInvalidPacket
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

			ctx := ctx.WithField("recipient", recipient)
			ctx.Debug("POST Request")

			buf := new(bytes.Buffer)
			buf.Write([]byte(payload))

			// Send request
			resp, err := a.Post(fmt.Sprintf("http://%s", recipient.Address.(string)), "application/json", buf)
			if err != nil {
				cherr <- err
				return
			}

			// Check response code
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
				ctx.WithField("response", resp.StatusCode).Warn("Unexpected response")
				cherr <- fmt.Errorf("Unexpected response from server: %s (%d)", resp.Status, resp.StatusCode)
				return
			}

			// Process response body
			raw := make([]byte, resp.ContentLength)
			n, err := resp.Body.Read(raw)
			defer resp.Body.Close()
			if err != nil && err != io.EOF {
				cherr <- err
				return
			}

			// Process packet
			var packet core.Packet
			if err := json.Unmarshal(raw[:n], &packet); err != nil {
				cherr <- err
				return
			}

			chresp <- packet
		}(recipient)
	}

	// Wait for each request to be done, and return
	wg.Wait()

	// Collect errors
	var errors []error
	for i := 0; i < len(cherr); i += 1 {
		err := <-cherr
		ctx.WithError(err).Error("POST Failed")
		errors = append(errors, err)
	}

	// Check responses
	if len(chresp) > 1 {
		ctx.WithField("response_count", len(chresp)).Error("Received Too many positive answers")
		return core.Packet{}, fmt.Errorf("Several positive answer from servers")
	}

	// Get packet
	select {
	case packet := <-chresp:
		return packet, nil
	default:
		if errors != nil {
			return core.Packet{}, fmt.Errorf("Errors: %v", errors)
		}
		ctx.Error("No response packet available")
		return core.Packet{}, fmt.Errorf("No response packet available")
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
	return core.Packet{}, nil, ErrNotImplemented
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

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	core "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

// Adapter type materializes an http adapter which implements the basic http protocol
type Adapter struct {
	http.Client                    // Adapter is also an http client
	ctx           log.Interface    // Just a logger, no one really cares about him.
	packets       chan PktReq      // Channel used to "transforms" incoming request to something we can handle concurrently
	recipients    []core.Recipient // Known recipient used for broadcast if any
	registrations chan RegReq      // Incoming registrations
	serveMux      *http.ServeMux   // Holds a references to the adapter servemux in order to dynamically define endpoints
}

// Handler defines endpoint-specific handler.
type Handler interface {
	Url() string
	Handle(w http.ResponseWriter, reg chan<- RegReq, req *http.Request)
}

// Message sent through the response channel of a pktReq or regReq
type MsgRes struct {
	StatusCode int    // The http status code to set as an answer
	Content    []byte // The response content.
}

// Message sent through the packets channel when an incoming request arrives
type PktReq struct {
	Packet []byte      // The actual packet that has been parsed
	Chresp chan MsgRes // A response channel waiting for an success or reject confirmation
}

// Message sent through the registration channel when an incoming registration arrives
type RegReq struct {
	Registration core.Registration
	Chresp       chan MsgRes
}

// NewAdapter constructs and allocates a new http adapter
func NewAdapter(port uint, recipients []core.Recipient, ctx log.Interface) (*Adapter, error) {
	a := Adapter{
		Client:        http.Client{},
		ctx:           ctx,
		packets:       make(chan PktReq),
		recipients:    recipients,
		registrations: make(chan RegReq),
		serveMux:      http.NewServeMux(),
	}

	go a.listenRequests(port)

	return &a, nil
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, recipients ...core.Recipient) ([]byte, error) {
	stats.MarkMeter("http_adapter.send")
	stats.UpdateHistogram("http_adapter.send_recipients", int64(len(recipients)))

	// Marshal the packet to raw binary data
	data, err := p.MarshalBinary()
	if err != nil {
		a.ctx.WithError(err).Warn("Invalid Packet")
		return nil, errors.New(ErrInvalidStructure, err)
	}

	// Try to define a more helpful context
	var devEUI *lorawan.EUI64
	var ctx log.Interface
	addressable, ok := p.(core.Addressable)
	if ok {
		if d, err := addressable.DevEUI(); err == nil {
			ctx = a.ctx.WithField("devEUI", devEUI)
			devEUI = &d
		}
	}
	if devEUI == nil {
		a.ctx.Warn("Unable to retrieve devEUI")
		ctx = a.ctx
	}
	ctx.Debug("Sending Packet")

	// Determine whether it's a broadcast or a direct send
	nb := len(recipients)
	isBroadcast := false
	if nb == 0 {
		// If no recipient was supplied, try with the known one, otherwise quit.
		recipients = a.recipients
		nb = len(recipients)
		isBroadcast = true
		if nb == 0 {
			return nil, errors.New(ErrFailedOperation, "No recipient found")
		}
	}

	// Prepare ground for parrallel http request
	cherr := make(chan error, nb)
	chresp := make(chan []byte, nb)
	wg := sync.WaitGroup{}
	wg.Add(nb)

	// Run each request
	for _, recipient := range recipients {
		go func(rawRecipient core.Recipient) {
			defer wg.Done()

			// Get the actual recipient
			recipient, ok := rawRecipient.(HttpRecipient)
			if !ok {
				ctx.WithField("recipient", rawRecipient).Warn("Unable to interpret recipient as httpRecipient")
				return
			}
			ctx := ctx.WithField("recipient", recipient.Url())

			// Send request
			ctx.Debugf("%s Request", recipient.Method())
			buf := new(bytes.Buffer)
			buf.Write(data)
			resp, err := a.Post(fmt.Sprintf("http://%s", recipient.Url()), "application/octet-stream", buf)
			if err != nil {
				cherr <- errors.New(ErrFailedOperation, err)
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
			switch resp.StatusCode {
			case http.StatusOK:
				ctx.Debug("Recipient registered for packet")
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil && err != io.EOF {
					cherr <- errors.New(ErrFailedOperation, err)
					return
				}
				chresp <- data
				if isBroadcast { // Generate registration on broadcast
					go func() {
						a.registrations <- RegReq{
							Registration: httpRegistration{
								recipient: rawRecipient,
								devEUI:    devEUI,
							},
							Chresp: nil,
						}
					}()
				}
			case http.StatusNotFound:
				ctx.Debug("Recipient not interested in packet")
				cherr <- errors.New(ErrWrongBehavior, "Recipient not interested")
			default:
				cherr <- errors.New(ErrFailedOperation, fmt.Sprintf("Unexpected response from server: %s (%d)", resp.Status, resp.StatusCode))
			}
		}(recipient)
	}

	// Wait for each request to be done
	stats.IncCounter("http_adapter.waiting_for_send")
	wg.Wait()
	stats.DecCounter("http_adapter.waiting_for_send")
	close(cherr)
	close(chresp)

	// Collect errors and see if everything went well
	var errored uint8
	for i := 0; i < len(cherr); i += 1 {
		err := <-cherr
		if err.(errors.Failure).Nature != ErrWrongBehavior {
			errored += 1
			ctx.WithError(err).Error("POST Failed")
		}
	}

	// Collect response
	if len(chresp) > 1 {
		return nil, errors.New(ErrWrongBehavior, "Received too many positive answers")
	}

	if len(chresp) == 0 && errored != 0 {
		return nil, errors.New(ErrFailedOperation, "No positive response from recipients but got unexpected answer")
	}

	if len(chresp) == 0 && errored == 0 {
		return nil, errors.New(ErrWrongBehavior, "No recipient gave a positive answer")
	}

	return <-chresp, nil
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() ([]byte, core.AckNacker, error) {
	p := <-a.packets
	return p.Packet, httpAckNacker{Chresp: p.Chresp}, nil
}

// NextRegistration implements the core.Adapter interface. Not implemented for this adapter.
//
// See broadcast and pubsub adapters for mechanisms to handle registrations.
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	r := <-a.registrations
	return r.Registration, regAckNacker{Chresp: r.Chresp}, nil
}

// Bind registers a handler to a specific endpoint
func (a *Adapter) Bind(h Handler) {
	a.ctx.WithField("url", h.Url()).Info("Register new endpoint")
	a.serveMux.HandleFunc(h.Url(), func(w http.ResponseWriter, req *http.Request) {
		a.ctx.WithField("url", h.Url()).Debug("Handle new request")
		h.Handle(w, a.registrations, req)
	})
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

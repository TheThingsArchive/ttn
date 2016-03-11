// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
)

// Adapter type materializes an http adapter which implements the basic http protocol
type Adapter struct {
	http.Client                    // Adapter is also an http client
	ctx           log.Interface    // Just a logger, no one really cares about him.
	packets       chan PktReq      // Channel used to "transforms" incoming request to something we can handle concurrently
	recipients    []core.Recipient // Known recipient used for broadcast if any
	registrations chan RegReq      // Incoming registrations
	serveMux      *http.ServeMux   // Holds a references to the adapter servemux in order to dynamically define endpoints
	net           string           // Address on which is listening the adapter http server
}

// Handler defines endpoint-specific handler.
type Handler interface {
	URL() string
	Handle(w http.ResponseWriter, chpkt chan<- PktReq, chreg chan<- RegReq, req *http.Request) error
}

// MsgRes are sent through the response channel of a pktReq or regReq
type MsgRes struct {
	StatusCode int    // The http status code to set as an answer
	Content    []byte // The response content.
}

// PktReq are sent through the packets channel when an incoming request arrives
type PktReq struct {
	Packet []byte      // The actual packet that has been parsed
	Chresp chan MsgRes // A response channel waiting for an success or reject confirmation
}

// RegReq are sent through the registration channel when an incoming registration arrives
type RegReq struct {
	Registration core.Registration
	Chresp       chan MsgRes
}

// NewAdapter constructs and allocates a new http adapter
func NewAdapter(net string, recipients []core.Recipient, ctx log.Interface) (*Adapter, error) {
	a := Adapter{
		Client:        http.Client{Timeout: 6 * time.Second},
		ctx:           ctx,
		packets:       make(chan PktReq),
		recipients:    recipients,
		registrations: make(chan RegReq),
		serveMux:      http.NewServeMux(),
		net:           net,
	}

	go a.listenRequests(net)

	return &a, nil
}

// Register implements the core.Subscriber interface
func (a *Adapter) Subscribe(r core.Registration) error {
	// 1. Type assertions and convertions
	jsonMarshaler, ok := r.(json.Marshaler)
	if !ok {
		return errors.New(errors.Structural, "Unable to marshal registration")
	}
	httpRecipient, ok := r.Recipient().(Recipient)
	if !ok {
		return errors.New(errors.Structural, "Invalid recipient")
	}

	// 2. Marshaling
	data, err := json.Marshal(struct {
		Recipient struct {
			Method string `json:"method"`
			URL    string `json:"url"`
		} `json:"recipient"`
		Registration json.Marshaler `json:"registration"`
	}{
		Recipient: struct {
			Method string `json:"method"`
			URL    string `json:"url"`
		}{
			Method: "POST",
			URL:    a.net,
		},
		Registration: jsonMarshaler,
	})
	if err != nil {
		return errors.New(errors.Structural, err)
	}
	buf := new(bytes.Buffer)
	buf.Write(data)

	// 3. Send Request
	req, err := http.NewRequest(httpRecipient.Method(), fmt.Sprintf("http://%s/end-devices", httpRecipient.URL()), buf)
	if err != nil {
		return errors.New(errors.Operational, err)
	}
	req.Header.Add("content-type", "application/json")
	resp, err := a.Do(req)
	if err != nil {
		return errors.New(errors.Operational, err)
	}
	defer resp.Body.Close()

	// 4. Handle response -> resp body isn't relevant
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		errData := make([]byte, resp.ContentLength)
		resp.Body.Read(errData)
		return errors.New(errors.Operational, string(errData))
	}
	return nil
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, recipients ...core.Recipient) ([]byte, error) {
	// Marshal the packet to raw binary data
	data, err := p.MarshalBinary()
	if err != nil {
		a.ctx.WithError(err).Warn("Invalid Packet")
		return nil, errors.New(errors.Structural, err)
	}

	// Try to define a more helpful context
	ctx := a.ctx.WithField("devEUI", p.DevEUI())
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
			return nil, errors.New(errors.Structural, "No recipient found")
		}
	}

	if isBroadcast {
		stats.MarkMeter("http_adapter.broadcast")
	} else {
		stats.MarkMeter("http_adapter.send")
	}
	stats.UpdateHistogram("http_adapter.send_recipients", int64(len(recipients)))

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
			recipient, ok := rawRecipient.(Recipient)
			if !ok {
				ctx.WithField("recipient", rawRecipient).Warn("Unable to interpret recipient as Recipient")
				return
			}
			ctx := ctx.WithField("recipient", recipient.URL())

			// Send request
			ctx.Debugf("%s Request", recipient.Method())
			buf := new(bytes.Buffer)
			buf.Write(data)
			req, err := http.NewRequest(recipient.Method(), fmt.Sprintf("http://%s/packets", recipient.URL()), buf)
			if err != nil {
				cherr <- errors.New(errors.Operational, err)
				return
			}
			req.Header.Add("content-type", "application/octet-stream")
			resp, err := a.Do(req)

			if err != nil {
				cherr <- errors.New(errors.Operational, err)
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
					cherr <- errors.New(errors.Operational, err)
					return
				}
				if len(data) > 0 {
					chresp <- data
				}
				if isBroadcast { // Generate registration on broadcast
					go func() {
						a.registrations <- RegReq{
							Registration: httpRegistration{
								recipient: rawRecipient,
								devEUI:    p.DevEUI(),
							},
							Chresp: nil,
						}
					}()
				}
			case http.StatusNotFound:
				ctx.Debug("Recipient not interested in packet")
				cherr <- errors.New(errors.NotFound, "Recipient not interested")
			default:
				cherr <- errors.New(errors.Operational, fmt.Sprintf("Unexpected response from server: %s (%d)", resp.Status, resp.StatusCode))
			}
		}(recipient)
	}

	// Wait for each request to be done
	stats.IncCounter("http_adapter.waiting_for_send")
	wg.Wait()
	stats.DecCounter("http_adapter.waiting_for_send")
	close(cherr)
	close(chresp)

	var errored uint8
	var notFound uint8
	for i := 0; i < len(cherr); i++ {
		err := <-cherr
		if err.(errors.Failure).Nature != errors.NotFound {
			errored++
			ctx.WithError(err).Warn("POST Failed")
		} else {
			notFound++
			ctx.WithError(err).Debug("Packet destination not found")
		}
	}

	// Collect response
	if len(chresp) > 1 {
		return nil, errors.New(errors.Behavioural, "Received too many positive answers")
	}

	if len(chresp) == 0 && errored > 0 {
		return nil, errors.New(errors.Operational, "No positive response from recipients but got unexpected answer")
	}

	if len(chresp) == 0 && notFound > 0 {
		return nil, errors.New(errors.NotFound, "No available recipient found")
	}

	if len(chresp) == 0 {
		return nil, nil
	}
	return <-chresp, nil
}

// GetRecipient implements the core.Adapter interface
func (a *Adapter) GetRecipient(raw []byte) (core.Recipient, error) {
	recipient := new(recipient)
	if err := recipient.UnmarshalBinary(raw); err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	return *recipient, nil
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
	a.ctx.WithField("url", h.URL()).Info("Register new endpoint")
	a.serveMux.HandleFunc(h.URL(), func(w http.ResponseWriter, req *http.Request) {
		ctx := a.ctx.WithField("url", h.URL())
		ctx.Debug("Handle new request")
		err := h.Handle(w, a.packets, a.registrations, req)
		if err != nil {
			ctx.WithError(err).Debug("Failed to handle the request")
		}
	})
}

// listenRequests handles incoming registration request sent through http to the adapter
func (a *Adapter) listenRequests(net string) {
	server := http.Server{
		Addr:    net,
		Handler: a.serveMux,
	}
	a.ctx.WithField("bind", net).Info("Starting Server")
	err := server.ListenAndServe()
	a.ctx.WithError(err).Warn("HTTP connection lost")
}

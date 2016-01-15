// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pubsub

import (
	"net/http"

	"github.com/TheThingsNetwork/ttn/core"
	httpadapter "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/apex/log"
)

type Adapter struct {
	*httpadapter.Adapter
	Parser
	ctx           log.Interface
	registrations chan regReq
}

type Parser interface {
	Parse(req *http.Request) (core.Registration, error)
}

type regReq struct {
	core.Registration             // The actual registration request
	response          chan regRes // A dedicated channel to send back a response (ack or nack)
}

type regRes struct {
	statusCode int    // The response status, 200 for ack 4xx for nack
	content    []byte // The response content
}

// NewAdapter constructs a new http adapter that also handle registrations via http requests
func NewAdapter(adapter *httpadapter.Adapter, parser Parser, ctx log.Interface) (*Adapter, error) {
	a := &Adapter{
		Adapter:       adapter,
		Parser:        parser,
		ctx:           ctx,
		registrations: make(chan regReq),
	}

	// So far we only supports one endpoint [PUT] /end-device/:devAddr
	a.RegisterEndpoint("/end-devices/", a.handlePutEndDevice)

	return a, nil
}

// NextRegistration implements the core.Adapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	request := <-a.registrations
	return request.Registration, regAckNacker{response: request.response}, nil
}

// handle request [PUT] on /end-device/:devAddr
func (a *Adapter) handlePutEndDevice(w http.ResponseWriter, req *http.Request) {
	ctx := a.Ctx.WithField("sender", req.RemoteAddr)
	ctx.Debug("Receiving new registration request")

	// Check the http method
	if req.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Unreckognized HTTP method. Please use [PUT] to register a device"))
		return
	}

	// Parse body and query params
	config, err := a.Parse(req)
	if err != nil {
		ctx.WithError(err).Warn("Received invalid request")
		httpadapter.BadRequest(w, err.Error())
		return
	}

	// Send the registration and wait for ack / nack
	response := make(chan regRes)
	a.registrations <- regReq{Registration: config, response: response}
	r, ok := <-response
	if !ok {
		ctx.Error("Core server not responding")
		httpadapter.BadRequest(w, "Core server not responding")
		return
	}
	w.WriteHeader(r.statusCode)
	w.Write(r.content)
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pubsub

import (
	"fmt"
	"net/http"

	"github.com/TheThingsNetwork/ttn/core"
	httpadapter "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/utils/log"
)

type Adapter struct {
	*httpadapter.Adapter
	Parser
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
func NewAdapter(port uint, parser Parser, loggers ...log.Logger) (*Adapter, error) {
	adapter, err := httpadapter.NewAdapter(loggers...)
	if err != nil {
		return nil, err
	}

	a := &Adapter{
		Adapter:       adapter,
		Parser:        parser,
		registrations: make(chan regReq),
	}

	go a.listenRegistration(port)

	return a, nil
}

// NextRegistration implements the core.Adapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	request := <-a.registrations
	return request.Registration, regAckNacker{response: request.response}, nil
}

// listenRegistration handles incoming registration request sent through http to the adapter
func (a *Adapter) listenRegistration(port uint) {
	// Create a server multiplexer to handle request
	serveMux := http.NewServeMux()

	// So far we only supports one endpoint [PUT] /end-device/:devAddr
	serveMux.HandleFunc("/end-device/", a.handlePutEndDevice)

	server := http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: serveMux,
	}
	a.Logf("Start listening on %d", port)
	err := server.ListenAndServe()
	a.Logf("HTTP connection lost: %v", err)
}

// fail logs the given failure and sends an appropriate response to the client
func (a *Adapter) badRequest(w http.ResponseWriter, msg string) {
	a.Logf("registration request rejected: %s", msg)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(msg))
}

// handle request [PUT] on /end-device/:devAddr
func (a *Adapter) handlePutEndDevice(w http.ResponseWriter, req *http.Request) {
	a.Logf("Receive new registration request")
	// Check the http method
	if req.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Unreckognized HTTP method. Please use [PUT] to register a device"))
		return
	}

	// Check Content-type
	if req.Header.Get("Content-Type") != "application/json" {
		a.badRequest(w, "Incorrect content type")
		return
	}

	// Parse body and query params
	config, err := a.Parse(req)
	if err != nil {
		a.badRequest(w, err.Error())
		return
	}

	// Send the registration and wait for ack / nack
	response := make(chan regRes)
	a.registrations <- regReq{Registration: config, response: response}
	r, ok := <-response
	if !ok {
		a.badRequest(w, "Core server not responding")
		return
	}
	w.WriteHeader(r.statusCode)
	w.Write(r.content)
}

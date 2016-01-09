// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package httpregister

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"net/http"
	"time"
)

type regReq struct {
	core.Registration             // The actual registration request
	response          chan regRes // A dedicated channel to send back a response (ack or nack)
}

type regRes struct {
	statusCode int    // The response status, 200 for ack 4xx for nack
	content    []byte // The response content
}

type regAckNacker struct {
	response chan regRes // A channel dedicated to send back a response
}

// Ack implements the core.Acker interface
func (r regAckNacker) Ack(p core.Packet) error {
	select {
	case r.response <- regRes{statusCode: http.StatusOK}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return ErrConnectionLost
	}
}

// Nack implements the core.Nacker interface
func (r regAckNacker) Nack(p core.Packet) error {
	select {
	case r.response <- regRes{
		statusCode: http.StatusConflict,
		content:    []byte("Unable to register the given device"),
	}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return ErrConnectionLost
	}
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
	a.log("Start listening on %d", port)
	err := server.ListenAndServe()
	a.log("HTTP connection lost: %v", err)
}

// fail logs the given failure and sends an appropriate response to the client
func (a *Adapter) badRequest(w http.ResponseWriter, msg string) {
	a.log("registration request rejected: %s", msg)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(msg))
}

// handle request [PUT] on /end-device/:devAddr
func (a *Adapter) handlePutEndDevice(w http.ResponseWriter, req *http.Request) {
	a.log("Receive new registration request")
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

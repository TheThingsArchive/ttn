// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package brk_hdl_http

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/thethingsnetwork/core"
	"io"
	"net/http"
	"regexp"
	"strings"
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

// listenRegistration handles incoming registration request send through http to the broker
func (a *Adapter) listenRegistration(port uint) {
	// Create a server multiplexer to handle request
	serveMux := http.NewServeMux()

	// So far we only supports one endpoint [PUT] /end-device/:devAddr
	serveMux.HandleFunc("/end-device/", a.handlePostEndDevice)

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
func (a *Adapter) handlePostEndDevice(w http.ResponseWriter, req *http.Request) {
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

	// Check the query parameter
	reg := regexp.MustCompile("end-device/([a-fA-F0-9]{8})$")
	query := reg.FindStringSubmatch(req.RequestURI)
	if len(query) < 2 {
		a.badRequest(w, "Incorrect end-device address format")
		return
	}
	devAddr, err := hex.DecodeString(query[1])
	if err != nil {
		a.badRequest(w, "Incorrect end-device address format")
		return
	}

	// Check configuration in body
	body := make([]byte, 256)
	n, err := req.Body.Read(body)
	if err != nil && err != io.EOF {
		a.badRequest(w, "Incorrect request body")
		return
	}
	params := &struct {
		Id     string `json:"app_id"`
		Url    string `json:"app_url"`
		NwsKey string `json:"nws_key"`
	}{}
	if err := json.Unmarshal(body[:n], params); err != nil {
		a.badRequest(w, "Incorrect body payload")
		return
	}

	nwsKey, err := hex.DecodeString(params.NwsKey)
	if err != nil || len(nwsKey) != 16 {
		a.badRequest(w, "Incorrect network sesssion key")
		return
	}

	params.Id = strings.Trim(params.Id, " ")
	params.Url = strings.Trim(params.Url, " ")
	if len(params.Id) <= 0 {
		a.badRequest(w, "Incorrect application id")
		return
	}
	if len(params.Url) <= 0 {
		a.badRequest(w, "Incorrect application url")
		return
	}

	// Create registration
	config := core.Registration{
		Handler: core.Recipient{Id: params.Id, Address: params.Url},
	}
	copy(config.NwsKey[:], nwsKey)
	copy(config.DevAddr[:], devAddr)

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

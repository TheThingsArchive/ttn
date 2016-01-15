// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/TheThingsNetwork/ttn/core"
)

func (a *Adapter) handlePostPacket(w http.ResponseWriter, req *http.Request) {
	ctx := a.Ctx.WithField("sender", req.RemoteAddr)

	ctx.Debug("Receiving new registration request")
	// Check the http method
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Unreckognized HTTP method. Please use [POST] to transfer a packet"))
		return
	}

	// Parse body and query params
	packet, err := a.Parse(req)
	if err != nil {
		ctx.WithError(err).Warn("Received invalid body in request")
		BadRequest(w, err.Error())
		return
	}

	// Send the packet and wait for ack / nack
	response := make(chan pktRes)
	a.packets <- pktReq{Packet: packet, response: response}
	r, ok := <-response
	if !ok {
		ctx.Error("Core server not responding")
		BadRequest(w, "Core server not responding")
		return
	}
	w.WriteHeader(r.statusCode)
	w.Write(r.content)
}

type JSONPacketParser struct{}

func (p JSONPacketParser) Parse(req *http.Request) (core.Packet, error) {
	// Check Content-type
	if req.Header.Get("Content-Type") != "application/json" {
		return core.Packet{}, fmt.Errorf("Received invalid content-type in request")
	}

	// Check configuration in body
	body := make([]byte, req.ContentLength)
	n, err := req.Body.Read(body)
	if err != nil && err != io.EOF {
		return core.Packet{}, err
	}
	packet := new(core.Packet)
	if err := json.Unmarshal(body[:n], packet); err != nil {
		return core.Packet{}, err
	}

	return *packet, nil
}

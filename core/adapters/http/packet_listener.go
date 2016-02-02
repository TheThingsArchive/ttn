// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"net/http"
)

// handlePostPacket defines an http handler over the adapter to handle POST request on /packets
func (a *Adapter) handlePostPacket(w http.ResponseWriter, req *http.Request) {
	ctx := a.ctx.WithField("sender", req.RemoteAddr)

	ctx.Debug("Receiving new packet")
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

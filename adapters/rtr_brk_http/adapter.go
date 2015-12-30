// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package rtr_brk_http
//
// Assume one endpoint url accessible through a POST http request
package rtr_brk_http

import (
	"bytes"
	"encoding/json"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/utils/log"
	"net/http"
)

type Adapter struct {
	Logger log.Logger
}

// New constructs a new Router-Broker-HTTP adapter
func New(router core.Router, port uint, broAddrs ...core.BrokerAddress) (*Adapter, error) {
	return nil, nil
}

// Connect implements the core.BrokerRouter interface
func (a *Adapter) Connect(router core.Router, port uint, broAddrs ...core.BrokerAddress) {
	a.log("Connects to router %+v", router)
}

// Broadcast implements the core.BrokerRouter interface
func (a *Adapter) Broadcast(packet semtech.Packet) {
}

// Forward implements the core.BrokerRouter interface
func (a *Adapter) Forward(router core.Router, packet semtech.Packet, broAddrs ...core.BrokerAddress) {
	if packet.Payload == nil || len(packet.Payload.RXPK) == 0 {
		a.log("Ignores irrelevant packet %+v", packet) // NOTE Should we trigger an error here ?
		return
	}

	client := http.Client{}
	for _, addr := range broAddrs {
		go func() {
			data := new(bytes.Buffer)
			rawJSON, err := json.Marshal(packet.Payload)
			if err != nil {
				a.log("Unable to marshal payload %+v", err)
				router.HandleError(core.ErrForward(err))
				return
			}

			_, err = data.Write(rawJSON)

			if err != nil {
				a.log("Unable to write raw JSON in buffer %+v", err)
				router.HandleError(core.ErrForward(err))
				return
			}

			resp, err := client.Post(string(addr), "application/json", data)

			if err != nil {
				a.log("Unable to send POST request %+v", err)
				router.HandleError(core.ErrForward(err))
				return
			}

			if resp.StatusCode != http.StatusOK {
				a.log("Unexpected answer from the broker %+v", err)
				router.HandleError(core.ErrForward(err))
				return
			}

			// NOTE Do We Care about the response ? The router is supposed to handle HTTP request
			// from the broker to handle packets or anything else ? Is it efficient ? Should
			// downlinks packets be sent back with the HTTP body response ? Its a 2 seconds frame...

			resp.Body.Close()
		}()
	}
}

// log is nothing more than a shortcut / helper to access the logger
func (a Adapter) log(format string, i ...interface{}) {
	if a.Logger == nil {
		return
	}
	a.Logger.Log(format, i...)
}

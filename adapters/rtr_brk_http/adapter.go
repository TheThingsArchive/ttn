// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package rtr_brk_http
//
// Assume one endpoint url accessible through a POST http request
package rtr_brk_http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/utils/log"
	"net/http"
	"sync"
)

type Adapter struct {
	Logger  log.Logger
	brokers []core.BrokerAddress
}

// New constructs a new Router-Broker-HTTP adapter
func New(router core.Router, port uint, broAddrs ...core.BrokerAddress) (*Adapter, error) {
	adapter := Adapter{
		Logger:  log.VoidLogger{},
		brokers: broAddrs,
	}

	if err := adapter.Connect(router, port, broAddrs...); err != nil {
		return nil, err
	}

	return &adapter, nil
}

// Connect implements the core.BrokerRouter interface
func (a *Adapter) Connect(router core.Router, port uint, broAddrs ...core.BrokerAddress) error {
	a.log("Connects to router %+v", router)
	return nil
}

// Broadcast implements the core.BrokerRouter interface
func (a *Adapter) Broadcast(router core.Router, payload semtech.Payload) {
	// Determine the devAddress associated to that payload
	if payload.RXPK == nil || len(payload.RXPK) == 0 { // NOTE are those conditions significantly different ?
		router.HandleError(core.ErrBroadcast(fmt.Errorf("Cannot broadcast given payload: %+v", payload)))
		return
	}
	var devAddr semtech.DeviceAddress
	var defaultDevAddr semtech.DeviceAddress
	// We check them all to be sure, but all RXPK should refer to the same End-Device
	for _, rxpk := range payload.RXPK {
		addr := rxpk.DevAddr()
		if addr == nil || (devAddr != defaultDevAddr && devAddr != *addr) {
			router.HandleError(core.ErrBroadcast(fmt.Errorf("Cannot broadcast given payload: %+v", payload)))
			return
		}
		devAddr = *addr
	}

	// Prepare ground to store brokers that are in charge
	register := make(chan core.BrokerAddress, len(a.brokers))
	wg := sync.WaitGroup{}
	wg.Add(len(a.brokers))

	client := http.Client{}
	for _, addr := range a.brokers {
		go func(addr core.BrokerAddress) {
			defer wg.Done()

			resp, err := post(client, string(addr), payload)

			if err != nil {
				a.log("Unable to send POST request %+v", err)
				router.HandleError(core.ErrBroadcast(err))
				return
			}

			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				a.log("Broker %+v handles packets coming from %+v", addr, devAddr)
				register <- addr
			case http.StatusNotFound: //NOTE Convention with the broker
				a.log("Broker %+v does not handle packets coming from %+v", addr, devAddr)
			default:
				a.log("Unexpected answer from the broker %+v", err)
				router.HandleError(core.ErrBroadcast(err))
			}
		}(addr)
	}

	go func() {
		wg.Wait()
		close(register)
		brokers := make([]core.BrokerAddress, 0)
		for addr := range register {
			brokers = append(brokers, addr)
		}
		if len(brokers) > 0 {
			router.RegisterDevice(devAddr, brokers...)
		}
	}()
}

// Forward implements the core.BrokerRouter interface
func (a *Adapter) Forward(router core.Router, payload semtech.Payload, broAddrs ...core.BrokerAddress) {
	client := http.Client{}
	for _, addr := range broAddrs {
		go func(url string) {
			resp, err := post(client, url, payload)

			if err != nil {
				a.log("Unable to send POST request %+v", err)
				router.HandleError(core.ErrForward(err))
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				a.log("Unexpected answer from the broker %+v", err)
				router.HandleError(core.ErrForward(err))
				return
			}

			// NOTE Do We Care about the response ? The router is supposed to handle HTTP request
			// from the broker to handle packets or anything else ? Is it efficient ? Should
			// downlinks packets be sent back with the HTTP body response ? Its a 2 seconds frame...

		}(string(addr))
	}
}

func post(client http.Client, url string, payload semtech.Payload) (*http.Response, error) {
	data := new(bytes.Buffer)
	rawJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if _, err := data.Write(rawJSON); err != nil {
		return nil, err
	}

	return client.Post(url, "application/json", data)
}

// log is nothing more than a shortcut / helper to access the logger
func (a Adapter) log(format string, i ...interface{}) {
	if a.Logger == nil {
		return
	}
	a.Logger.Log(format, i...)
}

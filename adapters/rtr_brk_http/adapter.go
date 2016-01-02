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
	mu      *sync.RWMutex // Guard brokers
}

func NewAdapter() Adapter {
	return Adapter{mu: &sync.RWMutex{}}
}

func (a *Adapter) ok() bool {
	return a != nil && a.mu != nil
}

// Listen implements the core.Adapter interface
func (a *Adapter) Listen(router core.Router, options interface{}) error {
	if !a.ok() {
		return core.ErrNotInitialized
	}

	switch options.(type) {
	case []core.BrokerAddress:
		a.mu.Lock()
		defer a.mu.Unlock()
		a.brokers = options.([]core.BrokerAddress)
		if len(a.brokers) == 0 {
			return core.ErrBadOptions
		}
	default:
		a.log("Invalid options provided: %v", options)
		return core.ErrBadOptions
	}

	return nil
}

// Broadcast implements the core.BrokerRouter interface
func (a *Adapter) Broadcast(router core.Router, payload semtech.Payload) error {
	if !a.ok() {
		return core.ErrNotInitialized
	}

	// Determine the devAddress associated to that payload
	if payload.RXPK == nil || len(payload.RXPK) == 0 { // NOTE are those conditions significantly different ?
		a.log("Cannot broadcast given payload: %+v", payload)
		return core.ErrInvalidPayload
	}

	devAddr, err := payload.UniformDevAddr()
	if err != nil {
		a.log("Cannot broadcast given payload: %+v", payload)
		return core.ErrInvalidPayload
	}

	// Prepare ground to store brokers that are in charge
	register := make(chan core.BrokerAddress, len(a.brokers))
	wg := sync.WaitGroup{}
	wg.Add(len(a.brokers))

	client := http.Client{}
	a.mu.RLock()
	for _, addr := range a.brokers {
		go func(addr core.BrokerAddress) {
			defer wg.Done()

			resp, err := post(client, string(addr), payload)

			if err != nil {
				a.log("Unable to send POST request %+v", err)
				router.HandleError(core.ErrBroadcast) // NOTE Mote information should be sent
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
				router.HandleError(core.ErrBroadcast) // NOTE More information should be sent
			}
		}(addr)
	}
	a.mu.RUnlock()

	go func() {
		wg.Wait()
		close(register)
		brokers := make([]core.BrokerAddress, 0)
		for addr := range register {
			brokers = append(brokers, addr)
		}
		if len(brokers) > 0 {
			router.RegisterDevice(*devAddr, brokers...)
		}
	}()

	return nil
}

// Forward implements the core.BrokerRouter interface
func (a *Adapter) Forward(router core.Router, payload semtech.Payload, broAddrs ...core.BrokerAddress) error {
	if !a.ok() {
		return core.ErrNotInitialized
	}

	if payload.RXPK == nil || len(payload.RXPK) == 0 { // NOTE are those conditions significantly different ?
		a.log("Cannot broadcast given payload: %+v", payload)
		return core.ErrInvalidPayload
	}

	client := http.Client{}
	a.mu.RLock()
	for _, addr := range broAddrs {
		go func(url string) {
			a.log("Send payload to %s", url)
			resp, err := post(client, url, payload)

			if err != nil {
				a.log("Unable to send POST request %+v", err)
				router.HandleError(core.ErrForward) // NOTE More information should be sent
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				a.log("Unexpected answer from the broker %+v", err)
				router.HandleError(core.ErrForward) // NOTE More information should be sent
				return
			}

			// NOTE Do We Care about the response ? The router is supposed to handle HTTP request
			// from the broker to handle packets or anything else ? Is it efficient ? Should
			// downlinks packets be sent back with the HTTP body response ? Its a 2 seconds frame...

		}(string(addr))
	}
	a.mu.RUnlock()

	return nil
}

// post regroups some logic used in both Forward and Broadcast methods
func post(client http.Client, host string, payload semtech.Payload) (*http.Response, error) {
	data := new(bytes.Buffer)
	rawJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if _, err := data.Write(rawJSON); err != nil {
		return nil, err
	}

	return client.Post(fmt.Sprintf("http://%s", host), "application/json", data)
}

// log is nothing more than a shortcut / helper to access the logger
func (a Adapter) log(format string, i ...interface{}) {
	if a.Logger == nil {
		return
	}
	a.Logger.Log(format, i...)
}

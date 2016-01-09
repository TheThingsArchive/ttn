// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package httpregister

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/utils/log"
	"net/http"
	"sync"
)

var ErrInvalidPort = fmt.Errorf("The given port is invalid")
var ErrInvalidPacket = fmt.Errorf("The given packet is invalid")
var ErrConnectionLost = fmt.Errorf("The connection has been lost")

type Adapter struct {
	Parser
	client        http.Client
	loggers       []log.Logger // 0 to several loggers to get feedback from the Adapter.
	registrations chan regReq  // Communication dedicated to incoming registration
}

type Parser interface {
	Parse(req *http.Request) (core.Registration, error)
}

// NewAdapter constructs and allocate a new Broker <-> Handler http adapter
func NewAdapter(port uint, parser Parser, loggers ...log.Logger) (*Adapter, error) {
	if port == 0 {
		return nil, ErrInvalidPort
	}

	a := Adapter{
		Parser:        parser,
		registrations: make(chan regReq),
		loggers:       loggers,
	}

	go a.listenRegistration(port)

	return &a, nil
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, r ...core.Recipient) error {
	// Generate payload from core packet
	m, err := json.Marshal(p.Metadata)
	if err != nil {
		return ErrInvalidPacket
	}
	pl, err := p.Payload.MarshalBinary()
	if err != nil {
		return ErrInvalidPacket
	}
	payload := fmt.Sprintf(`{"payload":"%s","metadata":%s}`, base64.StdEncoding.EncodeToString(pl), m)

	// Prepare ground for parrallel http request
	nb := len(r)
	cherr := make(chan error, nb)
	wg := sync.WaitGroup{}
	wg.Add(nb)

	// Run each request
	for _, recipient := range r {
		go func(recipient core.Recipient) {
			defer wg.Done()
			a.log("Post to %v", recipient)
			buf := new(bytes.Buffer)
			buf.Write([]byte(payload))
			resp, err := http.Post(fmt.Sprintf("http://%s", recipient.Address.(string)), "application/json", buf)
			if err != nil {
				// Non-blocking, buffered
				cherr <- err
				return
			}
			if resp.StatusCode != http.StatusOK {
				// Non-blocking, buffered
				cherr <- fmt.Errorf("Unexpected response from server: %s (%d)", resp.Status, resp.StatusCode)
				return
			}
		}(recipient)
	}

	// Wait for each request to be done, and return
	wg.Wait()
	var errors []error
	for i := 0; i < len(cherr); i += 1 {
		errors = append(errors, <-cherr)
	}
	if errors != nil {
		return fmt.Errorf("Errors: %v", errors)
	}
	return nil
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() (core.Packet, core.AckNacker, error) {
	// NOTE not implemented
	return core.Packet{}, nil, nil
}

// NextRegistration implements the core.Adapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	request := <-a.registrations
	return request.Registration, regAckNacker{response: request.response}, nil
}

func (a *Adapter) log(format string, i ...interface{}) {
	for _, logger := range a.loggers {
		logger.Log(format, i...)
	}
}

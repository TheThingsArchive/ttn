// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/utils/log"
	"net/http"
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
	metadata, err := json.Marshal(p.Metadata)
	if err != nil {
		return ErrInvalidPacket
	}

	payload, err := p.Payload.MarshalBinary()
	if err != nil {
		return ErrInvalidPacket
	}
	base64Payload := base64.StdEncoding.EncodeToString(payload)

	var errors []error
	for _, recipient := range r {
		buf := new(bytes.Buffer)
		buf.Write([]byte(fmt.Sprintf(`{"payload":"%s","metadata":%s}`, base64Payload, metadata)))
		a.log("Post to %v", recipient)
		resp, err := http.Post(fmt.Sprintf("http://%s", recipient.Address.(string)), "application/json", buf)
		if err != nil || resp.StatusCode != http.StatusOK {
			errors = append(errors, err)
		}
	}

	if errors != nil {
		return fmt.Errorf("Errors: %v", errors)
	}
	return nil
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() (core.Packet, core.AckNacker, error) {
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

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broadcast

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/TheThingsNetwork/ttn/core"
	httpadapter "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/apex/log"
)

// Broadcast matereializes a extended basic http adapter which will also generate registration based
// on request responses. Whenever the adapter broadcast a request, it will trigger a registration
// demand for each new valid recipient.
type Adapter struct {
	*httpadapter.Adapter                        // Composed of an original http adapter
	ctx                  log.Interface          // Just a logger
	recipients           []core.Recipient       // A predefined list of recipients towards which broadcast messages
	registrations        chan core.Registration // Communication channel responsible for registration management
}

var ErrBadOptions = fmt.Errorf("Bad options provided")
var ErrInvalidPacket = fmt.Errorf("The given packet is invalid")
var ErrSeveralPositiveAnswers = fmt.Errorf("Several positive response for a given packet")

// NewAdapter promotes an existing basic adapter to a broadcast adapter using a list a of known recipient
func NewAdapter(adapter *httpadapter.Adapter, recipients []core.Recipient, ctx log.Interface) (*Adapter, error) {
	if len(recipients) == 0 {
		return nil, ErrBadOptions
	}

	return &Adapter{
		Adapter:       adapter,
		ctx:           ctx,
		recipients:    recipients,
		registrations: make(chan core.Registration, len(recipients)),
	}, nil
}

// Send implements the Adapter interfaces
func (a *Adapter) Send(p core.Packet, r ...core.Recipient) (core.Packet, error) {
	if len(r) == 0 {
		a.ctx.Debug("No recipient provided. The packet will be broadcast")
		return a.broadcast(p)
	}

	packet, err := a.Adapter.Send(p, r...)
	if err == httpadapter.ErrInvalidPacket {
		return core.Packet{}, ErrInvalidPacket
	}
	return packet, err
}

// broadcast is merely a send where recipients are the predefined list used at instantiation time.
// Beside, a registration request will be triggered if one of the recipient reponses positively.
func (a *Adapter) broadcast(p core.Packet) (core.Packet, error) {
	// Generate payload from core packet
	m, err := json.Marshal(p.Metadata)
	if err != nil {
		return core.Packet{}, ErrInvalidPacket
	}
	pl, err := p.Payload.MarshalBinary()
	if err != nil {
		return core.Packet{}, ErrInvalidPacket
	}
	payload := fmt.Sprintf(`{"payload":"%s","metadata":%s}`, base64.StdEncoding.EncodeToString(pl), m)

	devAddr, err := p.DevAddr()
	if err != nil {
		return core.Packet{}, ErrInvalidPacket
	}

	// Prepare ground for parrallel http request
	nb := len(a.recipients)
	cherr := make(chan error, nb)
	chresp := make(chan core.Packet, nb)
	register := make(chan core.Recipient, nb)
	wg := sync.WaitGroup{}
	wg.Add(nb)

	// Run each request
	for _, recipient := range a.recipients {
		go func(recipient core.Recipient) {
			defer wg.Done()

			ctx := a.ctx.WithField("recipient", recipient)

			ctx.Debug("POST Request")

			buf := new(bytes.Buffer)
			buf.Write([]byte(payload))
			resp, err := a.Post(fmt.Sprintf("http://%s/packets", recipient.Address.(string)), "application/json", buf)
			if err != nil {
				cherr <- err
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				ctx.WithField("devAddr", devAddr).Debug("Recipient registered for packet")

				raw := make([]byte, resp.ContentLength)
				n, err := resp.Body.Read(raw)
				if err != nil && err != io.EOF {
					cherr <- err
					return
				}
				var packet core.Packet
				if err := json.Unmarshal(raw[:n], &packet); err != nil {
					cherr <- err
					return
				}

				register <- recipient
				chresp <- packet
			case http.StatusNotFound:
				ctx.WithField("devAddr", devAddr).Debug("Recipient not interested in packet")
			default:
				// Non-blocking, buffered
				cherr <- fmt.Errorf("Unexpected response from server: %s (%d)", resp.Status, resp.StatusCode)
			}
		}(recipient)
	}

	// Wait for each request to be done, and return
	wg.Wait()
	close(cherr)
	close(register)
	var errors []error
	for err := range cherr {
		errors = append(errors, err)
	}
	if errors != nil {
		return core.Packet{}, fmt.Errorf("Errors: %v", errors)
	}

	if len(chresp) > 1 { // NOTE We consider several positive responses as an error
		return core.Packet{}, ErrSeveralPositiveAnswers
	}

	for recipient := range register {
		a.registrations <- core.Registration{DevAddr: devAddr, Recipient: recipient}
	}

	select {
	case packet := <-chresp:
		return packet, nil
	default:
		return core.Packet{}, fmt.Errorf("Unexpected error. No response packet available")
	}
}

// NextRegistration implements the Adapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	registration := <-a.registrations
	return registration, voidAckNacker{}, nil
}

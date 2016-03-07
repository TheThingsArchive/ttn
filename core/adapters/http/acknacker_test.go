// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"testing"
	"time"

	//	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	errutil "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestHTTPAckNacker(t *testing.T) {
	{
		Desc(t, "Ack a nil packet")

		// Build
		chresp := make(chan MsgRes, 1)
		an := httpAckNacker{Chresp: chresp}

		// Operate
		err := an.Ack(nil)

		// Expectation
		want := &MsgRes{
			StatusCode: http.StatusOK,
			Content:    nil,
		}

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckResps(t, want, chresp)
	}

	// --------------------

	{
		Desc(t, "Ack on a nil chresp")

		// Build
		an := httpAckNacker{Chresp: nil}

		// Operate
		err := an.Ack(nil)

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckResps(t, nil, nil)
	}

	// --------------------

	{
		Desc(t, "Ack a valid packet")

		// Build
		chresp := make(chan MsgRes, 1)
		an := httpAckNacker{Chresp: chresp}
		p := mocks.NewMockPacket()
		p.OutMarshalBinary = []byte{14, 14, 14}

		// Operate
		err := an.Ack(p)

		// Expectation
		want := &MsgRes{
			StatusCode: http.StatusOK,
			Content:    p.OutMarshalBinary,
		}

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckResps(t, want, chresp)
	}

	// --------------------

	{
		Desc(t, "Ack an invalid packet")

		// Build
		chresp := make(chan MsgRes, 1)
		an := httpAckNacker{Chresp: chresp}
		p := mocks.NewMockPacket()
		p.Failures["MarshalBinary"] = errors.New(errors.Structural, "Mock Error")

		// Operate
		err := an.Ack(p)

		// Check
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckResps(t, nil, chresp)
	}

	// --------------------

	{
		Desc(t, "Don't consume chresp")

		// Build
		chresp := make(chan MsgRes)
		an := httpAckNacker{Chresp: chresp}

		// Operate
		cherr := make(chan error)
		go func() {
			cherr <- an.Ack(nil)
		}()

		// Check
		var err error
		select {
		case err = <-cherr:
		case <-time.After(time.Millisecond * 100):
		}
		errutil.CheckErrors(t, pointer.String(string(errors.Operational)), err)
	}

	// --------------------

	{
		Desc(t, "Nack")

		// Build
		chresp := make(chan MsgRes, 1)
		an := httpAckNacker{Chresp: chresp}

		// Operate
		err := an.Nack()

		// Expectation
		want := &MsgRes{
			StatusCode: http.StatusNotFound,
			Content:    []byte(errors.Structural),
		}

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckResps(t, want, chresp)
	}

	// --------------------

	{
		Desc(t, "Nack on a nil chresp")

		// Build
		an := httpAckNacker{Chresp: nil}

		// Operate
		err := an.Nack()

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckResps(t, nil, nil)
	}
}

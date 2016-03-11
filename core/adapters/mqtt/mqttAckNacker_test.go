// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	errutil "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestAckNacker(t *testing.T) {
	{
		Desc(t, "Ack a nil packet")

		// Build
		chresp := make(chan MsgRes, 1)
		an := mqttAckNacker{Chresp: chresp}

		// Operate
		err := an.Ack(nil)

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckResps(t, nil, chresp)
	}

	// --------------------

	{
		Desc(t, "Ack on a nil chresp")

		// Build
		an := mqttAckNacker{Chresp: nil}

		// Operate
		err := an.Ack(mocks.NewMockPacket())

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckResps(t, nil, nil)
	}

	// --------------------

	{
		Desc(t, "Ack a valid packet")

		// Build
		chresp := make(chan MsgRes, 1)
		an := mqttAckNacker{Chresp: chresp}
		p := mocks.NewMockPacket()
		p.OutMarshalBinary = []byte{14, 14, 14}

		// Operate
		err := an.Ack(p)

		// Expectation
		want := MsgRes(p.OutMarshalBinary)

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckResps(t, &want, chresp)
	}

	// --------------------

	{
		Desc(t, "Ack an invalid packet")

		// Build
		chresp := make(chan MsgRes, 1)
		an := mqttAckNacker{Chresp: chresp}
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
		Desc(t, "Don't consume chresp on Ack")

		// Build
		chresp := make(chan MsgRes)
		an := mqttAckNacker{Chresp: chresp}

		// Operate
		cherr := make(chan error)
		go func() {
			cherr <- an.Ack(mocks.NewMockPacket())
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
		Desc(t, "Nack no error")

		// Build
		chresp := make(chan MsgRes, 1)
		an := mqttAckNacker{Chresp: chresp}

		// Operate
		err := an.Nack(nil)

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckResps(t, nil, chresp)
	}

	// --------------------

	{
		Desc(t, "Nack on a nil chresp")

		// Build
		an := mqttAckNacker{Chresp: nil}

		// Operate
		err := an.Nack(nil)

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckResps(t, nil, nil)
	}
}

func CheckResps(t *testing.T, want *MsgRes, got chan MsgRes) {
	if want == nil {
		if len(got) == 0 {
			Ok(t, "Check Resps")
			return
		}
		Ko(t, "Expected no message response but got one")
	}

	if len(got) < 1 {
		Ko(t, "Expected one message but got none")
	}

	msg := <-got
	mocks.Check(t, *want, msg, "Resps")
}

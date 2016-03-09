// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	errutil "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	testutil "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestNext(t *testing.T) {
	{
		testutil.Desc(t, "Send a packet when no handler is defined")

		// Build
		addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:2016")
		conn, _ := net.DialUDP("udp", nil, addr)
		adapter, errNew := NewAdapter("0.0.0.0:2016", testutil.GetLogger(t, "Adapter"))
		errutil.CheckErrors(t, nil, errNew)

		// Operate
		<-time.After(time.Millisecond * 25)
		_, errWrite := conn.Write([]byte{1, 2, 3, 4})

		// Operate
		packet, errNext := tryNext(adapter)

		// Check
		errutil.CheckErrors(t, nil, errWrite)
		CheckPackets(t, nil, packet)
		errutil.CheckErrors(t, nil, errNext)
	}

	// --------------------

	{
		testutil.Desc(t, "Start adapter on a busy connection")

		// Build
		addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:2017")
		_, _ = net.ListenUDP("udp", addr)
		_, errNew := NewAdapter("0.0.0.0:2017", testutil.GetLogger(t, "Adapter"))

		// Check
		errutil.CheckErrors(t, pointer.String(string(errors.Operational)), errNew)
	}

	// --------------------

	{
		testutil.Desc(t, "Attach a handler to the adapter and fake udp reception")

		// Build
		addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:2018")
		conn, _ := net.DialUDP("udp", nil, addr)
		handler := &MockHandler{}
		adapter, errNew := NewAdapter("0.0.0.0:2018", testutil.GetLogger(t, "Adapter"))
		errutil.CheckErrors(t, nil, errNew)

		// Operate
		adapter.Bind(handler)
		_, errWrite := conn.Write([]byte{1, 2, 3, 4})
		<-time.After(time.Millisecond * 25)

		// Check
		errutil.CheckErrors(t, nil, errWrite)
		CheckPackets(t, []byte{1, 2, 3, 4}, handler.InMsg.Data)
	}

	// --------------------

	{
		testutil.Desc(t, "Send next data through the handler")

		// Build
		addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:2019")
		conn, _ := net.DialUDP("udp", nil, addr)
		handler := &MockHandler{}
		handler.OutMsgReq = []byte{14, 42, 14, 42}
		adapter, errNew := NewAdapter("0.0.0.0:2019", testutil.GetLogger(t, "Adapter"))
		errutil.CheckErrors(t, nil, errNew)

		// Operate
		adapter.Bind(handler)
		_, errWrite := conn.Write([]byte{1, 2, 3, 4})
		<-time.After(time.Millisecond * 25)
		packet, errNext := tryNext(adapter)

		// Check
		errutil.CheckErrors(t, nil, errWrite)
		errutil.CheckErrors(t, nil, errNext)
		CheckPackets(t, []byte{1, 2, 3, 4}, handler.InMsg.Data)
		CheckPackets(t, nil, handler.InChresp)
		CheckPackets(t, []byte{14, 42, 14, 42}, packet)
	}

	// --------------------

	{
		testutil.Desc(t, "Send next data back through the connection")

		// Build
		addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:2020")
		conn, _ := net.DialUDP("udp", nil, addr)
		read := make([]byte, 20)
		handler := &MockHandler{}
		handler.OutMsgUDP = []byte{14, 42, 14, 42}
		adapter, errNew := NewAdapter("0.0.0.0:2020", testutil.GetLogger(t, "Adapter"))
		errutil.CheckErrors(t, nil, errNew)

		// Operate
		adapter.Bind(handler)
		_, errWrite := conn.Write([]byte{1, 2, 3, 4})
		<-time.After(time.Millisecond * 25)
		packet, errNext := tryNext(adapter)
		n, errRead := conn.Read(read)

		// Check
		errutil.CheckErrors(t, nil, errWrite)
		errutil.CheckErrors(t, nil, errRead)
		errutil.CheckErrors(t, nil, errNext)
		CheckPackets(t, []byte{1, 2, 3, 4}, handler.InMsg.Data)
		CheckPackets(t, read[:n], []byte{14, 42, 14, 42})
		CheckPackets(t, nil, packet)
	}
}

func TestNotImplemented(t *testing.T) {
	{
		testutil.Desc(t, "NextRegistration ~> not implemented")

		// Build
		adapter, errNew := NewAdapter("0.0.0.0:2021", testutil.GetLogger(t, "Adapter"))
		errutil.CheckErrors(t, nil, errNew)

		// Operate
		_, _, errNext := adapter.NextRegistration()

		// Check
		errutil.CheckErrors(t, pointer.String(string(errors.Implementation)), errNext)
	}

	// --------------------

	{
		testutil.Desc(t, "Send ~> not implemented")

		// Build
		adapter, errNew := NewAdapter("0.0.0.0:2022", testutil.GetLogger(t, "Adapter"))
		errutil.CheckErrors(t, nil, errNew)

		// Operate
		_, errSend := adapter.Send(nil)

		// Check
		errutil.CheckErrors(t, pointer.String(string(errors.Implementation)), errSend)
	}

	// --------------------

	{
		testutil.Desc(t, "GetRecipient ~> not implemented")

		// Build
		adapter, errNew := NewAdapter("0.0.0.0:2023", testutil.GetLogger(t, "Adapter"))
		errutil.CheckErrors(t, nil, errNew)

		// Operate
		_, errGet := adapter.GetRecipient(nil)

		// Check
		errutil.CheckErrors(t, pointer.String(string(errors.Implementation)), errGet)
	}
}

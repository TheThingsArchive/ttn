// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"net"
	"testing"

	. "github.com/smartystreets/assertions"
)

func mkConn() *gatewayConn {
	return &gatewayConn{}
}

func TestSetConn(t *testing.T) {
	a := New(t)
	exp := &net.UDPConn{}
	c := mkConn()
	c.SetConn(exp)
	a.So(c.conn, ShouldEqual, exp)
}

func TestSetUplinkAddr(t *testing.T) {
	a := New(t)
	exp := &net.UDPAddr{}
	c := mkConn()
	c.SetUplinkAddr(exp)
	a.So(c.uplinkAddr, ShouldEqual, exp)
}

func TestSetDownlinkAddr(t *testing.T) {
	a := New(t)
	exp := &net.UDPAddr{}
	c := mkConn()
	c.SetDownlinkAddr(exp)
	a.So(c.downlinkAddr, ShouldEqual, exp)
}

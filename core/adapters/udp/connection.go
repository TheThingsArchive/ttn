// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"fmt"
	"net"
	"sync"
)

type replier interface {
	DestinationID() []byte
	WriteToUplink(data []byte) error
	WriteToDownlink(data []byte) error
}

type gatewayConn struct {
	sync.RWMutex
	gatewayID    []byte
	conn         *net.UDPConn
	uplinkAddr   *net.UDPAddr
	downlinkAddr *net.UDPAddr
}

func (c *gatewayConn) SetConn(conn *net.UDPConn) {
	c.Lock()
	defer c.Unlock()
	c.conn = conn
}

func (c *gatewayConn) SetUplinkAddr(addr *net.UDPAddr) {
	c.Lock()
	defer c.Unlock()
	c.uplinkAddr = addr
}

func (c *gatewayConn) SetDownlinkAddr(addr *net.UDPAddr) {
	c.Lock()
	defer c.Unlock()
	c.downlinkAddr = addr
}

func (c *gatewayConn) DestinationID() []byte {
	return c.gatewayID
}

func (c *gatewayConn) WriteToUplink(data []byte) error {
	c.RLock()
	defer c.RUnlock()
	if c.conn == nil || c.uplinkAddr == nil {
		return fmt.Errorf("Uplink connection unavailable.")
	}

	_, err := c.conn.WriteToUDP(data, c.uplinkAddr)
	return err
}

func (c *gatewayConn) WriteToDownlink(data []byte) error {
	c.RLock()
	defer c.RUnlock()
	if c.conn == nil || c.downlinkAddr == nil {
		return fmt.Errorf("Downlink connection unavailable.")
	}

	_, err := c.conn.WriteToUDP(data, c.downlinkAddr)
	return err
}

type gatewayPool struct {
	sync.Mutex
	connections map[[8]byte]*gatewayConn
}

func newPool() *gatewayPool {
	return &gatewayPool{
		connections: make(map[[8]byte]*gatewayConn),
	}
}

func (p *gatewayPool) GetOrCreate(gatewayID []byte) *gatewayConn {
	p.Lock()
	defer p.Unlock()
	var id [8]byte
	copy(id[:], gatewayID)
	if _, ok := p.connections[id]; !ok {
		p.connections[id] = &gatewayConn{gatewayID: gatewayID}
	}
	return p.connections[id]
}

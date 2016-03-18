// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"google.golang.org/grpc"
)

// Dialer abstracts the connection to grpc, or anything else
type Dialer interface {
	MarshalSafely() []byte
	Dial() (core.HandlerClient, Closer, error) // Dial actually attempts to dial a connection
}

// Closer is returned by a Dialer to give a hand on closing the dialed connection
type Closer interface {
	Close() error
}

// dialer implements the Dialer interface
type dialer struct {
	NetAddr string
}

// closer implements the Closer interface
type closer struct {
	Conn *grpc.ClientConn
}

// Close implements the Closer interface
func (c closer) Close() error {
	if err := c.Conn.Close(); err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

// NewDialer constructs a new dialer from a given net address
func NewDialer(netAddr []byte) Dialer {
	return &dialer{NetAddr: string(netAddr)}
}

// Dial implements the Dialer interface
func (d dialer) Dial() (core.HandlerClient, Closer, error) {
	conn, err := grpc.Dial(d.NetAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second*2))
	if err != nil {
		return nil, nil, errors.New(errors.Operational, err)
	}
	return core.NewHandlerClient(conn), closer{Conn: conn}, nil
}

// MarshalSafely implements the Dialer interface
func (d dialer) MarshalSafely() []byte {
	return []byte(d.NetAddr)
}

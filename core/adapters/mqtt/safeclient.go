// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"github.com/yosssi/gmq/mqtt/client"
)

// NOTE: All that code could be easily generated

// safeClient implements the mqtt.Client interface. It is completely concurrent-safe
type safeClient struct {
	chcmd chan<- interface{}
}

// Publish implements the Client interface
func (c safeClient) Publish(o *client.PublishOptions) error {
	cherr := make(chan error)
	c.chcmd <- cmdPublish{options: o, cherr: cherr}
	return <-cherr
}

// Terminate implements the Client interface
func (c safeClient) Terminate() {
	cherr := make(chan error)
	c.chcmd <- cmdTerminate{cherr: cherr}
	<-cherr
}

type cmdPublish struct {
	options *client.PublishOptions
	cherr   chan<- error
}

type cmdTerminate struct {
	cherr chan<- error
}

type cmdClient struct {
	options *client.Client
	cherr   chan<- error
}

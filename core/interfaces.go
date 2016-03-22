// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

type Router interface {
	RouterClient
}

type Broker interface {
	BrokerClient
	BrokerManagerClient
	BeginToken(token string) Broker
	EndToken()
}

type Handler interface {
	HandlerClient
	HandlerManagerClient
}

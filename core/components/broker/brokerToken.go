// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

// BeginToken implements the core.Broker interface
func (b *component) BeginToken(token string) Interface {
	// TODO - Acquire Mutex, Mutate Token
	return b
}

// EndToken implements the core.Broker interface
func (b *component) EndToken() {
	// TODO - Release Mutex
}

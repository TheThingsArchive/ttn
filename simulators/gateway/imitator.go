// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import ()

type Imitator interface {
	Mimic() error
}

func (g *Gateway) Mimic() error {
	return nil
}

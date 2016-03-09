// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/mocks"
)

// tryNext attempts to get a next packet from the adapter. It timeouts after a given delay if
// nothing is ready.
func tryNext(adapter *Adapter) ([]byte, error) {
	chresp := make(chan struct {
		Packet []byte
		Error  error
	})
	go func() {
		packet, an, err := adapter.Next()
		if err != nil {
			an.Nack(nil)
		} else {
			an.Ack(nil)
		}
		chresp <- struct {
			Packet []byte
			Error  error
		}{packet, err}
	}()

	select {
	case resp := <-chresp:
		return resp.Packet, resp.Error
	case <-time.After(time.Millisecond * 75):
		return nil, nil
	}
}

func CheckPackets(t *testing.T, want []byte, got []byte) {
	mocks.Check(t, want, got, "Packets")
}

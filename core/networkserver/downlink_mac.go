// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	pb_broker "github.com/TheThingsNetwork/api/broker"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
)

func (n *networkServer) handleDownlinkMAC(message *pb_broker.DownlinkMessage, dev *device.Device) error {
	if err := n.handleDownlinkADR(message, dev); err != nil {
		return err
	}
	return nil
}

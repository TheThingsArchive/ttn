// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"time"

	"github.com/spf13/viper"
)

func (n *networkServer) monitorNetworkServerStatus() {
	interval := viper.GetDuration("monitor-interval")

	t := time.Tick(interval * time.Second)
	if t == nil {
		panic("monitor-interval value is not valid")
	}

	for _ = range t {
		n.monitorStream.Send(n.GetStatus())
	}
}

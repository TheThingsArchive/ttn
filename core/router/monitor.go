// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"time"

	"github.com/spf13/viper"
)

func (r *router) monitorRouterStatus() {
	interval := viper.GetDuration("monitor-interval")

	t := time.Tick(interval * time.Second)
	if t == nil {
		panic("monitor-interval value is not valid")
	}

	for _ = range t {
		r.monitorStream.Send(r.GetStatus())
	}
}

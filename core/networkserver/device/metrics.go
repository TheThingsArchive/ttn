// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"github.com/prometheus/client_golang/prometheus"
)

var deviceCount = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	Namespace: "ttn",
	Subsystem: "networkserver",
	Name:      "registered_devices",
	Help:      "Registered devices.",
}, func() float64 {
	var count uint64
	for _, store := range stores {
		devices, err := store.Count()
		if err != nil {
			continue
		}
		count += uint64(devices)
	}
	return float64(count)
})

func init() {
	prometheus.Register(deviceCount)
}

var stores []*RedisDeviceStore

func countStore(store *RedisDeviceStore) {
	stores = append(stores, store)
}

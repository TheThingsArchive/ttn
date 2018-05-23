// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package application

import (
	"github.com/prometheus/client_golang/prometheus"
)

var applicationCount = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	Namespace: "ttn",
	Subsystem: "handler",
	Name:      "registered_applications",
	Help:      "Registered applications.",
}, func() float64 {
	var count uint64
	for _, store := range stores {
		applications, err := store.Count()
		if err != nil {
			continue
		}
		count += uint64(applications)
	}
	return float64(count)
})

func init() {
	prometheus.Register(applicationCount)
}

var stores []*RedisApplicationStore

func countStore(store *RedisApplicationStore) {
	stores = append(stores, store)
}

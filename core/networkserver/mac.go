// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"sort"

	pb_gateway "github.com/TheThingsNetwork/api/gateway"
)

const macCMD = "cmd" // For Tracing

type bySNR []*pb_gateway.RxMetadata

func (a bySNR) Len() int           { return len(a) }
func (a bySNR) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a bySNR) Less(i, j int) bool { return a[i].SNR < a[j].SNR }

func bestSNR(metadata []*pb_gateway.RxMetadata) float32 {
	if len(metadata) == 0 {
		return 0
	}
	sorted := bySNR(metadata)
	sort.Sort(sorted)
	return sorted[len(sorted)-1].SNR
}

var demodulationFloor = map[string]float32{
	"SF7BW125":  -7.5,
	"SF8BW125":  -10,
	"SF9BW125":  -12.5,
	"SF10BW125": -15,
	"SF11BW125": -17.5,
	"SF12BW125": -20,
	"SF7BW250":  -4.5,
}

func linkMargin(dataRate string, snr float32) float32 {
	if floor, ok := demodulationFloor[dataRate]; ok {
		return snr - floor
	}
	return 0
}

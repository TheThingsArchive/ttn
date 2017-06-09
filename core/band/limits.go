// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package band

import (
	"time"

	"github.com/TheThingsNetwork/go-utils/rate"
)

// UtilizationLimits limits the utilization to a fraction within a timeframe
type UtilizationLimits map[time.Duration]float64

func (l UtilizationLimits) size() (min, max time.Duration) {
	min = time.Minute
	max = time.Hour
	for d := range l {
		if max < d {
			max = d
		}
		if min > d {
			min = d
		}
	}
	return
}

// Progress towards the limit
func (l UtilizationLimits) Progress(rate rate.Counter) (progress float64) {
	now := time.Now()
	for d, l := range l {
		ns, _ := rate.Get(now, d)
		if p := (float64(ns) / float64(d)) / l; p > progress {
			progress = p
		}
	}
	return
}

// SubBandLimits limits the utilization of a sub-band
type SubBandLimits struct {
	appliesTo   func(frequency uint64) bool
	utilization rate.Counter
	limits      UtilizationLimits
}

// Add airtime to the utilization
func (l SubBandLimits) Add(airtime time.Duration) {
	l.utilization.Add(time.Now(), uint64(airtime))
}

// Progress towards the sub-band utilization limit
func (l SubBandLimits) Progress() float64 {
	return l.limits.Progress(l.utilization)
}

// New SubBandLimits instance
func (l SubBandLimits) New() SubBandLimits {
	return SubBandLimits{
		appliesTo:   l.appliesTo,
		utilization: rate.NewCounter(l.limits.size()),
		limits:      l.limits,
	}
}

// BandLimits limits the utilization of a band
type BandLimits []SubBandLimits

// Add airtime to a frequency
func (b BandLimits) Add(frequency uint64, airtime time.Duration) {
	for _, sub := range b {
		if sub.appliesTo(frequency) {
			sub.Add(airtime)
		}
	}
}

// Progress towards the frequency utilization limit
func (b BandLimits) Progress(frequency uint64) (progress float64) {
	for _, sub := range b {
		if sub.appliesTo(frequency) {
			subProgress := sub.Progress()
			if subProgress > progress {
				progress = subProgress
			}
		}
	}
	return
}

// New BandLimits instance
func (b BandLimits) New() BandLimits {
	newLimits := make(BandLimits, len(b))
	for i, l := range b {
		newLimits[i] = l.New()
	}
	return newLimits
}

func (b BandLimits) TimeOffAir(tx time.Duration) time.Duration {
	return 0
}

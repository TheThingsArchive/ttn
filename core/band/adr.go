// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package band

import "errors"

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

// ADRConfig contains configuration for Adaptive Data Rate
type ADRConfig struct {
	MinDataRate int
	MaxDataRate int
	MinTXPower  int
	MaxTXPower  int
}

// ErrADRUnavailable is returned when ADR is not available
var ErrADRUnavailable = errors.New("ADR Unavailable")

// ADRSettings gets the ADR settings given a dataRate, txPower, SNR and device margin
func (f *FrequencyPlan) ADRSettings(dataRate string, txPower int, snr float32, deviceMargin float32) (desiredDataRate string, desiredTxPower int, err error) {
	if f.ADR == nil {
		return dataRate, txPower, ErrADRUnavailable
	}
	margin := linkMargin(dataRate, snr) - deviceMargin
	drIdx, err := f.GetDataRateIndexFor(dataRate)
	if err != nil {
		return dataRate, txPower, err
	}
	nStep := int(margin / 3)

	// Increase the data rate with each step
	for nStep > 0 && drIdx < f.ADR.MaxDataRate {
		drIdx++
		nStep--
	}

	// Decrease the Tx power by 3 for each step, until min reached
	for nStep > 0 && txPower > f.ADR.MinTXPower {
		txPower -= 3
		nStep--
	}
	if txPower < f.ADR.MinTXPower {
		txPower = f.ADR.MinTXPower
	}

	// Increase the Tx power by 3 for each step, until max reached
	for nStep < 0 && txPower < f.ADR.MaxTXPower {
		txPower += 3
		nStep++
	}
	if txPower > f.ADR.MaxTXPower {
		txPower = f.ADR.MaxTXPower
	}

	desiredDataRate, err = f.GetDataRateStringForIndex(drIdx)
	if err != nil {
		return dataRate, txPower, err // This should maybe panic; it means that f.ADR is incosistent with f.DataRates
	}
	return desiredDataRate, txPower, nil
}

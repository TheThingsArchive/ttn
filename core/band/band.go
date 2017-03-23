// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package band

import (
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
	lora "github.com/brocaar/lorawan/band"
)

// FrequencyPlan includes band configuration and CFList
type FrequencyPlan struct {
	lora.Band
	ADR    *ADRConfig
	CFList *lorawan.CFList
}

func (f *FrequencyPlan) GetDataRateStringForIndex(drIdx int) (string, error) {
	dr, err := types.ConvertDataRate(f.DataRates[drIdx])
	if err != nil {
		return "", err
	}
	return dr.String(), nil
}

func (f *FrequencyPlan) GetDataRateIndexFor(dataRate string) (int, error) {
	dr, err := types.ParseDataRate(dataRate)
	if err != nil {
		return 0, err
	}
	return f.Band.GetDataRate(lora.DataRate{Modulation: lora.LoRaModulation, SpreadFactor: int(dr.SpreadingFactor), Bandwidth: int(dr.Bandwidth)})
}

func (f *FrequencyPlan) GetTxPowerIndexFor(txPower int) (int, error) {
	for i, power := range f.TXPower {
		if power == txPower {
			return i, nil
		}
	}
	return 0, errors.New("core/band: the given tx-power does not exist")
}

// Guess the region based on frequency
func Guess(frequency uint64) string {
	// Join frequencies
	switch {
	case frequency == 923200000 && frequency <= 923400000:
		// not considering AS_920_923 and AS_923_925 because we're not sure
		return pb_lorawan.FrequencyPlan_AS_923.String()
	case frequency == 922100000 || frequency == 922300000 || frequency == 922500000:
		return pb_lorawan.FrequencyPlan_KR_920_923.String()
	}

	// Existing Channels
	if region, ok := channels[int(frequency)]; ok {
		return region
	}

	// Everything Else: not allowed
	return ""
}

// Get the frequency plan for the given region
func Get(region string) (frequencyPlan FrequencyPlan, err error) {
	if fp, ok := frequencyPlans[region]; ok {
		return fp, nil
	}
	switch region {
	case pb_lorawan.FrequencyPlan_EU_863_870.String():
		frequencyPlan.Band, err = lora.GetConfig(lora.EU_863_870, false, lorawan.DwellTimeNoLimit)
		// TTN uses SF9BW125 in RX2
		frequencyPlan.RX2DataRate = 3
		// TTN frequency plan includes extra channels next to the default channels:
		frequencyPlan.UplinkChannels = []lora.Channel{
			lora.Channel{Frequency: 868100000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 868300000, DataRates: []int{0, 1, 2, 3, 4, 5, 6}}, // Also SF7BW250
			lora.Channel{Frequency: 868500000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 867100000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 867300000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 867500000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 867700000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 867900000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 868800000, DataRates: []int{7}}, // FSK 50kbps
		}
		frequencyPlan.DownlinkChannels = frequencyPlan.UplinkChannels
		frequencyPlan.CFList = &lorawan.CFList{867100000, 867300000, 867500000, 867700000, 867900000}
		frequencyPlan.ADR = &ADRConfig{MinDataRate: 0, MaxDataRate: 5, MinTXPower: 2, MaxTXPower: 14}
	case pb_lorawan.FrequencyPlan_US_902_928.String():
		frequencyPlan.Band, err = lora.GetConfig(lora.US_902_928, false, lorawan.DwellTime400ms)
	case pb_lorawan.FrequencyPlan_CN_779_787.String():
		frequencyPlan.Band, err = lora.GetConfig(lora.CN_779_787, false, lorawan.DwellTimeNoLimit)
	case pb_lorawan.FrequencyPlan_EU_433.String():
		frequencyPlan.Band, err = lora.GetConfig(lora.EU_433, false, lorawan.DwellTimeNoLimit)
	case pb_lorawan.FrequencyPlan_AU_915_928.String():
		frequencyPlan.Band, err = lora.GetConfig(lora.AU_915_928, false, lorawan.DwellTime400ms)
	case pb_lorawan.FrequencyPlan_CN_470_510.String():
		frequencyPlan.Band, err = lora.GetConfig(lora.CN_470_510, false, lorawan.DwellTimeNoLimit)
	case pb_lorawan.FrequencyPlan_AS_923.String():
		frequencyPlan.Band, err = lora.GetConfig(lora.AS_923, false, lorawan.DwellTime400ms)
	case pb_lorawan.FrequencyPlan_AS_920_923.String():
		frequencyPlan.Band, err = lora.GetConfig(lora.AS_923, false, lorawan.DwellTime400ms)
		frequencyPlan.UplinkChannels = []lora.Channel{
			lora.Channel{Frequency: 923200000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 923400000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 922200000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 922400000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 922600000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 922800000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 923000000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 922000000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 922100000, DataRates: []int{6}},
			lora.Channel{Frequency: 921800000, DataRates: []int{7}},
		}
		frequencyPlan.DownlinkChannels = frequencyPlan.UplinkChannels
		frequencyPlan.CFList = &lorawan.CFList{922200000, 922400000, 922600000, 922800000, 923000000}
	case pb_lorawan.FrequencyPlan_AS_923_925.String():
		frequencyPlan.Band, err = lora.GetConfig(lora.AS_923, false, lorawan.DwellTime400ms)
		frequencyPlan.UplinkChannels = []lora.Channel{
			lora.Channel{Frequency: 923200000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 923400000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 923600000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 923800000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 924000000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 924200000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 924400000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 924600000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 924500000, DataRates: []int{6}},
			lora.Channel{Frequency: 924800000, DataRates: []int{7}},
		}
		frequencyPlan.DownlinkChannels = frequencyPlan.UplinkChannels
		frequencyPlan.CFList = &lorawan.CFList{923600000, 923800000, 924000000, 924200000, 924400000}
	case pb_lorawan.FrequencyPlan_KR_920_923.String():
		frequencyPlan.Band, err = lora.GetConfig(lora.KR_920_923, false, lorawan.DwellTimeNoLimit)
		// TTN frequency plan includes extra channels next to the default channels:
		frequencyPlan.UplinkChannels = []lora.Channel{
			lora.Channel{Frequency: 922100000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 922300000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 922500000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 922700000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 922900000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 923100000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 923300000, DataRates: []int{0, 1, 2, 3, 4, 5}},
		}
		frequencyPlan.DownlinkChannels = frequencyPlan.UplinkChannels
		frequencyPlan.CFList = &lorawan.CFList{922700000, 922900000, 923100000, 923300000, 0}
	default:
		err = errors.NewErrInvalidArgument("Frequency Band", "unknown")
	}
	return
}

var frequencyPlans map[string]FrequencyPlan
var channels map[int]string

func init() {
	frequencyPlans = make(map[string]FrequencyPlan)
	channels = make(map[int]string)
	for _, r := range []pb_lorawan.FrequencyPlan{ // ordering is important here
		pb_lorawan.FrequencyPlan_EU_863_870,
		pb_lorawan.FrequencyPlan_US_902_928,
		pb_lorawan.FrequencyPlan_CN_779_787,
		pb_lorawan.FrequencyPlan_EU_433,
		pb_lorawan.FrequencyPlan_AS_923,
		pb_lorawan.FrequencyPlan_AS_920_923,
		pb_lorawan.FrequencyPlan_AS_923_925,
		pb_lorawan.FrequencyPlan_KR_920_923,
		pb_lorawan.FrequencyPlan_AU_915_928,
		pb_lorawan.FrequencyPlan_CN_470_510,
	} {
		region := r.String()
		frequencyPlans[region], _ = Get(region)
		for _, ch := range frequencyPlans[region].UplinkChannels {
			if len(ch.DataRates) > 1 { // ignore FSK channels
				if _, ok := channels[ch.Frequency]; !ok { // ordering indicates priority
					channels[ch.Frequency] = region
				}
			}
		}
	}
}

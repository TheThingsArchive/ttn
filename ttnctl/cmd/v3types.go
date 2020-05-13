// Copyright © 2020 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import "fmt"

type v3ApplicationIDs struct {
	ApplicationID string `json:"application_id,omitempty"`
}

type v3DeviceIDs struct {
	DeviceID       string           `json:"device_id"`
	ApplicationIDs v3ApplicationIDs `json:"application_ids,omitempty"`
	DevEUI         string           `json:"dev_eui,omitempty"`
	JoinEUI        string           `json:"join_eui,omitempty"`
	DevAddr        string           `json:"dev_addr,omitempty"`
}

type v3DeviceLocation struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Altitude  int32   `json:"altitude,omitempty"`
	Source    string  `json:"source,omitempty"`
}

type v3DeviceKey struct {
	Key string `json:"key"`
}

type v3DeviceRootKeys struct {
	AppKey v3DeviceKey `json:"app_key,omitempty"`
}

type v3DeviceSessionKeys struct {
	AppSKey     v3DeviceKey `json:"app_s_key,omitempty"`
	FNwkSIntKey v3DeviceKey `json:"f_nwk_s_int_key,omitempty"`
}

type v3DeviceSession struct {
	DevAddr       string              `json:"dev_addr"`
	Keys          v3DeviceSessionKeys `json:"keys"`
	LastFCntUp    uint32              `json:"last_f_cnt_up"`
	LastNFCntDown uint32              `json:"last_n_f_cnt_down"`
}

type v3RxDelay struct {
	Value string `json:"value,omitempty"`
}

type v3DeviceMACSettings struct {
	Rx1Delay                 v3RxDelay `json:"rx1_delay,omitempty"`
	Supports32BitFCnt        bool      `json:"supports_32_bit_f_cnt"`
	ResetsFCnt               bool      `json:"resets_f_cnt"`
	FactoryPresetFrequencies []uint64  `json:"factory_preset_frequencies,omitempty"`
}

type v3Device struct {
	IDs             v3DeviceIDs                 `json:"ids"`
	Name            string                      `json:"name,omitempty"`
	Description     string                      `json:"description,omitempty"`
	Attributes      map[string]string           `json:"attributes,omitempty"`
	DevAddr         string                      `json:"dev_addr,omitempty"`
	Locations       map[string]v3DeviceLocation `json:"locations,omitempty"`
	MACVersion      string                      `json:"lorawan_version"`
	PHYVersion      string                      `json:"lorawan_phy_version"`
	FrequencyPlanID string                      `json:"frequency_plan_id"`
	SupportsJoin    bool                        `json:"supports_join"`
	RootKeys        *v3DeviceRootKeys           `json:"root_keys,omitempty"`
	NetID           []byte                      `json:"net_id,omitempty"`
	MACSettings     v3DeviceMACSettings         `json:"mac_settings"`
	Session         *v3DeviceSession            `json:"session,omitempty"`
}

var (
	frequencyPlans = []string{
		"AS_920_923_LBT",
		"AS_920_923",
		"AS_923_925_LBT",
		"AS_923_925_TTN_AU",
		"AS_923_925",
		"AU_915_928_FSB_1",
		"AU_915_928_FSB_2",
		"AU_915_928_FSB_6",
		"CN_470_510_FSB_11",
		"EU_863_870_TTN",
		"EU_863_870",
		"IN_865_867",
		"KR_920_923_TTN",
		"RU_864_870_TTN",
		"US_902_928_FSB_1",
		"US_902_928_FSB_2",
	}
)

func getOption(options []string, option string) (string, error) {
	for _, o := range options {
		if o == option {
			return option, nil
		}
	}
	return "", fmt.Errorf("invalid_value")
}

// Copyright © 2019 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/api"
	"github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

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
	AppKey v3DeviceKey `json:"app_key"`
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

type v3DeviceMACSettings struct {
	Rx1Delay struct {
		Value string `json:"value"`
	} `json:"rx1_delay"`
	Rx2DataRateIndex struct {
		Value string `json:"value"`
	} `json:"rx2_data_rate_index"`
	Supports32BitFCnt        bool     `json:"supports_32_bit_f_cnt"`
	ResetsFCnt               bool     `json:"resets_f_cnt"`
	FactoryPresetFrequencies []uint64 `json:"factory_preset_frequencies,omitempty"`
}

type v3Device struct {
	IDs             v3DeviceIDs                 `json:"ids"`
	Description     string                      `json:"description,omitempty"`
	Attributes      map[string]string           `json:"attributes,omitempty"`
	Locations       map[string]v3DeviceLocation `json:"locations,omitempty"`
	MACVersion      string                      `json:"lorawan_version"`
	PHYVersion      string                      `json:"lorawan_phy_version"`
	FrequencyPlanID string                      `json:"frequency_plan_id"`
	SupportsJoin    bool                        `json:"supports_join"`
	RootKeys        *v3DeviceRootKeys           `json:"root_keys,omitempty"`
	NetID           []byte                      `json:"net_id,omitempty"`
	MACSettings     v3DeviceMACSettings         `json:"mac_settings"`
	// MACState?
	Session *v3DeviceSession `json:"session,omitempty"`
}

func exportV3Device(dev *handler.Device) *v3Device {
	v3Dev := &v3Device{}
	v3Dev.IDs.ApplicationIDs.ApplicationID = dev.AppID
	v3Dev.IDs.DeviceID = dev.DevID

	if dev.Latitude != 0 || dev.Longitude != 0 {
		v3Dev.Locations = map[string]v3DeviceLocation{
			"registered": v3DeviceLocation{
				Latitude:  float64(dev.Latitude),
				Longitude: float64(dev.Longitude),
				Altitude:  dev.Altitude,
				Source:    "SOURCE_REGISTRY",
			},
		}
	}

	v3Dev.Description = dev.Description
	v3Dev.Attributes = dev.Attributes

	lorawanDevice := dev.GetLoRaWANDevice()

	v3Dev.IDs.JoinEUI = lorawanDevice.GetAppEUI().String()
	v3Dev.IDs.DevEUI = lorawanDevice.GetDevEUI().String()
	if lorawanDevice.AppKey != nil {
		v3Dev.RootKeys = &v3DeviceRootKeys{
			AppKey: v3DeviceKey{Key: lorawanDevice.AppKey.String()},
		}
	}
	if devAddr := lorawanDevice.DevAddr; !devAddr.IsEmpty() {
		v3Dev.IDs.DevAddr = lorawanDevice.DevAddr.String()
		v3Dev.Session = &v3DeviceSession{}
		v3Dev.Session.DevAddr = lorawanDevice.DevAddr.String()
		if lorawanDevice.AppSKey != nil {
			v3Dev.Session.Keys.AppSKey = v3DeviceKey{Key: lorawanDevice.AppSKey.String()}
		}
		if lorawanDevice.NwkSKey != nil {
			v3Dev.Session.Keys.FNwkSIntKey = v3DeviceKey{Key: lorawanDevice.NwkSKey.String()}
		}
		v3Dev.Session.LastFCntUp = lorawanDevice.FCntUp
		v3Dev.Session.LastNFCntDown = lorawanDevice.FCntDown
	}

	v3Dev.MACSettings.Supports32BitFCnt = lorawanDevice.Uses32BitFCnt
	v3Dev.MACSettings.ResetsFCnt = lorawanDevice.DisableFCntCheck

	// Make behavior as similar as possible to v2

	v3Dev.MACVersion = "MAC_V1_0_2"
	v3Dev.PHYVersion = "PHY_V1_0_2_REV_B"
	v3Dev.SupportsJoin = true

	// TODO: Frequency Plan ID

	v3Dev.MACSettings.Rx1Delay.Value = "RX_DELAY_1" // TODO: Is this necessary?
	{                                               // TODO: If on EU, uses SF9 for RX2
		// v3Dev.MACSettings.Rx2DataRateIndex.Value = "DATA_RATE_3"
	}
	{ // TODO: Depending on region, set factory_preset_frequencies
		// v3Dev.MACSettings.FactoryPresetFrequencies = []uint64{
		// 	868100000, 868300000, 868500000,
		// 	867100000, 867300000, 867500000, 867700000, 867900000,
		// }
	}

	return v3Dev
}

var devicesExportCmd = &cobra.Command{
	Use:     "export [Device ID]",
	Short:   "Export a device",
	Long:    `ttnctl devices export exports a device to an external format.`,
	Example: `$ ttnctl devices export test | ttn-lw-cli end-devices create --application-id app`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		devID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(devID, "Device ID"); err != nil {
			ctx.Fatal(err.Error())
		}

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		format, _ := cmd.Flags().GetString("format")

		dev, err := manager.GetDevice(appID, devID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get existing device.")
		}
		switch format {
		case "v3":
			b, _ := json.Marshal(exportV3Device(dev))
			fmt.Println(string(b))
		}
	},
}

func init() {
	devicesCmd.AddCommand(devicesExportCmd)
	devicesExportCmd.Flags().String("format", "v3", "Formatting: v3")
}

package models

import (
	"fmt"
	"bytes"
	"errors"
  "github.com/spf13/viper"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/components/handler"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/TheThingsNetwork/ttn/api/auth"
	"golang.org/x/net/context"
	"encoding/json"
)

type Device struct {
	DevEUI   EUI     `json:"eui,omitempty"`
	AppEUI   EUI     `json:"app"`
	Type     string  `json:"type"`
	Address  []byte  `json:"address"`
	FCntUp   uint32  `json:"fcnt_up"`
	FCntDown uint32  `json:"fcnt_down"`
	Active   bool    `json:"activated"`
	AppSKey  SKey    `json:"app_secret"`
	NwkSKey  SKey    `json:"nwk_secret"`
}

type dev_ Device
func (dev Device) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
			dev_
			Type string `json:"@type"`
			Self string `json:"@self"`
		}{
			dev_:  dev_(dev),
			Type: "device",
			Self: fmt.Sprintf("/api/v1/applications/%s/devices/%s", WriteEUI(dev.AppEUI), WriteEUI(dev.DevEUI)),
		})
}

func getHandlerManager() core.AuthHandlerClient {
	handlerUri := viper.GetString("ttn-handler")
	h, err     := handler.NewClient(handlerUri)
	if err != nil {
		fmt.Println("Could not connect: %v", err)
	}
	return h
}

// Fetch devices for specific app
func DevicesForApp (accessToken auth.Token, appEui EUI) ([]Device, error) {
	println(string(accessToken))
	manager := getHandlerManager()
	res, err := manager.ListDevices(context.Background(), &core.ListDevicesHandlerReq{
		Token:  string(accessToken),
		AppEUI: []byte(appEui),
	})

	if err != nil {
		return nil, err
	}

	result := make([]Device, 0, len(res.ABP) + len(res.OTAA))

	for i, device := range res.ABP {
		result[i] = Device{
			DevEUI:   nil,
			AppEUI:   appEui,
			Type:     "personal",
			Address:  device.DevAddr,
			FCntUp:   device.FCntUp,
			FCntDown: device.FCntDown,
			Active:   true,
			AppSKey:  device.AppSKey,
			NwkSKey:  device.NwkSKey,
		}
	}

	n := len(res.ABP)
	for i, device := range res.OTAA {
		result[i + n] = Device{
			DevEUI:   device.DevEUI,
			AppEUI:   appEui,
			Type:     "dynamic",
			Address:  device.DevAddr,
			FCntUp:   device.FCntUp,
			FCntDown: device.FCntDown,
			Active:   len(device.DevAddr) == 0,
			AppSKey:  device.AppSKey,
			NwkSKey:  device.NwkSKey,
		}
	}

	return result, nil
}

func DeviceInfo(accessToken auth.Token, appEui EUI, devEui EUI) (dev Device, err error) {
	devices, err := DevicesForApp (accessToken, appEui)
	if err != nil {
		return dev, err
	}

	for _, device := range devices {
		if bytes.Equal(device.DevEUI, devEui) {
			return device, nil
		}
	}

	return dev, errors.New("Device with that id does not exist")
}


func RegisterOTAA(accessToken auth.Token, appEUI EUI, devEUI EUI, rest ...SKey) error {
	var appKey SKey
	switch len(rest) {
		case 0:
			// allow for random appKey
			appKey = SKey(random.Bytes(16))
		case 1:
			// get appKey from param
			appKey = rest[0]

		default:
			return errors.New("illegal amount of parameters passed to RegisterOTAA")
	}

	manager := getHandlerManager()
	res, err := manager.UpsertOTAA(context.Background(), &core.UpsertOTAAHandlerReq{
		Token:  string(accessToken),
		AppEUI: []byte(appEUI),
		DevEUI: []byte(devEUI),
		AppKey: appKey,
	})

	switch {
		case err != nil:
			return err
		case res == nil:
			return errors.New("unable to create device")
		default:
			return nil
	}
}

func RegisterABP(accessToken auth.Token, appEUI EUI, addr []byte, rest ...SKey) error {
	var appKey SKey
	var nwkKey SKey
	switch len(rest) {
		case 0:
			// generate both keys
			appKey = SKey(random.Bytes(16))
			nwkKey = SKey(random.Bytes(16))
		case 1:
			// get appKey from param, generate devKey
			appKey = rest[0]
			nwkKey = SKey(random.Bytes(16))
		case 2:
			// get both keys from params
			appKey = rest[0]
			nwkKey = rest[1]
		default:
			return errors.New("illegal amount of parameters passed to RegisterOTAA")
	}

	manager := getHandlerManager()
	res, err := manager.UpsertABP(context.Background(), &core.UpsertABPHandlerReq{
		Token:   string(accessToken),
		AppEUI:  []byte(appEUI),
		DevAddr: addr,
		AppSKey: appKey,
		NwkSKey: nwkKey,
	})

	switch {
		case err != nil:
			return err
		case res == nil:
			return errors.New("unable to create device")
		default:
			return nil
	}
}

func RegisterDefault(accessToken auth.Token, appEUI EUI, appKey SKey) error {
	manager  := getHandlerManager()
	_, err := manager.SetDefaultDevice(context.Background(), &core.SetDefaultDeviceReq{
		Token:   string(accessToken),
		AppEUI:  []byte(appEUI),
		AppKey:  []byte(appKey),
	})
	return err
}




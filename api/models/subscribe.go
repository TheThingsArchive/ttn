package models

import (
	"github.com/TheThingsNetwork/ttn/api/auth"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/spf13/viper"
	"fmt"
)


func ConnectMQTTClient(token auth.Token, appEUI EUI) (client mqtt.Client, err error) {
	broker := fmt.Sprintf("tcp://%s", viper.GetString("mqtt-broker"))

	app, err := GetApplication(token, appEUI)
	if err != nil {
		return client, err
	}

	client = mqtt.NewClient(nil, "ttnhttp", WriteEUI(app.EUI), string(app.AccessKeys[0]), broker)
	err    = client.Connect();
	return client, err
}

func SubscribeDevice (token auth.Token, appEUI EUI, devEUI EUI, cb func(dataUp core.DataUpAppReq)) (func (), error) {
	client, err := ConnectMQTTClient(token, appEUI)
	if err != nil {
		return func() {}, err
	}

	tok := client.SubscribeDeviceUplink(appEUI, devEUI, func (client mqtt.Client, appEUI []byte, devEUI []byte, dataUp core.DataUpAppReq) {
		cb(dataUp)
	})

	if tok.Wait(); tok.Error() != nil {
		return func() {}, err
	}

	return func () {
		client.Disconnect()
	}, nil
}

func Subscribe(token auth.Token, appEUI EUI, cb func(dataUp core.DataUpAppReq)) (func (), error) {
	client, err := ConnectMQTTClient(token, appEUI)
	if err != nil {
		return func() {}, err
	}

	tok := client.SubscribeDeviceUplink(appEUI, nil, func (client mqtt.Client, appEUI []byte, devEUI []byte, dataUp core.DataUpAppReq) {
		cb(dataUp)
	})

	if tok.Wait(); tok.Error() != nil {
		return func() {}, err
	}

	return func () {
		client.Disconnect()
	}, nil
}


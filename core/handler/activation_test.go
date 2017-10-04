// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_protocol "github.com/TheThingsNetwork/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
	gogo "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/assertions"
)

func TestHandleActivationChallenge(t *testing.T) {
	a := New(t)

	h := &handler{
		Component:    &component.Component{Ctx: GetLogger(t, "TestHandleActivationChallenge")},
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-activation-challenge"),
		devices:      device.NewRedisDeviceStore(GetRedisClient(), "handler-test-activation-challenge"),
		qEvent:       make(chan *types.DeviceEvent, 10),
	}
	h.InitStatus()

	appEUI, devEUI := types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}, types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8}
	appID, devID := appEUI.String(), devEUI.String()
	appKey := types.AppKey{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}
	dev := &device.Device{AppID: appID, DevID: devID}
	req := &pb_broker.ActivationChallengeRequest{AppID: appID, DevID: devID}

	// Device does not exist
	_, err := h.HandleActivationChallenge(req)
	a.So(err, ShouldNotBeNil)

	h.devices.Set(dev)
	defer func() { h.devices.Delete(appID, devID) }()

	// Device does not have AppKey
	_, err = h.HandleActivationChallenge(req)
	a.So(err, ShouldNotBeNil)

	dev.AppKey = appKey
	h.devices.Set(dev)

	// No LoRaWAN payload
	_, err = h.HandleActivationChallenge(req)
	a.So(err, ShouldNotBeNil)

	pld := lorawan.PHYPayload{MHDR: lorawan.MHDR{MType: lorawan.JoinRequest}}
	pld.MACPayload = &lorawan.JoinRequestPayload{AppEUI: lorawan.EUI64(appEUI), DevEUI: lorawan.EUI64(devEUI), DevNonce: [2]byte{1, 2}}
	req.Payload, _ = pld.MarshalBinary()

	pld.SetMIC(lorawan.AES128Key(appKey))

	res, err := h.HandleActivationChallenge(req)
	a.So(err, ShouldBeNil)

	res.UnmarshalPayload()
	a.So(res.GetMessage().GetLoRaWAN().MIC, ShouldResemble, pld.MIC[:])
}

func TestHandleActivation(t *testing.T) {
	a := New(t)

	h := &handler{
		Component:    &component.Component{Ctx: GetLogger(t, "TestHandleActivation")},
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-activation"),
		devices:      device.NewRedisDeviceStore(GetRedisClient(), "handler-test-activation"),
		qEvent:       make(chan *types.DeviceEvent, 10),
	}
	var eventsReceived uint
	go func() {
		for ev := range h.qEvent {
			h.Ctx.Infof("EVENT: %v", ev)
			eventsReceived++
		}
	}()
	h.InitStatus()

	devAddr := types.DevAddr{1, 2, 3, 4}
	appEUI, devEUI := types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}, types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8}
	appID, devID := appEUI.String(), devEUI.String()
	if os.Getenv("APP_ID") != "" {
		appID = os.Getenv("APP_ID")
	}
	appKey := types.AppKey{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}
	dev := &device.Device{
		AppID:  appID,
		DevID:  devID,
		AppEUI: appEUI,
		DevEUI: devEUI,
	}
	app := &application.Application{
		AppID: appID,
	}
	req := &pb_broker.DeduplicatedDeviceActivationRequest{
		AppID:  appID,
		DevID:  devID,
		AppEUI: appEUI,
		DevEUI: devEUI,
	}

	// No ResponseTemplate
	_, err := h.HandleActivation(req)
	a.So(err, ShouldNotBeNil)

	{
		req.ResponseTemplate = new(pb_broker.DeviceActivationResponse)
		req.ResponseTemplate.Message = new(pb_protocol.Message)
		msg := req.ResponseTemplate.Message.InitLoRaWAN()
		msg.MType = pb_lorawan.MType_JOIN_ACCEPT
		msg.Payload = &pb_lorawan.Message_JoinAcceptPayload{JoinAcceptPayload: &pb_lorawan.JoinAcceptPayload{}}
		req.ResponseTemplate.Payload = msg.PHYPayloadBytes()
		req.ResponseTemplate.DownlinkOption = new(pb_broker.DownlinkOption)
	}

	// Device does not exist
	_, err = h.HandleActivation(req)
	a.So(err, ShouldNotBeNil)

	h.applications.Set(app)
	defer func() { h.applications.Delete(appID) }()

	h.devices.Set(dev)
	defer func() { h.devices.Delete(appID, devID) }()

	// Device does not have AppKey
	_, err = h.HandleActivation(req)
	a.So(err, ShouldNotBeNil)

	dev.AppKey = appKey
	dev.AppKey[1] = 0xff
	h.devices.Set(dev)

	// No LoRaWAN activation metadata
	_, err = h.HandleActivation(req)
	a.So(err, ShouldNotBeNil)

	req.ActivationMetadata = &pb_protocol.ActivationMetadata{Protocol: &pb_protocol.ActivationMetadata_LoRaWAN{LoRaWAN: &pb_lorawan.ActivationMetadata{
		AppEUI:  appEUI,
		DevEUI:  devEUI,
		DevAddr: &devAddr,
	}}}

	// Invalid payload
	_, err = h.HandleActivation(req)
	a.So(err, ShouldNotBeNil)

	{
		req.Message = new(pb_protocol.Message)
		msg := req.Message.InitLoRaWAN()
		msg.MType = pb_lorawan.MType_JOIN_REQUEST
		msg.Payload = &pb_lorawan.Message_JoinRequestPayload{JoinRequestPayload: &pb_lorawan.JoinRequestPayload{
			AppEUI: appEUI,
			DevEUI: devEUI,
		}}
		phy := msg.PHYPayload()
		phy.SetMIC(lorawan.AES128Key(appKey))
		req.Payload, _ = phy.MarshalBinary()
	}

	// Wrong AppKey
	_, err = h.HandleActivation(req)
	a.So(err, ShouldNotBeNil)

	dev.AppKey = appKey
	h.devices.Set(dev)

	// Valid join
	res, err := h.HandleActivation(req)
	a.So(err, ShouldBeNil)
	a.So(res.ActivationMetadata.GetLoRaWAN().DevEUI, ShouldResemble, devEUI)

	// TODO: Check response
	// TODO: Check DB contents

	// DevNonce Re-use
	_, err = h.HandleActivation(req)
	a.So(err, ShouldNotBeNil)

	// Now we create a "default" device
	otherDevEUI := types.DevEUI{1, 1, 1, 1, 1, 1, 1, 1}
	dev = &device.Device{
		AppID:  appID,
		DevID:  "default",
		AppEUI: appEUI,
		AppKey: appKey,
	}

	h.devices.Set(dev)
	defer func() { h.devices.Delete(appID, "default") }()

	{
		req.DevEUI = otherDevEUI
		req.ActivationMetadata.GetLoRaWAN().DevEUI = otherDevEUI
		req.Message.GetLoRaWAN().GetJoinRequestPayload().DevEUI = otherDevEUI
		phy := req.Message.GetLoRaWAN().PHYPayload()
		phy.SetMIC(lorawan.AES128Key(appKey))
		req.Payload, _ = phy.MarshalBinary()
		req.DevID = "default"
	}

	// No access key set
	_, err = h.HandleActivation(req)
	a.So(err, ShouldNotBeNil)

	if token := os.Getenv("APP_TOKEN"); token != "" {
		app.RegisterOnJoinAccessKey = token
	} else {
		app.RegisterOnJoinAccessKey = "test.key"
	}
	h.applications.Set(app)

	// Can't get access key
	_, err = h.HandleActivation(req)
	a.So(err, ShouldNotBeNil)

	time.Sleep(200 * time.Millisecond)
	a.So(eventsReceived, ShouldEqual, 10) // 10 activation error events. One for each HandleActivation

	for _, env := range strings.Split("ACCOUNT_SERVER_PROTO ACCOUNT_SERVER_USERNAME ACCOUNT_SERVER_PASSWORD ACCOUNT_SERVER_URL APP_ID APP_TOKEN", " ") {
		if os.Getenv(env) == "" {
			t.Skipf("Skipping test that needs auth server: %s not configured", env)
		}
	}
	h.Config.AuthServers = map[string]string{
		"ttn-account-v2": fmt.Sprintf("%s://%s:%s@%s",
			os.Getenv("ACCOUNT_SERVER_PROTO"),
			os.Getenv("ACCOUNT_SERVER_USERNAME"),
			os.Getenv("ACCOUNT_SERVER_PASSWORD"),
			os.Getenv("ACCOUNT_SERVER_URL"),
		),
	}
	h.InitAuth()

	ctrl := gomock.NewController(t)
	ttnDeviceManager := pb_lorawan.NewMockDeviceManagerClient(ctrl)
	h.ttnDeviceManager = ttnDeviceManager
	ttnDeviceManager.EXPECT().SetDevice(gomock.Any(), gomock.Any()).Return(new(gogo.Empty), nil)

	res, err = h.HandleActivation(req)
	a.So(err, ShouldBeNil)
	a.So(res.ActivationMetadata.GetLoRaWAN().DevEUI, ShouldResemble, otherDevEUI)

	// TODO: Check response
	// TODO: Check DB contents

}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"testing"
	"time"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestDownlink(t *testing.T) {
	a := New(t)

	gtw := NewGateway(GetLogger(t, "TestDownlink"), "test", nil)
	gtw.schedule.Sync(0)

	now := time.Now()
	newUplink := func() *pb_router.UplinkMessage {
		return &pb_router.UplinkMessage{
			GatewayMetadata: &pb_gateway.RxMetadata{
				Time: now.UnixNano(),
			},
			ProtocolMetadata: &pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_Lorawan{Lorawan: &pb_lorawan.Metadata{}}},
		}
	}

	{
		uplink := newUplink()
		options, err := gtw.GetDownlinkOptions(uplink)
		a.So(err, ShouldNotBeNil)
		a.So(options, ShouldBeEmpty)
	}

	{
		_, err := gtw.GetDownlinkOption(869525000, time.Second)
		a.So(err, ShouldNotBeNil)
	}

	gtw.setFrequencyPlan("EU_863_870")
	gtw.frequencyPlan.RX1Delays = []time.Duration{
		time.Second,
		5 * time.Second,
	}

	{
		uplink := newUplink()
		options, err := gtw.GetDownlinkOptions(uplink)
		a.So(err, ShouldNotBeNil)
		a.So(options, ShouldBeEmpty)
	}

	{
		_, err := gtw.GetDownlinkOption(869525000, time.Second)
		a.So(err, ShouldNotBeNil)
	}

	gtw.schedule.Subscribe()
	defer gtw.schedule.Stop()

	{
		uplink := newUplink()
		options, err := gtw.GetDownlinkOptions(uplink)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[0].GetGatewayConfig().Timestamp, ShouldEqual, 2000000) // RX2
		a.So(options[1].GetGatewayConfig().Timestamp, ShouldEqual, 6000000) // RX2
		a.So(uplink.Trace, ShouldNotBeNil)
	}

	{
		uplink := newUplink()
		uplink.GetGatewayMetadata().Frequency = 868100000
		uplink.GetProtocolMetadata().GetLorawan().DataRate = "SF7BW125"
		uplink.GetProtocolMetadata().GetLorawan().CodingRate = "4/5"
		options, err := gtw.GetDownlinkOptions(uplink)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 4)
		a.So(options[0].GetGatewayConfig().Timestamp, ShouldEqual, 2000000) // RX2
		a.So(options[1].GetGatewayConfig().Timestamp, ShouldEqual, 6000000) // RX2
		a.So(options[2].GetGatewayConfig().Timestamp, ShouldEqual, 1000000) // RX1
		a.So(options[3].GetGatewayConfig().Timestamp, ShouldEqual, 5000000) // RX1
	}

	{
		option, err := gtw.GetDownlinkOption(869525000, 500*time.Millisecond)
		a.So(err, ShouldBeNil)
		a.So(option.PossibleConflicts, ShouldEqual, 1)
	}
}

func TestBuildDownlinkOptionsFrequency(t *testing.T) {
	a := New(t)

	gtw := NewGateway(GetLogger(t, "TestDownlink"), "test", nil)

	gtw.schedule.Subscribe()
	defer gtw.schedule.Stop()

	gtw.setFrequencyPlan("EU_863_870")
	gtw.frequencyPlan.RX1Delays = []time.Duration{time.Second}

	now := time.Now()
	newUplink := func() *pb_router.UplinkMessage {
		return &pb_router.UplinkMessage{
			GatewayMetadata: &pb_gateway.RxMetadata{
				Time: now.UnixNano(),
			},
			ProtocolMetadata: &pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_Lorawan{Lorawan: &pb_lorawan.Metadata{
				Modulation: pb_lorawan.Modulation_LORA,
				DataRate:   "SF7BW125",
				CodingRate: "4/5",
			}}},
		}
	}

	ttnEUFrequencies := []uint64{
		868100000,
		868300000,
		868500000,
		867100000,
		867300000,
		867500000,
		867700000,
		867900000,
	}
	for _, freq := range ttnEUFrequencies {
		up := newUplink()
		up.GatewayMetadata.Frequency = freq
		options, err := gtw.buildDownlinkOptions(up)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, freq)
		a.So(options[0].GatewayConfig.Frequency, ShouldEqual, 869525000)
	}

	// Unsupported frequencies use only RX2 for downlink
	up := newUplink()
	up.GatewayMetadata.Frequency = 867800000
	options, err := gtw.buildDownlinkOptions(up)
	a.So(err, ShouldBeNil)
	a.So(options, ShouldHaveLength, 1)

	gtw.setFrequencyPlan("US_902_928")
	gtw.frequencyPlan.RX1Delays = []time.Duration{time.Second}

	ttnUSFrequencies := map[uint64]uint64{
		903900000: 923300000,
		904100000: 923900000,
		904300000: 924500000,
		904500000: 925100000,
		904700000: 925700000,
		904900000: 926300000,
		905100000: 926900000,
		905300000: 927500000,
	}
	for upFreq, downFreq := range ttnUSFrequencies {
		up := newUplink()
		up.GatewayMetadata.Frequency = upFreq
		options, err := gtw.buildDownlinkOptions(up)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, downFreq)
		a.So(options[0].GatewayConfig.Frequency, ShouldEqual, 923300000)
	}

	gtw.setFrequencyPlan("AU_915_928")
	gtw.frequencyPlan.RX1Delays = []time.Duration{time.Second}

	ttnAUFrequencies := map[uint64]uint64{
		916800000: 923300000,
		917000000: 923900000,
		917200000: 924500000,
		917400000: 925100000,
		917600000: 925700000,
		917800000: 926300000,
		918000000: 926900000,
		918200000: 927500000,
	}
	for upFreq, downFreq := range ttnAUFrequencies {
		up := newUplink()
		up.GatewayMetadata.Frequency = upFreq
		options, err := gtw.buildDownlinkOptions(up)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, downFreq)
		a.So(options[0].GatewayConfig.Frequency, ShouldEqual, 923300000)
	}
}

func TestBuildDownlinkOptionsDataRate(t *testing.T) {
	a := New(t)

	gtw := NewGateway(GetLogger(t, "TestDownlink"), "test", nil)

	gtw.schedule.Subscribe()
	defer gtw.schedule.Stop()

	gtw.setFrequencyPlan("EU_863_870")
	gtw.frequencyPlan.RX1Delays = []time.Duration{time.Second}

	now := time.Now()
	newUplink := func() *pb_router.UplinkMessage {
		return &pb_router.UplinkMessage{
			GatewayMetadata: &pb_gateway.RxMetadata{
				Time: now.UnixNano(),
			},
			ProtocolMetadata: &pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_Lorawan{Lorawan: &pb_lorawan.Metadata{
				Modulation: pb_lorawan.Modulation_LORA,
				DataRate:   "SF7BW125",
				CodingRate: "4/5",
			}}},
		}
	}

	ttnEUDataRates := []string{
		"SF7BW125",
		"SF8BW125",
		"SF9BW125",
		"SF10BW125",
		"SF11BW125",
		"SF12BW125",
	}
	for _, dr := range ttnEUDataRates {
		up := newUplink()
		up.ProtocolMetadata.GetLorawan().DataRate = dr
		up.GatewayMetadata.Frequency = 868100000
		options, err := gtw.buildDownlinkOptions(up)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, dr)
		a.So(options[0].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF9BW125")
	}

	gtw.setFrequencyPlan("US_902_928")
	gtw.frequencyPlan.RX1Delays = []time.Duration{time.Second}

	// Test 500kHz channel
	up := newUplink()
	up.GatewayMetadata.Frequency = 904600000
	up.ProtocolMetadata.GetLorawan().DataRate = "SF8BW500"
	options, err := gtw.buildDownlinkOptions(up)
	a.So(err, ShouldBeNil)
	a.So(options, ShouldHaveLength, 2)
	a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF7BW500")

	ttnUSDataRates := map[string]string{
		"SF7BW125":  "SF7BW500",
		"SF8BW125":  "SF8BW500",
		"SF9BW125":  "SF9BW500",
		"SF10BW125": "SF10BW500",
	}
	for drUp, drDown := range ttnUSDataRates {
		up := newUplink()
		up.GatewayMetadata.Frequency = 903900000
		up.ProtocolMetadata.GetLorawan().DataRate = drUp
		options, err := gtw.buildDownlinkOptions(up)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, drDown)
	}

	gtw.setFrequencyPlan("AU_915_928")
	gtw.frequencyPlan.RX1Delays = []time.Duration{time.Second}

	// Test 500kHz channel
	up = newUplink()
	up.GatewayMetadata.Frequency = 917500000
	up.ProtocolMetadata.GetLorawan().DataRate = "SF8BW500"
	options, err = gtw.buildDownlinkOptions(up)
	a.So(err, ShouldBeNil)
	a.So(options, ShouldHaveLength, 2)
	a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF7BW500")

	ttnAUDataRates := map[string]string{
		"SF7BW125":  "SF7BW500",
		"SF8BW125":  "SF8BW500",
		"SF9BW125":  "SF9BW500",
		"SF10BW125": "SF10BW500",
	}
	for drUp, drDown := range ttnAUDataRates {
		up := newUplink()
		up.GatewayMetadata.Frequency = 916800000
		up.ProtocolMetadata.GetLorawan().DataRate = drUp
		options, err = gtw.buildDownlinkOptions(up)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, drDown)
	}

	gtw.setFrequencyPlan("CN_470_510")
	gtw.frequencyPlan.RX1Delays = []time.Duration{time.Second}

	ttnCNDataRates := []string{
		"SF7BW125",
		"SF8BW125",
		"SF9BW125",
		"SF10BW125",
		"SF11BW125",
		"SF12BW125",
	}
	for _, dr := range ttnCNDataRates {
		up := newUplink()
		up.GatewayMetadata.Frequency = 470300000
		up.ProtocolMetadata.GetLorawan().DataRate = dr
		options, err = gtw.buildDownlinkOptions(up)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, dr)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, 500300000)
		a.So(options[0].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF12BW125")
		a.So(options[0].GatewayConfig.Frequency, ShouldEqual, 505300000)
	}

	gtw.setFrequencyPlan("AS_923")
	gtw.frequencyPlan.RX1Delays = []time.Duration{time.Second}

	ttnASDataRates := map[string]string{
		"SF7BW125":  "SF7BW125",
		"SF8BW125":  "SF8BW125",
		"SF9BW125":  "SF9BW125",
		"SF10BW125": "SF10BW125",
		"SF11BW125": "SF10BW125", // MinDR = 2
		"SF12BW125": "SF10BW125", // MinDR = 2
	}
	for drUp, drDown := range ttnASDataRates {
		up := newUplink()
		up.GatewayMetadata.Frequency = 923200000
		up.ProtocolMetadata.GetLorawan().DataRate = drUp
		options, err = gtw.buildDownlinkOptions(up)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, drDown)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, 923200000)
		a.So(options[0].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF10BW125")
		a.So(options[0].GatewayConfig.Frequency, ShouldEqual, 923200000)
	}

	gtw.setFrequencyPlan("KR_920_923")
	gtw.frequencyPlan.RX1Delays = []time.Duration{time.Second}

	ttnKRDataRates := []string{
		"SF7BW125",
		"SF8BW125",
		"SF9BW125",
		"SF10BW125",
		"SF11BW125",
		"SF12BW125",
	}
	for _, dr := range ttnKRDataRates {
		up := newUplink()
		up.GatewayMetadata.Frequency = 922100000
		up.ProtocolMetadata.GetLorawan().DataRate = dr
		options, err = gtw.buildDownlinkOptions(up)
		a.So(err, ShouldBeNil)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, dr)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, 922100000)
		a.So(options[0].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF12BW125")
		a.So(options[0].GatewayConfig.Frequency, ShouldEqual, 921900000)
	}
}

func TestSubscribeUnsubscribeDownlink(t *testing.T) {
	a := New(t)
	gtw := NewGateway(GetLogger(t, "TestSubscribeUnsubscribeDownlink"), "test", nil)

	a.So(gtw.HasDownlink(), ShouldBeFalse)

	err := gtw.ScheduleDownlink("", &pb_router.DownlinkMessage{})
	a.So(err, ShouldNotBeNil)

	downlink := gtw.SubscribeDownlink()
	a.So(downlink, ShouldNotBeNil)

	err = gtw.ScheduleDownlink("doesnotexist", &pb_router.DownlinkMessage{})
	a.So(err, ShouldNotBeNil)

	err = gtw.ScheduleDownlink("", &pb_router.DownlinkMessage{Payload: []byte{1, 2, 3, 4}})
	a.So(err, ShouldBeNil)

	select {
	case msg := <-downlink:
		a.So(msg.Payload, ShouldResemble, []byte{1, 2, 3, 4})
	case <-time.After(time.Second):
		t.Fatal("Did not receive on downlink channel")
	}

	gtw.StopDownlink()

	select {
	case _, ok := <-downlink:
		a.So(ok, ShouldBeFalse)
	case <-time.After(time.Second):
		t.Fatal("Downlink channel was not closed")
	}

}

func TestHandleDownlink(t *testing.T) {
	a := New(t)
	gtw := NewGateway(GetLogger(t, "TestHandleDownlink"), "test", nil)

	oldLastSeen := gtw.lastSeen

	err := gtw.HandleDownlink(pb_router.RandomDownlinkMessage())
	a.So(err, ShouldBeNil)

	a.So(gtw.lastSeen, ShouldNotEqual, oldLastSeen)

	rate, err := gtw.downlink.Get(time.Now(), time.Second)
	a.So(err, ShouldBeNil)
	a.So(rate, ShouldNotEqual, 0)
}

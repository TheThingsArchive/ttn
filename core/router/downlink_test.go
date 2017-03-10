// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"sync"
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/monitor"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

// newReferenceDownlink returns a default uplink message
func newReferenceDownlink() *pb.DownlinkMessage {
	up := &pb.DownlinkMessage{
		Payload: make([]byte, 20),
		ProtocolConfiguration: &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{Lorawan: &pb_lorawan.TxConfiguration{
			CodingRate: "4/5",
			DataRate:   "SF7BW125",
			Modulation: pb_lorawan.Modulation_LORA,
		}}},
		GatewayConfiguration: &pb_gateway.TxConfiguration{
			Timestamp: 100,
			Frequency: 868100000,
		},
	}
	return up
}

func TestHandleDownlink(t *testing.T) {
	a := New(t)

	logger := GetLogger(t, "TestHandleDownlink")
	r := &router{
		Component: &component.Component{
			Ctx:     logger,
			Monitor: monitor.NewClient(monitor.DefaultClientConfig),
		},
		gateways: map[string]*gateway.Gateway{},
	}
	r.InitStatus()

	gtwID := "eui-0102030405060708"
	id, _ := r.getGateway(gtwID).Schedule.GetOption(0, 10*1000)
	err := r.HandleDownlink(&pb_broker.DownlinkMessage{
		Payload: []byte{},
		DownlinkOption: &pb_broker.DownlinkOption{
			GatewayId:      gtwID,
			Identifier:     id,
			ProtocolConfig: &pb_protocol.TxConfiguration{},
			GatewayConfig:  &pb_gateway.TxConfiguration{},
		},
	})

	a.So(err, ShouldBeNil)
}

func TestSubscribeUnsubscribeDownlink(t *testing.T) {
	a := New(t)

	logger := GetLogger(t, "TestSubscribeUnsubscribeDownlink")
	r := &router{
		Component: &component.Component{
			Ctx:     logger,
			Monitor: monitor.NewClient(monitor.DefaultClientConfig),
		},
		gateways: map[string]*gateway.Gateway{},
	}
	r.InitStatus()

	gtwID := "eui-0102030405060708"
	gateway.Deadline = 1 * time.Millisecond
	gtw := r.getGateway(gtwID)
	gtw.Schedule.Sync(0)
	id, _ := gtw.Schedule.GetOption(5000, 10*1000)

	ch, err := r.SubscribeDownlink(gtwID, "")
	a.So(err, ShouldBeNil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var gotDownlink bool
		for dl := range ch {
			gotDownlink = true
			a.So(dl.Payload, ShouldResemble, []byte{0x02})
		}
		a.So(gotDownlink, ShouldBeTrue)
		wg.Done()
	}()

	r.HandleDownlink(&pb_broker.DownlinkMessage{
		Payload: []byte{0x02},
		DownlinkOption: &pb_broker.DownlinkOption{
			GatewayId:      gtwID,
			Identifier:     id,
			ProtocolConfig: &pb_protocol.TxConfiguration{},
			GatewayConfig:  &pb_gateway.TxConfiguration{},
		},
	})

	// Wait for the downlink to arrive
	<-time.After(10 * time.Millisecond)

	err = r.UnsubscribeDownlink(gtwID, "")
	a.So(err, ShouldBeNil)

	wg.Wait()
}

func TestUplinkBuildDownlinkOptions(t *testing.T) {
	a := New(t)

	r := &router{}

	// If something is incorrect, it just returns an empty list
	up := &pb.UplinkMessage{}
	gtw := gateway.NewGateway(GetLogger(t, "TestUplinkBuildDownlinkOptions"), "eui-0102030405060708")
	options := r.buildDownlinkOptions(up, false, gtw)
	a.So(options, ShouldBeEmpty)

	// The reference gateway and uplink work as expected
	gtw, up = newReferenceGateway(t, "EU_863_870"), newReferenceUplink()
	options = r.buildDownlinkOptions(up, false, gtw)
	a.So(options, ShouldHaveLength, 2)
	a.So(options[1].Score, ShouldBeLessThan, options[0].Score)

	// Check Delay
	a.So(options[1].GatewayConfig.Timestamp, ShouldEqual, 1000100)
	a.So(options[0].GatewayConfig.Timestamp, ShouldEqual, 2000100)

	// Check Frequency
	a.So(options[1].GatewayConfig.Frequency, ShouldEqual, 868100000)
	a.So(options[0].GatewayConfig.Frequency, ShouldEqual, 869525000)

	// Check Power
	a.So(options[1].GatewayConfig.Power, ShouldEqual, 14)
	a.So(options[0].GatewayConfig.Power, ShouldEqual, 27)

	// Check Data Rate
	a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF7BW125")
	a.So(options[0].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF9BW125")

	// Check Coding Rate
	a.So(options[1].ProtocolConfig.GetLorawan().CodingRate, ShouldEqual, "4/5")
	a.So(options[0].ProtocolConfig.GetLorawan().CodingRate, ShouldEqual, "4/5")

	// And for joins we want a different delay (both RX1 and RX2) and DataRate (RX2)
	gtw, up = newReferenceGateway(t, "EU_863_870"), newReferenceUplink()
	options = r.buildDownlinkOptions(up, true, gtw)
	a.So(options[1].GatewayConfig.Timestamp, ShouldEqual, 5000100)
	a.So(options[0].GatewayConfig.Timestamp, ShouldEqual, 6000100)
	a.So(options[0].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF12BW125")
}

func TestUplinkBuildDownlinkOptionsFrequencies(t *testing.T) {
	a := New(t)

	r := &router{}

	// Unsupported frequencies use only RX2 for downlink
	gtw, up := newReferenceGateway(t, "EU_863_870"), newReferenceUplink()
	up.GatewayMetadata.Frequency = 869300000
	options := r.buildDownlinkOptions(up, false, gtw)
	a.So(options, ShouldHaveLength, 1)

	// Supported frequencies use RX1 (on the same frequency) for downlink
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
		up = newReferenceUplink()
		up.GatewayMetadata.Frequency = freq
		options := r.buildDownlinkOptions(up, false, gtw)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, freq)
	}

	// Unsupported frequencies use only RX2 for downlink
	gtw, up = newReferenceGateway(t, "US_902_928"), newReferenceUplink()
	up.GatewayMetadata.Frequency = 923300000
	options = r.buildDownlinkOptions(up, false, gtw)
	a.So(options, ShouldHaveLength, 1)

	// Supported frequencies use RX1 (on the same frequency) for downlink
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
		up = newReferenceUplink()
		up.GatewayMetadata.Frequency = upFreq
		options := r.buildDownlinkOptions(up, false, gtw)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, downFreq)
	}

	// Unsupported frequencies use only RX2 for downlink
	gtw, up = newReferenceGateway(t, "AU_915_928"), newReferenceUplink()
	up.GatewayMetadata.Frequency = 923300000
	options = r.buildDownlinkOptions(up, false, gtw)
	a.So(options, ShouldHaveLength, 1)

	// Supported frequencies use RX1 (on the same frequency) for downlink
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
		up = newReferenceUplink()
		up.GatewayMetadata.Frequency = upFreq
		options := r.buildDownlinkOptions(up, false, gtw)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, downFreq)
	}
}

func TestUplinkBuildDownlinkOptionsDataRate(t *testing.T) {
	a := New(t)

	r := &router{}

	gtw := newReferenceGateway(t, "EU_863_870")

	// Supported datarates use RX1 (on the same datarate) for downlink
	ttnEUDataRates := []string{
		"SF7BW125",
		"SF8BW125",
		"SF9BW125",
		"SF10BW125",
		"SF11BW125",
		"SF12BW125",
	}
	for _, dr := range ttnEUDataRates {
		up := newReferenceUplink()
		up.ProtocolMetadata.GetLorawan().DataRate = dr
		options := r.buildDownlinkOptions(up, false, gtw)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, dr)
	}

	gtw = newReferenceGateway(t, "US_902_928")

	// Test 500kHz channel
	up := newReferenceUplink()
	up.GatewayMetadata.Frequency = 904600000
	up.ProtocolMetadata.GetLorawan().DataRate = "SF8BW500"
	options := r.buildDownlinkOptions(up, false, gtw)
	a.So(options, ShouldHaveLength, 2)
	a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF7BW500")

	// Supported datarates use RX1 (on the same datarate) for downlink
	ttnUSDataRates := map[string]string{
		"SF7BW125":  "SF7BW500",
		"SF8BW125":  "SF8BW500",
		"SF9BW125":  "SF9BW500",
		"SF10BW125": "SF10BW500",
	}
	for drUp, drDown := range ttnUSDataRates {
		up := newReferenceUplink()
		up.GatewayMetadata.Frequency = 903900000
		up.ProtocolMetadata.GetLorawan().DataRate = drUp
		options := r.buildDownlinkOptions(up, false, gtw)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, drDown)
	}

	gtw = newReferenceGateway(t, "AU_915_928")

	// Test 500kHz channel
	up = newReferenceUplink()
	up.GatewayMetadata.Frequency = 917500000
	up.ProtocolMetadata.GetLorawan().DataRate = "SF8BW500"
	options = r.buildDownlinkOptions(up, false, gtw)
	a.So(options, ShouldHaveLength, 2)
	a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF7BW500")

	// Supported datarates use RX1 (on the same datarate) for downlink
	ttnAUDataRates := map[string]string{
		"SF7BW125":  "SF7BW500",
		"SF8BW125":  "SF8BW500",
		"SF9BW125":  "SF9BW500",
		"SF10BW125": "SF10BW500",
	}
	for drUp, drDown := range ttnAUDataRates {
		up := newReferenceUplink()
		up.GatewayMetadata.Frequency = 916800000
		up.ProtocolMetadata.GetLorawan().DataRate = drUp
		options := r.buildDownlinkOptions(up, false, gtw)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, drDown)
	}

	gtw = newReferenceGateway(t, "CN_470_510")

	// Supported datarates use RX1 (on the same datarate) for downlink
	ttnCNDataRates := []string{
		"SF7BW125",
		"SF8BW125",
		"SF9BW125",
		"SF10BW125",
		"SF11BW125",
		"SF12BW125",
	}
	for _, dr := range ttnCNDataRates {
		up := newReferenceUplink()
		up.GatewayMetadata.Frequency = 470300000
		up.ProtocolMetadata.GetLorawan().DataRate = dr
		options := r.buildDownlinkOptions(up, false, gtw)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, dr)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, 500300000)
		a.So(options[0].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF12BW125")
		a.So(options[0].GatewayConfig.Frequency, ShouldEqual, 505300000)
	}

	gtw = newReferenceGateway(t, "AS_923")

	// Supported datarates use RX1 (on the same datarate) for downlink
	ttnASDataRates := map[string]string{
		"SF7BW125":  "SF7BW125",
		"SF8BW125":  "SF8BW125",
		"SF9BW125":  "SF9BW125",
		"SF10BW125": "SF10BW125",
		"SF11BW125": "SF10BW125", // MinDR = 2
		"SF12BW125": "SF10BW125", // MinDR = 2
	}
	for drUp, drDown := range ttnASDataRates {
		up := newReferenceUplink()
		up.GatewayMetadata.Frequency = 923200000
		up.ProtocolMetadata.GetLorawan().DataRate = drUp
		options := r.buildDownlinkOptions(up, false, gtw)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, drDown)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, 923200000)
		a.So(options[0].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF10BW125")
		a.So(options[0].GatewayConfig.Frequency, ShouldEqual, 923200000)
	}

	gtw = newReferenceGateway(t, "KR_920_923")

	// Supported datarates use RX1 (on the same datarate) for downlink
	ttnKRDataRates := []string{
		"SF7BW125",
		"SF8BW125",
		"SF9BW125",
		"SF10BW125",
		"SF11BW125",
		"SF12BW125",
	}
	for _, dr := range ttnKRDataRates {
		up := newReferenceUplink()
		up.GatewayMetadata.Frequency = 922100000
		up.ProtocolMetadata.GetLorawan().DataRate = dr
		options := r.buildDownlinkOptions(up, false, gtw)
		a.So(options, ShouldHaveLength, 2)
		a.So(options[1].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, dr)
		a.So(options[1].GatewayConfig.Frequency, ShouldEqual, 922100000)
		a.So(options[0].ProtocolConfig.GetLorawan().DataRate, ShouldEqual, "SF12BW125")
		a.So(options[0].GatewayConfig.Frequency, ShouldEqual, 921900000)
	}

}

// Note: This test uses r.buildDownlinkOptions which in turn calls computeDownlinkScores
func TestComputeDownlinkScores(t *testing.T) {
	a := New(t)
	r := &router{}
	gtw := newReferenceGateway(t, "EU_863_870")
	refScore := r.buildDownlinkOptions(newReferenceUplink(), false, gtw)[1].Score

	// Lower RSSI -> worse score
	testSubject := newReferenceUplink()
	testSubject.GatewayMetadata.Rssi = -80.0
	testSubjectgtw := newReferenceGateway(t, "EU_863_870")
	testSubjectScore := r.buildDownlinkOptions(testSubject, false, testSubjectgtw)[1].Score
	a.So(testSubjectScore, ShouldBeGreaterThan, refScore)

	// Lower SNR -> worse score
	testSubject = newReferenceUplink()
	testSubject.GatewayMetadata.Snr = 2.0
	testSubjectgtw = newReferenceGateway(t, "EU_863_870")
	testSubjectScore = r.buildDownlinkOptions(testSubject, false, testSubjectgtw)[1].Score
	a.So(testSubjectScore, ShouldBeGreaterThan, refScore)

	// Slower DataRate -> worse score
	testSubject = newReferenceUplink()
	testSubject.ProtocolMetadata.GetLorawan().DataRate = "SF8BW125"
	testSubjectgtw = newReferenceGateway(t, "EU_863_870")
	testSubjectScore = r.buildDownlinkOptions(testSubject, false, testSubjectgtw)[1].Score
	a.So(testSubjectScore, ShouldBeGreaterThan, refScore)

	// Gateway used for Rx -> worse score
	testSubject1 := newReferenceUplink()
	testSubject2 := newReferenceUplink()
	testSubject2.GatewayMetadata.Timestamp = 10000000
	testSubject2.GatewayMetadata.Frequency = 868500000
	testSubjectgtw = newReferenceGateway(t, "EU_863_870")
	testSubjectgtw.Utilization.AddRx(newReferenceUplink())
	testSubjectgtw.Utilization.Tick()
	testSubject1Score := r.buildDownlinkOptions(testSubject1, false, testSubjectgtw)[1].Score
	testSubject2Score := r.buildDownlinkOptions(testSubject2, false, testSubjectgtw)[1].Score
	a.So(testSubject1Score, ShouldBeGreaterThan, refScore)          // Because of Rx in the gateway
	a.So(testSubject2Score, ShouldBeGreaterThan, refScore)          // Because of Rx in the gateway
	a.So(testSubject1Score, ShouldBeGreaterThan, testSubject2Score) // Because of Rx on the same channel

	// European Alarm Band
	// NOTE: This frequency is not part of the TTN DownlinkChannels. This test
	// case makes sure we don't allow Tx on the alarm bands even if someone
	// changes the frequency plan.
	testSubject = newReferenceUplink()
	testSubject.GatewayMetadata.Frequency = 869300000
	testSubjectgtw = newReferenceGateway(t, "EU_863_870")
	options := r.buildDownlinkOptions(testSubject, false, testSubjectgtw)
	a.So(options, ShouldHaveLength, 1) // RX1 Removed
	a.So(options[0].GatewayConfig.Frequency, ShouldNotEqual, 869300000)

	// European Duty-cycle Enforcement
	testSubject = newReferenceUplink()
	testSubjectgtw = newReferenceGateway(t, "EU_863_870")
	for i := 0; i < 5; i++ {
		testSubjectgtw.Utilization.AddTx(newReferenceDownlink())
	}
	testSubjectgtw.Utilization.Tick()
	options = r.buildDownlinkOptions(testSubject, false, testSubjectgtw)
	a.So(options, ShouldHaveLength, 1) // RX1 Removed
	a.So(options[0].GatewayConfig.Frequency, ShouldNotEqual, 868100000)

	// European Duty-cycle Preferences - Prefer RX1 for low SF
	testSubject = newReferenceUplink()
	testSubject.ProtocolMetadata.GetLorawan().DataRate = "SF7BW125"
	options = r.buildDownlinkOptions(testSubject, false, newReferenceGateway(t, "EU_863_870"))
	a.So(options[1].Score, ShouldBeLessThan, options[0].Score)
	testSubject.ProtocolMetadata.GetLorawan().DataRate = "SF8BW125"
	options = r.buildDownlinkOptions(testSubject, false, newReferenceGateway(t, "EU_863_870"))
	a.So(options[1].Score, ShouldBeLessThan, options[0].Score)

	// European Duty-cycle Preferences - Prefer RX2 for high SF
	testSubject.ProtocolMetadata.GetLorawan().DataRate = "SF9BW125"
	options = r.buildDownlinkOptions(testSubject, false, newReferenceGateway(t, "EU_863_870"))
	a.So(options[1].Score, ShouldBeGreaterThan, options[0].Score)
	testSubject.ProtocolMetadata.GetLorawan().DataRate = "SF10BW125"
	options = r.buildDownlinkOptions(testSubject, false, newReferenceGateway(t, "EU_863_870"))
	a.So(options[1].Score, ShouldBeGreaterThan, options[0].Score)
	testSubject.ProtocolMetadata.GetLorawan().DataRate = "SF11BW125"
	options = r.buildDownlinkOptions(testSubject, false, newReferenceGateway(t, "EU_863_870"))
	a.So(options[1].Score, ShouldBeGreaterThan, options[0].Score)
	testSubject.ProtocolMetadata.GetLorawan().DataRate = "SF12BW125"
	options = r.buildDownlinkOptions(testSubject, false, newReferenceGateway(t, "EU_863_870"))
	a.So(options[1].Score, ShouldBeGreaterThan, options[0].Score)

	// Scheduling Conflicts
	testSubject1 = newReferenceUplink()
	testSubject2 = newReferenceUplink()
	testSubject2.GatewayMetadata.Timestamp = 2000000
	testSubjectgtw = newReferenceGateway(t, "EU_863_870")
	testSubjectgtw.Schedule.GetOption(1000100, 50000)
	testSubject1Score = r.buildDownlinkOptions(testSubject1, false, testSubjectgtw)[1].Score
	testSubject2Score = r.buildDownlinkOptions(testSubject2, false, testSubjectgtw)[1].Score
	a.So(testSubject1Score, ShouldBeGreaterThan, refScore) // Scheduling conflict with RX1
	a.So(testSubject2Score, ShouldEqual, refScore)         // No scheduling conflicts
}

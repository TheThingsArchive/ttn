package gateway

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func getStatusMessage() (status *StatusMessage, smap map[string]string) {
	t := int64(1462201853428843766)
	return &StatusMessage{
			Timestamp:    12345,
			Time:         t,
			Ip:           []string{"169.50.131.24", "2a03:8180:1401:14f::2"},
			Platform:     "The Things Gateway",
			ContactEmail: "contact@email.net",
			Description:  "Description",
			Gps: &GPSMetadata{
				Time:      t,
				Latitude:  52.3737171,
				Longitude: 4.884567,
				Altitude:  9,
			},
			Rtt:  12,
			RxIn: 42,
			RxOk: 41,
			TxIn: 52,
			TxOk: 51,
		}, map[string]string{
			"timestamp":     "12345",
			"time":          "1462201853428843766",
			"ip":            "169.50.131.24,2a03:8180:1401:14f::2",
			"platform":      "The Things Gateway",
			"contact_email": "contact@email.net",
			"description":   "Description",
			"gps.time":      "1462201853428843766",
			"gps.latitude":  "52.37372",
			"gps.longitude": "4.884567",
			"gps.altitude":  "9",
			"rtt":           "12",
			"rx_in":         "42",
			"rx_ok":         "41",
			"tx_in":         "52",
			"tx_ok":         "51",
		}
}

func TestToStringMap(t *testing.T) {
	a := New(t)
	status, expected := getStatusMessage()
	smap, err := status.ToStringStringMap(StatusMessageProperties...)
	a.So(err, ShouldBeNil)
	a.So(smap, ShouldResemble, expected)
}

func TestFromStringMap(t *testing.T) {
	a := New(t)
	status := &StatusMessage{}
	expected, smap := getStatusMessage()
	err := status.FromStringStringMap(smap)
	a.So(err, ShouldBeNil)
	a.So(status, ShouldResemble, expected)
}

// TODO: Test error cases

package gateway

import (
	"strings"

	"github.com/TheThingsNetwork/ttn/core/storage"
)

func (m *StatusMessage) ToHSlice() *storage.HSlice {
	slice := storage.NewHSlice()
	if m.Timestamp != 0 {
		slice.SetUint32("timestamp", m.Timestamp)
	}
	if m.Time != 0 {
		slice.SetInt64("time", m.Time)
	}
	if len(m.Ip) > 0 {
		slice.SetString("ip", strings.Join(m.Ip, ","))
	}
	if m.Platform != "" {
		slice.SetString("platform", m.Platform)
	}
	if m.ContactEmail != "" {
		slice.SetString("contact_email", m.ContactEmail)
	}
	if m.Description != "" {
		slice.SetString("description", m.Description)
	}
	if m.Gps != nil {
		if m.Gps.Time != 0 {
			slice.SetInt64("gps_time", m.Gps.Time)
		}
		if m.Gps.Latitude != 0 {
			slice.SetFloat32("latitude", m.Gps.Latitude)
		}
		if m.Gps.Longitude != 0 {
			slice.SetFloat32("longitude", m.Gps.Longitude)
		}
		if m.Gps.Altitude != 0 {
			slice.SetInt32("altitude", m.Gps.Altitude)
		}
	}
	if m.Rtt != 0 {
		slice.SetUint32("rtt", m.Rtt)
	}
	if m.RxIn != 0 {
		slice.SetUint32("rx_in", m.RxIn)
	}
	if m.RxOk != 0 {
		slice.SetUint32("rx_ok", m.RxOk)
	}
	if m.TxIn != 0 {
		slice.SetUint32("tx_in", m.TxIn)
	}
	if m.TxOk != 0 {
		slice.SetUint32("tx_ok", m.TxOk)
	}
	return slice
}

func (m *StatusMessage) FromHSlice(slice *storage.HSlice) error {
	timestamp, err := slice.GetUint32("timestamp")
	if err == nil {
		m.Timestamp = timestamp
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	time, err := slice.GetInt64("time")
	if err == nil {
		m.Time = time
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	ip, err := slice.GetString("ip")
	if err == nil {
		m.Ip = strings.Split(ip, ",")
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	platform, err := slice.GetString("platform")
	if err == nil {
		m.Platform = platform
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	contactEmail, err := slice.GetString("contact_email")
	if err == nil {
		m.ContactEmail = contactEmail
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	description, err := slice.GetString("description")
	if err == nil {
		m.Description = description
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	gpsTime, err := slice.GetInt64("gps_time")
	if err == nil {
		if m.Gps == nil {
			m.Gps = &GPSMetadata{}
		}
		m.Gps.Time = gpsTime
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	latitude, err := slice.GetFloat32("latitude")
	if err == nil {
		if m.Gps == nil {
			m.Gps = &GPSMetadata{}
		}
		m.Gps.Latitude = latitude
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	longitude, err := slice.GetFloat32("longitude")
	if err == nil {
		if m.Gps == nil {
			m.Gps = &GPSMetadata{}
		}
		m.Gps.Longitude = longitude
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	altitude, err := slice.GetInt32("altitude")
	if err == nil {
		if m.Gps == nil {
			m.Gps = &GPSMetadata{}
		}
		m.Gps.Altitude = altitude
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	rtt, err := slice.GetUint32("rtt")
	if err == nil {
		m.Rtt = rtt
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	rxIn, err := slice.GetUint32("rx_in")
	if err == nil {
		m.RxIn = rxIn
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	rxOk, err := slice.GetUint32("rx_ok")
	if err == nil {
		m.RxOk = rxOk
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	txIn, err := slice.GetUint32("tx_in")
	if err == nil {
		m.TxIn = txIn
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	txOk, err := slice.GetUint32("tx_ok")
	if err == nil {
		m.TxOk = txOk
	} else if err != storage.ErrDoesNotExist {
		return err
	}
	return nil
}

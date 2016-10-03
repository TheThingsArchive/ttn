package gateway

import "github.com/TheThingsNetwork/ttn/utils/scramble"

func (m *GPSMetadata) Scramble() (err error) {
	if m.Latitude, err = scramble.Float32(m.Latitude, 10+.01*m.Latitude); err != nil {
		return err
	}
	if m.Longitude, err = scramble.Float32(m.Longitude, 10+.01*m.Longitude); err != nil {
		return err
	}
	if m.Altitude, err = scramble.Int32Delta64(m.Altitude, int64(10+m.Altitude/100)); err != nil {
		return err
	}
	return nil
}

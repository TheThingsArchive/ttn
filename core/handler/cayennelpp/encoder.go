// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cayennelpp

import protocol "github.com/TheThingsNetwork/go-cayenne-lib/cayennelpp"

type Encoder struct {
}

func (e *Encoder) Encode(fields map[string]interface{}, fPort uint8) ([]byte, error) {
	encoder := protocol.NewEncoder()
	for name, value := range fields {
		key, channel, err := parseName(name)
		if err != nil {
			continue
		}
		switch key {
		case digitalInputKey:
			if val, ok := value.(int); ok {
				encoder.AddDigitalInput(channel, uint8(val))
			}
		case digitalOutputKey:
			if val, ok := value.(int); ok {
				encoder.AddDigitalOutput(channel, uint8(val))
			}
		case analogInputKey:
			if val, ok := value.(float64); ok {
				encoder.AddAnalogInput(channel, float32(val))
			}
		case analogOutputKey:
			if val, ok := value.(float64); ok {
				encoder.AddAnalogOutput(channel, float32(val))
			}
		case luminosityKey:
			if val, ok := value.(int); ok {
				encoder.AddLuminosity(channel, uint16(val))
			}
		case presenceKey:
			if val, ok := value.(int); ok {
				encoder.AddPresence(channel, uint8(val))
			}
		case temperatureKey:
			if val, ok := value.(float64); ok {
				encoder.AddTemperature(channel, float32(val))
			}
		case relativeHumidityKey:
			if val, ok := value.(float64); ok {
				encoder.AddRelativeHumidity(channel, float32(val))
			}
		case accelerometerKey:
			if val, ok := value.(map[string]float64); ok {
				valX := val["x"]
				valY := val["y"]
				valZ := val["z"]
				encoder.AddAccelerometer(channel, float32(valX), float32(valY), float32(valZ))
			}
		case barometricPressureKey:
			if val, ok := value.(float64); ok {
				encoder.AddBarometricPressure(channel, float32(val))
			}
		case gyrometerKey:
			if val, ok := value.(map[string]float64); ok {
				valX := val["x"]
				valY := val["y"]
				valZ := val["z"]
				encoder.AddGyrometer(channel, float32(valX), float32(valY), float32(valZ))
			}
		case gpsKey:
			if val, ok := value.(map[string]float64); ok {
				valLatitude := val["latitude"]
				valLongitude := val["longitude"]
				valAltitude := val["altitude"]
				encoder.AddGPS(channel, float32(valLatitude), float32(valLongitude), float32(valAltitude))
			}
		}
	}
	return encoder.Bytes(), nil
}

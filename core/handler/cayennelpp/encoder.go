// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cayennelpp

import (
	protocol "github.com/TheThingsNetwork/go-cayenne-lib/cayennelpp"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
)

// Encoder is a CayenneLPP PayloadEncoder
type Encoder struct {
}

// Encode encodes the fields to CayenneLPP
func (e *Encoder) Encode(fields map[string]interface{}, fPort uint8) ([]byte, bool, error) {
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
			} else if val, ok := value.(float64); ok {
				encoder.AddDigitalInput(channel, uint8(val))
			}
		case digitalOutputKey:
			if val, ok := value.(int); ok {
				encoder.AddDigitalOutput(channel, uint8(val))
			} else if val, ok := value.(float64); ok {
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
			} else if val, ok := value.(float64); ok {
				encoder.AddLuminosity(channel, uint16(val))
			}
		case presenceKey:
			if val, ok := value.(int); ok {
				encoder.AddPresence(channel, uint8(val))
			} else if val, ok := value.(float64); ok {
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
			var valX, valY, valZ float64
			if val, ok := value.(map[string]float64); ok {
				valX = val["x"]
				valY = val["y"]
				valZ = val["z"]
			} else if val, ok := value.(map[string]interface{}); ok {
				valX, _ = val["x"].(float64)
				valY, _ = val["y"].(float64)
				valZ, _ = val["z"].(float64)
			}
			encoder.AddAccelerometer(channel, float32(valX), float32(valY), float32(valZ))
		case barometricPressureKey:
			if val, ok := value.(float64); ok {
				encoder.AddBarometricPressure(channel, float32(val))
			}
		case gyrometerKey:
			var valX, valY, valZ float64
			if val, ok := value.(map[string]float64); ok {
				valX = val["x"]
				valY = val["y"]
				valZ = val["z"]
			} else if val, ok := value.(map[string]interface{}); ok {
				valX, _ = val["x"].(float64)
				valY, _ = val["y"].(float64)
				valZ, _ = val["z"].(float64)
			}
			encoder.AddGyrometer(channel, float32(valX), float32(valY), float32(valZ))
		case gpsKey:
			var lat, lon, alt float64
			if val, ok := value.(map[string]float64); ok {
				lat = val["latitude"]
				lon = val["longitude"]
				alt = val["altitude"]
			} else if val, ok := value.(map[string]interface{}); ok {
				lat, _ = val["latitude"].(float64)
				lon, _ = val["longitude"].(float64)
				alt, _ = val["altitude"].(float64)
			}
			encoder.AddGPS(channel, float32(lat), float32(lon), float32(alt))
		}
	}
	return encoder.Bytes(), true, nil
}

// Log returns the log
func (e *Encoder) Log() []*pb_handler.LogEntry {
	return nil
}

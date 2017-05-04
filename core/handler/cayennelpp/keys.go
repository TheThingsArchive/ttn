// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cayennelpp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	valueKey              = "value"
	digitalInputKey       = "digital_in"
	digitalOutputKey      = "digital_out"
	analogInputKey        = "analog_in"
	analogOutputKey       = "analog_out"
	luminosityKey         = "luminosity"
	presenceKey           = "presence"
	temperatureKey        = "temperature"
	relativeHumidityKey   = "relative_humidity"
	accelerometerKey      = "accelerometer"
	barometricPressureKey = "barometric_pressure"
	gyrometerKey          = "gyrometer"
	gpsKey                = "gps"
)

func formatName(key string, channel uint8) string {
	return fmt.Sprintf("%s_%d", key, channel)
}

func parseName(name string) (string, uint8, error) {
	parts := strings.Split(name, "_")
	if len(parts) < 2 {
		return "", 0, errors.New("Invalid name")
	}
	key := strings.Join(parts[:len(parts)-1], "_")
	if key == "" {
		return "", 0, errors.New("Invalid key")
	}
	channel, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return "", 0, err
	}
	if channel < 0 || channel > 255 {
		return "", 0, errors.New("Invalid range")
	}
	return key, uint8(channel), nil
}

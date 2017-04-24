// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cayennelpp

import (
	"bytes"

	protocol "github.com/TheThingsNetwork/go-cayenne-lib/cayennelpp"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
)

// Decoder is a CayenneLPP PayloadDecoder
type Decoder struct {
	result map[string]interface{}
}

// Decodes decodes the CayenneLPP payload to fields
func (d *Decoder) Decode(payload []byte, fPort uint8) (map[string]interface{}, bool, error) {
	decoder := protocol.NewDecoder(bytes.NewBuffer(payload))
	d.result = make(map[string]interface{})
	if err := decoder.DecodeUplink(d); err != nil {
		return nil, false, err
	}
	return d.result, true, nil
}

// Log returns the log
func (d *Decoder) Log() []*pb_handler.LogEntry {
	return nil
}

func (d *Decoder) DigitalInput(channel, value uint8) {
	d.result[formatName(digitalInputKey, channel)] = value
}

func (d *Decoder) DigitalOutput(channel, value uint8) {
	d.result[formatName(digitalOutputKey, channel)] = value
}

func (d *Decoder) AnalogInput(channel uint8, value float32) {
	d.result[formatName(analogInputKey, channel)] = value
}

func (d *Decoder) AnalogOutput(channel uint8, value float32) {
	d.result[formatName(analogOutputKey, channel)] = value
}

func (d *Decoder) Luminosity(channel uint8, value uint16) {
	d.result[formatName(luminosityKey, channel)] = value
}

func (d *Decoder) Presence(channel, value uint8) {
	d.result[formatName(presenceKey, channel)] = value
}

func (d *Decoder) Temperature(channel uint8, celcius float32) {
	d.result[formatName(temperatureKey, channel)] = celcius
}

func (d *Decoder) RelativeHumidity(channel uint8, rh float32) {
	d.result[formatName(relativeHumidityKey, channel)] = rh
}

func (d *Decoder) Accelerometer(channel uint8, x, y, z float32) {
	d.result[formatName(accelerometerKey, channel)] = map[string]float32{
		"x": x,
		"y": y,
		"z": z,
	}
}

func (d *Decoder) BarometricPressure(channel uint8, hpa float32) {
	d.result[formatName(barometricPressureKey, channel)] = hpa
}

func (d *Decoder) Gyrometer(channel uint8, x, y, z float32) {
	d.result[formatName(gyrometerKey, channel)] = map[string]float32{
		"x": x,
		"y": y,
		"z": z,
	}
}

func (d *Decoder) GPS(channel uint8, latitude, longitude, altitude float32) {
	d.result[formatName(gpsKey, channel)] = map[string]float32{
		"latitude":  latitude,
		"longitude": longitude,
		"altitude":  altitude,
	}
}

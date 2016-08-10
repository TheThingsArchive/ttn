// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package application

import (
	"fmt"
	"time"
)

// Application contains the state of an application
type Application struct {
	AppID string
	// Decoder is a JavaScript function that accepts the payload as byte array and
	// returns an object containing the decoded values
	Decoder string
	// Converter is a JavaScript function that accepts the data as decoded by
	// Decoder and returns an object containing the converted values
	Converter string
	// Validator is a JavaScript function that validates the data is converted by
	// Converter and returns a boolean value indicating the validity of the data
	Validator string
	// Encoder is a JavaScript function that encode the data send on Downlink messages
	Encoder string

	UpdatedAt time.Time
}

// ApplicationProperties contains all properties of a Application that can be stored in Redis.
var ApplicationProperties = []string{
	"app_id",
	"decoder",
	"converter",
	"validator",
	"encoder",
	"updated_at",
}

// ToStringStringMap converts the given properties of an Application to a
// map[string]string for storage in Redis.
func (application *Application) ToStringStringMap(properties ...string) (map[string]string, error) {
	output := make(map[string]string)
	for _, p := range properties {
		property, err := application.formatProperty(p)
		if err != nil {
			return output, err
		}
		output[p] = property
	}
	return output, nil
}

// FromStringStringMap imports known values from the input to an Application.
func (application *Application) FromStringStringMap(input map[string]string) error {
	for k, v := range input {
		application.parseProperty(k, v)
	}
	return nil
}

func (application *Application) formatProperty(property string) (formatted string, err error) {
	switch property {
	case "app_id":
		formatted = application.AppID
	case "decoder":
		formatted = application.Decoder
	case "converter":
		formatted = application.Converter
	case "validator":
		formatted = application.Validator
	case "encoder":
		formatted = application.Encoder
	case "updated_at":
		if !application.UpdatedAt.IsZero() {
			formatted = application.UpdatedAt.UTC().Format(time.RFC3339Nano)
		}
	default:
		err = fmt.Errorf("Property %s does not exist in Application", property)
	}
	return
}

func (application *Application) parseProperty(property string, value string) error {
	if value == "" {
		return nil
	}
	switch property {
	case "app_id":
		application.AppID = value
	case "decoder":
		application.Decoder = value
	case "converter":
		application.Converter = value
	case "validator":
		application.Validator = value
	case "encoder":
		application.Encoder = value
	case "updated_at":
		val, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return err
		}
		application.UpdatedAt = val
	}
	return nil
}

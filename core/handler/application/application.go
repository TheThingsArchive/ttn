package application

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// Application contains the state of an application
type Application struct {
	AppEUI        types.AppEUI
	DefaultAppKey types.AppKey
	// Decoder is a JavaScript function that accepts the payload as byte array and
	// returns an object containing the decoded values
	Decoder string
	// Converter is a JavaScript function that accepts the data as decoded by
	// Decoder and returns an object containing the converted values
	Converter string
	// Validator is a JavaScript function that validates the data is converted by
	// Converter and returns a boolean value indicating the validity of the data
	Validator string
}

// ApplicationProperties contains all properties of a Application that can be stored in Redis.
var ApplicationProperties = []string{
	"app_eui",
	"default_app_key",
	"decoder",
	"converter",
	"validator",
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
		if property != "" {
			output[p] = property
		}
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
	case "app_eui":
		formatted = application.AppEUI.String()
	case "default_app_key":
		formatted = application.DefaultAppKey.String()
	case "decoder":
		formatted = application.Decoder
	case "converter":
		formatted = application.Converter
	case "validator":
		formatted = application.Validator
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
	case "app_eui":
		val, err := types.ParseAppEUI(value)
		if err != nil {
			return err
		}
		if !val.IsEmpty() {
			application.AppEUI = val
		}
	case "default_app_key":
		val, err := types.ParseAppKey(value)
		if err != nil {
			return err
		}
		if !val.IsEmpty() {
			application.DefaultAppKey = val
		}
	case "decoder":
		application.Decoder = value
	case "converter":
		application.Converter = value
	case "validator":
		application.Validator = value
	}
	return nil
}

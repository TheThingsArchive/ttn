package application

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// Application contains the state of an application
type Application struct {
	AppEUI            types.AppEUI
	HandlerID         string
	HandlerNetAddress string
}

// ApplicationProperties contains all properties of an Application that can be stored in Redis.
var ApplicationProperties = []string{
	"app_eui",
	"handler_id",
	"handler_net_address",
}

// ToStringStringMap converts the given properties of Application to a
// map[string]string for storage in Redis.
func (app *Application) ToStringStringMap(properties ...string) (map[string]string, error) {
	output := make(map[string]string)
	for _, p := range properties {
		property, err := app.formatProperty(p)
		if err != nil {
			return output, err
		}
		if property != "" {
			output[p] = property
		}
	}
	return output, nil
}

// FromStringStringMap imports known values from the input to a Application.
func (app *Application) FromStringStringMap(input map[string]string) error {
	for k, v := range input {
		app.parseProperty(k, v)
	}
	return nil
}

func (app *Application) formatProperty(property string) (formatted string, err error) {
	switch property {
	case "app_eui":
		formatted = app.AppEUI.String()
	case "handler_id":
		formatted = app.HandlerID
	case "handler_net_address":
		formatted = app.HandlerNetAddress
	default:
		err = fmt.Errorf("Property %s does not exist in Application", property)
	}
	return
}

func (app *Application) parseProperty(property string, value string) error {
	if value == "" {
		return nil
	}
	switch property {
	case "app_eui":
		val, err := types.ParseAppEUI(value)
		if err != nil {
			return err
		}
		app.AppEUI = val
	case "handler_id":
		app.HandlerID = value
	case "handler_net_address":
		app.HandlerNetAddress = value
	}
	return nil
}

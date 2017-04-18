package models

import (
	"fmt"
  "github.com/spf13/viper"
	"net/http"
	"encoding/json"
	"bytes"
	"errors"
	"strings"
	"github.com/TheThingsNetwork/ttn/api/auth"
	"net/url"
)


// The application type
type Application struct {
	EUI        EUI         `json:"eui"`
	Name       string      `json:"name"`
	Owner      string      `json:"owner"`
	Valid      bool        `json:"valid"`
	AccessKeys []accessKey `json:"accessKeys"`
}

// Marshalling Applications to JSON
type app_ Application
func (app Application) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
			app_
			Type  string `json:"@type"`
			Self  string `json:"@self"`
		}{
			app_:  app_(app),
			Type: "application",
			Self: fmt.Sprintf("/api/v1/applications/%X", app.EUI),
		})
}

// list all applications for a specific user
func ListApplications(accessToken auth.Token) ([]Application, error) {
	authServer := viper.GetString("ttn-account-server")
	url        := fmt.Sprintf("%s/applications", authServer)

	req, err := auth.NewAuthRequest(accessToken, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)
	var apps []Application
	if err := decoder.Decode(&apps); err != nil {
		return nil, err
	}

	return apps, nil
}

// get specific application
func GetApplication(accessToken auth.Token, eui EUI) (app Application, err error) {
	apps, err := ListApplications(accessToken)
	if err != nil {
		return app, err
	}

	for _, app := range apps {
		if bytes.Equal(app.EUI, eui) {
			return app, nil
		}
	}
	return app, errors.New("Application with that eui does not exist")
}


func CreateApplication(accessToken auth.Token, eui EUI, name string) (app Application, err error) {

	authServer := viper.GetString("ttn-account-server")
	uri        := fmt.Sprintf("%s/applications", authServer)

	values := url.Values{
		"name": {name},
		"eui":  {WriteEUI(eui)},
	}

	req, err := auth.NewAuthRequest(accessToken, "POST", uri, strings.NewReader(values.Encode()))
	if err != nil {
		return app, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	if err != nil {
		return app, errors.New("Failed to create application")
	}

	if resp.StatusCode != http.StatusCreated {
		// TODO: check why (eg. already exists / ...)
		return app, errors.New("Failed to create application")
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&app); err != nil {
		return app, err
	}

	// yay! nothing went wrong
	return app, nil
}


// Authorizes user to application
// TODO: should I check wether the user exists?
func AuthorizeUserToApplication(accessToken auth.Token, eui EUI, email string) error {

	authServer := viper.GetString("ttn-account-server")
	uri        := fmt.Sprintf("%s/applications/%s/authorize", authServer, WriteEUI(eui))

	values := url.Values{
		"email": {email},
	}

	req, err := auth.NewAuthRequest(accessToken, "PUT", uri, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	if err != nil {
		return errors.New("Failed to authorize user")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Failed to authorize user")
	}

	return nil
}

// Delete an application
// TODO: this doesn't work on the auth-server yet
func DeleteApplication(accessToken auth.Token, eui EUI) error {
	authServer := viper.GetString("ttn-account-server")
	uri        := fmt.Sprintf("%s/applications/%s", authServer, WriteEUI(eui))

	req, err := auth.NewAuthRequest(accessToken, "DELETE", uri, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	if err != nil {
		return errors.New("Failed to delete application")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Failed to delete application")
	}

	return nil
}


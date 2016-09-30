// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/apex/log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func getConfigLocation() (string, error) {
	cFile := viper.ConfigFileUsed()
	if cFile == "" {
		dir, err := homedir.Dir()
		if err != nil {
			return "", fmt.Errorf("Could not get homedir: %s", err.Error())
		}
		expanded, err := homedir.Expand(dir)
		if err != nil {
			return "", fmt.Errorf("Could not get homedir: %s", err.Error())
		}
		cFile = path.Join(expanded, ".ttnctl.yaml")
	}
	return cFile, nil
}

// ReadConfig reads the config file
func ReadConfig() (map[string]interface{}, error) {
	cFile, err := getConfigLocation()
	if err != nil {
		return nil, err
	}

	c := make(map[string]interface{})

	// Read config file
	bytes, err := ioutil.ReadFile(cFile)
	if err == nil {
		err = yaml.Unmarshal(bytes, &c)
	}
	if err != nil {
		return nil, fmt.Errorf("Could not read configuration file: %s", err.Error())
	}

	return c, nil
}

// WriteConfigFile writes the config file
func WriteConfigFile(data map[string]interface{}) error {
	cFile, err := getConfigLocation()
	if err != nil {
		return err
	}

	// Write config file
	d, err := yaml.Marshal(&data)
	if err != nil {
		return fmt.Errorf("Could not generate configiguration file contents: %s", err.Error())
	}
	err = ioutil.WriteFile(cFile, d, 0644)
	if err != nil {
		return fmt.Errorf("Could not write configiguration file: %s", err.Error())
	}

	return nil
}

// SetConfig sets the specified fields in the config file.
func SetConfig(data map[string]interface{}) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}
	for key, value := range data {
		config[key] = value
	}
	return WriteConfigFile(config)
}

// GetAppEUI returns the AppEUI that must be set in the command options or config
func GetAppEUI(ctx log.Interface) types.AppEUI {
	appEUIString := viper.GetString("app-eui")
	if appEUIString == "" {
		ctx.Fatal("Missing AppEUI. You should select an application to use with \"ttnctl applications select\"")
	}
	eui, err := types.ParseAppEUI(appEUIString)
	if err != nil {
		ctx.WithError(err).Fatal("Invalid AppEUI")
	}
	return eui
}

// GetAppID returns the AppID that must be set in the command options or config
func GetAppID(ctx log.Interface) string {
	appID := viper.GetString("app-id")
	if appID == "" {
		ctx.Fatal("Missing AppID. You should select an application to use with \"ttnctl applications select\"")
	}
	return appID
}

// GetConfigFile returns the location of the configuration file.
// It checks the following (in this order):
// the --config flag
// $XDG_CONFIG_HOME/ttnctl/config.yml (if $XDG_CONFIG_HOME is set)
// $HOME/.ttnctl.yml
func GetConfigFile() string {
	file := viper.GetString("config")
	if file != "" {
		return file
	}

	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg != "" {
		return path.Join(xdg, "ttnctl", "config.yml")
	}

	return path.Join(os.Getenv("HOME"), ".ttnctl.yml")
}

// GetDataDir returns the location of the data directory used for
// sotring data.
// It checks the following (in this order):
// the --data flag
// $XDG_DATA_HOME/ttnctl (if $XDG_DATA_HOME is set)
// $XDG_CACHE_HOME/ttnctl (if $XDG_CACHE_HOME is set)
// $HOME/.ttnctl
func GetDataDir() string {
	file := viper.GetString("data")
	if file != "" {
		return file
	}

	xdg := os.Getenv("XDG_DATA_HOME")
	if xdg != "" {
		return path.Join(xdg, "ttnctl")
	}

	xdg = os.Getenv("XDG_CACHE_HOME")
	if xdg != "" {
		return path.Join(xdg, "ttnctl")
	}

	return path.Join(os.Getenv("HOME"), ".ttnctl")
}

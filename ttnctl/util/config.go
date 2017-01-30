// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/spf13/viper"
)

const (
	appFilename = "app"
	euiKey      = "eui"
	idKey       = "id"
)

// GetConfigFile returns the location of the configuration file.
// It checks the following (in this order):
// the --config flag
// $XDG_CONFIG_HOME/ttnctl/config.yml (if $XDG_CONFIG_HOME is set)
// $HOME/.ttnctl.yml
func GetConfigFile() string {
	flag := viper.GetString("config")

	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg != "" {
		xdg = path.Join(xdg, "ttnctl", "config.yml")
	}

	home := os.Getenv("HOME")
	homeyml := ""
	homeyaml := ""

	if home != "" {
		homeyml = path.Join(home, ".ttnctl.yml")
		homeyaml = path.Join(home, ".ttnctl.yaml")
	}

	try_files := []string{
		flag,
		xdg,
		homeyml,
		homeyaml,
	}

	// find a file that exists, and use that
	for _, file := range try_files {
		if file != "" {
			if _, err := os.Stat(file); err == nil {
				return file
			}
		}
	}

	// no file found, set up correct fallback
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		return xdg
	} else {
		return homeyml
	}
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

func readData(file string) map[string]interface{} {
	fullpath := path.Join(GetDataDir(), file)

	c := make(map[string]interface{})

	// Read config file
	data, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return c
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return c
	}

	return c
}

func writeData(file string, data map[string]interface{}) error {
	fullpath := path.Join(GetDataDir(), file)

	// Generate yaml contents
	d, err := yaml.Marshal(&data)
	if err != nil {
		return fmt.Errorf("Could not generate configiguration file contents: %s", err.Error())
	}

	// Write to file
	err = ioutil.WriteFile(fullpath, d, 0644)
	if err != nil {
		return fmt.Errorf("Could not write configuration file: %s", err.Error())
	}

	return nil
}

func setData(file, key string, data interface{}) error {
	config := readData(file)
	config[key] = data
	return writeData(file, config)
}

// GetAppEUI returns the AppEUI that must be set in the command options or config
func GetAppEUI(ctx ttnlog.Interface) types.AppEUI {
	appEUIString := viper.GetString("app-eui")
	if appEUIString == "" {
		appData := readData(appFilename)
		eui, ok := appData[euiKey].(string)
		if !ok {
			ctx.Fatal("Invalid AppEUI in config file")
		}
		appEUIString = eui
	}

	if appEUIString == "" {
		ctx.Fatal("Missing AppEUI. You should select an application to use with \"ttnctl applications select\"")
	}

	eui, err := types.ParseAppEUI(appEUIString)
	if err != nil {
		ctx.WithError(err).Fatal("Invalid AppEUI")
	}
	return eui
}

// SetApp stores the app EUI preference
func SetAppEUI(ctx ttnlog.Interface, appEUI types.AppEUI) {
	err := setData(appFilename, euiKey, appEUI.String())
	if err != nil {
		ctx.WithError(err).Fatal("Could not save app EUI")
	}
}

// GetAppID returns the AppID that must be set in the command options or config
func GetAppID(ctx ttnlog.Interface) string {
	appID := viper.GetString("app-id")
	if appID == "" {
		appData := readData(appFilename)
		id, ok := appData[idKey].(string)
		if !ok {
			ctx.Fatal("Invalid appID in config file.")
		}
		appID = id
	}

	if appID == "" {
		ctx.Fatal("Missing AppID. You should select an application to use with \"ttnctl applications select\"")
	}
	return appID
}

// SetApp stores the app ID and app EUI preferences
func SetApp(ctx ttnlog.Interface, appID string, appEUI types.AppEUI) {
	config := readData(appFilename)
	config[idKey] = appID
	config[euiKey] = appEUI.String()
	err := writeData(appFilename, config)
	if err != nil {
		ctx.WithError(err).Fatal("Could not save app preference")
	}
}

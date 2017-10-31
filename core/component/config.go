// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"github.com/spf13/viper"
)

// Config is the configuration for this component
type Config struct {
	AuthServers map[string]string
	KeyDir      string
	UseTLS      bool
}

// ConfigFromViper imports configuration from Viper
func ConfigFromViper() Config {
	return Config{
		AuthServers: viper.GetStringMapString("auth-servers"),
		KeyDir:      viper.GetString("key-dir"),
		UseTLS:      viper.GetBool("tls"),
	}
}

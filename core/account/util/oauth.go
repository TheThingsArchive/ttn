// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"fmt"

	"golang.org/x/oauth2"
)

// MakeConfig creates an oauth.Config object based on the necessary parameters,
// redirectURL can be left empty if not needed
func MakeConfig(server string, clientID string, clientSecret string, redirectURL string) oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: fmt.Sprintf("%s/users/token", server),
		AuthURL:  fmt.Sprintf("%s/users/authorize", server),
	}

	return oauth2.Config{
		ClientID:     "ttnctl",
		ClientSecret: "",
		Endpoint:     endpoint,
		RedirectURL:  redirectURL,
	}
}

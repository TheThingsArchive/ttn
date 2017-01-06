// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"errors"
	"strings"

	"github.com/TheThingsNetwork/ttn/api"
	"google.golang.org/grpc"
)

// Dial dials the component represented by this Announcement
func (a *Announcement) Dial() (*grpc.ClientConn, error) {
	if a.NetAddress == "" {
		return nil, errors.New("Can not dial this component")
	}
	if a.Certificate == "" {
		return api.Dial(strings.Split(a.NetAddress, ",")[0])
	}
	return api.DialWithCert(strings.Split(a.NetAddress, ",")[0], a.Certificate)
}

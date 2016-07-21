package discovery

import (
	"strings"

	"github.com/TheThingsNetwork/ttn/api"
	"google.golang.org/grpc"
)

// Dial dials the component represented by this Announcement
func (a *Announcement) Dial() (*grpc.ClientConn, error) {
	return api.DialWithCert(strings.Split(a.NetAddress, ",")[0], a.Certificate)
}

package broker

import (
	"errors"
	"strings"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
)

// ByScore is used to sort a list of DownlinkOptions based on Score
type ByScore []*pb.DownlinkOption

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].Score < a[j].Score }

func (b *broker) HandleDownlink(downlink *pb.DownlinkMessage) error {
	var err error
	downlink, err = b.ns.Downlink(b.Component.GetContext(), downlink)
	if err != nil {
		return err
	}

	var routerID string
	if id := strings.Split(downlink.DownlinkOption.Identifier, ":"); len(id) == 2 {
		routerID = id[0]
	} else {
		return errors.New("ttn/broker: Invalid downlink option")
	}

	router, err := b.getRouter(routerID)
	if err != nil {
		return err
	}

	router <- downlink

	return nil
}

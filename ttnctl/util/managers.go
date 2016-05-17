package util

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/components/handler"
	"github.com/apex/log"
	"github.com/spf13/viper"
)

func GetHandlerManager(ctx log.Interface) core.AuthHandlerClient {
	cli, err := handler.NewClient(viper.GetString("ttn-handler"))
	if err != nil {
		ctx.Fatalf("Could not connect: %v", err)
	}
	return cli
}

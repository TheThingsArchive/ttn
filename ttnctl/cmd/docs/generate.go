package main

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/cmd"
	"github.com/TheThingsNetwork/ttn/utils/docs"
)

func main() {
	fmt.Println(`# API Reference

Control The Things Network from the command line.
`)
	fmt.Print(docs.Generate(cmd.RootCmd))
}

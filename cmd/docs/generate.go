package main

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/cmd"
	"github.com/TheThingsNetwork/ttn/utils/docs"
)

func main() {
	fmt.Println(`# API Reference

The Things Network's backend servers.
`)
	fmt.Print(docs.Generate(cmd.RootCmd))
}

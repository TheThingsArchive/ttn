package parse

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/robertkrimen/otto/parser"
)

// Port returns the port from an address
func Port(address string) (uint, error) {
	parts := strings.Split(address, ":")

	length := len(parts)
	if length < 2 {
		return 0, errors.New("Could not parse the port: malformated address")
	}

	port, err := strconv.Atoi(parts[length-1])
	if err != nil {
		return 0, err
	}
	if port < 0 {
		return 0, errors.New("Invalid port number")
	}

	return uint(port), nil
}

// PayloadFunction parses the given js code an verify that the syntax is valid
func PayloadFunction(code string) error {
	_, err := parser.ParseFunction("", code)
	if err != nil {
		return fmt.Errorf("Syntax Error: %s", err.Error())
	}
	return nil
}

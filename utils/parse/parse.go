package parse

import (
	"errors"
	"strconv"
	"strings"
)

// Port returns the port from an address
func Port(address string) (uint, error) {
	parts := strings.SplitN(address, ":", 2)
	if len(parts) != 2 {
		return 0, errors.New("Could not parse the port: malformated address")
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	if port < 0 {
		return 0, errors.New("Invalid port number")
	}

	return uint(port), nil
}

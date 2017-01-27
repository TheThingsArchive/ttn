package util

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/TheThingsNetwork/ttn/core/types"
)

var stdin = bufio.NewReader(os.Stdin)
var ErrInvalidPayload = errors.New("Invalid payload (Should be hexadecimal)")

func IsErrOutOfRange(err error) bool {
	if ok := strings.HasSuffix(err.Error(), "value out of range"); !ok {
		return false
	}
	return true
}

func ReadPort() (uint8, error) {
	fmt.Print("Port: ")
	input, err := ReadInput(stdin)
	if err != nil {
		return 0, err
	}

	port, err := parsePort(input)
	if err != nil {
		return 0, err
	}

	return port, nil
}

func parsePort(input string) (uint8, error) {
	port, err := strconv.ParseUint(input, 10, 8)
	if err != nil {
		if IsErrOutOfRange(err) {
			return 0, fmt.Errorf(fmt.Sprintf("The port number is too big (Should be uint8): %s", err.Error()))
		}
		return 0, err
	}

	return uint8(port), nil
}

func ReadFields() (map[string]interface{}, error) {
	fmt.Print("Fields: ")
	input, err := ReadInput(stdin)
	if err != nil {
		return nil, err
	}

	parsedFields, err := parseFields(input)
	if err != nil {
		return nil, err
	}

	return parsedFields, nil
}

func parseFields(input string) (map[string]interface{}, error) {
	parsedFields := make(map[string]interface{})
	err := json.Unmarshal([]byte(input), &parsedFields)
	if err != nil {
		return nil, err
	}

	return parsedFields, nil
}

func ReadPayload() ([]byte, error) {
	fmt.Print("Payload: ")
	input, err := ReadInput(stdin)
	if err != nil {
		return nil, err
	}

	payload, err := parsePayload(input)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func parsePayload(input string) ([]byte, error) {
	if len(input)%2 != 0 {
		return nil, ErrInvalidPayload
	}

	payload, err := types.ParseHEX(input, len(input)/2)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// ReadInput allows to read an input line from the user
func ReadInput(reader *bufio.Reader) (string, error) {
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(input, "\n"), nil
}

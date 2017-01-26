package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var stdin = bufio.NewReader(os.Stdin)

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

	port, err := strconv.ParseUint(input, 10, 8)
	if err != nil {
		if IsErrOutOfRange(err) {
			return 0, fmt.Errorf(fmt.Sprintf("The port number is too big (Should be int8): %s", err.Error()))
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

	parsedFields := make(map[string]interface{})
	err = json.Unmarshal([]byte(input), &parsedFields)
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

	payload := []byte(input)

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

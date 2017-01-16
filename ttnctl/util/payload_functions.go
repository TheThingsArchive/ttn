package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/TheThingsNetwork/ttn/utils/parse"
	"github.com/apex/log"
	"github.com/robertkrimen/otto"
)

var (
	ErrInvalidDecoderReturn       = errors.New("The decoder should return an object of fields")
	ErrInvalidValidatorReturn     = errors.New("The validator should return a boolean")
	ErrInvalidEncoderReturn       = errors.New("The encoder should return an array of a buffer of bytes")
	ErrInvalidConverterReturn     = errors.New("The converter should return an object")
	ErrInvalidPayloadFunctionType = errors.New("This type of payload function is not supported")
	ErrUndefinedReturn            = errors.New("The function returned an undefined value")
)

func ValidatePayload(ctx log.Interface, code, payloadType string) (string, error) {
	// We first parse the function to return if the AST cannnot be built (syntax error)
	ctx.Info("Parsing function...")
	err := parse.PayloadFunction(code)
	if err != nil {
		return "", err
	}
	ctx.Info("Function parsed successfully: syntax checked")

	fmt.Printf("\nDo you want to test the function to detect runtime errors ?\n")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		return code, nil
	}

	fmt.Println("Write your function call and provide arguments to test it")
	fmt.Println("Note: Use the built-in function JSON.stringify() to provide json objects as parameters: E.g: JSON.stingify({ valid: argument })")

	switch payloadType {
	case "decoder":
		fmt.Println(`
function Decoder(bytes, port) {
  // Instructions
}

// The function call to test the Decoder
Decoder([10, 23, 35], 3)`)
	case "converter":
		fmt.Println(`
function Converter(decoded, port) {
  // Instructions
}

// The function call to test the Converter
Converter(JSON.stringify({ foo: bar }), 3)`)
	case "validator":
		fmt.Println(`
function Validator(converted, port) {
  // Instructions
}

// The function call to test the Validator
Validator(JSON.stringify({ foo: bar }), 3)`)
	case "encoder":
		fmt.Println(`
function Encoder(object, port) {
  // Instructions
}

// The function call to test the Encoder
Encoder(JSON.stringify({ foo: bar }), 3)`)
	default:
		return "", ErrInvalidPayloadFunctionType
	}

	fmt.Println("########## Write your testing function call here and end with Ctrl+D (EOF):")

	testSignature := ReadFunction(ctx)

	// If no syntax error, we run the function
	ctx.Info("Testing...")
	err = testPayload(ctx, code, testSignature, payloadType)
	if err != nil {
		return "", err
	}

	ctx.Info("The test is successful, the given function is valid")
	return code, nil
}

func testPayload(ctx log.Interface, code, testSignature, payloadType string) error {
	// testEnvironment puts the test entry after the function declaration
	// to be able to run the function with the provided values. That way we don't have to do data conversion
	// function definition and test call are placed in a js environment and executed directly.
	testEnvironment := fmt.Sprintf(`
    %s

    %s
    `, code, testSignature)

	vm := otto.New()
	value, err := vm.Run(testEnvironment)
	if err != nil {
		return err
	}

	if value.IsDefined() {
		switch payloadType {
		case "decoder":
			if !value.IsObject() {
				return ErrInvalidDecoderReturn
			}
		case "converter":
			if !value.IsObject() {
				return ErrInvalidConverterReturn
			}
		case "validator":
			if !value.IsBoolean() {
				return ErrInvalidValidatorReturn
			}
		case "encoder":
			if value.Class() != "Array" {
				return ErrInvalidEncoderReturn
			}
		default:
			return ErrInvalidPayloadFunctionType
		}
		return nil
	}

	return ErrUndefinedReturn
}

func ReadFunction(ctx log.Interface) string {
	content, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		ctx.WithError(err).Fatal("Could not read function from STDIN.")
	}
	return strings.TrimSpace(string(content))
}

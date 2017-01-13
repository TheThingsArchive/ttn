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

	fmt.Printf("\n Test the function to detect runtime errors \n")
	fmt.Println("Provide the signature of the payload function with test values")
	fmt.Println(`
Note:
1) Use single quotes for strings: E.g: 'this is a valid string'
2) Use the built-in function JSON.stringify() to provide json objects parameters: E.g: JSON.stingify({ valid: argument })
3) The provided signature should match the previous function declaration: E.g: MyFunc('entry', 123) will allow us to test the function called MyFunc() and which takes 2 arguments.

########## Write your testing entry here and end with Ctrl+D (EOF):`)

	testSignature := ReadFunction(ctx)

	// If no syntax error, we run the function
	ctx.Info("Testing...")
	err = testPayload(ctx, code, testSignature, payloadType)
	if err != nil {
		return "", err
	}

	ctx.Info("The test is successful, the given function is a valid payload")
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

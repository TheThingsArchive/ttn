package util

import (
	"os"
	"testing"

	cliHandler "github.com/TheThingsNetwork/go-utils/handlers/cli"
	"github.com/apex/log"
	. "github.com/smartystreets/assertions"
)

func GetLogger(t *testing.T, tag string) log.Interface {
	logger := &log.Logger{
		Handler: cliHandler.New(os.Stdout),
	}
	return logger.WithField("tag", tag)
}

func TestTestPayload(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestPayload")

	// Test Decoder
	{
		// Valid syntax but wrong return type
		function := `
			function Decoder(bytes, port) {
				return 1;
			}
		`

		testSignature := `
			Decoder(12, 'test')
		`

		err := testPayload(ctx, function, testSignature, "decoder")
		a.So(err, ShouldResemble, ErrInvalidDecoderReturn)

		// Unknown reference
		function = `
			function Decoder(bytes, port) {
				var decoded = {};
				return test; // Undefined reference 'test'
			}
		`

		testSignature = `
			Decoder(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "decoder")
		a.So(err, ShouldNotBeNil)

		// Syntax error
		function = `
			function Decoder(bytes, port) {
				va decoded = {}; // syntax error
				return test; // Undefined reference 'test'
			}
		`

		testSignature = `
			Decoder(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "decoder")
		a.So(err, ShouldNotBeNil)

		// Test signature doesn't fit any previous function declaration
		function = `
			function Decoder(bytes, port) {
				var decoded = {};
				return decoded;
			}
		`

		testSignature = `
			Decode(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "decoder")
		a.So(err, ShouldNotBeNil)

		// Valid test and function
		function = `
			function Decoder(bytes, port) {
				var decoded = { "name": "John" }
				return decoded;
			}
		`

		testSignature = `
			Decoder(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "decoder")
		a.So(err, ShouldBeNil)
	}

	// Test Converter
	{
		// Valid syntax but wrong return type
		function := `
			function Converter(bytes, port) {
				return 1;
			}
		`

		testSignature := `
			Converter(12, 'test')
		`

		err := testPayload(ctx, function, testSignature, "converter")
		a.So(err, ShouldResemble, ErrInvalidConverterReturn)

		// Undefined reference
		function = `
			function Converter(bytes, port) {
				var decoded = {};
				var converted = { "name": "John" };
				return test; // Undefined reference 'test'
			}
		`

		testSignature = `
			Converter(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "converter")
		a.So(err, ShouldNotBeNil)

		// Syntax error
		function = `
			function Converter(bytes, port) {
				va decoded = {}; // syntax error
				var converted = { "name": "John" };
				return converted;
			}
		`

		testSignature = `
			Decoder(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "converter")
		a.So(err, ShouldNotBeNil)

		// Test signature doesn't fit any previous function declaration
		function = `
			function Converter(bytes, port) {
				var decoded = {};
				var converted = { "name": "John" };
				return converted;
			}
		`

		testSignature = `
			Convert(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "converter")
		a.So(err, ShouldNotBeNil)

		// Valid test and function
		function = `
			function Converter(bytes, port) {
				var decoded = {};
				var converted = { "name": "John" };
				return converted;
			}
		`

		testSignature = `
			Converter(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "converter")
		a.So(err, ShouldBeNil)
	}

	// Test Encoder
	{
		// Valid syntax but wrong return type
		function := `
			function Encoder(bytes, port) {
				return 1;
			}
		`

		testSignature := `
			Encoder(12, 'test')
		`

		err := testPayload(ctx, function, testSignature, "encoder")
		a.So(err, ShouldResemble, ErrInvalidEncoderReturn)

		// Undefined reference
		function = `
			function Encoder(bytes, port) {
				var encoded = [ 10, 12, 30 ];
				return test; // Undefined reference 'test'
			}
		`

		testSignature = `
			Encoder(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "encoder")
		a.So(err, ShouldNotBeNil)

		// Syntax error
		function = `
			function Encoder(bytes, port) {
				va encoder = [ 10, 12, 30]; // syntax error
				return encoder;
			}
		`

		testSignature = `
			Encoder(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "encoder")
		a.So(err, ShouldNotBeNil)

		// Test signature doesn't fit any previous function declaration
		function = `
			function Converter(bytes, port) {
				var encoded = [ 10, 12, 30 ];
				return encoded;
			}
		`

		testSignature = `
			Encode(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "encoder")
		a.So(err, ShouldNotBeNil)

		// Valid test and function
		function = `
			function Encoder(bytes, port) {
				var encoded = [ 10, 12, 30 ];
				return encoded;
			}
		`

		testSignature = `
			Encoder(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "encoder")
		a.So(err, ShouldBeNil)
	}

	// Test Validator
	{
		// Valid syntax but wrong return type
		function := `
			function Validator(bytes, port) {
				return 12;
			}
		`

		testSignature := `
			Validator(12, 'test')
		`

		err := testPayload(ctx, function, testSignature, "validator")
		a.So(err, ShouldResemble, ErrInvalidValidatorReturn)

		// Undefined reference
		function = `
			function Validator(bytes, port) {
				var validated = false;
				return test; // Undefined reference 'test'
			}
		`

		testSignature = `
			validator(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "validator")
		a.So(err, ShouldNotBeNil)

		// Syntax error
		function = `
			function Validator(bytes, port) {
				va validated = false; // syntax error
				return validated;
			}
		`

		testSignature = `
			Validator(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "validator")
		a.So(err, ShouldNotBeNil)

		// Test signature doesn't fit any previous function declaration
		function = `
			function Validator(bytes, port) {
				var validated = true;
				return validated;
			}
		`

		testSignature = `
			Validate(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "validator")
		a.So(err, ShouldNotBeNil)

		// Valid test and function
		function = `
			function Validator(bytes, port) {
				var validated = true;
				return validated;
			}
		`

		testSignature = `
			Validator(12, 'test')
		`

		err = testPayload(ctx, function, testSignature, "validator")
		a.So(err, ShouldBeNil)
	}
}

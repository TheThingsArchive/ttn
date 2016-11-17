// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func debug(msgs ...string) {
	if logtostderr {
		s := strings.Join(msgs, " ")
		fmt.Fprintln(os.Stderr, "protoc-gen-ttndoc: ", s)
	}
}

func failWithError(err error, msgs ...string) {
	s := strings.Join(msgs, " ") + ":" + err.Error()
	log.Print("protoc-gen-ttndoc: error:", s)
	os.Exit(1)
}

func fail(msgs ...string) {
	s := strings.Join(msgs, " ")
	log.Print("protoc-gen-ttndoc: fail:", s)
	os.Exit(1)
}

var logtostderr bool

// Read from standard input
func main() {
	g := generator.New()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		failWithError(err, "reading input")
	}

	if err := proto.Unmarshal(data, g.Request); err != nil {
		failWithError(err, "parsing input proto")
	}

	if len(g.Request.FileToGenerate) == 0 {
		fail("no files to generate")
	}

	g.CommandLineParameters(g.Request.GetParameter())

	if val, ok := g.Param["logtostderr"]; ok && val == "true" {
		logtostderr = true
	}

	tree := buildTree(g.Request.GetProtoFile())

	packageLocations := make(map[string]string)
	for _, file := range g.Request.GetProtoFile() {
		loc := file.GetName()[:strings.LastIndex(file.GetName(), "/")]
		packageLocations[file.GetPackage()] = loc
	}

	selectedServices := make(map[string]string)
	for k, v := range g.Param {
		switch {
		case k == "logtostderr":
			if v == "true" {
				logtostderr = true
			}
		case strings.HasPrefix(k, "."):
			selectedServices[k] = v
		}
	}

	for serviceKey := range selectedServices {
		service, ok := tree.services[serviceKey]
		if !ok {
			fail("Service", serviceKey, "unknown")
		}

		usedMessages := make(map[string]*message)
		usedEnums := make(map[string]*enum)

		content := new(bytes.Buffer)

		fmt.Fprintf(content, "# %s API Reference\n\n", service.GetName())
		if service.comment != "" {
			fmt.Fprintf(content, "%s\n\n", service.comment)
		}

		if len(service.methods) > 0 {
			fmt.Fprint(content, "## Methods\n\n")

			for _, method := range service.methods {
				fmt.Fprintf(content, "### `%s`\n\n", method.GetName())
				if method.comment != "" {
					fmt.Fprintf(content, "%s\n\n", method.comment)
				}
				if method.inputStream {
					fmt.Fprintf(content, "- Client stream of [`%s`](#%s)\n", method.input.GetName(), heading(method.input.key))
				} else {
					fmt.Fprintf(content, "- Request: [`%s`](#%s)\n", method.input.GetName(), heading(method.input.key))
				}
				useMessage(tree, method.input, usedMessages, usedEnums)
				if method.outputStream {
					fmt.Fprintf(content, "- Server stream of [`%s`](#%s)\n", method.output.GetName(), heading(method.input.key))
				} else {
					fmt.Fprintf(content, "- Response: [`%s`](#%s)\n", method.output.GetName(), heading(method.input.key))
				}
				useMessage(tree, method.output, usedMessages, usedEnums)

				fmt.Fprintln(content)

				if len(method.endpoints) != 0 {
					if len(method.endpoints) == 1 {
						fmt.Fprint(content, "### HTTP Endpoint\n\n")
					} else {
						fmt.Fprint(content, "### HTTP Endpoints\n\n")
					}
					for _, endpoint := range method.endpoints {
						fmt.Fprintf(content, "- `%s` `%s`\n", endpoint.method, endpoint.url)
					}
					fmt.Fprintln(content)
				}
			}
		}

		if len(usedMessages) > 0 {
			fmt.Fprint(content, "## Messages\n\n")

			var messageKeys []string
			for key := range usedMessages {
				messageKeys = append(messageKeys, key)
			}
			sort.Strings(messageKeys)

			for _, messageKey := range messageKeys {
				message := usedMessages[messageKey]
				fmt.Fprintf(content, "### `%s`\n\n", message.key)
				if strings.HasPrefix(messageKey, ".google") {
					fmt.Fprintf(content, "%s\n\n", strings.SplitAfter(message.comment, ".")[0])
				} else if message.comment != "" {
					fmt.Fprintf(content, "%s\n\n", message.comment)
				}
				if len(message.fields) > 0 {
					fmt.Fprintln(content, "| Field Name | Type | Description |")
					fmt.Fprintln(content, "| ---------- | ---- | ----------- |")
					for idx, field := range message.fields {
						if field.isOneOf {
							if idx == 0 || !message.fields[idx-1].isOneOf || message.fields[idx-1].GetOneofIndex() != field.GetOneofIndex() {
								oneof := message.GetOneof(field.GetOneofIndex())
								if len(oneof.fields) > 1 {
									fmt.Fprintf(content, "| **%s** | **oneof %d** | one of the following %d |\n", oneof.GetName(), len(oneof.fields), len(oneof.fields))
								}
							}
						} else {
							if idx > 0 && message.fields[idx-1].isOneOf {
								oneof := message.GetOneof(message.fields[idx-1].GetOneofIndex())
								if len(oneof.fields) > 1 {
									fmt.Fprintf(content, "| **%s** | **end oneof %d** |  |\n", oneof.GetName(), len(oneof.fields))
								}
							}
						}

						var fieldType string
						if field.repeated {
							fieldType += "_repeated_ "
						}
						typ := strings.ToLower(strings.TrimPrefix(field.GetType().String(), "TYPE_"))
						switch typ {
						case "message":
							friendlyType := field.GetTypeName()[strings.LastIndex(field.GetTypeName(), ".")+1:]
							fieldType += fmt.Sprintf("[`%s`](#%s)", friendlyType, heading(field.GetTypeName()))
						case "enum":
							// TODO(htdvisser): test this
							if enum, ok := tree.enums[field.GetTypeName()]; ok {
								usedEnums[field.GetTypeName()] = enum
							}
							friendlyType := field.GetTypeName()[strings.LastIndex(field.GetTypeName(), ".")+1:]
							fieldType += fmt.Sprintf("[`%s`](#%s)", friendlyType, heading(field.GetTypeName()))
						default:
							fieldType += fmt.Sprintf("`%s`", typ)
						}
						fmt.Fprintf(content, "| `%s` | %s | %s |\n", field.GetName(), fieldType, strings.Replace(field.comment, "\n", " ", -1))
					}
					fmt.Fprintln(content)
				}
			}
		}

		if len(usedEnums) > 0 {
			fmt.Fprint(content, "## Used Enums\n\n")

			var enumKeys []string
			for key := range usedEnums {
				enumKeys = append(enumKeys, key)
			}
			sort.Strings(enumKeys)

			for _, enumKey := range enumKeys {
				enum := usedEnums[enumKey]

				fmt.Fprintf(content, "### `%s`\n\n", enum.key)
				if enum.comment != "" {
					fmt.Fprintf(content, "%s\n\n", enum.comment)
				}

				if len(enum.values) > 0 {
					fmt.Fprintln(content, "| Value | Description |")
					fmt.Fprintln(content, "| ----- | ----------- |")
					for _, value := range enum.values {
						fmt.Fprintf(content, "| `%s` | %s |\n", value.GetName(), strings.Replace(value.comment, "\n", " ", -1))
					}
					fmt.Fprintln(content)
				}
			}
		}

		packageService := strings.TrimPrefix(service.key, ".")
		packageName := packageService[:strings.Index(packageService, ".")]
		location, ok := packageLocations[packageName]
		if !ok {
			fail("Could not find location of package", packageName)
		}
		fileName := path.Join(location, service.GetName()+".md")
		contentString := content.String()
		g.Response.File = append(g.Response.File, &plugin.CodeGeneratorResponse_File{
			Name:    &fileName,
			Content: &contentString,
		})
	}

	// Send back the results.
	data, err = proto.Marshal(g.Response)
	if err != nil {
		failWithError(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		failWithError(err, "failed to write output proto")
	}
}

func stringPtr(str string) *string {
	return &str
}

func heading(str string) string {
	return strings.ToLower(strings.NewReplacer(".", "").Replace(str))
}

func useMessage(tree *tree, msg *message, messages map[string]*message, enums map[string]*enum) {
	messages[msg.key] = msg
	for _, msg := range msg.nested {
		useMessage(tree, msg, messages, enums)
	}
	for _, field := range msg.fields {
		if msg, ok := tree.messages[field.GetTypeName()]; ok {
			useMessage(tree, msg, messages, enums)
		}
		if enum, ok := tree.enums[field.GetTypeName()]; ok {
			enums[enum.key] = enum
		}
	}
}

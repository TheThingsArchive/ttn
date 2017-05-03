// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

var tmpl = template.New("ttndoc")

var docTemplate = `{{ if .Packages }}# API Reference

{{ range .Services }}{{ template "service" . }}
{{ end }}
{{ if .Messages }}## Messages

{{ range .Messages }}{{ template "message" . }}
{{ end }}{{ end }}
{{ if .Enums }}## Enums

{{ range .Enums }}{{ template "enum" . }}
{{ end }}{{ end }}{{ end }}
`

var serviceTemplate = `## {{ .Name }}

{{ if .Comment }}{{ .Comment }}
{{ end }}
{{ if .Methods }}### Methods

{{ range .Methods }}{{ template "method" . }}
{{ end }}{{ end }}`

var methodTemplate = `#### {{ .Name | TrimPrefix | Code }}

{{ if .Comment }}{{ .Comment }}
{{ end }}
- Request: {{ if .Input.Stream }}stream of {{ end }}{{ .Input.Message.Name | Anchor }}
- Response: {{ if .Output.Stream }}stream of {{ end }}{{ .Output.Message.Name | Anchor }}

{{ if .HTTPEndpoints }}Available HTTP Endpoints:

{{ range .HTTPEndpoints }}- {{ .RequestMethod | Code }} {{ .RequestPath | Code }}
{{ end }}

Input:

{{ .Input.Message.Default | JSON | Codeblock "json" }}

Output:

{{ .Output.Message.Default | JSON | Codeblock "json" }}
{{ end }}
`

var messageTemplate = `### {{ .Name | TrimPrefix | Code }}

{{ if .Comment }}{{ .Comment }}
{{ end }}
{{ if .Fields }}| **Name** | **Type** | **Description** |
| -------- | -------- | --------------- |
{{ range .Fields }}{{ template "field" . }}
{{ end }}{{ end }}
`

var fieldTemplate = `| {{ .Name | TrimPrefix | Code }} | {{ if .Repeated }}_repeated_ {{ end }}{{ .Type | Anchor }} | {{ .Comment | WithoutNewlines }} |`

var enumTemplate = `### {{ .Name | TrimPrefix | Code }}

{{ if .Comment }}{{ .Comment }}
{{ end }}
{{ if .Values }}
| **Name** | **Description** |
| -------- | --------------- |
{{ range .Values }}{{ template "enum-value" . }}
{{ end }}{{ end }}`

var enumValueTemplate = `| {{ .Name | TrimPrefix | Code }} | {{ .Comment | WithoutNewlines }} |`

func trimPrefix(in string) string    { return in[strings.LastIndex(in, ".")+1:] }
func wrapCode(in string) string      { return fmt.Sprintf("`%s`", in) }
func replaceAnchor(in string) string { return strings.ToLower(strings.NewReplacer(".", "").Replace(in)) }

func init() {
	newline := regexp.MustCompile(`\n+`)
	tmpl.Funcs(template.FuncMap{
		"TrimPrefix": trimPrefix,
		"Code":       wrapCode,
		"Codeblock":  func(format, data string) string { return fmt.Sprintf("```%s\n%s\n```", format, data) },
		"Anchor": func(in string) string {
			if strings.HasPrefix(in, ".") {
				return fmt.Sprintf("[`%s`](#%s)", trimPrefix(in), replaceAnchor(in))
			}
			return wrapCode(in)
		},
		"WithoutNewlines": func(in string) string { return newline.ReplaceAllString(in, " ") },
		"JSON": func(in interface{}) string {
			out, _ := json.MarshalIndent(in, "", "  ")
			return string(out)
		},
	})
	template.Must(tmpl.Parse(docTemplate))
	template.Must(tmpl.New("service").Parse(serviceTemplate))
	template.Must(tmpl.New("method").Parse(methodTemplate))
	template.Must(tmpl.New("message").Parse(messageTemplate))
	template.Must(tmpl.New("field").Parse(fieldTemplate))
	template.Must(tmpl.New("enum").Parse(enumTemplate))
	template.Must(tmpl.New("enum-value").Parse(enumValueTemplate))
}

func (d *TTNDoc) Render() (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fatih/structs"
)

// StringStringMapEncoder encodes the given properties of the input to a map[string]string for storage in Redis
type StringStringMapEncoder func(input interface{}, properties ...string) (map[string]string, error)

type isZeroer interface {
	IsZero() bool
}

type isEmptier interface {
	IsEmpty() bool
}

func buildDefaultStructEncoder(tagName string) StringStringMapEncoder {
	if tagName == "" {
		tagName = defaultTagName
	}

	return func(input interface{}, properties ...string) (map[string]string, error) {
		vmap := make(map[string]string)
		s := structs.New(input)
		s.TagName = tagName
		if len(properties) == 0 {
			properties = s.Names()
		}
		for _, field := range s.Fields() {
			if !field.IsExported() {
				continue
			}

			if !stringInSlice(field.Name(), properties) {
				continue
			}

			tagName, opts := parseTag(field.Tag(tagName))
			if tagName == "" || tagName == "-" {
				continue
			}

			val := field.Value()

			if opts.Has("omitempty") {
				if field.IsZero() {
					continue
				}
				if z, ok := val.(isZeroer); ok && z.IsZero() {
					continue
				}
				if z, ok := val.(isEmptier); ok && z.IsEmpty() {
					continue
				}
			}

			if v, ok := val.(string); ok {
				vmap[tagName] = v
				continue
			}

			if !field.IsZero() {
				if m, ok := val.(encoding.TextMarshaler); ok {
					txt, err := m.MarshalText()
					if err != nil {
						return nil, err
					}
					vmap[tagName] = string(txt)
					continue
				}
				if m, ok := val.(json.Marshaler); ok {
					txt, err := m.MarshalJSON()
					if err != nil {
						return nil, err
					}
					vmap[tagName] = string(txt)
					continue
				}
			}

			if field.Kind() == reflect.String {
				vmap[tagName] = fmt.Sprint(val)
				continue
			}

			if txt, err := json.Marshal(val); err == nil {
				vmap[tagName] = string(txt)
				continue
			}

			vmap[tagName] = fmt.Sprintf("%v", val)
		}
		return vmap, nil
	}
}

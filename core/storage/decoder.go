// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/fatih/structs"
)

// StringStringMapDecoder is used to decode a map[string]string to a struct
type StringStringMapDecoder func(input map[string]string) (interface{}, error)

func decodeToType(typ reflect.Kind, value string) interface{} {
	switch typ {
	case reflect.String:
		return value
	case reflect.Bool:
		v, _ := strconv.ParseBool(value)
		return v
	case reflect.Int:
		v, _ := strconv.ParseInt(value, 10, 64)
		return int(v)
	case reflect.Int8:
		return int8(decodeToType(reflect.Int, value).(int))
	case reflect.Int16:
		return int16(decodeToType(reflect.Int, value).(int))
	case reflect.Int32:
		return int32(decodeToType(reflect.Int, value).(int))
	case reflect.Int64:
		return int64(decodeToType(reflect.Int, value).(int))
	case reflect.Uint:
		v, _ := strconv.ParseUint(value, 10, 64)
		return uint(v)
	case reflect.Uint8:
		return uint8(decodeToType(reflect.Uint, value).(uint))
	case reflect.Uint16:
		return uint16(decodeToType(reflect.Uint, value).(uint))
	case reflect.Uint32:
		return uint32(decodeToType(reflect.Uint, value).(uint))
	case reflect.Uint64:
		return uint64(decodeToType(reflect.Uint, value).(uint))
	case reflect.Float64:
		v, _ := strconv.ParseFloat(value, 64)
		return v
	case reflect.Float32:
		return float32(decodeToType(reflect.Float64, value).(float64))
	}
	return nil
}

func unmarshalToType(typ reflect.Type, value string) (val interface{}, err error) {
	// If we get a pointer in, we'll return a pointer out
	if typ.Kind() == reflect.Ptr {
		val = reflect.New(typ.Elem()).Interface()
	} else {
		val = reflect.New(typ).Interface()
	}
	defer func() {
		if err == nil && typ.Kind() != reflect.Ptr {
			val = reflect.Indirect(reflect.ValueOf(val)).Interface()
		}
	}()

	// If we can just assign the value, return the value
	if typ.AssignableTo(reflect.TypeOf(value)) {
		return value, nil
	}

	// Try Unmarshalers
	if um, ok := val.(encoding.TextUnmarshaler); ok {
		if err = um.UnmarshalText([]byte(value)); err == nil {
			return val, nil
		}
	}
	if um, ok := val.(json.Unmarshaler); ok {
		if err = um.UnmarshalJSON([]byte(value)); err == nil {
			return val, nil
		}
	}

	// Try conversion
	if typ.ConvertibleTo(reflect.TypeOf(value)) {
		return reflect.ValueOf(value).Convert(typ).Interface(), nil
	}

	// Try JSON
	if err = json.Unmarshal([]byte(value), val); err == nil {
		return val, nil
	}

	// Return error if we have one
	if err != nil {
		return nil, err
	}

	return val, fmt.Errorf("No way to unmarshal \"%s\" to %s", value, typ.Name())
}

// buildDefaultStructDecoder is used by the RedisMapStore
func buildDefaultStructDecoder(base interface{}) StringStringMapDecoder {
	return func(input map[string]string) (output interface{}, err error) {
		baseType := reflect.TypeOf(base)
		// If we get a pointer in, we'll return a pointer out
		if baseType.Kind() == reflect.Ptr {
			output = reflect.New(baseType.Elem()).Interface()
		} else {
			output = reflect.New(baseType).Interface()
		}
		defer func() {
			if err == nil && baseType.Kind() != reflect.Ptr {
				output = reflect.Indirect(reflect.ValueOf(output)).Interface()
			}
		}()

		s := structs.New(output)
		for _, field := range s.Fields() {
			if !field.IsExported() {
				continue
			}

			tagName, _ := parseTag(field.Tag(tagName))
			if tagName == "" || tagName == "-" {
				continue
			}
			if str, ok := input[tagName]; ok {
				baseField, _ := baseType.FieldByName(field.Name())

				var val interface{}
				switch field.Kind() {
				case reflect.Struct, reflect.Array, reflect.Interface, reflect.Slice, reflect.Ptr:
					var err error
					val, err = unmarshalToType(baseField.Type, str)
					if err != nil {
						return nil, err
					}
				default:
					val = decodeToType(field.Kind(), str)
				}

				if val == nil {
					continue
				}

				if !baseField.Type.AssignableTo(reflect.TypeOf(val)) && baseField.Type.ConvertibleTo(reflect.TypeOf(val)) {
					val = reflect.ValueOf(val).Convert(baseField.Type).Interface()
				}

				if err := field.Set(val); err != nil {
					return nil, err
				}
			}
		}
		return output, nil
	}
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package pointer provides helper method to quickly define pointer from basic go types
package pointer

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// String creates a pointer to a string from a string value
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

// Int creates a pointer to an int from an int value
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// Uint creates a pointer to an unsigned int from an unsigned int value
func Uint(v uint) *uint {
	p := new(uint)
	*p = v
	return p
}

// Float64 creates a pointer to a float64 from a float64 value
func Float64(v float64) *float64 {
	p := new(float64)
	*p = v
	return p
}

// Bool creates a pointer to a boolean from a boolean value
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

// Time creates a pointer to a time.Time from a time.Time value
func Time(v time.Time) *time.Time {
	p := new(time.Time)
	*p = v
	return p
}

// DumpStruct prints the content of a struct of pointers
func DumpPStruct(s interface{}, multiline bool) string {
	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Struct {
		return "Not a struct"
	}

	nl := ","
	str := fmt.Sprintf("%s{", v.Type().Name())
	if multiline {
		nl = "\n\t"
		str += nl
	}

	for k := 0; k < v.NumField(); k += 1 {
		name := v.Type().Field(k).Name
		if name[0] == strings.ToLower(name)[0] { // Unexported field
			continue
		}
		i := v.Field(k).Interface()
		key := fmt.Sprintf("%v", v.Type().Field(k).Name)

		switch t := i.(type) {
		case *bool:
			if t != nil {
				str += fmt.Sprintf("%s:%+v%s", key, *t, nl)
			}
		case *int:
			if t != nil {
				str += fmt.Sprintf("%s:%+v%s", key, *t, nl)
			}
		case *uint:
			if t != nil {
				str += fmt.Sprintf("%s:%+v%s", key, *t, nl)
			}
		case *string:
			if t != nil {
				str += fmt.Sprintf("%s:%+v%s", key, *t, nl)
			}
		case *float64:
			if t != nil {
				str += fmt.Sprintf("%s:%+v%s", key, *t, nl)
			}
		case *time.Time:
			if t != nil {
				str += fmt.Sprintf("%s:%+v%s", key, *t, nl)
			}
		default:
			str += fmt.Sprintf("%s:unknown%s", key, nl)
		}
	}

	if multiline {
		str += "\n}"
	} else {
		str += "}"
	}
	return str
}

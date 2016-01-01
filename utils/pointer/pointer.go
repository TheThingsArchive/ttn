// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package pointer provides helper method to quickly define pointer from basic go types
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
func DumpPStruct(s interface{}) {
	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Struct {
		fmt.Printf("Unable to dump: Not a struct.")
		return
	}

	for k := 0; k < v.NumField(); k += 1 {
		name := v.Type().Field(k).Name
		if name[0] == strings.ToLower(name)[0] { // Unexported field
			continue
		}
		i := v.Field(k).Interface()
		fmt.Printf("%v: ", v.Type().Field(k).Name)

		switch t := i.(type) {
		case *bool:
			if t == nil {
				fmt.Printf("nil\n")
			} else {
				fmt.Printf("%+v\n", *t)
			}
		case *int:
			if t == nil {
				fmt.Printf("nil\n")
			} else {
				fmt.Printf("%+v\n", *t)
			}
		case *uint:
			if t == nil {
				fmt.Printf("nil\n")
			} else {
				fmt.Printf("%+v\n", *t)
			}
		case *string:
			if t == nil {
				fmt.Printf("nil\n")
			} else {
				fmt.Printf("%+v\n", *t)
			}
		case *float64:
			if t == nil {
				fmt.Printf("nil\n")
			} else {
				fmt.Printf("%+v\n", *t)
			}
		case *time.Time:
			if t == nil {
				fmt.Printf("nil\n")
			} else {
				fmt.Printf("%+v\n", *t)
			}
		default:
			fmt.Printf("unknown\n")
		}
	}
}

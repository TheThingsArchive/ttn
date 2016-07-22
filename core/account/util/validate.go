// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"reflect"

	"github.com/asaskevich/govalidator"
)

func validateSlice(val interface{}) error {
	s := reflect.ValueOf(val)

	for i := 0; i < s.Len(); i++ {
		if err := Validate(s.Index(i).Interface()); err != nil {
			return fmt.Errorf("In slice at index %v: %s", i, err)
		}
	}
	return nil
}

// Validate recursivly validates most structures using govalidator
// struct tags. It currently works for slices, structs and pointers.
func Validate(val interface{}) error {
	switch reflect.TypeOf(val).Kind() {
	case reflect.Slice:
		// validate a slice
		err := validateSlice(val)
		return err
	case reflect.Ptr:
		// try to get the valu from the ptr
		v := reflect.ValueOf(val).Elem()
		if v.CanInterface() {
			return Validate(v.Interface())
		}
	case reflect.Struct:
		// when it is a struct jsut validate
		_, err := govalidator.ValidateStruct(val)
		return err
	}
	return nil
}

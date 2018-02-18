// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package reflector

import (
	"reflect"

	"github.com/dihedron/go-log/log"
)

func StructFromPointer(obj interface{}) interface{} {
	// extract the struct from the pointer
	rv := reflect.ValueOf(obj)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		log.Debugf("%v -> %v", rv.Kind(), rv.Type())
		rv = rv.Elem()
	}
	log.Debugf("%v -> %v", rv.Kind(), rv.Type())
	obj = rv.Interface()
	return obj
}

func GetFields(obj interface{}) []reflect.Value {
	fields := make([]reflect.Value, 0)

	obj = StructFromPointer(obj)

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	for i := 0; i < t.NumField(); i++ {
		value := v.Field(i)

		switch value.Kind() {
		case reflect.Struct:
			fields = append(fields, GetFields(value.Interface())...)
		default:
			fields = append(fields, value)
		}
	}

	return fields
}

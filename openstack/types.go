// Copyright 2017 Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package openstack

import (
	"fmt"
)

// String returns a string pointer referencing a copy of the given value.
func String(value string) *string {
	ptr := new(string)
	*ptr = value
	return ptr
}

// Int returns an int pointer referencing a copy of the given value.
func Int(value int) *int {
	ptr := new(int)
	*ptr = value
	return ptr
}

// Bool returns a bool pointer referencing a copy of the given value.
func Bool(value bool) *bool {
	ptr := new(bool)
	*ptr = value
	return ptr

}

// Float32 returns a a float pointer referencing a copy of the given value.
func Float32(value float32) *float32 {
	ptr := new(float32)
	*ptr = value
	return ptr

}

// Float64 returns a float pointer referencing a copy of the given value.
func Float64(value float64) *float64 {
	ptr := new(float64)
	*ptr = value
	return ptr
}

// Stringf allocates a string formatted according to the given parameters
// and returns a reference to it
func Stringf(format string, args ...interface{}) *string {
	value := fmt.Sprintf(format, args...)
	return &value
}

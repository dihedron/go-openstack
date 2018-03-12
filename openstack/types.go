// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"fmt"
	"time"
)

// Bool returns a pointer to a boolean value that is safe
// to be used in OpenStack API structs.
func Bool(value bool) *bool {
	return &value
}

// String returns a string pointer that is safe to be
// used in OpenStack API structs.
func String(value string) *string {
	return &value
}

// Int returns a pointer to an int value that is safe
// to be used in OpenStack API structs.
func Int(value int) *int {
	return &value
}

// Int8 returns a pointer to an int8 value that is safe
// to be used in OpenStack API structs.
func Int8(value int8) *int8 {
	return &value
}

// Int16 returns a pointer to an int16 value that is safe
// to be used in OpenStack API structs.
func Int16(value int16) *int16 {
	return &value
}

// Int32 returns a pointer to an int32 value that is safe
// to be used in OpenStack API structs.
func Int32(value int32) *int32 {
	return &value
}

// Int64 returns a pointer to an int64 value that is safe
// to be used in OpenStack API structs.
func Int64(value int64) *int64 {
	return &value
}

// UInt returns a pointer to an uint value that is safe
// to be used in OpenStack API structs.
func UInt(value uint) *uint {
	return &value
}

// UInt8 returns a pointer to an uint8 value that is safe
// to be used in OpenStack API structs.
func UInt8(value uint8) *uint8 {
	return &value
}

// UInt16 returns a pointer to an uint16 value that is safe
// to be used in OpenStack API structs.
func UInt16(value uint16) *uint16 {
	return &value
}

// UInt32 returns a pointer to an uint32 value that is safe
// to be used in OpenStack API structs.
func UInt32(value uint32) *uint32 {
	return &value
}

// UInt64 returns a pointer to an uint64 value that is safe
// to be used in OpenStack API structs.
func UInt64(value uint64) *uint64 {
	return &value
}

// UIntPtr returns a pointer to an uintptr value that is safe
// to be used in OpenStack API structs.
func UIntPtr(value uintptr) *uintptr {
	return &value
}

// Byte returns a pointer to a byte value that is safe
// to be used in OpenStack API structs.
func Byte(value byte) *byte {
	return &value
}

// Rune returns a pointer to a rune value that is safe
// to be used in OpenStack API structs.
func Rune(value rune) *rune {
	return &value
}

// Float32 returns a pointer to a float32 value that is safe
// to be used in OpenStack API structs.
func Float32(value float32) *float32 {
	return &value
}

// Float64 returns a pointer to a float64 value that is safe
// to be used in OpenStack API structs.
func Float64(value float64) *float64 {
	return &value
}

// Complex64 returns a pointer to a complex64 value that is safe
// to be used in OpenStack API structs.
func Complex64(value complex64) *complex64 {
	return &value
}

// Complex128 returns a pointer to a complex128 value that is safe
// to be used in OpenStack API structs.
func Complex128(value complex128) *complex128 {
	return &value
}

// StringSlice returns a pointer to a []string value that is safe
// to be used in OpenStack API structs.
func StringSlice(value []string) *[]string {
	return &value
}

// ISO8601 is the format of OpenStack timestamps.
const ISO8601 string = "2006-01-02T15:04:05.000000Z"

// Operator is the type of operators used for comparisons.
type Operator int8

// String returns the operator as a string prepresentation; this is used in
// HTTP query parameters that represent time filters.
func (op Operator) String() string {
	switch op {
	case EQ:
		return "eq"
	case LT:
		return "lt"
	case LTE:
		return "lte"
	case GT:
		return "gt"
	case GTE:
		return "gte"
	case NE:
		return "ne"
	}
	return ""
}

const (
	// EQ is the constant used to indicate that some entity is "equal to" some
	// other reference or provided value.
	EQ Operator = iota
	// LT is the constant used to indicate that some entity is "less than" some
	// other reference or provided value.
	LT
	// LTE is the constant used to indicate that some entity is "less than or
	// equal to" some other reference or provided value.
	LTE
	// GTE is the constant used to indicate that some entity is "greater than or
	// equal to" some other reference or provided value.
	GTE
	// GT is the constant used to indicate that some entity is "greater than"
	// some other reference or provided value.
	GT
	// NE is the constant used to indicate that some entity is "not equal to"
	// some other reference or provided value.
	NE
)

// TimeFilter is used to provide time-based filters in API calls, e.g. retrieving
// only those users whose passwords expire after (GT) a certain date.
type TimeFilter struct {
	Timestamp time.Time
	Operator  Operator
}

// String returns a TimeFilter as an acceptable query parameter.
func (tf TimeFilter) String() string {
	return fmt.Sprintf("%v:%v", tf.Operator, tf.Timestamp.Format(ISO8601))
}

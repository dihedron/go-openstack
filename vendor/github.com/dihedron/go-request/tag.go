// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package request

import (
	"strings"
)

// Tag represents the value of a tag on a tagged struct field.
type Tag struct {
	tag string
}

// NewTag creates a new Tag struct from the given tag value.
func NewTag(tag string) Tag {
	return Tag{tag: tag}
}

// Name extracts the name of the tag, that is the first value in the list of
// comma-separated values that can be given in a tagged struct field.
func (t Tag) Name() string {
	tokens := strings.Split(t.tag, ",")
	if len(tokens) > 0 {
		return strings.TrimSpace(tokens[0])
	}
	return ""
}

// IsMissing returns whether the tag on the field is missing/invalid (i.e. empty).
func (t Tag) IsMissing() bool {
	return strings.TrimSpace(t.tag) == ""
}

// IsIgnore returns whether the tagged field struct contains the "-" (dash)
// character that is usually employed to say that the field should be ignored
// with respect to this specific tag handling (e.g. `json:"-"` means "do not
// marshal when writing to JSON).
func (t Tag) IsIgnore() bool {
	for _, token := range strings.Split(t.tag, ",") {
		if strings.TrimSpace(token) == "-" {
			return true
		}
	}
	return false
}

// IsOmitEmpty returns whether the tag should be ignored when it contains an empty
// value (e.g. a null pointer).
func (t Tag) IsOmitEmpty() bool {
	for _, token := range strings.Split(t.tag, ",") {
		if strings.TrimSpace(token) == "omitempty" {
			return true
		}
	}
	return false
}

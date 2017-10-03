// Copyright 2017 Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"bytes"
	"fmt"
)

// Error represents an OpenStack-specific type of error; it may
// contain debugging or context information.
type Error struct {
	Err  error
	Info map[string]interface{}
}

// Error returns a textual representation of the error.
func (e Error) Error() string {
	var buffer bytes.Buffer
	buffer.WriteString(e.Err.Error())
	if e.Info != nil {
		buffer.WriteString(" where {")
		first := true
		for k, v := range e.Info {
			if first {
				first = false
				buffer.WriteString(fmt.Sprintf(" %q => %v", k, v))
			} else {
				buffer.WriteString(fmt.Sprintf(", %q => %v", k, v))
			}
		}
		buffer.WriteString(" }")
	}
	return buffer.String()
}

const (
	// StatusCode represents the HTTP status code attribute in OpenStackErrors
	// resulting from API calls.
	StatusCode string = "Status-Code"
)

var (
	// Success represents a successful outcome of an API call.
	Success = Errorf("Success").Where(StatusCode, 200)

	// ErrorInvalidReference is used whenever an invalid object
	// reference is passed to an API.
	ErrorInvalidReference = Errorf("Invalid object reference")

	// ErrorInvalidInput is used whenever an invalid input value or
	// parameter is passed to an API.
	ErrorInvalidInput = Errorf("Invalid input value")

	// ErrorBadRequest is returned when the service endpoint failed to
	// parse the request as expected. One of the following errors occurred:
	// * a required attribute was missing;
	// * an attribute that is not allowed was specified, such as an ID on a
	//   POST request in a basic CRUD operation;
	// * an attribute of an unexpected data type was specified.
	ErrorBadRequest = Errorf("Bad request").Where(StatusCode, 400)

	// ErrorUnauthorized is returned when one of the following errors occurred:
	// * authentication was not performed;
	// * the specified X-Auth-Token header is not valid;
	// * the authentication credentials are not valid.
	ErrorUnauthorized = Errorf("Unauthorized").Where(StatusCode, 401)

	// ErrorForbidden is returned when the identity was successfully
	// authenticated but it is not authorized to perform the requested
	// action.
	ErrorForbidden = Errorf("Forbidden").Where(StatusCode, 403)

	// ErrorNotFound is returned when an operation failed because a
	// referenced entity cannot be found by ID. For a POST request,
	// the referenced entity might be specified in the request body
	// rather than in the resource path.
	ErrorNotFound = Errorf("Not Found").Where(StatusCode, 404)

	// ErrorConflict is returned when a POST or PATCH operation failed.
	// For example, a client tried to update a unique attribute for an entity,
	// which conflicts with that of another entity in the same collection.
	// Or, a client issued a create operation twice on a collection with a
	// user-defined, unique attribute. For example, a client made a POST /users
	// request two times for the unique, user-defined name attribute for a user
	// entity.
	ErrorConflict = Errorf("Conflict").Where(StatusCode, 409)
)

// Errorf returns an OpenStack Error object.
func Errorf(text string, args ...interface{}) Error {
	return Error{
		Err: fmt.Errorf(text, args...),
	}
}

// Where adds contextual information to the Error.
func (e Error) Where(key string, info interface{}) Error {
	ne := Error{
		Err: e.Err,
	}
	ne.Info = make(map[string]interface{})
	if e.Info != nil {
		for k, v := range e.Info {
			ne.Info[k] = v
		}
	}
	ne.Info[key] = info
	return ne
}

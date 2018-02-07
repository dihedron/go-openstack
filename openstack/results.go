// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"fmt"
	"net/http"
)

// Result represents in compact form the result of an HTTP API call.
type Result struct {
	Code        int
	Status      string
	Description string
	Data        []byte
}

// String returns a string representation of an HTTP result.
func (r Result) String() string {
	return fmt.Sprintf("%d - %s (%d bytes)", r.Code, r.Status, len(r.Data))
}

// NewResult maps the status code in an HTTP Response to the corresponding
// API result.
func NewResult(response *http.Response, data []byte) *Result {
	var r Result
	switch response.StatusCode {
	case http.StatusOK: // 200
		r = Success
	case http.StatusCreated: // 201
		r = Created
	// case http.StatusAccepted: // 202
	// case http.StatusNonAuthoritativeInfo: // 203
	case http.StatusNoContent: // 204
		r = NoContent
	// case http.ResetContent: // 205
	// case http.StatusPartialContent: // 206
	// case http.StatusMulticase: // 207
	// case http.StatusAlreadyReported : // 208
	// case http.StatusIMUsed: // 226
	case 400:
		r = BadRequest
	case 401:
		r = Unauthorized
	case 403:
		r = Forbidden
	case 404:
		r = NotFound
	case 405:
		r = MethodNotAllowed
	case 409:
		r = Conflict
	case 413:
		r = RequestEntityTooLarge
	case 415:
		r = UnsupportedMediaType
	case 503:
		r = ServiceUnavailable
	default:
		r = Result{
			Description: "Unknown error.",
		}
	}
	r.Data = data
	return &r
}

// IsInformational returns whether the HTTP status code represents an informational
// message (class 1xx).
func (r Result) IsInformational() bool {
	return r.Code >= 100 && r.Code < 200
}

// IsSuccess returns whether the HTTP status code represents a successful request
// (class 2xx).
func (r Result) IsSuccess() bool {
	return r.Code >= 200 && r.Code < 300
}

// IsRedirection returns whether the HTTP status code represents a redirect
// response (class 3xx).
func (r Result) IsRedirection() bool {
	return r.Code >= 300 && r.Code < 400
}

// IsClientError returns whether the HTTP status code represents a processing
// error induced by the client via the request (class 4xx).
func (r Result) IsClientError() bool {
	return r.Code >= 400 && r.Code < 500
}

// IsServerError returns whether the HTTP status code represents a server-side
// error (class 5xx).
func (r Result) IsServerError() bool {
	return r.Code >= 500 && r.Code < 600
}

// IsUnofficial returns whether the the HHTP status code is out of known ranges
// (calss 6xx or more).
func (r Result) IsUnofficial() bool {
	return r.Code >= 600
}

var (
	// Success means that the HTTP API request was successful.
	Success = Result{
		Code:        200,
		Status:      "OK",
		Description: "Request was successful.",
	}

	// Created means that the resource was created and is ready to use.
	Created = Result{
		Code:        201,
		Status:      "Created",
		Description: "Resource was created and is ready to use.",
	}

	// NoContent means that there is no data associated with the requested resource;
	// this is typical with HEAD requests.
	NoContent = Result{
		Code:        204,
		Status:      "No Content",
		Description: "There is no data associated with the requested resource.",
	}

	// BadRequest means that some content in the HTTP API request was invalid.
	BadRequest = Result{
		Code:        400,
		Status:      "Bad Request",
		Description: "Some content in the request was invalid.",
	}

	// Unauthorized means that the user must authenticate before making a request.
	Unauthorized = Result{
		Code:        401,
		Status:      "Unauthorized",
		Description: "User must authenticate before making a request.",
	}

	// Forbidden means that someolicy does not allow current user to do this operation.
	Forbidden = Result{
		Code:        403,
		Status:      "Forbidden",
		Description: "Policy does not allow current user to do this operation.",
	}

	// NotFound means that the requested resource could not be found.
	NotFound = Result{
		Code:        404,
		Status:      "Not Found",
		Description: "The requested resource could not be found.",
	}

	// MethodNotAllowed means that the API call method is not valid for this endpoint.
	MethodNotAllowed = Result{
		Code:        405,
		Status:      "Method Not Allowed",
		Description: "Method is not valid for this endpoint.",
	}

	// Conflict means that A POST or PATCH operation failed; for example, a client tried to update
	// a unique attribute for an entity, which conflicts with that of another entity in the same
	// collection. Or, a client issued a create operation twice on a collection with a user-defined,
	// unique attribute. For example, a client made a POST /users request two times for the unique,
	// user-defined name attribute for a user entity.
	Conflict = Result{
		Code:        409,
		Status:      "Conflict",
		Description: "A POST or PATCH operation failed.",
	}

	// RequestEntityTooLarge means that the request is larger than the server is willing
	// or able to process.
	RequestEntityTooLarge = Result{
		Code:        413,
		Status:      "Request Entity Too Large",
		Description: "The request is larger than the server is willing or able to process.",
	}

	// UnsupportedMediaType means that the request entity has a media type which the server or resource
	// does not support.
	UnsupportedMediaType = Result{
		Code:        415,
		Status:      "Unsupported Media Type",
		Description: "The request entity has a media type which the server or resource does not support.",
	}

	// ServiceUnavailable is a server-side error that is mostly caused by service configuration
	// errors which prevents the service from successful start up.
	ServiceUnavailable = Result{
		Code:        503,
		Status:      "Service Unavailable",
		Description: "Service is not available.",
	}
)

// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/dihedron/go-log"
	"github.com/dihedron/go-request"
)

// API is the archetype and base struct of all API services.
type API struct {

	// client is the base SDK client, that takes care of sending requests and
	// retrieving responses from the server; moreover it contains some context
	// information such as User-Agent and access to other services. This
	// reference is only for convenience, it makes the public API more user-
	// friendly.
	client *Client

	// builder is the generator for service-specific requests; it provides the
	// base URL and some common parameters and headers; APIs can derive their
	// own sub-builders by specifying a different and more specific path, plus
	// their own sets of headers and query parameters.
	builder *request.Builder
}

// Checker is used internally by the Invoke method to decide whether the call was
// successful or it failed.
type Checker func(response *http.Response) bool

// StatusCodeIn is a factory method that returns a Checker function based on the
//  HTTP response status codes: if the actual status code is one of the provided
// "success" values, the check succeeds, otherwise it fails.
func StatusCodeIn(values ...int) Checker {
	return func(response *http.Response) bool {
		for _, value := range values {
			if value == response.StatusCode {
				return true
			}
		}
		return false
	}
}

// Invoke calls an API endpoint at the given path; if the receiver already has a
// base path configured, the given "url" can be relative to it; it can also be a
// full URI; the HTTP "method" identifies the kind of API request. The request
// is prepared by the provided "builder" using the information contained in the
// "input" parameter, which can have tagged fields for query parameters
// (`parameter`), for HTTP headers (`header`) and for the request entity in the
// body (`json`); the response is handled by a default handler and translated
// into values stored into the "output" struct according to their tagging:
// `header` for headers and `json` for body. Both the "input" and the "output"
// parameters must be structs, or the method panics.
func (api *API) Invoke(method string, url string, authenticated bool, checker Checker, input interface{}, output interface{}, failure interface{}) (*Result, error) {

	//log.Debugf("calling method %q on URL %q", method, url)

	request, err := api.PrepareRequest(method, url, authenticated, input)
	if err != nil {
		log.Errorf("error creating request: %v", err)
		return nil, err
	}

	log.Debugf("sending request to %q...", request.URL.EscapedPath())
	t0 := time.Now()
	response, err := api.client.HTTPClient.Do(request)
	if err != nil {
		log.Errorf("error sending request: %v", err)
		return nil, err
	}
	log.Debugf("response received in %v", time.Now().Sub(t0))

	defer response.Body.Close()

	var result *Result
	if checker != nil && checker(response) {
		log.Debugf("handling response as success")
		result, err = api.HandleResponse(response, output)
		result.OK = true
	} else {
		log.Debugf("handling response as failure")
		result, err = api.HandleResponse(response, failure)
	}
	if err != nil {
		log.Errorf("error handling response: %v", err)
	}
	return result, err
}

// PrepareRequest uses information in the input struct to populate HTTP query
// parameters (any field that is tagged with `parameter` will become a parameter),
// headers (fields tagged with `header` will be used to populate request headers)
// and the entity in the request body (fields tagged with `json`). All three are
// optional; if this is the case, pass nil for "input".
func (api *API) PrepareRequest(method string, url string, authenticated bool, input interface{}) (*http.Request, error) {

	builder := api.builder.New(method, url)

	// add authentication header if requested
	if authenticated {
		token := api.client.Authenticator.GetToken()
		if token == nil || token.Value == nil {
			log.Errorf("no valid token available for authenticated call")
			return nil, fmt.Errorf("no valid token for authenticated call")
		}
		builder.Add().Header("X-Auth-Token", *token.Value)
	}

	if input != nil {
		switch reflect.ValueOf(input).Kind() {
		case reflect.Struct:
			// do nothing, input is already a struct, thus it's ok
		case reflect.Ptr:
			// override input by the value it points to if it's a struct
			if reflect.ValueOf(input).Elem().Kind() == reflect.Struct {
				input = reflect.ValueOf(input).Elem().Interface()
			} else {
				panic("only structs can be passed as API input")
			}
		default:
			panic("only structs can be passed as API input")
		}

		// add query parameters, headers and request entity
		builder.
			Add().
			QueryParametersFrom(input).
			VariablesFrom(input).
			HeadersFrom(input).
			WithJSONEntity(input)

		log.Infof("request:\n%v\nwith entity:\n%s", builder, log.ToJSON(input))
	}
	return builder.Make()
}

// HandleResponse parses the HTTP response to an API call and populates the
// "output" struct fields; fields tagged with `header` will be populated using
// the corresponding header value(s) if present; fields tagged with `json` will
// be populated by unmarshalling JSON values in the response. Both are optional;
// if so, pass in nil for "output".
func (api *API) HandleResponse(response *http.Response, output interface{}) (*Result, error) {

	log.Infof("status code: %q", response.Status)

	// read the response data into a buffer
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("error reading response raw data: %v", err)
		return nil, err
	}
	if len(data) == 0 {
		log.Debugf("no payload in response")
		// } else {
		// 	log.Debugf("the response payload is:\n%s\n", string(data))
	}

	if output != nil {

		if reflect.ValueOf(output).Kind() == reflect.Ptr && reflect.ValueOf(output).Elem().Kind() == reflect.String {
			log.Debugf("handling API output as plain string (can set: %t)", reflect.ValueOf(output).Elem().CanSet())
			reflect.ValueOf(output).Elem().SetString(string(data))
			return NewResult(response, data), nil
		}

		if reflect.TypeOf(output).Elem().Kind() != reflect.Struct {
			panic("only structs and *string can be passed as API output")
		}

		// extract headers into the output struct fields tagged with `header` and
		// the response entity into the struct fields tagged with `json`
		t := reflect.TypeOf(output).Elem()
		v := reflect.ValueOf(output).Elem()
		//log.Debugf("%T, %T, %d", t, v, v.Kind)
		for i := 0; i < t.NumField(); i++ {
			if tag := t.Field(i).Tag.Get("header"); tag != "" && tag != "-" {
				value := v.Field(i)
				if value.Kind() == reflect.Ptr {
					value.Set(reflect.New(value.Type().Elem()))
					value.Elem().SetString(response.Header.Get(tag))
				} else if value.Kind() == reflect.String {
					// TODO: test
					value.SetString(response.Header.Get(tag))
				} else {
					// there is an error????
					log.Warnf("invalid field type in output struct: %q", t.Field(i).Name)
				}
				log.Infof("header: %q => %q", t.Field(i).Name, response.Header.Get(tag))
			}
		}

		if len(data) > 0 {
			buffer := bytes.NewBuffer(data)
			if err := json.NewDecoder(buffer).Decode(output); err != nil {
				log.Errorf("error decoding response into entity: %v", err)
				return NewResult(response, data), err
			}
			log.Infof("response entity:\n%s", log.ToJSON(output))
		}
	}
	return NewResult(response, data), nil
}

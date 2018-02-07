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
	"strings"

	"github.com/dghubble/sling"
	"github.com/dihedron/go-openstack/log"
)

// API is the archetype and base struct of all API services.
type API struct {

	// client is the base SDK client, that takes care of sending requests and
	// retrieving responses from the server; moreover it contains some context
	// information such as User-Agent and access to other services. This
	// reference is only for convenience, it makes the public API more user-
	// friendly.
	client *Client

	// requestor is the generator for service-specific requests; it uses the base.
	requestor *sling.Sling
}

// Invoke calls an API endpoint at the given path; if the receiver already has a
// base path configured, the given "url" can be relative to it; it can also be a
// full URI; the HTTP "method" identifies the kind of API request. The request
// is prepared by the provided "builder" using the information contained in the
// "input" parameter, which can have tagged fields for query parameters (`url`),
// for HTTP headers (`header`) and for the request entity in the body (`json`);
// the response is handled by the user-provided "handler" and translated into
// headers and an entity as per the tags in the "output" interface. Both the
// "input" and the "output" parameters would usually be structs (although no
// check is performed since it can also be used by the user-provided builder,
// which could expect anything. If no "builder" or no "handler" is provided, the
// method uses their default implementations which relies on tags as noted above.
func (api *API) Invoke(method string, url string, authenticated bool, input interface{}, output interface{}) (*Result, error) {

	log.Debugf("API.Invoke: calling method %q on URL %q", method, url)

	request, err := api.PrepareRequest(method, url, authenticated, input)
	if err != nil {
		log.Errorf("API.Invoke: error creating request: %v", err)
		return nil, err
	}

	response, err := api.client.HTTPClient.Do(request)
	if err != nil {
		log.Errorf("API.Invoke: error sending request: %v", err)
		return nil, err
	}

	defer response.Body.Close()

	result, err := api.HandleResponse(response, output)
	if err != nil {
		log.Errorf("API.Invoke: error handling response: %v", err)
	}
	return result, err
}

// PrepareRequest uses information in the input struct to populate HTTP query
// parameters (any field that is tagged with `url` will become a parameter),
// headers (fields tagged with `header` will be used to pupolate request headers)
// and the entity in the request body (fields tagged with `json`). All three are
// optional; if this is the case, pass nil for "input".
func (api *API) PrepareRequest(method string, url string, authenticated bool, input interface{}) (*http.Request, error) {

	log.Debugf("API.PrepareRequest: preparing %s request for %s (authenticated: %t)", strings.ToUpper(method), url, authenticated)

	sling := api.requestor.New().Method(method).Path(url)

	// add authentication header if requested
	if authenticated {
		token := api.client.Authenticator.GetToken()
		if token == nil || token.Value == nil {
			log.Errorf("API.PrepareRequest: no valid token available for authenticated call")
			return nil, fmt.Errorf("no valid token for authenticated call")
		}
		log.Debugf("API.PrepareRequest: adding authentication token (X-Auth-Token): %s", ZipString(*token.Value, 16))
		sling.Add("X-Auth-Token", *token.Value)
	}

	if input != nil {
		if reflect.TypeOf(input).Elem().Kind() != reflect.Struct {
			panic("API.PrepareRequest: only structs can be passed as API input")
		}

		// add query parameters
		sling.QueryStruct(input)

		// add headers
		t := reflect.TypeOf(input).Elem()
		v := reflect.ValueOf(input).Elem()
		for i := 0; i < t.NumField(); i++ {
			if tag := t.Field(i).Tag.Get("header"); tag != "" {
				value := reflect.ValueOf(v.Field(i).Interface()).String()
				log.Debugf("API.PrepareRequest: adding header %q => %v", tag, value)
				sling.Add(tag, value)
			}
		}

		// add entity to request body
		sling.BodyJSON(input)
	}

	log.Debugf("API.PrepareRequest: Sling is now %v", sling)

	return sling.Request()
}

// HandleResponse parses the HTTP response to an API call and populates the
// "output" struct fields; fields tagged with `header` will be populated using
// the corresponding header value if present; fields tagged with `json` will be
// populated by unmasìrshalling JSON values in the response. Both are optional;
// if so, pass in nil for "output".
func (api *API) HandleResponse(response *http.Response, output interface{}) (*Result, error) {

	// read the response data into a buffer
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("API.HandleResponse: error reading response raw data: %v", err)
		return nil, err
	}
	if len(data) > 0 {
		log.Debugf("API.HandleResponse: the response payload is:\n%s\n", string(data))
	} else {
		log.Debugf("API.HandleResponse: no payload in response")
	}

	if output != nil {
		if reflect.TypeOf(output).Elem().Kind() != reflect.Struct {
			panic("API.HandleResponse: only structs can be passed as API output")
		}

		// extract headers into the output struct and entity or error into the
		// corresponding struct field() tagged with `entity:"success"` or
		// `entity:"failure"` respectively)
		t := reflect.TypeOf(output).Elem()
		v := reflect.ValueOf(output).Elem()
		//log.Debugf("API.HandleResponse: %T, %T, %d", t, v, v.Kind)
		for i := 0; i < t.NumField(); i++ {
			if tag := t.Field(i).Tag.Get("header"); tag != "" {
				value := v.Field(i)
				if value.Kind() == reflect.Ptr {
					value.Set(reflect.New(value.Type().Elem()))
					value.Elem().SetString(response.Header.Get(tag))
				} else if value.Kind() == reflect.String {
					// TODO: test
					value.SetString(response.Header.Get(tag))
				} else {
					// there is an error????
				}
			}
		}

		if len(data) > 0 {
			buffer := bytes.NewBuffer(data)
			if err := json.NewDecoder(buffer).Decode(output); err != nil {
				log.Errorf("Client.HandleResponse: error decoding response into entity: %v", err)
				return NewResult(response, data), err
			}
			log.Debugf("API.HandleResponse: deserialised entity is:\n%s\n", log.ToJSON(output))
		}
	}

	return NewResult(response, data), nil
}

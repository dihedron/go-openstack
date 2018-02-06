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

	// client is the base SDK client, that takes care of sending requests
	// and retrieving responses from the server; moreover it contains some
	// context information such as User-Agent and access to other services.
	// This reference is only for convenience, it makes the public API more
	// appealing.
	client *Client

	// requestor is the generator for service-specific requests.
	requestor *sling.Sling
}

// RequestBuilder is the signature of a function that, given a Sling,
// fills in the information to turn it into an http.Request ready to
// be submitted by the HTTP client. Its task is that of adding headers,
// query parameters and a request entity according to the information
// provided in the "input" parameter, then sealing the Sling and returning
// the http.Request object ready for submittal.
type RequestBuilder func(sling *sling.Sling, input interface{}) (*http.Request, error)

// ResponseHandler is the signature of a function that, given an http.Response,
// extracts from it the information pertaining to the specific API call: in some
// cases, this can be a few header values, under other circumstances in can be
// an entity, or a combination of the two. The keys parameter, which can be null,
// specifies the headers to extract from the response; the "output" parameter is
// the struct to be used as a template for decoding the JSON response in the
// response payload into an entity and to store the values of the headers; the
// values of headers are stored into fields marked with a "header" tag.
type ResponseHandler func(response *http.Response, output interface{}) (Result, []byte, error)

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
func (api *API) Invoke(method string, url string, authenticated bool, builder RequestBuilder, input interface{}, handler ResponseHandler, output interface{}, success []int) (*Result, []byte, error) {

	log.Debugf("API.Invoke: calling method %q on URL %q", method, url)

	sling := api.requestor.New().Method(method).Path(url)

	if authenticated {
		token := api.client.Authenticator.GetToken()
		if token == nil || token.Value == nil {
			log.Errorf("API.Invoke: no valid token available for authenticated call")
			return nil, nil, fmt.Errorf("no valid token for authenticated call")
		}
		log.Debugf("API.Invoke: adding token %s", ZipString(*token.Value, 16))
		sling.Add("X-Auth-Token", *token.Value)
	}

	log.Debugf("API.Invoke: Sling is now %v", sling)

	if builder == nil {
		builder = DefaultRequestBuilder
	}

	request, err := builder(sling, input)
	if err != nil {
		log.Errorf("API.Invoke: error creating request: %v", err)
		return nil, nil, err
	}

	response, err := api.client.HTTPClient.Do(request)
	if err != nil {
		log.Errorf("API.Invoke: error sending request: %v", err)
		return nil, nil, err
	}

	defer response.Body.Close()

	if handler == nil {
		handler = DefaultResponseHandler
	}

	result, data, err := handler(response, output)
	if err != nil {
		log.Errorf("API.Invoke: error handling response: %v", err)
		return &result, data, err
	}

	if !result.IsInformational() && !result.IsSuccess() {
		log.Warnf("API.Invoke: status code indicates some problem: %v", result)
	}

	return &result, data, nil
}

// DefaultRequestQueryBuilder is the function used by the default builder to
// populate request query parameters using reflection on fields tagged with "url"
// in the "input" struct. This function can also be used by custom implementations
// of RequestBuilders, in order not to have to reinvent the wheel if one has to
// change only how headers or the request body is built whilst accepting the
// default logic for query parameters.
func DefaultRequestQueryBuilder(sling *sling.Sling, input interface{}) *sling.Sling {
	return sling.QueryStruct(input)
}

// DefaultRequestHeadersBuilder is the function used by the default builder to
// populate request headers using reflection on fields tagged with "header" in
// the "input" struct. This function can also be used by custom implementations
// of RequestBuilders, in order not to have to reinvent the wheel if one has to
// change only how query parameters or the request body is built whilst accepting
// the default logic for headers.
func DefaultRequestHeadersBuilder(sling *sling.Sling, input interface{}) *sling.Sling {
	t := reflect.TypeOf(input).Elem()
	v := reflect.ValueOf(input).Elem()
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("header")
		if len(strings.TrimSpace(tag)) > 0 {
			value := reflect.ValueOf(v.Field(i).Interface()).String()
			log.Warnf("API.DefaultRequestBuilder: adding header %q => %v", tag, value)
			sling.Add(tag, value)
		}
	}
	return sling
}

// DefaultRequestEntityBuilder is the function used by the default builder to
// create the JSON entity sent in the request body using reflection on fields
// tagged with "entity" in the "input" struct. This function can also be used by
// custom implementations of RequestBuilders, in order not to have to reinvent
// the wheel if one has to change only how the entity is built whilst accepting
// the default logic for query parameters and headers.
func DefaultRequestEntityBuilder(sling *sling.Sling, input interface{}) *sling.Sling {
	t := reflect.TypeOf(input).Elem()
	v := reflect.ValueOf(input).Elem()
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("entity")
		if len(strings.TrimSpace(tag)) > 0 {
			//value := reflect.ValueOf(v.Field(i).Interface()).String()
			value := reflect.ValueOf(v.Field(i).Elem())
			log.Debugf("API.DefaultRequestEntityBuilder: adding entity %v (%T)", value, value)
			//return sling.BodyJSON(value)
		}
	}

	return sling.BodyJSON(input)
}

// DefaultRequestBuilder fills the Sling with all the necessary information taken
// from the provided opts parameter; opts must be a struct having fields properly
// tagged with "url" (for query parameters), "header" (for HTTP request headers)
// or "entity" (for request entity payload). Mixing the three types of tags is
// supported only for top-level struct elements (shallow scanning). The function
// returns an http.Request object ready for submittal to the API endpoint.
func DefaultRequestBuilder(sling *sling.Sling, input interface{}) (*http.Request, error) {

	if reflect.TypeOf(input).Elem().Kind() != reflect.Struct {
		panic("API.DefaultRequestBuilder: only structs can be passed as API options")
	}
	sling = DefaultRequestQueryBuilder(sling, input)
	sling = DefaultRequestHeadersBuilder(sling, input)
	sling = DefaultRequestEntityBuilder(sling, input)
	return sling.Request()
}

// DefaultResponseHeadersHandler is the function used by the default handler to extract
// the value of a sub-set of response headers into a map; the function is up for use
// by custom implementations of ResponseHandler's so one doesn't have to reinvent the wheel
// simply because a custom entity building logic (different from that of the default handler)
// is needed.
func DefaultResponseHeadersHandler(response *http.Response, output interface{}) error /*map[string][]string*/ {

	t := reflect.TypeOf(output).Elem()
	v := reflect.ValueOf(output).Elem()
	//log.Debugf("API.DefaultResponseHeadersHandler: %T, %T, %d", t, v, v.Kind)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("header")
		if len(strings.TrimSpace(tag)) > 0 {
			value := v.Field(i)
			if value.Kind() == reflect.Ptr {
				value.Set(reflect.New(value.Type().Elem()))
				value.Elem().SetString(response.Header.Get(tag))
			} else if value.Kind() == reflect.String {
				// TODO: test
				value.SetString("ciao")
			} else {
				// there is an error????
			}
		}
	}
	return nil
}

// DefaultResponseEntityHandler is the function used by the default handler to extract
// the valus of the reponse payload as a JSON entity; the function is up for use by custom
// implementations of ResponseHandler's so one doesn't have to reinvent the wheel simply
// because a custom headers extraction logic (different from that of the default handler)
// is needed.
func DefaultResponseEntityHandler(response *http.Response, output interface{}) ([]byte, error) {

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("API.DefaultResponseEntityHandler: error reading response raw data: %v", err)
		return nil, err
	}

	log.Debugf("API.DefaultResponseEntityHandler: the response payload is:\n%s\n", string(data))

	if output != nil {
		buffer := bytes.NewBuffer(data)
		if err := json.NewDecoder(buffer).Decode(output); err != nil {
			log.Errorf("Client.DefaultResponseEntityHandler: error decoding response into entity: %v", err)
			return data, err
		}
		log.Debugf("API.DefaultResponseEntityHandler: deserialised entity is:\n%s\n", log.ToJSON(output))
		return data, nil
	}

	return data, nil
}

// DefaultResponseHandler translates an API call response into a set of
// header values and an entity; headers are extracted from the response
// HTTP headers using the given set of keys; the entity is extracted from
// the HTTP response payload using the entity struct as the base structure
// to fill information into.
func DefaultResponseHandler(response *http.Response, output interface{}) (Result, []byte, error) {
	err := DefaultResponseHeadersHandler(response, output)
	data, err := DefaultResponseEntityHandler(response, output)
	result := NewResult(response)
	return result, data, err
}

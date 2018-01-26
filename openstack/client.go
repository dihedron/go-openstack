// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/dihedron/go-openstack/log"
)

const (
	// SDKVersion is the version of the current library.
	SDKVersion string = "0.0.1"

	// DefaultUserAgent is the default User-Agent string set by the SDK.
	DefaultUserAgent string = "go-openstack/" + SDKVersion
)

// Client is the go-openstack SDK client.
type Client struct {

	// HTTPClient is the HTTP Client used for connectiong to the API endpoints.
	HTTPClient http.Client

	// UserAgent is the User-Agent header value sent to the server.
	UserAgent string

	// Identity is the Identity service API wrapper.
	Identity *IdentityAPI
	// other services here
}

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

	// TODO: may want to add token????
}

// NewDefaultClient returns a new instance of a go-openstack SDK client,
// with sensible defaults for the http.Ckient and the user agent string;
// the Keystone URL must be provided.
func NewDefaultClient(catalogURL string) (*Client, error) {
	return NewClient(catalogURL, nil, nil)
}

// NewClient returns a new instance of a go-openstack SDK client;
// the first parameter is compulsory and represents the URL of the
// Keystone instance from which both the authorization Token and the
// catalog of active services can be retrieved; the others are optional
// and, if null, are automaticelly filled with sensible defaults.
func NewClient(catalogURL string, httpClient *http.Client, userAgent *string) (*Client, error) {

	if len(strings.TrimSpace(catalogURL)) == 0 {
		catalogURL = os.Getenv("OS_AUTH_URL")
	}

	if catalogURL == "" {
		log.Errorln("NewClient: no catalog URL, please provide URL of Keystone server either explicitly or as OS_AUTH_URL")
		return nil, fmt.Errorf("no valid catalog URL")
	}

	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		}
	}

	if userAgent == nil {
		userAgent = String(DefaultUserAgent)
	}

	client := &Client{
		HTTPClient: *httpClient,
		UserAgent:  *userAgent,
	}

	client.Identity = &IdentityAPI{
		API{
			client:    client,
			requestor: sling.New().Base(catalogURL).Set("User-Agent", *userAgent).Client(httpClient),
		},
	}

	// NOTE: other APIs will be dynamically added once we have
	// access to the catalog via an authenticated Keystore request

	return client, nil
}

// RequestBuilder is the signature of a function that, given a Sling,
// fills in the information to turn it into an http.Request ready to
// be submitted by the HTTP client. Its task is that of adding headers,
// query parameters and a request entity according to the information
// provided in the opts parameter, then sealing the Sling and returning
// the http.Request object ready for submittal.
type RequestBuilder func(sling *sling.Sling, opts interface{}) (*http.Request, error)

// ResponseHandler is the signature of a function that, given an
// http.Response, extracts from it the information pertaining to the
// specific API call: in some cases, this can be a few header values,
// under other circumstances in can be an entity, or a combination of
// the two. The keys parameter, which can be null, specifies the headers
// to extract from the response; the entity paraeter is the struct to be
// used as a template for decoding the JSON response in the response
// payload.
type ResponseHandler func(response *http.Response, keys []string, entity interface{}) (Result, map[string][]string, interface{}, error)

// Invoke calls an API endpoint at the given path (under the base path
// provided by the api receiver) with the given HTTP method; the request
// is prepared by the given builder using the information contained in
// the opts parameter; the response is handled by the user-provided handler
// and translated into headers and an entity. The input opts parameter would
// usually point to a struct (although no check is performed since its usage
// is restructed to the user-provided builder, which can arrange a protocol
// of its interest with the API consumer) whose fileds are annotated with
// "url" and "header" values; the request entity can itself be embedded inside
// the opts struct as a "json"-annotated struct closely matching the expected
// request entity payload. If no builder or handler is provided, the method
// uses their default implementations.
func (api *API) Invoke(method string, url string, opts interface{}, keys []string, entity interface{}, builder RequestBuilder, handler ResponseHandler) (map[string][]string, *Result, error) {

	log.Debugf("API.Invoke: calling method %q on URL %q", method, url)

	sling := api.requestor.New().Method(method).Path(url)

	log.Debugf("API.Invoke: Sling is now %v", sling)

	if builder == nil {
		builder = DefaultRequestBuilder
	}

	request, err := builder(sling, opts)
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

	result, headers, entity, err := handler(response, keys, entity)
	if err != nil {
		log.Errorf("API.Invoke: error handling response: %v", err)
		return nil, &result, err
	}

	if !result.IsInformational() && !result.IsSuccess() {
		log.Warnf("API.Invoke: status code indicates some problem: %v", result)
	}

	for key, values := range headers {
		log.Debugf("API.Invoke: header %q => %q", key, values)
	}

	return headers, &result, nil
}

// DefaultRequestQueryBuilder is the function used by the default builder to populate
// request query parameters using reflection on fields tagged with "url" in the input
// opts struct. This function is up for use by custom implementations of RequestBuilders,
// in order not to have to reinvent the wheel if one has to change only how headers or
// the request body is built whilst accepting the default logic for query parameters.
func DefaultRequestQueryBuilder(sling *sling.Sling, opts interface{}) *sling.Sling {
	return sling.QueryStruct(opts)
}

// DefaultRequestHeadersBuilder is the function used by the default builder to populate
// request headers using reflection on fields tagged with "header" in the input
// opts struct. This function is up for use by custom implementations of RequestBuilders,
// in order not to have to reinvent the wheel if one has to change only how query
// parameters or the request body is built whilst accepting the default logic for headers.
func DefaultRequestHeadersBuilder(sling *sling.Sling, opts interface{}) *sling.Sling {
	t := reflect.TypeOf(opts).Elem()
	v := reflect.ValueOf(opts).Elem()
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("header")
		if len(strings.TrimSpace(tag)) > 0 {
			value := reflect.ValueOf(v.Field(i).Interface()).String()
			log.Debugf("Client.DefaultRequestBuilder: adding header %q => %q", tag, value)
			sling.Add(tag, value)
		}
	}
	return sling
}

// DefaultRequestEntityBuilder is the function used by the default builder to create
// the JSON entity sent in the request body using reflection on fields tagged with
// "json" in the input opts struct. This function is up for use by custom implementations
// of RequestBuilders, in order not to have to reinvent the wheel if one has to change
// only how the entity is built whilst accepting the default logic for query parameters
// and headers.
func DefaultRequestEntityBuilder(sling *sling.Sling, opts interface{}) *sling.Sling {
	return sling.BodyJSON(opts)
}

// DefaultRequestBuilder fills the Sling with all the necessary information taken
// from the provided opts parameter; opts must be a struct having fields properly
// tagged with "url" (for query parameters), "header" (for HTTP request headers)
// or "json" (for request entity payload). Mixing the three types of tags is
// supported only for top-level struct elements (shallow scanning). The function
// returns an http.Request object ready for submittal to the API endpoint.
func DefaultRequestBuilder(sling *sling.Sling, opts interface{}) (*http.Request, error) {

	if reflect.TypeOf(opts).Elem().Kind() != reflect.Struct {
		panic("Client.DefaultRequestBuilder: only structs can be passed as API options")
	}
	sling = DefaultRequestQueryBuilder(sling, opts)
	sling = DefaultRequestHeadersBuilder(sling, opts)
	sling = DefaultRequestEntityBuilder(sling, opts)
	return sling.Request()
}

// DefaultResponseHeadersHandler is the function used by the default handler to extract
// the value of a sub-set of response headers into a map; the function is up for use
// by custom implementations of ResponseHandler's so one doesn't have to reinvent the wheel
// simply because a custom entity building logic (different from that of the default handler)
// is needed.
func DefaultResponseHeadersHandler(response *http.Response, keys []string) map[string][]string {
	var headers map[string][]string

	if keys != nil {
		headers = map[string][]string{}
		for _, key := range keys {
			headers[key] = append(headers[key], response.Header.Get(key))
		}
	}
	return headers
}

// DefaultResponseEntityHandler is the function used by the default handler to extract
// the valus of the reponse payload as a JSON entity; the function is up for use by custom
// implementations of ResponseHandler's so one doesn't have to reinvent the wheel simply
// because a custom headers extraction logic (different from that of the default handler)
// is needed.
func DefaultResponseEntityHandler(response *http.Response, entity interface{}) (interface{}, error) {

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Client.DefaultResponseEntityHandler: error reading response raw data: %v", err)
		return nil, err
	}

	log.Debugf("Client.DefaultResponseEntityHandler: the response payload is:\n%s\n", string(data))

	if entity != nil {
		buffer := bytes.NewBuffer(data)
		if err := json.NewDecoder(buffer).Decode(entity); err != nil {
			log.Errorf("Client.DefaultResponseEntityHandler: error decoding response into entity: %v", err)
			return nil, err
		}
		log.Debugf("Client.DefaultResponseEntityHandler: deserialised entity is:\n%s\n", log.ToJSON(entity))
		return entity, nil
	}

	return data, nil
}

// DefaultResponseHandler translates an API call response into a set of
// header values and an entity; headers are extracted from the response
// HTTP headers using the given set of keys; the entity is extracted from
// the HTTP response payload using the entity struct as the base structure
// to fill information into.
func DefaultResponseHandler(response *http.Response, keys []string, entity interface{}) (Result, map[string][]string, interface{}, error) {
	headers := DefaultResponseHeadersHandler(response, keys)
	entity, err := DefaultResponseEntityHandler(response, entity)
	result := NewResult(response)
	return result, headers, entity, err
}

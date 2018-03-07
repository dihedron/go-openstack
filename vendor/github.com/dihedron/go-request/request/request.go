// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package request

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/structs"
)

type operation int8

const (
	add operation = iota
	set
	del
	rem
)

// Builder is the HTTP request builder; it can be used to create child request
// factories, with specialised BaseURLs or other parameters; when a sub-builder
// is generated, it will share the Method and BaseURL by value and all other
// firlds by pointer, so any change in sub-factories will affect the parent too.
type Builder struct {

	// method is the HTTP method to be used for requests generated by this
	// builder.
	method string

	// url is the base URL for generating HTTP requests.
	url string

	// op is used internally to provide a flowing API to header and query parameters
	// maipulation methods.
	op operation

	// headers is a set of header values for HTTP request headers; special headers
	// such as "User-Agent" and "Content-Type" are stored here.
	headers http.Header

	// query is a set of values set in the URL as query parameters.
	parameters url.Values

	// entity is the entity provider; it will be used to generate the request
	// entity as an io.Reader. Moreover, it will be queried to set the request
	// content type.
	body io.Reader
}

// New returns a new request builder; the URL can be omitted and specified
// later via Base() or Path().
func New(url string) *Builder {
	return &Builder{
		method:     http.MethodGet,
		url:        url,
		headers:    map[string][]string{},
		parameters: map[string][]string{},
	}
}

// New clones the current builder and can optionally specify the request method
// and/or the request URL.
func (f *Builder) New(method, url string) *Builder {
	clone := &Builder{
		method:     f.method,
		url:        f.url,
		headers:    map[string][]string{},
		parameters: map[string][]string{},
		body:       f.body,
	}
	if method != "" {
		clone.method = strings.ToUpper(method)
	}
	if url != "" {
		clone.Path(url)
	}
	for key, values := range f.headers {
		if _, ok := clone.headers[key]; !ok {
			clone.headers[key] = []string{}
		}
		clone.headers[key] = append(clone.headers[key], values...)
	}
	for key, values := range f.parameters {
		if _, ok := clone.parameters[key]; !ok {
			clone.parameters[key] = []string{}
		}
		clone.parameters[key] = append(clone.parameters[key], values...)
	}
	return clone
}

// Base sets the base URL. If you intend to extend the url with Path, the URL
// should be specified with a trailing slash.
func (f *Builder) Base(url string) *Builder {
	f.url = url
	return f
}

// Path overrides the builder URL; absolute and relative URLs can be used.
// TODO: improve documentation showing relative paths
func (f *Builder) Path(path string) *Builder {
	baseURL, baseErr := url.Parse(f.url)
	pathURL, pathErr := url.Parse(path)
	if baseErr == nil && pathErr == nil {
		f.url = baseURL.ResolveReference(pathURL).String()
	}
	return f
}

// Method sets the default HTTP method for factoory-generated requests.
func (f *Builder) Method(method string) *Builder {
	if method != "" {
		f.method = strings.ToUpper(strings.TrimSpace(method))
	}
	return f
}

// UserAgent sets the user agent information in the request builder; the previous
// value is discarded.
func (f *Builder) UserAgent(userAgent string) *Builder {
	return f.Set().Header("User-Agent", userAgent)
}

// ContentType sets the content type information in the request builder; the
// previous value is discarded.
func (f *Builder) ContentType(contentType string) *Builder {
	return f.Set().Header("Content-Type", contentType)
}

// Add is used to provide a fluent API by which it is possible to add query
// parameters and headers without having many different methods or intermediate
// objects; this method relies on an internal Builder field (named op), which
// will be set to the "add" value and will instruct the following QueryParameter()
// and Header() methods to add the passed values to the current set for the given
// key.
func (f *Builder) Add() *Builder {
	f.op = add
	return f
}

// Set is used to provide a fluent API by which it is possible to replace query
// parameters and headers without having many different methods or intermediate
// objects; this method relies on an internal Builder field (named op), which
// will be set to the "set" value and will instruct the following QueryParameter()
// and Header() methods to replace the current set of values for the given key
// with the passed values.
func (f *Builder) Set() *Builder {
	f.op = set
	return f
}

// Del is used to provide a fluent API by which it is possible to replace query
// parameters and headers without having many different methods or intermediate
// objects; this method relies on an internal Builder field (named op), which
// will be set to the "set" value and will instruct the following QueryParameter()
// and Header() methods to replace the current set of values for the given key
// with the passed values.
func (f *Builder) Del() *Builder {
	f.op = del
	return f
}

// Remove is used to provide a fluent API by which it is possible to remove the
// values of query parameters and headers whose keys match a regular exception.
func (f *Builder) Remove() *Builder {
	f.op = rem
	return f
}

// QueryParameter adds, sets or removes the given set of values to the URL's query
// parameters; if the query parameter is being removed, there is no need to specify
// any value; if the query parameter is being reset, the key is regarded as a
// regular expression.
func (f *Builder) QueryParameter(key string, values ...string) *Builder {
	if f.op == add {
		for _, value := range values {
			f.parameters.Add(key, value)
		}
	} else if f.op == set {
		f.parameters.Del(key)
		for _, value := range values {
			f.parameters.Add(key, value)
		}
	} else if f.op == del {
		f.parameters.Del(key)
	} else if f.op == rem {
		re := regexp.MustCompile(key)
		for key := range f.parameters {
			if re.MatchString(key) {
				defer f.parameters.Del(key)
			}
		}
	}
	return f
}

// QueryParametersFrom adds, sets or removes values extracted from a struct (and
// tagged with "parameter") or from a map[string][]string to the URL's query
// parameters; if the query parameters are being removed, there is no need to
// specify any value in the input struct/map; if the query parameters are being
// reset, the keys are regarded as regular expressions.
func (f *Builder) QueryParametersFrom(source interface{}) *Builder {
	for key, values := range getValuesFrom("parameter", source) {
		f.QueryParameter(key, values...)
	}
	return f
}

// Header adds, sets or removes the given set of values to the URL's headers; if
// the header is being removed, there is no need to specify any value; if the
// header is being reset, the key is regarded as a regular expression.
func (f *Builder) Header(key string, values ...string) *Builder {
	if f.op == add {
		for _, value := range values {
			f.headers.Add(key, value)
		}
	} else if f.op == set {
		f.headers.Del(key)
		for _, value := range values {
			f.headers.Add(key, value)
		}
	} else if f.op == del {
		f.headers.Del(key)
	} else if f.op == rem {
		re := regexp.MustCompile(key)
		for key := range f.headers {
			if re.MatchString(key) {
				defer f.headers.Del(key)
			}
		}
	}
	return f
}

// HeadersFrom adds, sets or removes values extracted from a struct (and tagged
// with "header") or from a map[string][]string to the URL's headers; if the
// headers are being removed, there is no need to  specify any value in the input
// struct/map; if the headers are being reset, the keys are regarded as regular
// expressions.
func (f *Builder) HeadersFrom(source interface{}) *Builder {
	for key, values := range getValuesFrom("header", source) {
		f.Header(key, values...)
	}
	return f
}

// WithEntity sets the io.Reader from which the request body (payload) will be
// read; if nil is passed, the request will have no payload; the Content-Type
// MUST be provoded separately.
func (f *Builder) WithEntity(entity io.Reader) *Builder {
	f.body = entity
	return f
}

// WithJSONEntity sets an io.Reader that returns a JSON fragment as per the
// input struct; if no Content-Type has been set already, the method will
// automatically set it to "application/json".
func (f *Builder) WithJSONEntity(entity interface{}) *Builder {

	switch reflect.ValueOf(entity).Kind() {
	case reflect.Struct:
		// do nothing, entity is already a struct, thus it's ok
	case reflect.Ptr:
		// override entity by the value it points to if it's a struct
		if reflect.ValueOf(entity).Elem().Kind() == reflect.Struct {
			entity = reflect.ValueOf(entity).Elem().Interface()
		} else {
			panic("only structs can be passed as source for JSON entities")
		}
	default:
		panic("only structs can be passed as source for JSON entities")
	}

	data, err := json.Marshal(entity)
	if err != nil {
		return nil
	}

	if f.headers.Get("Content-Type") == "" {
		f.ContentType("application/json")
	}

	f.body = bytes.NewReader(data)
	return f
}

// WithXMLEntity sets an io.Reader that returns an XML fragment as per the
// input struct; if no Content-Type has been set already, the method will
// automatically set it to "application/xml".
func (f *Builder) WithXMLEntity(entity interface{}) *Builder {

	switch reflect.ValueOf(entity).Kind() {
	case reflect.Struct:
		// do nothing, entity is already a struct, thus it's ok
	case reflect.Ptr:
		// override entity by the value it points to if it's a struct
		if reflect.ValueOf(entity).Elem().Kind() == reflect.Struct {
			entity = reflect.ValueOf(entity).Elem().Interface()
		} else {
			panic("only structs can be passed as source for XML entities")
		}
	default:
		panic("only structs can be passed as source for XML entities")
	}

	data, err := xml.Marshal(entity)
	if err != nil {
		return nil
	}

	if f.headers.Get("Content-Type") == "" {
		f.ContentType("text/xml")
	}

	f.body = bytes.NewReader(data)
	return f
}

// Get sets the builder method to "GET" and returns an http.Request.
func (f *Builder) Get() *Builder {
	return f.Method(http.MethodGet)
}

// Post sets the builder method to "POST" and returns an http.Request.
func (f *Builder) Post() *Builder {
	return f.Method(http.MethodPost)
}

// Put sets the builder method to "PUT" and returns an http.Request.
func (f *Builder) Put() *Builder {
	return f.Method(http.MethodPut)
}

// Patch sets the builder method to "PATCH" and returns an http.Request.
func (f *Builder) Patch() *Builder {
	return f.Method(http.MethodPatch)
}

// Delete sets the builder method to "DELETE" and returns an http.Request.
func (f *Builder) Delete() *Builder {
	return f.Method(http.MethodDelete)
}

// Head sets the builder method to "HEAD" and returns an http.Request.
func (f *Builder) Head() *Builder {
	return f.Method(http.MethodHead)
}

// Trace sets the builder method to "TRACE" and returns an http.Request.
func (f *Builder) Trace() *Builder {
	return f.Method(http.MethodTrace)
}

// Options sets the builder method to "OPTIONS" and returns an http.Request.
func (f *Builder) Options() *Builder {
	return f.Method(http.MethodOptions)
}

// Connect sets the builder method to "CONNECT" and returns an http.Request.
func (f *Builder) Connect() *Builder {
	return f.Method(http.MethodConnect)
}

// Make creates a new http.Request from the information available in the Builder.
func (f *Builder) Make() (*http.Request, error) {

	// parse URL to validate
	url, err := url.Parse(f.url)
	if err != nil {
		return nil, err
	}

	// augment URL with additional query parameters
	url, err = addQueryParameters(url, f.parameters)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(f.method, url.String(), f.body)
	if err != nil {
		return nil, err
	}

	request.Header = f.headers

	return request, nil
}

func getValuesFrom(tag string, source interface{}) map[string][]string {
	var m map[string][]string
	switch reflect.ValueOf(source).Kind() {
	case reflect.Struct:
		m = getValuesFromStruct(tag, source)
	case reflect.Map:
		var ok bool
		if m, ok = source.(map[string][]string); !ok {
			panic("only structs and maps can be passed as sources")
		}
	case reflect.Ptr:
		if reflect.ValueOf(source).Elem().Kind() == reflect.Struct {
			source = reflect.ValueOf(source).Elem().Interface()
			m = getValuesFromStruct(tag, source)
		} else if reflect.ValueOf(source).Elem().Kind() == reflect.Map {
			source = reflect.ValueOf(source).Elem().Interface()
			var ok bool
			if m, ok = source.(map[string][]string); !ok {
				panic("only structs and maps can be passed as sources")
			}
		} else {
			panic("only structs and maps can be passed as sources")
		}
	default:
		panic("only structs and maps can be passed as sources")
	}
	return m
}

func getValuesFromStruct(tag string, source interface{}) map[string][]string {
	result := map[string][]string{}
	for key, values := range scan(tag, source) {
		for _, value := range values {
			if _, ok := result[key]; !ok {
				result[key] = []string{}
			}
			if reflect.ValueOf(value).Kind() == reflect.Ptr {
				result[key] = append(result[key], fmt.Sprintf("%v", reflect.ValueOf(value).Elem().Interface()))
			} else {
				result[key] = append(result[key], fmt.Sprintf("%v", value))
			}
		}
	}
	return result
}

func addQueryParameters(requestURL *url.URL, parameters url.Values) (*url.URL, error) {
	qp, err := url.ParseQuery(requestURL.RawQuery)
	if err != nil {
		return nil, err
	}
	// encodes query structs into a url.Values map and merges maps
	for key, values := range parameters {
		for _, value := range values {
			qp.Add(key, value)
		}
	}
	// url.Values formats to a sorted "url encoded" string, e.g. "key=val&foo=bar"
	requestURL.RawQuery = qp.Encode()
	return requestURL, nil
}

// scan is the actual workhorse method: it scans the source struct for tagged
// headers and extracts their values; if any embedded or child struct is
// encountered, it is scanned for values.
func scan(key string, source interface{}) map[string][]interface{} {
	result := map[string][]interface{}{}
	for _, field := range structs.Fields(source) {
		if field.IsEmbedded() || field.Kind() == reflect.Struct ||
			(field.Kind() == reflect.Ptr && reflect.ValueOf(field.Value()).Elem().Kind() == reflect.Struct) {
			for k, v := range scan(key, field.Value()) {
				if values, ok := result[k]; ok {
					result[k] = append(values, v...)
				} else {
					result[k] = v
				}
			}
		} else {
			tag := field.Tag(key)
			if tag != "" {
				value := field.Value()
				if values, ok := result[tag]; ok {
					result[tag] = append(values, value)
				} else {
					result[tag] = []interface{}{value}
				}
			}
		}
	}
	return result
}

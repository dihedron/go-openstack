// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"net"
	"net/http"
	"time"
)

// DefaultUserAgent is the default User-Agent string set by the SDK.
const DefaultUserAgent string = "go-openstack 0.0.1"

// Client is the go-openstack SDK client.
type Client struct {
	// HTTPClient is the HTTP Client used for connectiong to the API endpoints.
	HTTPClient http.Client
	// UserAgent is the User-Agent header value sent to the server.
	UserAgent string
	// Identity is the Identity API wrapper.
	Identity *IdentityAPI
}

// NewDefaultClient returns a new instance of a go-openstack SDK client,
// with sensible defaults for the http.Ckient and the user agent string;
// the Keystone URL must be provided.
func NewDefaultClient(url string) (*Client, error) {
	return NewClient(url, nil, nil)
}

// NewClient returns a new instance of a go-openstack SDK client;
// the first parameter is compulsory and represents the URL of the
// Keystone instance; the others are optional and, if null, are
// automaticelly filled with sensible defaults.
func NewClient(url string, client *http.Client, agent *string) (*Client, error) {
	if client == nil {
		client = &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		}
	}

	if agent == nil {
		agent = String(DefaultUserAgent)
	}

	return &Client{
		HTTPClient: *client,
		UserAgent:  *agent,
		Identity:   newIdentityAPI(url, client, *agent),
	}, nil
}

type Result struct {
	Status  int
	Payload interface{}
}

/*
type ErrorHandler func(res *http.Response)

func (c *Client) Get(url string, headers *map[string][]string, query interface{}, body interface{}, eh ErrorHandler) (interface{}, error) {
	return "", nil
}

func (c *Client) Post(url string, headers *map[string][]string, query interface{}, body interface{}, eh ErrorHandler) (interface{}, error) {
	c.New().Post(url).QueryStruct(query)

	.BodyJSON(body).Request(); err == nil {
		res, err := i.client.Do(req)
		if err != nil {
			log.Errorf("Identity.CreateToken: error sending request: %v", err)
			return err
		}
		defer res.Body.Close()

		body := &createTokenResponseBody{}
		json.NewDecoder(res.Body).Decode(body)
		b, _ := json.MarshalIndent(body, "", "  ")
		fmt.Printf("RESPONSE HEADER:\n%s\nRESPONSE BODY:\n%s\n", res.Header.Get("X-Subject-Token"), b)
	}
}
*/

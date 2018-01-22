// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"net"
	"net/http"
	"time"
)

const DefaultUserAgent string = "go-openstach 0.0.1"

// Client is the go-openstack SDK client.
type Client struct {
	// The HTTP Client used for connectiong to the API endpoints.
	HTTPClient http.Client
	// The User-Agent string sent to the server.
	Agent string
	// The Identity API wrapper.
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
		Agent:      *agent,
		Identity:   newIdentityAPI(url, client, *agent),
	}, nil
}

/*
func (c *Client) Post(user string, headers *map[string][]string, query interface{}, body interface{}) (interface{}, error) {
	return "", nil
}
*/

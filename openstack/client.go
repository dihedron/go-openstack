// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"fmt"
	"net"
	"net/http"
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

	// StartURL is the address of the Keystone server used to receive the first
	// authentication Token and then proceed to browsing the Catalog.
	StartURL string

	// UserAgent is the User-Agent header value sent to the server.
	UserAgent string

	// Identity is the Identity API wrapper.
	Identity *IdentityService
	// other services here
}

// NewDefaultClient returns a new instance of a go-openstack SDK client,
// with sensible defaults for the http.Ckient and the user agent string;
// the Keystone URL must be provided.
func NewDefaultClient(url string) (*Client, error) {
	return NewClient(url, nil, nil)
}

// NewClient returns a new instance of a go-openstack SDK client;
// the first parameter is compulsory and represents the URL of the
// Keystone instance from which both the authorization Token and the
// catalog of active services can be retrieved; the others are optional
// and, if null, are automaticelly filled with sensible defaults.
func NewClient(startURL string, httpClient *http.Client, userAgent *string) (*Client, error) {

	if len(strings.TrimSpace(startURL)) == 0 {
		log.Errorln("NewClient: invalid KeyStone URL")
		return nil, fmt.Errorf("invalid keystone URL")
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
		//		StartURL:   startURL,
		//		UserAgent:  *userAgent,
		// may want to add token????
	}

	client.Identity = &IdentityService{
		Client:         client,
		RequestFactory: sling.New().Base(startURL).Set("User-Agent", *userAgent).Client(httpClient),
	}

	return client, nil
}

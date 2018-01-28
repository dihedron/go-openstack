// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"fmt"
	"net"
	"net/http"
	"os"
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

	// Authenticator is the Identity service API wrapper used for the first
	// authentication and to retrieve the API /services) catalog; it is
	// treated in a special way since it is the only service that can be
	// accessed without an authorisation; moreover it returns the list of
	// all the other services, and publicy
	Authenticator *Authenticator

	// other services here

	//services map[]
}

// NewDefaultClient returns a new instance of a go-openstack SDK client,
// with sensible defaults for the http.Ckient and the user agent string;
// the Keystone URL must be provided.
func NewDefaultClient() *Client {
	return NewClient(nil, nil)
}

// NewClient returns a new instance of a go-openstack SDK client; the
// httpClient parameter allows to use one's own implementation of a
// http.Client, e.g. to support custom mechanisms for TLS etc.; the second
// parameter allows to specify one's own User-Agent string. If any of the
// parameters is omitted (that is, nil), sensible defaults are automatically
// provided by the SDK.
func NewClient(httpClient *http.Client, userAgent *string) *Client {

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
		Authenticator: &Authenticator{
			Identity: &IdentityV3API{
				API{
					client: nil, // initialise later (*) with pointer to this same struct
					//requestor: sling.New().Base(catalogURL).Set("User-Agent", *userAgent).Client(httpClient),
					requestor: sling.New().Set("User-Agent", *userAgent).Client(httpClient),
				},
			},
			TokenValue: nil,
			TokenInfo:  nil,
		},
	}
	// (*) initialised here!
	client.Authenticator.Identity.client = client

	// NOTE: other APIs will be dynamically added once we have
	// access to the catalog via an authenticated Keystore request

	return client
}

//
// ConnectTo configures the client for connection to the given URL; this URL
// represents the address of the Keystone instance from which both the
// authorization Token and the catalog of active services will be retrieved.
func (c *Client) ConnectTo(catalogURL string) (*Client, error) {
	if len(strings.TrimSpace(catalogURL)) == 0 {
		catalogURL = os.Getenv("OS_AUTH_URL")
	}

	if catalogURL == "" {
		log.Errorln("NewClient: no catalog URL, please provide URL of Keystone server either explicitly or as OS_AUTH_URL")
		return nil, fmt.Errorf("no valid catalog URL")
	}

	c.Authenticator.Identity.API.requestor.Base(catalogURL)

	return c, nil
}

// GetServices returns the set of currently supported services as per the
// catalog; each service is defined by a name, a type and a set of endpoints,
// each of which has an URL and specifies whether it is a public, administra-
// tive or internal interface and the region to which it belongs.
func (c *Client) GetServices() *[]Service {
	return c.Authenticator.TokenInfo.Catalog
}

// IdentityV3 returns an IdentityV3API service reference.
func (c *Client) IdentityV3() *IdentityV3API {
	// TODO
	return nil
}

// {
// 	"id": "c73e65c7a9bf4b0c931b1f11e2f62071",
// 	"name": "keystone",
// 	"type": "identity",
// 	"endpoints": [
// 	  {
// 		"id": "4d856dd8a69c4aefb83d88a32c2106ba",
// 		"interface": "public",
// 		"region": "RegionOne",
// 		"region_id": "RegionOne",
// 		"url": "http://192.168.56.101/identity"
// 	  },
// 	  {
// 		"id": "ed9e3bdcd3e14f92aab4e04dbe75044b",
// 		"interface": "admin",
// 		"region": "RegionOne",
// 		"region_id": "RegionOne",
// 		"url": "http://192.168.56.101/identity"
// 	  }
// 	]
// },

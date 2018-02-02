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

	// This is the profile, that is the set of service, along with version, URL
	// and interfaces reuqested by the user; this is used to apply a filter to
	// the set of services and endpoints in the catalog.
	Profile *Profile

	// This is the set of available services; it is populated as soon as the
	// client performs a logon to the identity service and retrieves the catalog.
	Services map[string]interface{}
}

// NewDefaultClient returns a new instance of a go-openstack SDK client,
// with sensible defaults for the http.Ckient and the user agent string;
// the Keystone URL must be provided.
func NewDefaultClient(authURL string) *Client {
	return NewClient(authURL, nil, nil)
}

// NewClient returns a new instance of a go-openstack SDK client; the
// httpClient parameter allows to use one's own implementation of a
// http.Client, e.g. to support custom mechanisms for TLS etc.; the second
// parameter allows to specify one's own User-Agent string. If any of the
// parameters is omitted (that is, nil), sensible defaults are automatically
// provided by the SDK.
func NewClient(authURL string, httpClient *http.Client, userAgent *string) *Client {

	if len(strings.TrimSpace(authURL)) == 0 {
		authURL = os.Getenv("OS_AUTH_URL")
		log.Debugf("NewClient: identity service URL through $OS_AUTH_URL\n")
	}

	if authURL == "" {
		log.Errorln("NewClient: no identity service URL, please provide URL of identity service either explicitly or through $OS_AUTH_URL")
		return nil
	}

	log.Debugf("NewClient: connecting to identity service at %q\n", authURL)

	if httpClient == nil {
		log.Debugln("NewClient: connecting using library-provided HTTP client")
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

	log.Debugf("NewClient: HTTP client will present itself as %q\n", userAgent)

	client := &Client{
		HTTPClient: *httpClient,
		UserAgent:  *userAgent,
		Services:   map[string]interface{}{},
	}

	// now we've got a reference to client and we can finally
	// initialise the authenticator with a backref to it
	client.Authenticator = &Authenticator{
		AuthURL: String(authURL),
		Identity: &IdentityV3API{
			API{
				client:    client,
				requestor: sling.New().Set("User-Agent", *userAgent).Client(httpClient).Base(NormaliseURL(authURL)),
			},
		},
		TokenValue: nil,
		TokenInfo:  nil,
	}

	// NOTE: other APIs will be dynamically added once we have
	// access to the catalog via an authenticated Keystore request

	return client
}

// Connect attempts to perform a login to an identity service already configured
// via a call to For; the opts parameter is the set of values needed for logging
// in to the identity service.
func (c *Client) Connect(opts *LoginOpts) error {

	if c.Authenticator.AuthURL == nil {
		log.Errorf("Client.Connect: no identity service URL configured")
		return fmt.Errorf("no valid identity service URL")
	}

	c.Authenticator.Logout()

	err := c.Authenticator.Login(opts)
	if err != nil {
		log.Errorf("Client.Connect: error logging in to the identity service at %q", c.Authenticator.AuthURL)
		return err
	}

	if c.Authenticator.TokenInfo.Catalog == nil {
		log.Errorf("Client.Connect: no catalog info available")
		return fmt.Errorf("no catalog information available from identity service")
	}

	for _, service := range *c.Authenticator.TokenInfo.Catalog {
		// log.Debugf("Client.Connect: checking service %q (%q)\n", *service.Type, *service.Name)

		//outer:
		for _, endpoint := range *service.Endpoints {
			// log.Debugf("Client.Connect: checking endpoint, interface %q, region %q, URL %q\n", *endpoint.Interface, *endpoint.Region, *endpoint.URL)
			// inner:
			if c.Profile != nil {
				// look for a match between a service and a filter before proceeding
				log.Debugln("Client.Connect: applying filters to catalog")

			inner:
				for _, filter := range c.Profile.Filters {
					// log.Debugf("Client.Connect: does filter type %q, interface %q, region %q, URL %q match?\n", *filter.Type, *endpoint.Interface, *endpoint.Region, *endpoint.URL)
					if *service.Type != *filter.Type {
						continue inner
					}
					if *endpoint.Interface != *filter.Interface {
						continue inner
					}
					if *endpoint.Region != *filter.Region {
						continue inner
					}
					if *endpoint.URL != *filter.EndpointURL {
						continue inner
					}

					log.Debugf("Client.Connect: service %q (type: %q, interface %q, region %q, URL %q) matches filter, adding to catalog\n", *service.Name, *service.Type, *endpoint.Interface, *endpoint.Region, *endpoint.URL)
				}
			}

			switch *service.Type {
			case "identity":
				c.Services[*service.Type] = IdentityV3API{
					API{
						client:    c,
						requestor: sling.New().Set("User-Agent", c.UserAgent).Client(&c.HTTPClient).Base(NormaliseURL(*endpoint.URL)),
					},
				}
			default:
				log.Debugf("Client.Connect: unsupported service %q (type: %q)\n", *service.Name, *service.Type)
			}
		}
	}

	return nil
}

// Close closes the client and releases the identity token; it can be used
// to defer client cleanup.
func (c *Client) Close() error {
	c.Services = map[string]interface{}{}
	return c.Authenticator.Logout()
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
	for k, v := range c.Services {
		if k == "identity" {
			api := v.(IdentityV3API)
			return &api
		}
	}
	return nil
}

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

	"github.com/dihedron/go-log/log"
	"github.com/dihedron/sling"
)

const (
	// SDKVersion is the version of the current library.
	SDKVersion string = "0.0.1"

	// DefaultUserAgent is the default User-Agent string set by the SDK.
	DefaultUserAgent string = "go-openstack/" + SDKVersion
)

// Client is the go-openstack SDK client.
type Client struct {

	// HTTPClient is the HTTP Client used for connecting to the API endpoints.
	HTTPClient http.Client

	// UserAgent is the User-Agent header value sent to the server by the client.
	UserAgent string

	// Authenticator is an Identity V3 API wrapper used for authentication and
	// to retrieve the list of endpoints associated with the session token; it
	// is a bit "special" because it is the only service that can be accessed
	// without an authorisation; moreover it returns the list of all the other
	// services available through the authentication token.
	Authenticator *Authenticator

	// This is the profile, that is the set of service, along with version, URL
	// and interfaces reuqested by the user; this is used to apply a filter to
	// the set of services and endpoints in the catalog.
	Profile *Profile

	// This is the set of available services; it is populated as soon as the
	// client performs a logon to the identity service and retrieves the catalog.
	Services map[string]interface{}
}

// NewDefaultClient returns a new instance of a go-openstack SDK client, with
// sensible defaults for the http.Ckient and the user agent string; the Keystone
// URL must be provided either explicitly or through the $OS_AUTH_URL variable.
func NewDefaultClient(authURL string) *Client {
	return NewClient(authURL, nil, nil)
}

// NewClient returns a new instance of a go-openstack SDK client; the httpClient
// parameter allows to use one's own implementation of a http.Client, e.g. to
// support custom mechanisms for TLS etc.; the second parameter allows to specify
// one's own User-Agent string; the Keystone URL must be provided either
// explicitly or through the $OS_AUTH_URL variable. If any of the parameters is
// omitted (that is, it is left is nil), sensible defaults are automatically
// provided by the SDK.
func NewClient(authURL string, httpClient *http.Client, userAgent *string) *Client {

	if len(strings.TrimSpace(authURL)) == 0 {
		authURL = os.Getenv("OS_AUTH_URL")
		log.Debugf("identity service URL through $OS_AUTH_URL\n")
	}

	if authURL == "" {
		log.Errorln("no identity service URL, please provide URL of identity service either explicitly or through $OS_AUTH_URL")
		return nil
	}

	log.Debugf("connecting to identity service at %q\n", authURL)

	if httpClient == nil {
		log.Debugln("connecting using library-provided HTTP client")
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

	log.Debugf("HTTP client will present itself as %q\n", *userAgent)

	client := &Client{
		HTTPClient: *httpClient,
		UserAgent:  *userAgent,
		Services:   map[string]interface{}{},
	}

	// now we've got a reference to client and we can finally initialise the
	// authenticator with a backref to it
	client.Authenticator = &Authenticator{
		AuthURL: String(authURL),
		Identity: &IdentityV3API{
			API{
				client:    client,
				requestor: sling.New().Set("User-Agent", *userAgent).Client(httpClient).Base(NormaliseURL(authURL)),
			},
		},
		token: nil,
	}

	return client
}

// Connect attempts to perform a login to an identity service already configured
// via a call to For; the opts parameter is the set of values needed for logging
// in to the identity service.
func (c *Client) Connect(opts *LoginOptions) error {

	if c.Authenticator.AuthURL == nil {
		log.Errorf("no identity service URL configured")
		return fmt.Errorf("no valid identity service URL")
	}

	c.Authenticator.Logout()

	err := c.Authenticator.Login(opts)
	if err != nil {
		log.Errorf("error logging in to the identity service at %q", c.Authenticator.AuthURL)
		return err
	}

	if c.Authenticator.GetCatalog() == nil {
		log.Warnf("no catalog info available (maybe it's an unscoped login?)")

		// add the identity service anyway, otherwise we may not be able to ever
		// get access to the catalog
		c.Services["identity"] = *(c.Authenticator.Identity)
		return nil
	}

	for _, service := range *c.Authenticator.GetCatalog() {
		// log.Debugf("checking service %q (%q)\n", *service.Type, *service.Name)

		//outer:
		for _, endpoint := range *service.Endpoints {
			// log.Debugf("checking endpoint, interface %q, region %q, URL %q\n", *endpoint.Interface, *endpoint.Region, *endpoint.URL)
			// inner:
			if c.Profile != nil {
				// look for a match between a service and a filter before proceeding
				log.Debugln("applying filters to catalog")

			inner:
				for _, filter := range c.Profile.Filters {
					// log.Debugf("does filter type %q, interface %q, region %q, URL %q match?\n", *filter.Type, *endpoint.Interface, *endpoint.Region, *endpoint.URL)
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

					log.Debugf("service %q (type: %q, interface %q, region %q, URL %q) matches filter, adding to catalog\n", *service.Name, *service.Type, *endpoint.Interface, *endpoint.Region, *endpoint.URL)
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
				log.Debugf("unsupported service %q (type: %q)\n", *service.Name, *service.Type)
			}
		}
	}

	return nil
}

// Close closes the client and releases the identity token; it can be used to
// defer client cleanup.
func (c *Client) Close() error {
	log.Debugf("closing client")
	c.Services = map[string]interface{}{}
	return c.Authenticator.Logout()
}

// GetServices returns the set of currently supported services as per the
// catalog; each service is defined by a name, a type and a set of endpoints,
// each of which has an URL and specifies whether it is a public, administra-
// tive or internal interface and the region to which it belongs.
func (c *Client) GetServices() *[]Service {
	return c.Authenticator.GetCatalog()
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

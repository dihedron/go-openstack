// Copyright 2017 Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"net/url"
	"os"

	"github.com/dihedron/go-openstack/log"
)

// IdentityService implements the identity functionalities
// of OpenStack.
type IdentityService struct {
	Service
}

// RegisterIdentityService activates in the current session an
// IdentityService provider which must be available at the given
// endpoint URL; the URL contains the base of the
func (s *Connection) RegisterIdentityService(endpoint string) error {
	if s == nil {
		log.Errorln("invalid Session reference")
		return ErrorInvalidReference
	}
	if endpoint == "" {
		log.Errorln("invalid input URL")
		return ErrorInvalidInput
	}

	// now parse the given URL and extract the base part and the version, if any
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Errorf("error parsing URL %q: %v\n", endpoint, err)
		return err
	}
	i := &IdentityService{
		Service{
			session: s,
			base:    NormalizeURL(u.String()),
		},
	}
	if u.Path != "" {
		i.Service.endpoint = NormalizeURL(endpoint)
	}
	log.Debugf("endpoint: %q, base URL: %q", i.Service.endpoint, i.Service.base)

	// register service provider
	s.Identity = i
	return nil
}

// AuthOptions represents the parameters passed to the API to
// authenticate against the remote service.
type AuthOptions struct {
	AuthURL    *string
	UserID     *string
	UserName   *string
	Password   *string
	TenantID   *string
	TenantName *string
	DomainID   *string
	DomainName *string
}

// NewAuthOptions returns an empty set of AuthOptions.
func NewAuthOptions() *AuthOptions {
	return &AuthOptions{}
}

// FromEnv initialises the given AuthOptions structure using
// information from the environment; no validation is performed;
// unset variables leave a nil reference whereas empty variables
// have a reference to an empty value.
func (o *AuthOptions) FromEnv() *AuthOptions {
	if value, ok := os.LookupEnv("OS_AUTH_URL"); ok {
		o.AuthURL = String(value)
	}
	if value, ok := os.LookupEnv("OS_USERID"); ok {
		o.UserID = String(value)
	}
	if value, ok := os.LookupEnv("OS_USERNAME"); ok {
		o.UserName = String(value)
	}
	if value, ok := os.LookupEnv("OS_PASSWORD"); ok {
		o.Password = String(value)
	}
	if value, ok := os.LookupEnv("OS_TENANT_ID"); ok {
		o.TenantID = String(value)
	}
	if value, ok := os.LookupEnv("OS_TENANT_NAME"); ok {
		o.TenantName = String(value)
	}
	if value, ok := os.LookupEnv("OS_DOMAIN_ID"); ok {
		o.DomainID = String(value)
	}
	if value, ok := os.LookupEnv("OS_DOMAIN_NAME"); ok {
		o.DomainName = String(value)
	}
	return o
}

// IsValid returns whether the structure contains the minimum
// information needed to attempt an authentication request.
func (o *AuthOptions) IsValid() (bool, error) {
	if o == nil {
		return false, ErrorInvalidReference
	}
	if o.AuthURL == nil || *o.AuthURL == "" {
		return false, ErrorInvalidInput.Where("AuthURL", "must not be null")
	}
	if (o.UserName == nil || *o.UserName == "") && (o.UserID == nil || *o.UserID == "") {
		return false, ErrorInvalidInput.Where("UserID or UserName", "at least one must not be null")
	}
	if o.Password == nil || *o.Password == "" {
		return false, ErrorInvalidInput.Where("Password", "must not be null")
	}
	return true, nil
}

/*
func (i *IdentityService) AuthenticateByPassword(opts *AuthOptions) error {
	if opts == nil || opts.AuthURL == nil {
		return ErrorInvalidInput
	}
	u, err := url.Parse(*opts.AuthURL)
	if err != nil {
		return err
	}

	i.base = NormalizeURL(u.String())

	if u.Path != "" {
		i.endpoint = NormalizeURL(*opts.AuthURL)
	}

	versions := []*Version{
		{ID: v20, Priority: 20, Suffix: "/v2.0/"},
		{ID: v30, Priority: 30, Suffix: "/v3/"},
	}

	i.client.NegotiateVersion(i.endpoint, i.base, versions)

	return nil
}
*/

/*
{
  "versions": {
    "values": [
      {
        "status": "stable",
        "updated": "2016-10-06T00:00:00Z",
        "media-types": [
          {
            "base": "application/json",
            "type": "application/vnd.openstack.identity-v3+json"
          }
        ],
        "id": "v3.7",
        "links": [
          {
            "href": "http://10.114.10.2:5000/v3/",
            "rel": "self"
          }
        ]
      },
      {
        "status": "deprecated",
        "updated": "2016-08-04T00:00:00Z",
        "media-types": [
          {
            "base": "application/json",
            "type": "application/vnd.openstack.identity-v2.0+json"
          }
        ],
        "id": "v2.0",
        "links": [
          {
            "href": "http://10.114.10.2:5000/v2.0/",
            "rel": "self"
          },
          {
            "href": "http://docs.openstack.org/",
            "type": "text/html",
            "rel": "describedby"
          }
        ]
      }
    ]
  }
}
*/

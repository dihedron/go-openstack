// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dghubble/sling"
	"github.com/dihedron/go-openstack/log"
)

// Authentication contains the identity entity used to authenticate users
// and issue tokens against a Keystone instance; it can be scoped (when
// either Project or Domain is specified), implicitly unscoped ()
type Authentication struct {
	Identity *Identity   `json:"identity,omitempty"`
	Scope    interface{} `json:"scope,omitempty"`
}

type Domain struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type Endpoint struct {
	ID        *string `json:"id,omitempty"`
	Interface *string `json:"interface,omitempty"`
	Region    *string `json:"region,omitempty"`
	RegionID  *string `json:"region_id,omitempty"`
	URL       *string `json:"url,omitempty"`
}

type Identity struct {
	Methods  *[]string `json:"methods,omitempty"`
	Password *Password `json:"password,omitempty"`
	Token    *Token    `json:"token,omitempty"`
}

type Password struct {
	User *User `json:"user,omitempty"`
}

type Project struct {
	ID     *string `json:"id,omitempty"`
	Name   *string `json:"name,omitempty"`
	Domain *Domain `json:"domain,omitempty"`
}

type Role struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type Scope struct {
	Project *Project `json:"project,omitempty"`
	Domain  *Domain  `json:"domain,omitempty"` // either one or the other: if both, BadRequest!
}

type Service struct {
	ID        *string     `json:"id,omitempty"`
	Name      *string     `json:"name,omitempty"`
	Type      *string     `json:"type,omitempy"`
	Endpoints *[]Endpoint `json:"endpoints,omitempty"`
}

type Token struct {
	ID           *string    `json:"id,omitempty"`
	IssuedAt     *string    `json:"issued_at,omitempty"`
	ExpiresAt    *string    `json:"expires_at,omitempty"`
	User         *User      `json:"user,omitempty"`
	Roles        *[]Role    `json:"roles,omitempty"`
	Methods      *[]string  `json:"methods,omitempty"`
	AuditIDs     *[]string  `json:"audit_ids,omitempty"`
	Project      *Project   `json:"project,omitempty"`
	IsDomain     *bool      `json:"is_domain,omitempty"`
	IsAdminToken *bool      `json:"is_admin_token,omitempty"`
	Catalog      *[]Service `json:"catalog,omitempty"`
}

type User struct {
	ID                *string `json:"id,omitempty"`
	Name              *string `json:"name,omitempty"`
	Domain            *Domain `json:"domain,omitempty"`
	Password          *string `json:"password,omitempty"`
	PasswordExpiresAt *string `json:"password_expires_at,omitempty"`
}

// IdentityAPI represents the identity API providing all services
// regarding authentication, authentication, role and reosurce management.
// See https://developer.openstack.org/api-ref/identity/v3/
type IdentityAPI struct {
	api    *sling.Sling
	client *http.Client
}

/*
 * AUTHENTICATION AND TOKEN MANAGEMENT
 */

const (
	// CreateTokenMethodPassword is the constant used for password-based
	// authentication onto the Keystone server.
	CreateTokenMethodPassword string = "password"
	// CreateTokenMethodToken is the constant used for token-based
	// authentication onto the Keystone server.
	CreateTokenMethodToken string = "token"
)

// CreateTokenOpts contains the set of parameters and options used to
// perform an authentication (create an authentication token).
type CreateTokenOpts struct {
	Method           string
	NoCatalog        bool
	UserID           *string
	UserName         *string
	UserDomainID     *string
	UserDomainName   *string
	UserPassword     *string
	TokenID          *string
	ScopeProjectID   *string
	ScopeProjectName *string
	ScopeDomainID    *string
	ScopeDomainName  *string
	UnscopedToken    *bool
}

type createTokenRequestQuery struct {
	NoCatalog bool `url:"nocatalog,omitempty"`
}

type createTokenRequestBody struct {
	Auth *Authentication `json:"auth,omitempty"`
}

type createTokenResponseBody struct {
	Token *Token `json:"token,omitempty"`
}

// CreateToken uses the provided parameters to authenticate the client to the
// Keystone server and receive a token.
func (i IdentityAPI) CreateToken(opts *CreateTokenOpts) error {

	body := &createTokenRequestBody{
		Auth: &Authentication{
			Identity: &Identity{
				Methods: &[]string{
					opts.Method,
				},
			},
		},
	}

	if opts.Method == CreateTokenMethodPassword {
		if opts.UserID != nil && len(strings.TrimSpace(*opts.UserID)) > 0 {
			body.Auth.Identity.Password = &Password{
				User: &User{
					ID:       opts.UserID,
					Password: opts.UserPassword,
				},
			}
		} else {
			body.Auth.Identity.Password = &Password{
				User: &User{
					Name:     opts.UserName,
					Password: opts.UserPassword,
				},
			}
			if opts.UserDomainID != nil && len(strings.TrimSpace(*opts.UserDomainID)) > 0 {
				body.Auth.Identity.Password.User.Domain = &Domain{
					ID: opts.UserDomainID,
				}
			} else {
				body.Auth.Identity.Password.User.Domain = &Domain{
					Name: opts.UserDomainName,
				}
			}
		}
	} else if opts.Method == CreateTokenMethodToken {
		if opts.TokenID != nil && len(strings.TrimSpace(*opts.TokenID)) > 0 {
			body.Auth.Identity.Token = &Token{
				ID: opts.TokenID,
			}
		}
	}

	// manage scoped/unscoped token requests
	if opts.ScopeProjectID != nil && len(strings.TrimSpace(*opts.ScopeProjectID)) > 0 {
		body.Auth.Scope = &Scope{
			Project: &Project{
				ID: opts.ScopeProjectID,
			},
		}
	} else if opts.ScopeProjectName != nil && len(strings.TrimSpace(*opts.ScopeProjectName)) > 0 {
		scope := &Scope{
			Project: &Project{
				Name: opts.ScopeProjectName,
			},
		}

		if opts.ScopeDomainID != nil && len(strings.TrimSpace(*opts.ScopeDomainID)) > 0 {
			scope.Project.Domain = &Domain{
				ID: opts.ScopeDomainID,
			}
		} else {
			scope.Project.Domain = &Domain{
				Name: opts.ScopeDomainName,
			}
		}
		body.Auth.Scope = scope
	} else {
		if opts.ScopeDomainID != nil && len(strings.TrimSpace(*opts.ScopeDomainID)) > 0 {
			body.Auth.Scope = &Scope{
				Domain: &Domain{
					ID: opts.ScopeDomainID,
				},
			}
		} else if opts.ScopeDomainName != nil && len(strings.TrimSpace(*opts.ScopeDomainName)) > 0 {
			body.Auth.Scope = &Scope{
				Domain: &Domain{
					Name: opts.ScopeDomainName,
				},
			}
		} else if opts.UnscopedToken != nil && *opts.UnscopedToken {
			// all values are null: the request is unscoped
			body.Auth.Scope = String("unscoped")
		}
	}
	b, _ := json.MarshalIndent(body, "", "  ")
	log.Debugf("Identity.CreateToken: request body is\n%s\n", b)

	query := &createTokenRequestQuery{
		NoCatalog: opts.NoCatalog,
	}

	if req, err := i.api.New().Post("/identity/v3/auth/tokens").QueryStruct(query).BodyJSON(body).Request(); err == nil {
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

	return nil
}

/*
 * PRIVATE METHODS
 */

// newIdentityAPI ceates a new instance of the Indentity API wrapper; the
// URL parameter is the URL of the Keystone service providing the service;
// the http.Client is the HTTP client (provided by the user or in its default
// implementation) used to perform the API requests, and the agent is the
// User-Agent header sent along with the requests.
func newIdentityAPI(url string, client *http.Client, agent string) *IdentityAPI {
	if strings.TrimSpace(url) == "" {
		panic("invalid url")
	}
	id := &IdentityAPI{
		api:    sling.New().Base(url).Set("User-Agent", agent).Client(client),
		client: client,
	}

	return id
}

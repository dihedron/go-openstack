// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"encoding/json"
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

// Project represents a container that groups or isolates resources or identity
// objects; depending on the service operator, a project might map to a customer,
// account, organization, or tenant.
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

// Service represents an OpenStack service, such as Compute (nova), Object Storage
// (swift), or Image service (glance), that provides one or more endpoints through
// which users can access resources and perform operations.
type Service struct {
	ID        *string     `json:"id,omitempty"`
	Name      *string     `json:"name,omitempty"`
	Type      *string     `json:"type,omitempy"`
	Endpoints *[]Endpoint `json:"endpoints,omitempty"`
}

// Token represents An alpha-numeric text string that enables access to OpenStack
// APIs and resources and all associated metadata. A token may be revoked at any
// time and is valid for a finite duration. While OpenStack Identity supports
// token-based authentication in this release, it intends to support additional
// protocols in the future. OpenStack Identity is an integration service that does
// not aspire to be a full-fledged identity store and management solution.
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

// User is a digital representation of a person, system, or service that uses
// OpenStack cloud services. The Identity service validates that incoming requests
// are made by the user who claims to be making the call. Users have a login and
// can access resources by using assigned tokens. Users can be directly assigned
// to a particular project and behave as if they are contained in that project.
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
	factory *sling.Sling
	client  *http.Client
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

/*
 * CREATE TOKEN
 */

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

// CreateToken uses the provided parameters to authenticate the client to the
// Keystone server and receive a token.
func (api IdentityAPI) CreateToken(opts *CreateTokenOpts) (string, *Token, error) {

	query, _ := initCreateTokenRequestQuery(opts)

	// no headers in request!

	body, _ := initCreateTokenRequestBody(opts)

	log.Debugf("Identity.CreateToken: request body is\n%s\n", log.ToJSON(body))

	var err error
	if req, err := api.factory.New().Post("/identity/v3/auth/tokens").QueryStruct(query).BodyJSON(body).Request(); err == nil {
		res, err := api.client.Do(req)
		if err != nil {
			log.Errorf("Identity.CreateToken: error sending request: %v", err)
			return "", nil, err
		}
		defer res.Body.Close()

		if res.StatusCode == 201 {
			body := &createTokenResponseBody{}
			json.NewDecoder(res.Body).Decode(body)

			header := res.Header.Get("X-Subject-Token")

			log.Debugf("Identity.CreateToken: token value:\n%s\n", header)
			log.Debugf("Identity.CreateToken: token info:\n%s\n", log.ToJSON(body))
			return header, body.Token, nil
		}

		err = FromResponse(res)
		log.Debugf("Identity.CreateToken: API call unsuccessful: %v", err)
		return "", nil, err
	}

	log.Errorf("Identity.CreateToken: error creating request: %v\n", err)
	return "", nil, err
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

// initCreateTokenRequestQuery creates the struct used to pass the request
// options that go on the query string.
func initCreateTokenRequestQuery(opts *CreateTokenOpts) (interface{}, error) {
	return &createTokenRequestQuery{
		NoCatalog: opts.NoCatalog,
	}, nil
}

// initCreateTokenRequestHeaders creates a pmap of header values to be
// passed to the server along with the request.
func initCreateTokenRequestHeaders(opts *CreateTokenOpts) (map[string][]string, error) {
	return map[string][]string{}, nil
}

// initCreateTokenRequestBody creates the structure representing the request
// entity; the struct will be automatically serialised to JSON by the client.
func initCreateTokenRequestBody(opts *CreateTokenOpts) (interface{}, error) {

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
	return body, nil
}

/*
 * VALIDATE AND GET TOKEN INFO
 */

// ValidateTokenOpts contains the set of parameters and options used to
// perform the valudation of a token on the Identity server.
type ValidateTokenOpts struct {
	NoCatalog    bool
	AllowExpired bool
	SubjectToken string
}

// ValidateToken uses the provided parameters to validate the given token and retrieve
// information about it from the Identity server; this API requires a valid admin
// token.
func (api IdentityAPI) ValidateToken(token string, opts *ValidateTokenOpts) (*Token, error) {
	query, _ := initValidateTokenRequestQuery(opts)

	headers, _ := initValidateTokenRequestHeaders(token, opts)

	// no entities in body!

	log.Debugf("Identity.ValidateToken: checking subject token:\n%s\n", opts.SubjectToken)

	var err error
	sling := api.factory.New().Get("/identity/v3/auth/tokens").QueryStruct(query)
	for key, values := range headers {
		for _, value := range values {
			sling.Add(key, value)
		}
	}
	if req, err := sling.Request(); err == nil {
		res, err := api.client.Do(req)
		if err != nil {
			log.Errorf("Identity.ValidateToken: error sending request: %v", err)
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode == 200 {
			body := &validateTokenResponseBody{}
			json.NewDecoder(res.Body).Decode(body)

			log.Debugf("Identity.ValidateToken: token info:\n%s\n", log.ToJSON(body))
			return body.Token, nil
		}

		err = FromResponse(res)
		log.Debugf("Identity.ValidateToken: API call unsuccessful: %v", err)
		return nil, err
	}

	log.Errorf("Identity.ValidateToken: error creating request: %v\n", err)
	return nil, err
}

type validateTokenRequestQuery struct {
	NoCatalog    bool `url:"nocatalog,omitempty"`
	AllowExpired bool `url:"allow_expired,omitempty"`
}

type validateQueryRequestHeaders map[string][]string

type validateTokenRequestBody struct{}

type validateTokenResponseBody struct {
	Token *Token `json:"token,omitempty"`
}

// initValidateTokenRequestQuery creates the struct used to pass the request
// options that go on the query string.
func initValidateTokenRequestQuery(opts *ValidateTokenOpts) (interface{}, error) {
	return &validateTokenRequestQuery{
		NoCatalog:    opts.NoCatalog,
		AllowExpired: opts.AllowExpired,
	}, nil
}

// initValidateTokenRequestHeaders creates a map of header values to be
// passed to the server along with the request.
func initValidateTokenRequestHeaders(token string, opts *ValidateTokenOpts) (map[string][]string, error) {
	return map[string][]string{
		"X-Auth-Token": []string{
			token,
		},
		"X-Subject-Token": []string{
			opts.SubjectToken,
		},
	}, nil
}

// initValidateTokenRequestBody creates the structure representing the request
// entity; the struct will be automatically serialised to JSON by the client.
func initValidateTokenRequestBody(opts *ValidateTokenOpts) (interface{}, error) {
	return nil, nil
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
		factory: sling.New().Base(url).Set("User-Agent", agent).Client(client),
		client:  client,
	}

	return id
}

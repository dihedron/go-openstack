// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"net/http"
	"os"
	"strings"

	"github.com/dihedron/go-log"
)

// CreateTokenOptions provides all the options for password-based, token-based
// and app-credentials-based logon, both scoped  and unscoped, with and without
// the associated catalog of available services.
type CreateTokenOptions struct {
	TokenID          *string
	UserID           *string
	UserName         *string
	UserDomainID     *string
	UserDomainName   *string
	UserPassword     *string
	Secret           *string
	AppCredentialID  *string
	ScopeProjectID   *string
	ScopeProjectName *string
	ScopeDomainID    *string
	ScopeDomainName  *string
	UnscopedToken    *bool
	NoCatalog        *bool
	Authenticated    bool
}

/*
 * CREATE TOKEN
 */

// CreateToken uses the provided parameters to authenticate the client or the
// application to the Keystone server and receive a token; authentication can be
// performed via username and password ("password" method), via an existing
// token ("token" method, e.g. when an unscoped token is already available), or
// via a pre-existing secret issued by Keystone ("application_credential" method,
// used to authenticate an application to the platform as if it were interacting
// on behalf of a user and authorising it to a subset of the user's resources
// without sharing the user's credentials).
func (api *IdentityV3API) CreateToken(opts *CreateTokenOptions) (*Token, *Result, error) {

	input := &struct {
		NoCatalog *bool           `parameter:"nocatalog,omitempty" header:"-" json:"-"`
		Auth      *Authentication `parameter:"-" header:"-" json:"auth,omitempty"`
	}{
		NoCatalog: opts.NoCatalog,
	}

	if opts.UserPassword != nil && len(*opts.UserPassword) > 0 {
		log.Debugf("logging in by password")
		input.Auth = &Authentication{
			Identity: &Identity{
				Methods: &[]string{
					"password",
				},
				Password: &Password{
					User: &User{
						ID:       opts.UserID,
						Name:     opts.UserName,
						Password: opts.UserPassword,
						Domain: &Domain{
							ID:   opts.UserDomainID,
							Name: opts.UserDomainName,
						},
					},
				},
			},
		}
	} else if opts.TokenID != nil && len(*opts.TokenID) > 0 {
		log.Debugf("logging in by token")
		input.Auth = &Authentication{
			Identity: &Identity{
				Methods: &[]string{
					"token",
				},
				Token: &Token{
					ID: opts.TokenID,
				},
			},
		}
	} else if opts.AppCredentialID != nil && len(*opts.AppCredentialID) > 0 {
		log.Debugf("logging in by app credential")
		input.Auth = &Authentication{
			Identity: &Identity{
				Methods: &[]string{
					"application_credential",
				},
				AppCredential: &AppCredential{
					ID: opts.AppCredentialID,
					User: &User{
						ID:   opts.UserID,
						Name: opts.UserName,
						Domain: &Domain{
							ID:   opts.UserDomainID,
							Name: opts.UserDomainName,
						},
					},
				},
			},
		}
	}

	input.Auth.Scope = initCreateTokenOptionsScope(opts)

	// log.Debugf("entity in request body is\n%s\n", log.ToJSON(input))

	output := &struct {
		SubjectToken *string `parameter:"-" header:"X-Subject-Token" json:"-"`
		Token        *Token  `parameter:"-" header:"-" json:"token,omitempy"`
	}{}

	result, err := api.Invoke(http.MethodPost, "./v3/auth/tokens", opts.Authenticated, input, output)
	log.Debugf("result is %v (%v)", result, err)
	if output.SubjectToken != nil {
		output.Token.Value = output.SubjectToken
	}
	return output.Token, result, err
}

// initCreateTokenOptionsScope initialises the Scope section of the Authentication
// object in the HTTP request entity; there are a few priority rules for scoping:
// for details see the OpenStack Identity v3 documentation.
func initCreateTokenOptionsScope(opts *CreateTokenOptions) interface{} {
	// manage scoped/unscoped token requests
	if opts.ScopeProjectID != nil && len(strings.TrimSpace(*opts.ScopeProjectID)) > 0 {
		return &Scope{
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
		return scope
	} else {
		if opts.ScopeDomainID != nil && len(strings.TrimSpace(*opts.ScopeDomainID)) > 0 {
			return &Scope{
				Domain: &Domain{
					ID: opts.ScopeDomainID,
				},
			}
		} else if opts.ScopeDomainName != nil && len(strings.TrimSpace(*opts.ScopeDomainName)) > 0 {
			return &Scope{
				Domain: &Domain{
					Name: opts.ScopeDomainName,
				},
			}
		} else if opts.UnscopedToken != nil && *opts.UnscopedToken {
			// all values are null: the request is unscoped
			return String("unscoped")
		}
	}
	return nil
}

// CreateTokenFromEnv uses the information in the environment to authenticate the
// client to the Keystore server and receive a token.
func (api *IdentityV3API) CreateTokenFromEnv() (*Token, *Result, error) {
	opts := &CreateTokenOptions{
		UserName:       String(os.Getenv("OS_USERNAME")),
		UserPassword:   String(os.Getenv("OS_PASSWORD")),
		UserDomainName: String(os.Getenv("OS_USER_DOMAIN_NAME")),
	}

	if os.Getenv("OS_PROJECT_NAME") != "" && os.Getenv("OS_PROJECT_DOMAIN_NAME") != "" {
		opts.ScopeProjectName = String(os.Getenv("OS_PROJECT_NAME"))
		opts.ScopeDomainName = String(os.Getenv("OS_PROJECT_DOMAIN_NAME"))
	} else {
		opts.UnscopedToken = Bool(true)
	}

	return api.CreateToken(opts)
}

/*
 * VALIDATE AND GET TOKEN INFO
 */

// ReadTokenOptions contains the set of parameters and options used to perform the
// validation of a token on the Identity server.
type ReadTokenOptions struct {
	NoCatalog    *bool  `parameter:"nocatalog,omitempty" header:"-" json:"-"`
	AllowExpired *bool  `parameter:"allow_expired,omitempty" header:"-" json:"-"`
	SubjectToken string `parameter:"-" header:"X-Subject-Token" json:"-"`
}

// ReadToken uses the provided parameters to read the given token and retrieve
// information about it from the Identity server; this API requires a valid admin
// token.
func (api *IdentityV3API) ReadToken(opts *ReadTokenOptions) (*Token, *Result, error) {
	output := &struct {
		Token        *Token  `parameter:"-" header:"-" json:"token,omitempy"`
		SubjectToken *string `parameter:"-" header:"X-Subject-Token" json:"-"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/tokens", true, opts, output)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		output.Token.Value = output.SubjectToken
		return output.Token, result, err
	}

	log.Debugf("header is %s\n", *output.SubjectToken)

	return nil, result, err
}

/*
 * CHECK TOKEN
 */

// CheckTokenOptions contains the set of parameters and options used to perform the
// validation of a token on the Identity server.
type CheckTokenOptions struct {
	AllowExpired *bool  `parameter:"allow_expired,omitempty" header:"-" json:"-"`
	SubjectToken string `parameter:"-" header:"X-Subject-Token" json:"-"`
}

// CheckToken uses the provided parameters to check the given token and retrieve
// information about it from the Identity server; this API requires a valid admin
// token.
func (api *IdentityV3API) CheckToken(opts *CheckTokenOptions) (bool, *Result, error) {
	result, err := api.Invoke(http.MethodHead, "./v3/auth/tokens", true, opts, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 || result.Code == 204 {
		return true, result, err
	}
	return false, result, err
}

/*
 * DELETE TOKEN
 */

// DeleteTokenOptions contains the set of parameters and options used to perform
// the deletion of a token on the Identity server.
type DeleteTokenOptions struct {
	SubjectToken string `parameter:"-" header:"X-Subject-Token" json:"-"`
}

// DeleteToken uses the provided parameters to delete the given token; the token
// is immediately invalid regardless of the value in the expires_at attribute;
// this API requires a valid admin token.
func (api *IdentityV3API) DeleteToken(opts *DeleteTokenOptions) (bool, *Result, error) {
	result, err := api.Invoke(http.MethodDelete, "./v3/auth/tokens", true, opts, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 || result.Code == 204 {
		return true, result, err
	}
	return false, result, err
}

/*
 * GET CATALOG
 */

// ReadCatalog retrieves the catalog associated with the given authorisation
// token; the catalog is returned even if the token was issued withouth a catalog
// (?nocataog=true).
func (api *IdentityV3API) ReadCatalog() (*[]Service, *Result, error) {
	output := &struct {
		Catalog *[]Service `header:"-" json:"catalog,omitempty"`
		Links   *Links     `header:"-" json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/catalog", true, nil, output)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Catalog, result, err
	}
	return nil, result, err
}

/*
 * LIST PROJECTS
 */

// ListProjects returns the list of projects that are available to be scoped to
// based on the X-Auth-Token provided in the request. The structure of the
// response is exactly the same as listing projects for a user.
func (api *IdentityV3API) ListProjects() (*[]Project, *Result, error) {
	output := &struct {
		Projects *[]Project `header:"-" json:"projects,omitempty"`
		Links    *Links     `header:"-" json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/projects", true, nil, output)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Projects, result, err
	}
	return nil, result, err
}

/*
 * LIST DOMAINS
 */

// ListDomains returns the list of domains that are available to be scoped to
// based on the X-Auth-Token provided in the request. The structure is the same
// as listing domains.
func (api *IdentityV3API) ListDomains() (*[]Domain, *Result, error) {
	output := &struct {
		Domains *[]Domain `header:"-" json:"domains,omitempty"`
		Links   *Links    `header:"-" json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/domains", true, nil, output)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Domains, result, err
	}
	return nil, result, err
}

/*
 * LIST SYSTEMS
 */

// ListSystems returns the list of systems that are available to be scoped to
// based on the X-Auth-Token provided in the request.
func (api *IdentityV3API) ListSystems() (*[]System, *Result, error) {
	output := &struct {
		Systems *[]System `header:"-" json:"system,omitempty"`
		Links   *Links    `header:"-" json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/system", true, nil, output)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Systems, result, err
	}
	return nil, result, err
}

/*
 * LIST USERS
 */

// ListUsersOptions provides all the options available for filtering the list
// of registered users.
type ListUsersOptions struct {
	DomainID           *string     `parameter:"domain_id,omitempty" header:"-" json:"-"`
	Enabled            *bool       `parameter:"enabled,omitempty" header:"-" json:"-"`
	IdentityProviderID *string     `parameter:"idp_id,omitempty" header:"-" json:"-"`
	Name               *string     `parameter:"name,omitempty" header:"-" json:"-"`
	PasswordExpiresAt  *TimeFilter `parameter:"password_expires_at,omitempty" header:"-" json:"-"`
	ProtocolID         *string     `parameter:"protocol_id,omitempty" header:"-" json:"-"`
	UniqueID           *string     `parameter:"unique_id,omitempty" header:"-" json:"-"`
}

// ListUsers returns the list of users on the system.
func (api *IdentityV3API) ListUsers(opts *ListUsersOptions) (*[]User, *Result, error) {
	output := &struct {
		Users *[]User `header:"-" json:"users,omitempty"`
		Links *Links  `header:"-" json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/users", true, opts, output)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Users, result, err
	}
	return nil, result, err
}

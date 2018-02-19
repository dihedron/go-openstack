package openstack

import (
	"net/http"
	"os"
	"strings"

	"github.com/dihedron/go-log/log"
)

// IdentityV3API represents the identity API ver. 3, providing support for
// authentication, authorization, role and resource management.
// See https://developer.openstack.org/api-ref/identity/v3/
type IdentityV3API struct {
	API
}

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
	NoCatalog        bool `url:"nocatalog,omitempty"`
	Authenticated    bool
}

/*
 * CREATE TOKEN
 */

// CreateToken uses the provided parameters to authenticate the client or the
// application to the Keystone server and receive a token; application can be
// performed via username and password ("password" method), via an existing
// token ("token" method, e.g. when an unscoped token is already available), or
// via a pre-existing secret issued by Keystone ("application_credential" method,
// used to authenticate applications to the platform as if it were interacting on
// behalf of a user and authorising it to a  subset of the user's resources
// without sharing the user's credentials).
func (api *IdentityV3API) CreateToken(opts *CreateTokenOptions) (*Token, *Result, error) {

	input := &struct {
		NoCatalog bool            `url:"nocatalog,omitempty" json:"-"`
		Auth      *Authentication `json:"auth,omitempty"`
	}{}

	input.NoCatalog = opts.NoCatalog

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
		input.NoCatalog = opts.NoCatalog
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

	log.Debugf("entity in request body is\n%s\n", log.ToJSON(input))

	output := &struct {
		SubjectToken *string `header:"X-Subject-Token" json:"-"`
		Token        *Token  `json:"token,omitempy"`
	}{}

	log.Debugf("before invoking API")

	result, err := api.Invoke(http.MethodPost, "./v3/auth/tokens", opts.Authenticated, input, output)
	log.Debugf("result is %v (%v)", result, err)
	if output.SubjectToken != nil {
		output.Token.Value = output.SubjectToken
		return output.Token, result, err
	}
	return output.Token, result, err
}

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

// ReadTokenOpts contains the set of parameters and options used to perform the
// valudation of a token on the Identity server.
type ReadTokenOpts struct {
	NoCatalog    bool   `url:"nocatalog,omitempty" json:"-"`
	AllowExpired bool   `url:"allow_expired,omitempty" json:"-"`
	SubjectToken string `header:"X-Subject-Token" json:"-"`
}

// ReadToken uses the provided parameters to read the given token and retrieve
// information about it from the Identity server; this API requires a valid admin
// token.
func (api *IdentityV3API) ReadToken(opts *ReadTokenOpts) (*Token, *Result, error) {
	output := &struct {
		Token        *Token  `json:"token,omitempy"`
		SubjectToken *string `header:"X-Subject-Token" json:"-"`
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

// CheckTokenOpts contains the set of parameters and options used to perform the
// validation of a token on the Identity server.
type CheckTokenOpts struct {
	AllowExpired bool   `url:"allow_expired,omitempty" json:"-"`
	SubjectToken string `header:"X-Subject-Token" json:"-"`
}

// CheckToken uses the provided parameters to check the given token and retrieve
// information about it from the Identity server; this API requires a valid admin
// token.
func (api *IdentityV3API) CheckToken(opts *CheckTokenOpts) (bool, *Result, error) {
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

// DeleteTokenOpts contains the set of parameters and options used to perform the
// deletion of a token on the Identity server.
type DeleteTokenOpts struct {
	SubjectToken string `header:"X-Subject-Token" json:"-"`
}

// DeleteToken uses the provided parameters to delete the given token; the token
// is immediately invalid regardless of the value in the expires_at attribute;
// this API requires a valid admin token.
func (api *IdentityV3API) DeleteToken(opts *DeleteTokenOpts) (bool, *Result, error) {
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
		Catalog *[]Service `json:"catalog,omitempty"`
		Links   *Links     `json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/catalog", true, nil, output)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Catalog, result, err
	}
	return nil, result, err
}

/*
 * GET PROJECTS
 */

// ReadProjects returns the list of projects that are available to be scoped to
// based on the X-Auth-Token provided in the request. The structure of the
// response is exactly the same as listing projects for a user.
func (api *IdentityV3API) ReadProjects() (*[]Project, *Result, error) {
	output := &struct {
		Projects *[]Project `json:"projects,omitempty"`
		Links    *Links     `json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/projects", true, nil, output)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Projects, result, err
	}
	return nil, result, err
}

/*
 * GET PROJECTS
 */

// ReadDomains returns the list of domains that are available to be scoped to
// based on the X-Auth-Token provided in the request. The structure is the same
// as listing domains.
func (api *IdentityV3API) ReadDomains() (*[]Domain, *Result, error) {
	output := &struct {
		Domains *[]Domain `json:"domains,omitempty"`
		Links   *Links    `json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/domains", true, nil, output)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Domains, result, err
	}
	return nil, result, err
}

/*
 * GET SYSTEMS
 */

// ReadSystems returns the list of systems that are available to be scoped to
// based on the X-Auth-Token provided in the request.
func (api *IdentityV3API) ReadSystems() (*[]System, *Result, error) {
	output := &struct {
		Systems *[]System `json:"system,omitempty"`
		Links   *Links    `json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/system", true, nil, output)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Systems, result, err
	}
	return nil, result, err
}

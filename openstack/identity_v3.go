package openstack

import (
	"net/http"
	"os"
	"strings"

	"github.com/dihedron/go-openstack/log"
)

// IdentityV3API represents the identity API ver. 3, providing support for
// authentication, authorization, role and resource management.
// See https://developer.openstack.org/api-ref/identity/v3/
type IdentityV3API struct {
	API
}

// CreateTokenOpts contains the set of parameters and options used to perform an
// authentication (create an authentication token).
type CreateTokenOpts struct {
	NoCatalog        bool `url:"nocatalog,omitempty"`
	Method           string
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

/*
 * CREATE TOKEN
 */

// CreateToken uses the provided parameters to authenticate the client to the
// Keystone server and receive a token.
func (api *IdentityV3API) CreateToken(opts *CreateTokenOpts) (*Token, *Result, error) {

	input := &struct {
		NoCatalog bool            `url:"nocatalog,omitempty" json:"-"`
		Auth      *Authentication `json:"auth,omitempty"`
	}{
		NoCatalog: opts.NoCatalog,
		Auth: &Authentication{
			Identity: &Identity{
				Methods: &[]string{
					opts.Method,
				},
			},
		},
	}

	if opts.Method == "password" {
		if opts.UserID != nil && len(strings.TrimSpace(*opts.UserID)) > 0 {
			input.Auth.Identity.Password = &Password{
				User: &User{
					ID:       opts.UserID,
					Password: opts.UserPassword,
				},
			}
		} else {
			input.Auth.Identity.Password = &Password{
				User: &User{
					Name:     opts.UserName,
					Password: opts.UserPassword,
				},
			}
			if opts.UserDomainID != nil && len(strings.TrimSpace(*opts.UserDomainID)) > 0 {
				input.Auth.Identity.Password.User.Domain = &Domain{
					ID: opts.UserDomainID,
				}
			} else {
				input.Auth.Identity.Password.User.Domain = &Domain{
					Name: opts.UserDomainName,
				}
			}
		}
	} else if opts.Method == "token" {
		if opts.TokenID != nil && len(strings.TrimSpace(*opts.TokenID)) > 0 {
			input.Auth.Identity.Token = &Token{
				ID: opts.TokenID,
			}
		}
	}

	// manage scoped/unscoped token requests
	if opts.ScopeProjectID != nil && len(strings.TrimSpace(*opts.ScopeProjectID)) > 0 {
		input.Auth.Scope = &Scope{
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
		input.Auth.Scope = scope
	} else {
		if opts.ScopeDomainID != nil && len(strings.TrimSpace(*opts.ScopeDomainID)) > 0 {
			input.Auth.Scope = &Scope{
				Domain: &Domain{
					ID: opts.ScopeDomainID,
				},
			}
		} else if opts.ScopeDomainName != nil && len(strings.TrimSpace(*opts.ScopeDomainName)) > 0 {
			input.Auth.Scope = &Scope{
				Domain: &Domain{
					Name: opts.ScopeDomainName,
				},
			}
		} else if opts.UnscopedToken != nil && *opts.UnscopedToken {
			// all values are null: the request is unscoped
			input.Auth.Scope = String("unscoped")
		}
	}

	log.Debugf("IdentityV3.CreateTokenRequestBuilder: entity in request body is\n%s\n", log.ToJSON(input))

	output := &struct {
		SubjectToken *string `header:"X-Subject-Token" json:"-"`
		Token        *Token  `json:"token,omitempy"`
	}{}

	result, err := api.Invoke(http.MethodPost, "./v3/auth/tokens", false, input, output)
	if output.SubjectToken != nil {
		output.Token.Value = output.SubjectToken
		return output.Token, result, err
	}
	return output.Token, result, err
}

// CreateTokenFromEnv uses the information in the environment to authenticate the
// client to the Keystore server and receive a token.
func (api *IdentityV3API) CreateTokenFromEnv() (*Token, *Result, error) {
	opts := &CreateTokenOpts{
		Method:         "password",
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
func (api *IdentityV3API) ReadToken(opts *ReadTokenOpts) (bool, *Result, error) {
	output := &struct {
		Token        *Token  `json:"token,omitempy"`
		SubjectToken *string `header:"X-Subject-Token" json:"-"`
	}{}

	result, err := api.Invoke(http.MethodPost, "./v3/auth/tokens", true, opts, output)
	if result.Code == 200 {
		return true, result, err
	}
	log.Debugf("IdentityV3.ReadToken: header is %q\n", output.SubjectToken)

	return false, result, err
}

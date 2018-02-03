package openstack

import (
	"net/http"
	"os"
	"strings"

	"github.com/dghubble/sling"
	"github.com/dihedron/go-openstack/log"
)

// IdentityV3API represents the identity API ver. 3, providing support
// for authentication, authorization, role and resource management.
// See https://developer.openstack.org/api-ref/identity/v3/
type IdentityV3API struct {
	API
}

// CreateTokenOpts contains the set of parameters and options used to
// perform an authentication (create an authentication token).
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
func (api *IdentityV3API) CreateToken(opts *CreateTokenOpts) (string, *Token, *Result, error) {

	wrapper := &struct {
		Token        *Token  `json:"token,omitempy"`
		SubjectToken *string `header:"X-Subject-Token"`
	}{}

	result, _, err := api.Invoke(http.MethodPost, "./v3/auth/tokens", false, opts, wrapper, CreateTokenRequestBuilder, nil)
	if wrapper.SubjectToken != nil {
		return *wrapper.SubjectToken, wrapper.Token, result, err
	}
	return "", wrapper.Token, result, err
}

// CreateTokenFromEnv uses the information in the environment to authenticate the
// client to the Keystore server and receive a token.
func (api *IdentityV3API) CreateTokenFromEnv() (string, *Token, *Result, error) {
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

// CreateTokenRequestBuilder is a specialised version of a RequestBuilder
// specifically designed to prepare the request entity for create toke requests
// under a set of different circumstances including scoped/unscoped authentication
// and password- or token-based requests.
func CreateTokenRequestBuilder(sling *sling.Sling, opts interface{}) (request *http.Request, err error) {

	sling = DefaultRequestQueryBuilder(sling, opts)
	sling = DefaultRequestHeadersBuilder(sling, opts)

	info := opts.(*CreateTokenOpts)

	entity := &struct {
		Auth *Authentication `json:"auth,omitempty"`
	}{
		Auth: &Authentication{
			Identity: &Identity{
				Methods: &[]string{
					info.Method,
				},
			},
		},
	}

	if info.Method == "password" {
		if info.UserID != nil && len(strings.TrimSpace(*info.UserID)) > 0 {
			entity.Auth.Identity.Password = &Password{
				User: &User{
					ID:       info.UserID,
					Password: info.UserPassword,
				},
			}
		} else {
			entity.Auth.Identity.Password = &Password{
				User: &User{
					Name:     info.UserName,
					Password: info.UserPassword,
				},
			}
			if info.UserDomainID != nil && len(strings.TrimSpace(*info.UserDomainID)) > 0 {
				entity.Auth.Identity.Password.User.Domain = &Domain{
					ID: info.UserDomainID,
				}
			} else {
				entity.Auth.Identity.Password.User.Domain = &Domain{
					Name: info.UserDomainName,
				}
			}
		}
	} else if info.Method == "token" {
		if info.TokenID != nil && len(strings.TrimSpace(*info.TokenID)) > 0 {
			entity.Auth.Identity.Token = &Token{
				ID: info.TokenID,
			}
		}
	}

	// manage scoped/unscoped token requests
	if info.ScopeProjectID != nil && len(strings.TrimSpace(*info.ScopeProjectID)) > 0 {
		entity.Auth.Scope = &Scope{
			Project: &Project{
				ID: info.ScopeProjectID,
			},
		}
	} else if info.ScopeProjectName != nil && len(strings.TrimSpace(*info.ScopeProjectName)) > 0 {
		scope := &Scope{
			Project: &Project{
				Name: info.ScopeProjectName,
			},
		}

		if info.ScopeDomainID != nil && len(strings.TrimSpace(*info.ScopeDomainID)) > 0 {
			scope.Project.Domain = &Domain{
				ID: info.ScopeDomainID,
			}
		} else {
			scope.Project.Domain = &Domain{
				Name: info.ScopeDomainName,
			}
		}
		entity.Auth.Scope = scope
	} else {
		if info.ScopeDomainID != nil && len(strings.TrimSpace(*info.ScopeDomainID)) > 0 {
			entity.Auth.Scope = &Scope{
				Domain: &Domain{
					ID: info.ScopeDomainID,
				},
			}
		} else if info.ScopeDomainName != nil && len(strings.TrimSpace(*info.ScopeDomainName)) > 0 {
			entity.Auth.Scope = &Scope{
				Domain: &Domain{
					Name: info.ScopeDomainName,
				},
			}
		} else if info.UnscopedToken != nil && *info.UnscopedToken {
			// all values are null: the request is unscoped
			entity.Auth.Scope = String("unscoped")
		}
	}

	log.Debugf("IdentityV3.CreateTokenRequestBuilder: entity in request body is\n%s\n", log.ToJSON(entity))

	return sling.BodyJSON(entity).Request()
}

/*
 * VALIDATE AND GET TOKEN INFO
 */

// ReadTokenOpts contains the set of parameters and options used to
// perform the valudation of a token on the Identity server.
type ReadTokenOpts struct {
	NoCatalog    bool   `url:"nocatalog,omitempty"`
	AllowExpired bool   `url:"allow_expired,omitempty"`
	SubjectToken string `header:"X-Subject-Token"`
}

// ReadToken uses the provided parameters to read the given token and retrieve
// information about it from the Identity server; this API requires a valid admin
// token.
func (api *IdentityV3API) ReadToken(opts *ReadTokenOpts) (bool, *Result, error) {
	wrapper := &struct {
		Token        *Token  `json:"token,omitempy"`
		SubjectToken *string `header:"X-Subject-Token"`
	}{}

	result, _, err := api.Invoke(http.MethodPost, "./v3/auth/tokens", true, opts, wrapper, nil, nil)
	if result.Code == 200 {
		return true, result, err
	}
	log.Debugf("IdentityV3.ReadToken: header is %q\n", wrapper.SubjectToken)

	return false, result, err
}

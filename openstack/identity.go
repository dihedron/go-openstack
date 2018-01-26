package openstack

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dghubble/sling"
	"github.com/dihedron/go-openstack/log"
)

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

// LoginOpts is a subset of CreateTokenOpts; it assumes some defaults and is
// used when invoking the Client's Login method.
type LoginOpts struct {
	UserName         *string
	UserDomainName   *string
	UserPassword     *string
	ScopeProjectName *string
	ScopeDomainName  *string
	UnscopedLogin    *bool
	// TokenID????
}

/*
 * LOGIN
 */

// Login performs a login using the given options and sets the returned token
// inside the Client, so it can be automatically set as a request header in the
// following calls; moreover this method parses the catalog and initialises all
// the other available service APIs using the retrieved endpoints.
func (c Client) Login(opts *LoginOpts) error {
	opts2 := &CreateTokenOpts{
		NoCatalog:        false,
		Method:           "password",
		UserName:         opts.UserName,
		UserDomainName:   opts.UserDomainName,
		UserPassword:     opts.UserPassword,
		ScopeProjectName: opts.ScopeProjectName,
		ScopeDomainName:  opts.ScopeDomainName,
		UnscopedToken:    opts.UnscopedLogin,
	}

	token, info, _, err := c.Identity.CreateToken(opts2)
	if err != nil {
		log.Errorf("Client.Login: login failed: %v", err)
		return err
	}

	c.authToken = String(token)

	if info.Catalog == nil {
		log.Errorf("Client.Login: no catalog info available")
		return fmt.Errorf("no catalog information available from identity service")
	}

	for _, service := range *info.Catalog {
		log.Debugf("Client.Login: initialising service %s (type: %s, id: %s)", *service.Name, *service.Type, *service.ID)
	}

	return nil
}

/*
 * CREATE TOKEN
 */

// CreateToken uses the provided parameters to authenticate the client to the
// Keystone server and receive a token.
func (api IdentityAPI) CreateToken(opts *CreateTokenOpts) (string, *Token, *Result, error) {

	type wrapper struct {
		Token *Token `json:"token,omitempy"`
	}

	wrapped := &wrapper{}
	headers, result, err := api.Invoke(http.MethodPost, "/identity/v3/auth/tokens", opts, []string{"X-Subject-Token"}, wrapped, CreateTokenRequestBuilder, nil)
	if tokens, ok := headers["X-Subject-Token"]; ok {
		if len(tokens) > 0 {
			return headers["X-Subject-Token"][0], wrapped.Token, result, err
		}
	}
	return "", wrapped.Token, result, err
	//return "", token, result, err
}

// CreateTokenFromEnv uses the information in the environment to authenticate the
// client to the Keystore server and receive a token.
func (api IdentityAPI) CreateTokenFromEnv() (string, *Token, *Result, error) {
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

	log.Debugf("Identity.CreateTokenRequestBuilder: entity in request body is\n%s\n", log.ToJSON(entity))

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
func (api IdentityAPI) ReadToken(opts *ReadTokenOpts) (*Token, *Result, error) {
	return nil, nil, nil
}

// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"fmt"

	"github.com/dihedron/go-openstack/log"
)

// Authenticator uses the services of an IdentityAPI to get the first access
// token and after that the complete catalog of available APIs (services
// with their endpoints, interface types and regions). it provides some degree
// of abstraction over the IdentityAPI, e.g. it checks when a token is about
// to expire and makes sure it is reissued before an API call; moreover, it
// populates the client's internal references to all other available services
// as per the catalog returned by the IdentityAPI.
type Authenticator struct {
	// Identity is a reference to the identity service; at the resent version only
	// Identity v3 is supported; future versions should be based on an interface{}
	// and support both v2 and v3 API versions.
	Identity *IdentityV3API // TODO: switch to interface{}

	// Token is the token released at login by the Identity service; it
	// must be set in all authenticated API requests to gain access to protected
	// resources.
	TokenValue *string

	// TokenInfo contains all the information about the current token, as reported
	// by the adentity service hen the token is issued; it can be used to check for
	// expiration.
	TokenInfo *Token
}

// LoginOpts is a subset of CreateTokenOpts; it assumes some defaults and is
// used when invoking the AuthenticationAPI's Login method; it can be filled
// with values taken from the process enviroment.
type LoginOpts struct {
	UserName         *string
	UserDomainName   *string
	UserPassword     *string
	ScopeProjectName *string
	ScopeDomainName  *string
	UnscopedLogin    *bool
	// TokenID????
}

// Login performs a logon using the given options and sets the returned token
// as Token Value, in order for it to be available to be be automatically set
// as a request header ("X-Auth-Token") in the following procteted API calls;
// moreover this method parses the catalog and initialises all the other available
// service API references using the information about services, their versions and
// available endpoints.
func (auth *Authenticator) Login(opts *LoginOpts) error {
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

	token, info, _, err := auth.Identity.CreateToken(opts2)
	if err != nil {
		log.Errorf("AuthenticationAPI.Login: login failed: %v", err)
		return err
	}

	log.Debugf("AuthenticationAPI.Login: token value is %q, token info is:\n%s\n", token, log.ToJSON(info))

	auth.TokenValue = String(token)
	auth.TokenInfo = info

	if info.Catalog == nil {
		log.Errorf("AuthenticationAPI.Login: no catalog info available")
		return fmt.Errorf("no catalog information available from identity service")
	}

	for _, service := range *info.Catalog {
		log.Debugf("AuthenticationAPI.Login: initialising service %s (type: %s, id: %s)", *service.Name, *service.Type, *service.ID)
	}

	return nil
}

// Logout invalidates the current authentication token so that all succeding
// API calls will fail as unauthorised.
func (auth *Authenticator) Logout() error {
	if auth.TokenValue != nil {
		log.Debugf("AuthenticationAPI.Logout: invalidating authentication token %s", *auth.TokenValue)
		// TODO: api.DeleteToken()
		auth.TokenValue = nil
		auth.TokenInfo = nil
	}
	return nil
}

// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"strings"
	"sync"
	"time"

	"github.com/dihedron/go-log/log"
)

// Authenticator uses the services of an IdentityAPIV3 to get the first access
// token and, after that, it retrieves the complete catalog of available APIs
// (services with their endpoints, interface types and regions). It provides some
// degree of abstraction over the IdentityAPI, e.g. it checks when a token is
// about to expire and makes sure it is reissued by setting a times that will
// fire a few seconds before the token expiration and unleash a goroutine to
// have it renewed; this mechanism makes sure that an API call can always rely
// on the token being current; moreover, it populates the client's internal
// references to all other available services as per the catalog returned by the
// IdentityAPI.
type Authenticator struct {
	// AuthURL is the URL at which the authentication service can be reached, i.e.
	// the URL of the public Keystone endpoint used for the first authentication.
	AuthURL *string

	// Identity is a reference to the identity service; at the present version
	// only Identity v3 is supported; future versions may support both v2 and v3
	// API versions.
	Identity *IdentityV3API

	// token contains all the information about the current token, as reported
	// by the Identity service when the token is issued; it contains both the
	// token value, which must be set in all authenticated API requests to gain
	// access to protected resources, and the metadata which can be used e.g. to
	// check for expiration.
	token *Token

	// tokenMutex guards TokenValue and TokenInfo against concurrent read and write
	// accesses, e.g. when a token is being reissued by the background goroutine.
	mutex sync.RWMutex

	// tokenTimer is a timer that is set to file a few seconds before the
	// current tokenValue expires (as per the information in tokenInfo); a
	// goroutine is set to listen on it and to have the token reissued by the
	// identity server automatically via the Login() method and the current
	// token value; the scope is unchanged.
	timer *time.Timer
}

// LoginOptions is a subset of Identity V3's CreateToken[*]Options; it assumes
// some defaults and is used when invoking the IdentityV3API's CreateToken method;
// it can be filled with values taken from the process enviroment.
type LoginOptions struct {
	// UserName, UserDomainID and UserDomainName are used for password- and
	// application credential-based authentication.
	UserName       *string
	UserDomainID   *string
	UserDomainName *string

	// ScopeProjectID, ScopeProjectName, ScopeDomainID, ScopeDomainName and
	// UnscopedLogin specify the scope of
	// the requested authentication token; these parameters are in common between
	// password- and token-based logins.
	ScopeProjectID   *string
	ScopeProjectName *string
	ScopeDomainID    *string
	ScopeDomainName  *string
	UnscopedLogin    *bool

	// UserPassword is used for password-based authentication.
	UserPassword *string

	// TokenID is an existing valid token; when this value is not nil, the
	// token-based authentication method is used; the most common case when this
	// happens is to reissue a token that is about to expire.
	TokenID *string

	//
	AppCredentialID *string
	Secret          *string
}

// Login performs a logon using the given options and sets the returned token
// as Token's Value field, in order for it to be available to be be automatically
// set as a request header ("X-Auth-Token") in the following protected API calls;
// moreover this method parses the catalog and initialises all the other available
// service API references using the information about services, their versions and
// available endpoints fo the current token.
func (auth *Authenticator) Login(opts *LoginOptions) error {
	var err error

	log.Debugf("logging in")

	var cto *CreateTokenOptions

	if opts.TokenID != nil && len(strings.TrimSpace(*opts.TokenID)) > 0 {
		cto = &CreateTokenOptions{
			NoCatalog:        false,
			ScopeProjectID:   opts.ScopeProjectID,
			ScopeProjectName: opts.ScopeProjectName,
			ScopeDomainID:    opts.ScopeDomainID,
			ScopeDomainName:  opts.ScopeDomainName,
			UnscopedToken:    opts.UnscopedLogin,
			TokenID:          opts.TokenID,
		}
		log.Debugf("performing token-based authentication (%s)", ZipString(*opts.TokenID, 10))
	} else if opts.UserPassword != nil && len(strings.TrimSpace(*opts.UserPassword)) > 0 {
		cto = &CreateTokenOptions{
			NoCatalog:        false,
			ScopeProjectID:   opts.ScopeProjectID,
			ScopeProjectName: opts.ScopeProjectName,
			ScopeDomainID:    opts.ScopeDomainID,
			ScopeDomainName:  opts.ScopeDomainName,
			UnscopedToken:    opts.UnscopedLogin,
			UserName:         opts.UserName,
			UserDomainName:   opts.UserDomainName,
			UserPassword:     opts.UserPassword,
		}
		log.Debugf("performing password-based authentication (%s\\%s:%s)", *opts.UserDomainName, *opts.UserName, *opts.UserPassword)
	} else if opts.AppCredentialID != nil && len(strings.TrimSpace(*opts.AppCredentialID)) > 0 && opts.Secret != nil && len(strings.TrimSpace(*opts.Secret)) > 0 {
		cto = &CreateTokenOptions{
			NoCatalog:        false,
			ScopeProjectID:   opts.ScopeProjectID,
			ScopeProjectName: opts.ScopeProjectName,
			ScopeDomainID:    opts.ScopeDomainID,
			ScopeDomainName:  opts.ScopeDomainName,
			UnscopedToken:    opts.UnscopedLogin,
			UserName:         opts.UserName,
			UserDomainName:   opts.UserDomainName,
			AppCredentialID:  opts.AppCredentialID,
			Secret:           opts.Secret,
		}
		log.Debugf("performing app-credential-based authentication (%s:%s)", *opts.AppCredentialID, *opts.Secret)
	}

	token, _, err := auth.Identity.CreateToken(cto)
	if err != nil {
		log.Errorf("login failed: %v", err)
		return err
	}

	log.Debugf("token value is %s, token info is:\n%s\n", *token.Value, log.ToJSON(token))

	// now store that info inside the current authenticator and start the
	// background goroutine that will automatically reissue the token when it
	// is about to expire.
	auth.mutex.Lock()
	defer auth.mutex.Unlock()
	auth.token = token

	// TODO: re-enable
	if auth.token.ExpiresAt != nil {
		log.Debugf("setting timer for token refresh")
		if expiryDate, err := time.Parse(ISO8601, *auth.token.ExpiresAt); err == nil {
			when := expiryDate.Sub(time.Now().Add(5 * time.Second))
			log.Debugf("re-authentication timer will fire in %v", when)
			auth.timer = time.NewTimer(5 * time.Second)
			go func(lo *LoginOptions) {
				log.Debugf("starting re-authentication timer...")
				<-auth.timer.C
				log.Debugf("re-authentication timer logging in again...")
				auth.Login(lo)
			}(opts)
		} else {
			log.Errorf("error parsing date: %v", err)
		}
	}

	log.Debugf("done logging in")
	return nil
}

// Logout invalidates the current authentication token so that all succeding
// API calls will fail as unauthorised.
func (auth *Authenticator) Logout() error {
	log.Debugf("logging out by invalidating authentication token")
	token := auth.GetToken()
	if token == nil {
		log.Debugf("already logged out")
		return nil
	}
	value := token.Value
	if value != nil {
		log.Debugf("invalidating authentication token %s", ZipString(*value, 16))
		auth.mutex.Lock()
		defer auth.mutex.Unlock()
		if auth.timer != nil {
			log.Debugf("stopping timer")
			if !auth.timer.Stop() {
				// drain the timer, as per the docs
				<-auth.timer.C
			}
		}
		if auth.GetToken().Value != nil {
			opts := &DeleteTokenOptions{
				SubjectToken: *(auth.GetToken().Value),
			}
			auth.Identity.DeleteToken(opts)
		}
		auth.token.Value = nil
		auth.token = nil
	}
	log.Debugf("logged out")
	return nil
}

// GetToken returns the current Token information; this includes both data (the
// "Value" field) and metadata (such as its expiration date and the set of
// services and endpoints associated with the token).
func (auth *Authenticator) GetToken() *Token {
	auth.mutex.RLock()
	defer auth.mutex.RUnlock()
	return auth.token
}

// GetCatalog returns the set of services and endpoints associated with the
// current token.
func (auth *Authenticator) GetCatalog() *[]Service {
	auth.mutex.RLock()
	defer auth.mutex.RUnlock()
	if auth.token != nil {
		return auth.token.Catalog
	}
	return nil
}

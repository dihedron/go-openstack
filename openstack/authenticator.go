// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"strings"
	"sync"
	"time"

	"github.com/dihedron/go-openstack/log"
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
	// AuthURL is the URL at which the authentication service can be reaced, i.e.
	// the URL of the public Keystone endpoint used for the first authentication.
	AuthURL *string

	// Identity is a reference to the identity service; at the present version
	// only Identity v3 is supported; future versions may support both v2 and v3
	// API versions.
	Identity *IdentityV3API

	// tokenValue is the token released at login by the Identity service; it
	// must be set in all authenticated API requests to gain access to protected
	// resources.
	tokenValue *string

	// tokenInfo contains all the information about the current token, as reported
	// by the Identity service hen the token is issued; it can be used to check for
	// expiration.
	tokenInfo *Token

	// tokenMutex guards TokenValue and TokenInfo against concurrent read and write
	// accesses, e.g. when a token is being reissued by the background goroutine.
	tokenMutex sync.RWMutex

	// tokenTimer is a timer that is set to file a few seconds before the
	// current tokenValue expires (as per the information in tokenInfo); a
	// goroutine is set to listen on it and to have the token reissued by the
	// identity server automatically via the Login() method and the current
	// token value; the scope is unchanged.
	tokenTimer *time.Timer
}

// LoginOpts is a subset of CreateTokenOpts; it assumes some defaults and is
// used when invoking the IdentityAPIV's CreateToken method; it can be filled
// with values taken from the process enviroment.
type LoginOpts struct {
	// UserName, UserDomainName and UserPassword are used for password-based
	// authentication; this is not the preferred method and is not employed if
	// TokenID is not nil.
	UserName       *string
	UserDomainName *string
	UserPassword   *string

	// TokenID is an existing valid token; when this value is not nil, the
	// token-based authentication method is used; the most common case when this
	// happens is to reissue a token that is about to expire.
	TokenID *string

	// ScopeProjectName, ScopeDomainName and UnscopedLogin specify the scope of
	// the requested authentication token; these parameters are in common between
	// password- and token-based logins.
	ScopeProjectName *string
	ScopeDomainName  *string
	UnscopedLogin    *bool
}

// Login performs a logon using the given options and sets the returned token
// as Token Value, in order for it to be available to be be automatically set
// as a request header ("X-Auth-Token") in the following procteted API calls;
// moreover this method parses the catalog and initialises all the other available
// service API references using the information about services, their versions and
// available endpoints.
func (auth *Authenticator) Login(opts *LoginOpts) error {
	var err error
	opts2 := &CreateTokenOpts{
		NoCatalog:        false,
		ScopeProjectName: opts.ScopeProjectName,
		ScopeDomainName:  opts.ScopeDomainName,
		UnscopedToken:    opts.UnscopedLogin,
	}

	if opts.TokenID != nil && len(strings.TrimSpace(*opts.TokenID)) > 0 {
		opts2.Method = "token"
		opts2.TokenID = opts.TokenID
		log.Debugf("Authenticator.Login: performing token-based authentication (%s)", ZipString(*opts.TokenID, 10))
	} else {
		opts2.Method = "password"
		opts2.UserName = opts.UserName
		opts2.UserDomainName = opts.UserDomainName
		opts2.UserPassword = opts.UserPassword
		log.Debugf("Authenticator.Login: performing password-based authentication (%s\\%s:%s)", opts.UserDomainName, opts.UserName, opts.UserPassword)
	}

	token, info, _, err := auth.Identity.CreateToken(opts2)
	if err != nil {
		log.Errorf("Authenticator.Login: login failed: %v", err)
		return err
	}

	log.Debugf("Authenticator.Login: token value is %q, token info is:\n%s\n", token, log.ToJSON(info))

	// now store that info inside the current authenticator and start the
	// background goroutine that will automatically reissue the token when it
	// is about to expire.
	auth.tokenMutex.Lock()
	defer auth.tokenMutex.Unlock()
	auth.tokenValue = String(token)
	auth.tokenInfo = info
	if auth.tokenInfo.ExpiresAt != nil {
		log.Debugf("Authenticator.Login: setting timer for token refresh")
		if expiryDate, err := time.Parse(ISO8601, *auth.tokenInfo.ExpiresAt); err == nil {
			when := expiryDate.Sub(time.Now().Add(30 * time.Second))
			log.Debugf("Authenticator.Login: timer will fire in %v", when)
			auth.tokenTimer = time.NewTimer(when)
			//auth.tokenTimer = time.NewTimer(5 * time.Second)
			opts3 := &LoginOpts{
				TokenID:          auth.tokenValue,
				ScopeProjectName: opts.ScopeDomainName,
				ScopeDomainName:  opts.ScopeDomainName,
				UnscopedLogin:    opts.UnscopedLogin,
			}
			go func() {
				<-auth.tokenTimer.C
				auth.Login(opts3)
			}()
		} else {
			log.Errorf("Authenticator.Login: error parsing date: %v", err)
		}
	}

	return nil
}

func (auth *Authenticator) GetTokenValue() *string {
	auth.tokenMutex.RLock()
	defer auth.tokenMutex.RUnlock()
	return auth.tokenValue
}

func (auth *Authenticator) GetTokenInfo() *Token {
	auth.tokenMutex.RLock()
	defer auth.tokenMutex.RUnlock()
	return auth.tokenInfo
}

func (auth *Authenticator) GetCatalog() *[]Service {
	auth.tokenMutex.RLock()
	defer auth.tokenMutex.RUnlock()
	if auth.tokenInfo != nil {
		return auth.tokenInfo.Catalog
	}
	return nil
}

// Logout invalidates the current authentication token so that all succeding
// API calls will fail as unauthorised.
func (auth *Authenticator) Logout() error {
	value := auth.GetTokenValue()
	if value != nil {
		log.Debugf("Authenticator.Logout: invalidating authentication token %s", value)
		auth.tokenMutex.Lock()
		defer auth.tokenMutex.Unlock()
		if auth.tokenTimer != nil {
			auth.tokenTimer.Stop()
		}
		// TODO: api.DeleteToken()
		auth.tokenValue = nil
		auth.tokenInfo = nil
	}
	return nil
}

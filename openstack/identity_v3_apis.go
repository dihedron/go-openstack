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

	//log.Debugf("entity in request body is\n%s\n", log.ToJSON(input))

	output := &struct {
		SubjectToken *string `parameter:"-" header:"X-Subject-Token" json:"-"`
		Token        *Token  `parameter:"-" header:"-" json:"token,omitempy"`
	}{}

	failure := String("")
	result, err := api.Invoke(http.MethodPost, "./v3/auth/tokens", opts.Authenticated, StatusCodeIn(201), input, output, failure)
	log.Debugf("result is %q (%v, %t)", result, err, result.OK)
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

// RetrieveTokenOptions contains the set of parameters and options used to perform the
// validation of a token on the Identity server.
type RetrieveTokenOptions struct {
	NoCatalog    *bool  `parameter:"nocatalog,omitempty" header:"-" json:"-"`
	AllowExpired *bool  `parameter:"allow_expired,omitempty" header:"-" json:"-"`
	SubjectToken string `parameter:"-" header:"X-Subject-Token" json:"-"`
}

// RetrieveToken uses the provided parameters to read the given token and retrieve
// information about it from the Identity server; this API requires a valid admin
// token.
func (api *IdentityV3API) RetrieveToken(opts *RetrieveTokenOptions) (*Token, *Result, error) {
	output := &struct {
		Token        *Token  `parameter:"-" header:"-" json:"token,omitempy"`
		SubjectToken *string `parameter:"-" header:"X-Subject-Token" json:"-"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/tokens", true, StatusCodeIn(200), opts, output, nil)
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
	result, err := api.Invoke(http.MethodHead, "./v3/auth/tokens", true, StatusCodeIn(200), opts, nil, nil)
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
	// TODO: Docs state it's a 201 (created) upon success, weird! Shouldn't it be 204?
	result, err := api.Invoke(http.MethodDelete, "./v3/auth/tokens", true, StatusCodeIn(201), opts, nil, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 || result.Code == 204 {
		return true, result, err
	}
	return false, result, err
}

/*
 * RETRIEVE CATALOG
 */

// RetrieveCatalog retrieves the catalog associated with the given authorisation
// token; the catalog is returned even if the token was issued withouth a catalog
// (?nocataog=true).
func (api *IdentityV3API) RetrieveCatalog() (*[]Service, *Result, error) {
	output := &struct {
		Catalog *[]Service `header:"-" json:"catalog,omitempty"`
		Links   *Links     `header:"-" json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/auth/catalog", true, StatusCodeIn(200), nil, output, nil)
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

	result, err := api.Invoke(http.MethodGet, "./v3/auth/projects", true, StatusCodeIn(200), nil, output, nil)
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

	result, err := api.Invoke(http.MethodGet, "./v3/auth/domains", true, StatusCodeIn(200), nil, output, nil)
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

	result, err := api.Invoke(http.MethodGet, "./v3/auth/system", true, StatusCodeIn(200), nil, output, nil)
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
// of registered users (see https://developer.openstack.org/api-ref/identity/v3/#list-users)
type ListUsersOptions struct {
	DomainID           *string     `parameter:"domain_id,omitempty" header:"-" json:"-"`
	Enabled            *bool       `parameter:"enabled,omitempty" header:"-" json:"-"`
	IdentityProviderID *string     `parameter:"idp_id,omitempty" header:"-" json:"-"`
	Name               *string     `parameter:"name,omitempty" header:"-" json:"-"`
	PasswordExpiresAt  *TimeFilter `parameter:"password_expires_at,omitempty" header:"-" json:"-"`
	ProtocolID         *string     `parameter:"protocol_id,omitempty" header:"-" json:"-"`
	UniqueID           *string     `parameter:"unique_id,omitempty" header:"-" json:"-"`
}

// ListUsers returns the list of users on the system (see also
// https://developer.openstack.org/api-ref/identity/v3/#list-users)
func (api *IdentityV3API) ListUsers(opts *ListUsersOptions) (*[]User, *Result, error) {
	output := &struct {
		Users *[]User `header:"-" json:"users,omitempty"`
		Links *Links  `header:"-" json:"links,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodGet, "./v3/users", true, StatusCodeIn(200), opts, output, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Users, result, err
	}
	return nil, result, err
}

/*
 * CREATE USER
 */

// CreateUserOptions provides all the options available for creating a new user
// (see https://developer.openstack.org/api-ref/identity/v3/#create-user).
type CreateUserOptions struct {
	User *User `parameter:"-" header:"-" json:"user"`
}

// CreateUser creates a new user; for implementation details see also
// https://developer.openstack.org/api-ref/identity/v3/#create-user.
func (api *IdentityV3API) CreateUser(opts *CreateUserOptions) (*User, *Result, error) {
	output := &struct {
		User *User `header:"-" json:"user,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodPost, "./v3/users", true, StatusCodeIn(201), opts, output, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 201 {
		return output.User, result, err
	}
	return nil, result, err
}

/*
 * RETRIEVE USER
 */

// RetrieveUser retrieves the information about the user identified by the given
// user id; see also https://developer.openstack.org/api-ref/identity/v3/#show-user-details
func (api *IdentityV3API) RetrieveUser(userid string) (*User, *Result, error) {
	input := &struct {
		UserID string `parameter:"-" header:"-" variable:"userid" json:"-"`
	}{
		UserID: userid,
	}
	output := &struct {
		User *User `header:"-" json:"user,omitempty"`
	}{}
	result, err := api.Invoke(http.MethodGet, "./v3/users/{userid}", true, StatusCodeIn(200), input, output, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.User, result, err
	}
	return nil, result, err
}

/*
 * UPDATE USER
 */

// UpdateUserOptions provides all the options available for updating an existing user
// (see https://developer.openstack.org/api-ref/identity/v3/#update-user).
type UpdateUserOptions struct {
	UserID string `parameter:"-" header:"-" variable:"userid" json:"-"`
	User   *User  `parameter:"-" header:"-" json:"user"`
}

// UpdateUser updates an existing user; for implementation details see also
// https://developer.openstack.org/api-ref/identity/v3/#update-user.
func (api *IdentityV3API) UpdateUser(opts *UpdateUserOptions) (*User, *Result, error) {
	output := &struct {
		User *User `header:"-" json:"user,omitempty"`
	}{}

	result, err := api.Invoke(http.MethodPatch, "./v3/users/{userid}", true, StatusCodeIn(200), opts, output, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.User, result, err
	}
	return nil, result, err
}

/*
 * DELETE USER
 */

// DeleteUser removes the user identified by the given user id; see also
// https://developer.openstack.org/api-ref/identity/v3/#show-user-details
func (api *IdentityV3API) DeleteUser(userid string) (bool, *Result, error) {
	input := &struct {
		UserID string `parameter:"-" header:"-" variable:"userid" json:"-"`
	}{
		UserID: userid,
	}
	// output := &struct {
	// 	User *User `header:"-" json:"user,omitempty"`
	// }{}
	result, err := api.Invoke(http.MethodDelete, "./v3/users/{userid}", true, StatusCodeIn(200), input, nil, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 204 {
		return true, result, err
	}
	return false, result, err
}

/*
 * LIST USER GROUPS
 */

// ListUserGroups list groups to which a user belongs; see also
// https://developer.openstack.org/api-ref/identity/v3/#list-groups
func (api *IdentityV3API) ListUserGroups(userid string) (*[]Group, *Result, error) {
	input := &struct {
		UserID string `parameter:"-" header:"-" variable:"userid" json:"-"`
	}{
		UserID: userid,
	}
	output := &struct {
		Groups *[]Group `header:"-" json:"groups,omitempty"`
		Links  *Links   `header:"-" json:"links,omitempty"`
	}{}
	result, err := api.Invoke(http.MethodGet, "./v3/users/{userid}/groups", true, StatusCodeIn(200), input, output, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Groups, result, err
	}
	return nil, result, err
}

/*
 * LIST USER PROJECTS
 */

// ListUserProjects list projects for the given user; see also
// https://developer.openstack.org/api-ref/identity/v3/#list-projects-for-user
func (api *IdentityV3API) ListUserProjects(userid string) (*[]Project, *Result, error) {
	input := &struct {
		UserID string `parameter:"-" header:"-" variable:"userid" json:"-"`
	}{
		UserID: userid,
	}
	output := &struct {
		Projects *[]Project `header:"-" json:"projects,omitempty"`
		Links    *Links     `header:"-" json:"links,omitempty"`
	}{}
	result, err := api.Invoke(http.MethodGet, "./v3/users/{userid}/projects", true, StatusCodeIn(200), input, output, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.Projects, result, err
	}
	return nil, result, err
}

/*
 * CHANGE USER PASSWORD
 */

// ChangeUserPassword changes the password for a user; see also
// https://developer.openstack.org/api-ref/identity/v3/#change-password-for-user
func (api *IdentityV3API) ChangeUserPassword(userid, oldPassword, newPassword string) (bool, *Result, error) {
	input := &struct {
		UserID string `parameter:"-" header:"-" variable:"userid" json:"-"`
		User   *User  `parameter:"-" header:"-" variable:"-" json:"user,omitmepty"`
	}{
		UserID: userid,
		User: &User{
			Password:    String(newPassword),
			OldPassword: String(oldPassword),
		},
	}
	// note: this call does not require authentication
	result, err := api.Invoke(http.MethodPost, "./v3/users/{userid}/password", false, StatusCodeIn(201), input, nil, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 204 {
		return true, result, err
	}
	return false, result, err
}

/*
 * CREATE APPLICATION CREDENTIAL FOR USER
 */

// CreateUserAppCredentialOptions is the set of options used to create an
// application credential for the given user.
type CreateUserAppCredentialOptions struct {
	UserID        string         `parameter:"-" header:"-" variable:"userid" json:"-"`
	AppCredential *AppCredential `parameter:"-" header:"-" variable:"-" json:"application_credential"`
}

// CreateUserAppCredential creates an application credential for a user on the
// project to which they are currently scoped; see also
// https://developer.openstack.org/api-ref/identity/v3/#create-application-credential
func (api *IdentityV3API) CreateUserAppCredential(opts *CreateUserAppCredentialOptions) (*AppCredential, *Result, error) {

	output := &struct {
		AppCredential *AppCredential `header:"-" json:"projects,omitempty"`
		Links         *Links         `header:"-" json:"links,omitempty"`
	}{}
	result, err := api.Invoke(http.MethodPost, "./v3/users/{userid}/application_credentials", true, StatusCodeIn(201), opts, output, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 201 {
		return output.AppCredential, result, err
	}
	return nil, result, err
}

/*
 * LIST USER APPLICATION CREDENTIALS
 */

// ListUserAppCredentials list all application credentials for a user; see also
// https://developer.openstack.org/api-ref/identity/v3/#list-application-credentials
func (api *IdentityV3API) ListUserAppCredentials(userid string) (*[]AppCredential, *Result, error) {
	input := &struct {
		UserID string `parameter:"-" header:"-" variable:"userid" json:"-"`
	}{
		UserID: userid,
	}
	output := &struct {
		AppCredentials *[]AppCredential `header:"-" json:"application_credentials,omitempty"`
		Links          *Links           `header:"-" json:"links,omitempty"`
	}{}
	result, err := api.Invoke(http.MethodGet, "./v3/users/{userid}/application_credentials", true, StatusCodeIn(200), input, output, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.AppCredentials, result, err
	}
	return nil, result, err
}

/*
 * RETRIEVE USER APPLICATION CREDENTIAL
 */

// RetrieveUserAppCredential show details of an application credential; see also
// https://developer.openstack.org/api-ref/identity/v3/#show-application-credential-details
func (api *IdentityV3API) RetrieveUserAppCredential(userid, appcredid string) (*AppCredential, *Result, error) {
	input := &struct {
		UserID          string `parameter:"-" header:"-" variable:"userid" json:"-"`
		AppCredentialID string `parameter:"-" header:"-" variable:"appcredid" json:"-"`
	}{
		UserID:          userid,
		AppCredentialID: appcredid,
	}
	output := &struct {
		AppCredential *AppCredential `header:"-" json:"application_credential,omitempty"`
	}{}
	result, err := api.Invoke(http.MethodGet, "./v3/users/{userid}/application_credentials/{appcredid}", true, StatusCodeIn(200), input, output, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 200 {
		return output.AppCredential, result, err
	}
	return nil, result, err
}

/*
 * DELETE USER APPLICATION CREDENTIAL
 */

// DeleteUserAppCredential removes an application credential; see also
// https://developer.openstack.org/api-ref/identity/v3/#delete-application-credential
func (api *IdentityV3API) DeleteUserAppCredential(userid, appcredid string) (bool, *Result, error) {
	input := &struct {
		UserID          string `parameter:"-" header:"-" variable:"userid" json:"-"`
		AppCredentialID string `parameter:"-" header:"-" variable:"appcredid" json:"-"`
	}{
		UserID:          userid,
		AppCredentialID: appcredid,
	}
	result, err := api.Invoke(http.MethodDelete, "./v3/users/{userid}/application_credentials/{appcredid}", true, StatusCodeIn(200), input, nil, nil)
	log.Debugf("result is %v (%v)", result, err)
	if result.Code == 204 {
		return true, result, err
	}
	return false, result, err
}

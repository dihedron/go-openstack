// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/dihedron/go-log"
	"github.com/dihedron/go-openstack/openstack"
)

// OpenStack represents the OpenStack client.
type OpenStack struct {
	client *openstack.Client
}

func newClient() *OpenStack {
	endpoint := os.Getenv("OS_AUTH_URL")
	if endpoint == "" {
		if len(os.Args) >= 2 {
			endpoint = os.Args[1]
		} else {
			endpoint = "http://192.168.56.101/identity/" // my shiny devstack :-)
		}
	}
	client := &OpenStack{
		client: openstack.NewDefaultClient(endpoint),
	}
	return client
}

func (os *OpenStack) close() error {
	if os.client != nil {
		return os.client.Close()
	}
	return nil
}

func (os *OpenStack) doScopedLoginTo(project string) {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                         SCOPED LOGIN                              |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts := &openstack.LoginOptions{
		UserName:         openstack.String("admin"),
		UserDomainName:   openstack.String("Default"),
		UserPassword:     openstack.String("password"),
		ScopeProjectName: openstack.String(project),
		ScopeDomainName:  openstack.String("Default"),
	}
	os.client.Connect(opts)
}

func (os *OpenStack) doUnscopedLogin() {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                        UNSCOPED LOGIN                             |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts := &openstack.LoginOptions{
		UserName:       openstack.String("admin"),
		UserDomainName: openstack.String("Default"),
		UserPassword:   openstack.String("password"),
		UnscopedLogin:  openstack.Bool(true),
	}
	os.client.Connect(opts)
}

func (os *OpenStack) createToken() *openstack.Token {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                         CREATE TOKEN                              |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts := &openstack.CreateTokenOptions{
		NoCatalog:        openstack.Bool(true),
		Authenticated:    true,
		ScopeProjectName: openstack.String("admin"),
		ScopeDomainName:  openstack.String("Default"),
		UserName:         openstack.String("admin"),
		UserDomainName:   openstack.String("Default"),
		UserPassword:     openstack.String("password"),
	}

	token, result, err := os.client.IdentityV3().CreateToken(opts)
	log.Debugf("token is %q\n", *token.Value)
	log.Debugf("token is\n%s\n", log.ToJSON(token))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return token
}

func (os *OpenStack) readToken(tokenValue string) *openstack.Token {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          READ TOKEN                               |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts := &openstack.RetrieveTokenOptions{
		AllowExpired: openstack.Bool(true),
		NoCatalog:    openstack.Bool(false),
		SubjectToken: tokenValue,
	}
	token, result, err := os.client.IdentityV3().RetrieveToken(opts)
	log.Debugf("token is %q\n", *token.Value)
	log.Debugf("token is\n%s\n", log.ToJSON(token))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return token
}

func (os *OpenStack) checkToken(tokenValue string) bool {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          CHECK TOKEN                              |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts := &openstack.CheckTokenOptions{
		AllowExpired: openstack.Bool(true),
		SubjectToken: tokenValue,
	}
	ok, result, err := os.client.IdentityV3().CheckToken(opts)
	log.Debugf("token valid: %t\n", ok)
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return ok
}

func (os *OpenStack) deleteToken(tokenValue string) bool {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          DELETE TOKEN                             |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts := &openstack.DeleteTokenOptions{
		SubjectToken: tokenValue,
	}
	ok, result, err := os.client.IdentityV3().DeleteToken(opts)
	log.Debugf("token valid: %t\n", ok)
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return ok
}

func (os *OpenStack) createAppCredential() *openstack.Token {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                        APP CREDENTIALS                            |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts := &openstack.CreateTokenOptions{
		//NoCatalog:        openstack.Bool(false),
		Authenticated:    true,
		ScopeProjectName: openstack.String("admin"),
		ScopeDomainName:  openstack.String("Default"),
		UserName:         openstack.String("admin"),
		UserDomainName:   openstack.String("Default"),
		UserPassword:     openstack.String("password"),
	}

	token, result, err := os.client.IdentityV3().CreateToken(opts)
	log.Debugf("token is %q\n", *token.Value)
	log.Debugf("token is\n%s\n", log.ToJSON(token))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return token
}

func (os *OpenStack) getCatalog() *[]openstack.Service {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          GET CATALOG                              |")
	log.Debugf("+-------------------------------------------------------------------+")

	catalog, result, err := os.client.IdentityV3().RetrieveCatalog()
	log.Debugf("catalog is:\n%s\n", log.ToJSON(catalog))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return catalog
}

func (os *OpenStack) listProjects() *[]openstack.Project {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                         LIST PROJECTS                             |")
	log.Debugf("+-------------------------------------------------------------------+")

	projects, result, err := os.client.IdentityV3().ListProjects()
	log.Debugf("projects are:\n%s\n", log.ToJSON(projects))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return projects
}

func (os *OpenStack) listDomains() *[]openstack.Domain {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          LIST DOMAINS                             |")
	log.Debugf("+-------------------------------------------------------------------+")

	domains, result, err := os.client.IdentityV3().ListDomains()
	log.Debugf("domains are:\n%s\n", log.ToJSON(domains))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return domains
}

func (os *OpenStack) listSystems() *[]openstack.System {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          GET SYSTEMS                              |")
	log.Debugf("+-------------------------------------------------------------------+")

	systems, result, err := os.client.IdentityV3().ListSystems()
	log.Debugf("systems are:\n%s\n", log.ToJSON(systems))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return systems
}

func (os *OpenStack) listUsers() *[]openstack.User {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          LIST USERS                               |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts := &openstack.ListUsersOptions{
		Enabled: openstack.Bool(true),
		//Name:    openstack.String("neutron"),
		// PasswordExpiresAt: &openstack.TimeFilter{
		// 	Operator:  openstack.GT,
		// 	Timestamp: time.Now(),
		// },
	}

	//log.Debugf("entity is:\n%s", log.ToJSON(opts))

	log.Debugf("invoking API...")

	users, result, err := os.client.IdentityV3().ListUsers(opts)
	if users != nil {
		for _, user := range *users {
			log.Debugf("user: %s", log.ToJSON(user))
		}
	}
	// log.Debugf("users is %q\n", *token.Value)
	// log.Debugf("token is\n%s\n", log.ToJSON(token))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return users

}

func (os *OpenStack) readUser(userid string) *openstack.User {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          READ USER                                |")
	log.Debugf("+-------------------------------------------------------------------+")

	log.Debugf("invoking API...")

	user, result, err := os.client.IdentityV3().RetrieveUser(userid)
	if user != nil {
		log.Debugf("user: %s", log.ToJSON(user))
	}
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return user

}

func (os *OpenStack) listUserGroups(userid string) *[]openstack.Group {
	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                        LIST USER GROUPS                           |")
	log.Debugf("+-------------------------------------------------------------------+")

	groups, result, err := os.client.IdentityV3().ListUserGroups(userid)
	log.Debugf("groups are:\n%s\n", log.ToJSON(groups))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}
	return groups
}

// https://developer.openstack.org/sdks/python/openstacksdk/users/profile.html#openstack.profile.Profile
func main() {

	log.SetLevel(log.DBG)
	log.SetStream(os.Stdout, true)
	log.SetTimeFormat("15:04:05.000")

	sdk := newClient()
	//defer sdk.close()

	//client.LoadProfileFrom("./my-profile.json")

	sdk.doScopedLoginTo("admin")
	token1 := sdk.createToken()
	token2 := sdk.readToken(*token1.Value)
	if sdk.checkToken(*token2.Value) {
		log.Debugf("token is OK")
	} else {
		log.Debugf("token is KO")
	}
	sdk.listProjects()
	sdk.listDomains()
	// TODO: the following requires queens
	sdk.listSystems()

	sdk.listUsers()

	user := sdk.readUser("a744cae9f0f7490d98e127b851c80857")
	sdk.listUserGroups(*user.ID)

	// token2 is a copy of token1
	// sdk.deleteToken(*token1.Value)

	// time.Sleep(10 * time.Second)
}

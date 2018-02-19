// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"time"

	"github.com/dihedron/go-log/log"
	"github.com/dihedron/go-openstack/openstack"
)

// https://developer.openstack.org/sdks/python/openstacksdk/users/profile.html#openstack.profile.Profile
func main() {

	log.SetLevel(log.DBG)
	log.SetStream(os.Stdout)
	log.SetTimeFormat("15:04:05.000")

	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                             LOGIN                                 |")
	log.Debugf("+-------------------------------------------------------------------+")

	endpoint := os.Getenv("OS_AUTH_URL")
	if endpoint == "" {
		if len(os.Args) >= 2 {
			endpoint = os.Args[1]
		} else {
			endpoint = "http://192.168.56.101/identity/" // my shiny devstack :-)
		}
	}

	opts1 := &openstack.LoginOptions{
		UserName:       openstack.String("admin"),
		UserDomainName: openstack.String("Default"),
		UserPassword:   openstack.String("password"),
		//UnscopedLogin:  openstack.Bool(true),
		//ScopeProjectName: openstack.String("admin"),
		ScopeProjectName: openstack.String("demo"),
		ScopeDomainName:  openstack.String("Default"),
	}

	client := openstack.NewDefaultClient(endpoint)
	//client.LoadProfileFrom("./my-profile.json")
	client.Connect(opts1)
	defer client.Close()

	//time.Sleep(10 * time.Second)

	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                         CREATE TOKEN                              |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts2 := &openstack.CreateTokenOptions{
		NoCatalog:        false,
		Authenticated:    true,
		ScopeProjectName: openstack.String("admin"),
		ScopeDomainName:  openstack.String("Default"),
		UserName:         openstack.String("admin"),
		UserDomainName:   openstack.String("Default"),
		UserPassword:     openstack.String("password"),
	}

	token, result, err := client.IdentityV3().CreateToken(opts2)
	log.Debugf("token is %q\n", *token.Value)
	log.Debugf("token is\n%s\n", log.ToJSON(token))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}

	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          READ TOKEN                               |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts3 := &openstack.ReadTokenOpts{
		AllowExpired: true,
		NoCatalog:    false,
		SubjectToken: *token.Value,
	}
	token, result, err = client.IdentityV3().ReadToken(opts3)
	log.Debugf("token is %q\n", *token.Value)
	log.Debugf("token is\n%s\n", log.ToJSON(token))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}

	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          CHECK TOKEN                              |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts4 := &openstack.CheckTokenOpts{
		AllowExpired: true,
		SubjectToken: *token.Value,
	}
	ok, result, err := client.IdentityV3().CheckToken(opts4)
	log.Debugf("token valid: %t\n", ok)
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}

	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          GET CATALOG                              |")
	log.Debugf("+-------------------------------------------------------------------+")

	catalog, result, err := client.IdentityV3().ReadCatalog()
	log.Debugf("catalog is:\n%s\n", log.ToJSON(catalog))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}

	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                         GET PROJECTS                              |")
	log.Debugf("+-------------------------------------------------------------------+")

	projects, result, err := client.IdentityV3().ReadProjects()
	log.Debugf("projects are:\n%s\n", log.ToJSON(projects))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}

	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          GET DOMAINS                              |")
	log.Debugf("+-------------------------------------------------------------------+")

	domains, result, err := client.IdentityV3().ReadDomains()
	log.Debugf("domains are:\n%s\n", log.ToJSON(domains))
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}

	time.Sleep(10 * time.Second)

	// log.Debugf("+-------------------------------------------------------------------+")
	// log.Debugf("|                          GET SYSTEMS                              |")
	// log.Debugf("+-------------------------------------------------------------------+")

	// systems, result, err := client.IdentityV3().ReadSystems()
	// log.Debugf("systems are:\n%s\n", log.ToJSON(systems))
	// log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	// if err != nil {
	// 	log.Debugf("call resulted in %v\n", err)
	// }

	log.Debugf("+-------------------------------------------------------------------+")
	log.Debugf("|                          DELETE TOKEN                             |")
	log.Debugf("+-------------------------------------------------------------------+")

	opts5 := &openstack.DeleteTokenOptions{
		SubjectToken: *token.Value,
	}
	ok, result, err = client.IdentityV3().DeleteToken(opts5)
	log.Debugf("token valid: %t\n", ok)
	log.Debugf("result is %d (%s)\n", result.Code, result.Status)
	if err != nil {
		log.Debugf("call resulted in %v\n", err)
	}

	//client.Rea

	//client.InitProfile()
	//client.SaveProfileTo("./go-openstack-profile.json")

	// ropts := &openstack.ReadTokenOpts{
	// 	AllowExpired: true,
	// 	NoCatalog:    false,
	// 	SubjectToken: *client.Authenticator.TokenValue,
	// }
	// if ok, _, _ := client.Identity.ReadToken(rops); ok {
	// 	log.Debug("token read")
	// }

	// client.Authenticator.Login(opts)

	//client.Authenticator.Logout()

	// copts := &openstack.CreateTokenOpts{
	// 	/*
	// 		Method: openstack.CreateTokenMethodToken,
	// 		TokenID:         openstack.String("gAAAAABaZgmbPZtoEyuTzJXmggwMAyjLZSiknQJPeR4m1FQaL0dpv1nvvVZvd-B3PORQnRqXrR3OevmRKvMqrXwiam02xElVJXOQHKkExqpTK4kkBnttb-kZRxyS3AJLTLjOr7rxzGP2jw7OwGfOclzNxRIRZF00Ha88ApD0iNFKBczP9PBv4A8"),
	// 		ScopeDomainName: openstack.String("Default"),
	// 	*/

	// 	Method: "password",
	// 	//NoCatalog:      true,
	// 	UserName:       openstack.String("admin"),
	// 	UserDomainName: openstack.String("Default"),
	// 	UserPassword:   openstack.String("password"),
	// 	//ScopeProjectID: openstack.String("0877bbc0712043639e29f026cd56b9c7"),
	// 	/*
	// 		//ScopeProjectName: openstack.String("admin"),
	// 		//ScopeDomainName:  openstack.String("Default"),
	// 		//ScopeProjectName: openstack.String("demo"),
	// 		//ScopeDomainID:    openstack.String("default"),
	// 		//UnscopedToken: openstack.Bool(true),
	// 	*/
	// }
	// token, header, _, _ := client.Identity.CreateToken(copts)
	// log.Debugf("token: %s\ntoken info:\n%s\n", header, log.ToJSON(token))

	// log.Debugf("-----------------------------------------------------\n")

	// ropts := &openstack.ReadTokenOpts{
	// 	SubjectToken: token,
	// }
	// client.Identity.ReadToken(token, ropts)

	// log.Debugf("-----------------------------------------------------\n")

	// vopts := &openstack.CheckTokenOpts{
	// 	SubjectToken: token,
	// }
	// client.Identity.CheckToken(token, vopts)

	// log.Debugf("-----------------------------------------------------\n")

	// token2, _, _ := client.Identity.CreateToken(copts)
	// dopts := &openstack.DeleteTokenOpts{
	// 	SubjectToken: token,
	// }
	// client.Identity.DeleteToken(token2, dopts)
	// client.Identity.CheckToken(token2, vopts)

}

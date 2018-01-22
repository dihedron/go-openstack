// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/dihedron/go-openstack/log"
	"github.com/dihedron/go-openstack/openstack"
)

// https://developer.openstack.org/sdks/python/openstacksdk/users/profile.html#openstack.profile.Profile
func main() {

	endpoint := "http://192.168.56.101" // my shiny devstack :-)
	if len(os.Args) >= 2 {
		endpoint = os.Args[1]
	}

	log.SetLevel(log.DBG)
	log.SetStream(os.Stdout)
	log.SetTimeFormat("15:04:05.000")

	client, _ := openstack.NewDefaultClient(endpoint)

	opts := &openstack.CreateTokenOpts{
		/*
			Method: openstack.CreateTokenMethodToken,
			TokenID:         openstack.String("gAAAAABaZgmbPZtoEyuTzJXmggwMAyjLZSiknQJPeR4m1FQaL0dpv1nvvVZvd-B3PORQnRqXrR3OevmRKvMqrXwiam02xElVJXOQHKkExqpTK4kkBnttb-kZRxyS3AJLTLjOr7rxzGP2jw7OwGfOclzNxRIRZF00Ha88ApD0iNFKBczP9PBv4A8"),
			ScopeDomainName: openstack.String("Default"),
		*/

		Method: openstack.CreateTokenMethodPassword,
		//NoCatalog:      true,
		UserName:       openstack.String("admin"),
		UserDomainName: openstack.String("Default"),
		UserPassword:   openstack.String("password"),
		//ScopeProjectID: openstack.String("0877bbc0712043639e29f026cd56b9c7"),
		/*
			//ScopeProjectName: openstack.String("admin"),
			//ScopeDomainName:  openstack.String("Default"),
			//ScopeProjectName: openstack.String("demo"),
			//ScopeDomainID:    openstack.String("default"),
			//UnscopedToken: openstack.Bool(true),
		*/
	}
	client.Identity.CreateToken(opts)
}

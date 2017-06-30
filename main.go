// Copyright 2017 Andrea Funtò. All rights reserved.
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

	if len(os.Args) < 2 {
		log.Errorln("usage: go-openstack <keystone-url>")
	}
	endpoint := os.Args[1]

	log.SetLevel(log.DBG)
	log.SetStream(os.Stdout)
	log.SetTimeFormat("15:04:05.000")

	log.Debugln("hallo!!!!")
	log.Infoln("hallo!!!!")
	log.Warnln("hallo!!!!")
	log.Errorln("hallo!!!!")
	log.Debugf("hallo!!!!")
	log.Infof("hallo!!!!")
	log.Warnf("hallo!!!!")
	log.Errorf("hallo!!!!")

	log.Debugf("%d elemens in %s\n", 3, "hallo")
	session := openstack.NewDefaultConnection()
	session.RegisterIdentityService(endpoint)
	if versions, err := session.Identity.GetVersions(); err == nil {
		for _, version := range versions {
			log.Infof("identity service supports: %v\n", version)
		}
	}

	//client.Identity.AuthenticateByPassword(opts)

}

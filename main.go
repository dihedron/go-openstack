// Copyright 2017 Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
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

	/*
		client := httpclient.Defaults(httpclient.Map{
			httpclient.OPT_USERAGENT: DefaultUserAgent + "/" + LibraryVersion,
			"Accept-Language":        "en-us",
		})
		client.Post("http://192.168.56.101/identity/v3/auth/tokens")

		//client.Identity.AuthenticateByPassword(opts)
	*/
}


type Identity struct {
	url string
}


type Data struct {
	Auth struct {
		Identity struct {
			Methods []string `json:"methods,omitempty"`,
			Password struct {
				User struct {
					Name string `json:"name,omitempty"`,
					Domain struct {
						Name string `json:"name,omitempty"`,
					} `json:"domain,omitempty"`,
					Password string `json:"password,omitempty`,
				} `json:"user,omitempty`,
			} `json:"password,omitempty"`,
		} `json:"identity,omitempty"`,
	} `json:"auth,omitempty"`,
}
/*
{
    "auth": {
        "identity": {
            "methods": [
                "password"
            ],
            "password": {
                "user": {
                    "name": "admin",
                    "domain": {
                        "name": "Default"
                    },
                    "password": "devstacker"
                }
            }
        }
    }
}
*/

func (i Identity) GetAuthToken(opts GetTokenOpts) {
	//url := "http://restapi3.apiary.io/notes"
	fmt.Println("URL:>", url)

	b := Data {
		Auth: {
			Identity: {
				
			}
		}
	}

	var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

}

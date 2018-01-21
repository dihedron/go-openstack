// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/dghubble/sling"
	"github.com/dihedron/go-openstack/log"
	"github.com/dihedron/go-openstack/openstack"
)

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
			"password": "password"
		  }
		}
	  }
	}
  }
*/

// see https://developer.openstack.org/api-ref/identity/v3/#identity-api-operations

/*
 *****************************************
 */

type AuthenticationRequestQuery struct {
	NoCatalog bool `url:"nocatalog,omitempty"`
}

type AuthenticationRequestBody struct {
	Auth *openstack.Authentication `json:"auth,omitempty"`
}

type AuthenticationResponseBody struct {
	Token *openstack.Token `json:"token,omitempty"`
}

// https://developer.openstack.org/sdks/python/openstacksdk/users/profile.html#openstack.profile.Profile
func main() {

	if len(os.Args) < 2 {
		log.Errorln("usage: go-openstack <keystone-url>")
	}
	//endpoint := os.Args[1]

	log.SetLevel(log.DBG)
	log.SetStream(os.Stdout)
	log.SetTimeFormat("15:04:05.000")
	/*

		log.Debugln("hallo!!!!")
		log.Infoln("hallo!!!!")
		log.Warnln("hallo!!!!")
		log.Errorln("hallo!!!!")
		log.Debugf("hallo!!!!")
		log.Infof("hallo!!!!")
		log.Warnf("hallo!!!!")
		log.Errorf("hallo!!!!")

	*/

	var client = &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	keystone := sling.New().Base("http://192.168.56.101").Client(client)

	query := &AuthenticationRequestQuery{
		NoCatalog: false,
	}

	body := &AuthenticationRequestBody{
		Auth: &openstack.Authentication{
			Identity: &openstack.Identity{
				Methods: openstack.StringSlice([]string{
					"password",
				}),
				Password: &openstack.Password{
					User: openstack.User{
						Name: openstack.String("admin"),
						Domain: &openstack.Domain{
							Name: openstack.String("Default"),
						},
						Password: openstack.String("password"),
					},
				},
			},
		},
	}
	//req, err := githubBase.New().Post(path).BodyJSON(body)

	if req, err := keystone.New().Post("/identity/v3/auth/tokens").QueryStruct(query).BodyJSON(body).Request(); err == nil {
		resp, err := client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		body := &AuthenticationResponseBody{}
		json.NewDecoder(resp.Body).Decode(body)
		b, _ := json.MarshalIndent(body, "", "  ")
		fmt.Printf("RESPONSE HEADER:\n%s\nRESPONSE BODY:\n%s\n", resp.Header.Get("X-Subject-Token"), b)
	}

	/*

		log.Debugf("%d elemens in %s\n", 3, "hallo")
		session := openstack.NewDefaultConnection()
		session.RegisterIdentityService(endpoint)
		if versions, err := session.Identity.GetVersions(); err == nil {
			for _, version := range versions {
				log.Infof("identity service supports: %v\n", version)
			}
		}
	*/

	/*
		client := httpclient.Defaults(httpclient.Map{
			httpclient.OPT_USERAGENT: DefaultUserAgent + "/" + LibraryVersion,
			"Accept-Language":        "en-us",
		})
		client.Post("http://192.168.56.101/identity/v3/auth/tokens")

		//client.Identity.AuthenticateByPassword(opts)
	*/
}

/*
type Identity struct {
	url string
}
*/

/*
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
*/
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
/*
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
*/

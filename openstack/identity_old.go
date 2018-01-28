// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

/*
 * AUTHENTICATION AND TOKEN MANAGEMENT
 */

/*
 * CREATE TOKEN
 */

// const (
// 	// CreateTokenMethodPassword is the constant used for password-based
// 	// authentication onto the Keystone server.
// 	CreateTokenMethodPassword string = "password"
// 	// CreateTokenMethodToken is the constant used for token-based
// 	// authentication onto the Keystone server.
// 	CreateTokenMethodToken string = "token"
// )

// // CreateTokenOpts contains the set of parameters and options used to
// // perform an authentication (create an authentication token).
// type CreateTokenOpts struct {
// 	Method           string
// 	NoCatalog        bool
// 	UserID           *string
// 	UserName         *string
// 	UserDomainID     *string
// 	UserDomainName   *string
// 	UserPassword     *string
// 	TokenID          *string
// 	ScopeProjectID   *string
// 	ScopeProjectName *string
// 	ScopeDomainID    *string
// 	ScopeDomainName  *string
// 	UnscopedToken    *bool
// }

// // CreateToken uses the provided parameters to authenticate the client to the
// // Keystone server and receive a token.
// func (api IdentityAPI) CreateToken(opts *CreateTokenOpts) (string, *Token, error) {

// 	query, _ := initCreateTokenRequestQuery(opts)

// 	// no headers in request!

// 	body, _ := initCreateTokenRequestBody(opts)

// 	log.Debugf("Identity.CreateToken: request body is\n%s\n", log.ToJSON(body))

// 	var err error
// 	if req, err := api.RequestFactory.New().Post("/identity/v3/auth/tokens").QueryStruct(query).BodyJSON(body).Request(); err == nil {
// 		res, err := api.Client.HTTPClient.Do(req)
// 		if err != nil {
// 			log.Errorf("Identity.CreateToken: error sending request: %v", err)
// 			return "", nil, err
// 		}
// 		defer res.Body.Close()

// 		if res.StatusCode == 201 {
// 			body := &createTokenResponseBody{}
// 			json.NewDecoder(res.Body).Decode(body)

// 			header := res.Header.Get("X-Subject-Token")

// 			log.Debugf("Identity.CreateToken: token value:\n%s\n", header)
// 			log.Debugf("Identity.CreateToken: token info:\n%s\n", log.ToJSON(body))
// 			return header, body.Token, nil
// 		}

// 		err = FromResponse(res)
// 		log.Debugf("Identity.CreateToken: API call unsuccessful: %v", err)
// 		return "", nil, err
// 	}

// 	log.Errorf("Identity.CreateToken: error creating request: %v\n", err)
// 	return "", nil, err
// }

// type createTokenRequestQuery struct {
// 	NoCatalog bool `url:"nocatalog,omitempty"`
// }

// type createTokenRequestBody struct {
// 	Auth *Authentication `json:"auth,omitempty"`
// }

// type createTokenResponseBody struct {
// 	Token *Token `json:"token,omitempty"`
// }

// // initCreateTokenRequestQuery creates the struct used to pass the request
// // options that go on the query string.
// func initCreateTokenRequestQuery(opts *CreateTokenOpts) (interface{}, error) {
// 	return &createTokenRequestQuery{
// 		NoCatalog: opts.NoCatalog,
// 	}, nil
// }

// // initCreateTokenRequestHeaders creates a pmap of header values to be
// // passed to the server along with the request.
// func initCreateTokenRequestHeaders(opts *CreateTokenOpts) (map[string][]string, error) {
// 	return map[string][]string{}, nil
// }

// // initCreateTokenRequestBody creates the structure representing the request
// // entity; the struct will be automatically serialised to JSON by the client.
// func initCreateTokenRequestBody(opts *CreateTokenOpts) (interface{}, error) {

// 	body := &createTokenRequestBody{
// 		Auth: &Authentication{
// 			Identity: &Identity{
// 				Methods: &[]string{
// 					opts.Method,
// 				},
// 			},
// 		},
// 	}

// 	if opts.Method == CreateTokenMethodPassword {
// 		if opts.UserID != nil && len(strings.TrimSpace(*opts.UserID)) > 0 {
// 			body.Auth.Identity.Password = &Password{
// 				User: &User{
// 					ID:       opts.UserID,
// 					Password: opts.UserPassword,
// 				},
// 			}
// 		} else {
// 			body.Auth.Identity.Password = &Password{
// 				User: &User{
// 					Name:     opts.UserName,
// 					Password: opts.UserPassword,
// 				},
// 			}
// 			if opts.UserDomainID != nil && len(strings.TrimSpace(*opts.UserDomainID)) > 0 {
// 				body.Auth.Identity.Password.User.Domain = &Domain{
// 					ID: opts.UserDomainID,
// 				}
// 			} else {
// 				body.Auth.Identity.Password.User.Domain = &Domain{
// 					Name: opts.UserDomainName,
// 				}
// 			}
// 		}
// 	} else if opts.Method == CreateTokenMethodToken {
// 		if opts.TokenID != nil && len(strings.TrimSpace(*opts.TokenID)) > 0 {
// 			body.Auth.Identity.Token = &Token{
// 				ID: opts.TokenID,
// 			}
// 		}
// 	}

// 	// manage scoped/unscoped token requests
// 	if opts.ScopeProjectID != nil && len(strings.TrimSpace(*opts.ScopeProjectID)) > 0 {
// 		body.Auth.Scope = &Scope{
// 			Project: &Project{
// 				ID: opts.ScopeProjectID,
// 			},
// 		}
// 	} else if opts.ScopeProjectName != nil && len(strings.TrimSpace(*opts.ScopeProjectName)) > 0 {
// 		scope := &Scope{
// 			Project: &Project{
// 				Name: opts.ScopeProjectName,
// 			},
// 		}

// 		if opts.ScopeDomainID != nil && len(strings.TrimSpace(*opts.ScopeDomainID)) > 0 {
// 			scope.Project.Domain = &Domain{
// 				ID: opts.ScopeDomainID,
// 			}
// 		} else {
// 			scope.Project.Domain = &Domain{
// 				Name: opts.ScopeDomainName,
// 			}
// 		}
// 		body.Auth.Scope = scope
// 	} else {
// 		if opts.ScopeDomainID != nil && len(strings.TrimSpace(*opts.ScopeDomainID)) > 0 {
// 			body.Auth.Scope = &Scope{
// 				Domain: &Domain{
// 					ID: opts.ScopeDomainID,
// 				},
// 			}
// 		} else if opts.ScopeDomainName != nil && len(strings.TrimSpace(*opts.ScopeDomainName)) > 0 {
// 			body.Auth.Scope = &Scope{
// 				Domain: &Domain{
// 					Name: opts.ScopeDomainName,
// 				},
// 			}
// 		} else if opts.UnscopedToken != nil && *opts.UnscopedToken {
// 			// all values are null: the request is unscoped
// 			body.Auth.Scope = String("unscoped")
// 		}
// 	}
// 	return body, nil
// }

// /*
//  * VALIDATE AND GET TOKEN INFO
//  */

// // ReadTokenOpts contains the set of parameters and options used to
// // perform the valudation of a token on the Identity server.
// type ReadTokenOpts struct {
// 	NoCatalog    bool
// 	AllowExpired bool
// 	SubjectToken string
// }

// // ReadToken uses the provided parameters to read the given token and retrieve
// // information about it from the Identity server; this API requires a valid admin
// // token.
// func (api IdentityAPI) ReadToken(token string, opts *ReadTokenOpts) (*Token, error) {
// 	query, _ := initReadTokenRequestQuery(opts)

// 	headers, _ := initReadTokenRequestHeaders(token, opts)

// 	// no entities in body!

// 	log.Debugf("Identity.ReadToken: reading subject token:\n%s\n", opts.SubjectToken)

// 	var err error
// 	sling := api.RequestFactory.New().Get("/identity/v3/auth/tokens").QueryStruct(query)
// 	for key, values := range headers {
// 		for _, value := range values {
// 			sling.Add(key, value)
// 		}
// 	}
// 	if req, err := sling.Request(); err == nil {
// 		res, err := api.Client.HTTPClient.Do(req)
// 		if err != nil {
// 			log.Errorf("Identity.ReadToken: error sending request: %v", err)
// 			return nil, err
// 		}
// 		defer res.Body.Close()

// 		if res.StatusCode == 200 {
// 			body := &readTokenResponseBody{}
// 			json.NewDecoder(res.Body).Decode(body)

// 			log.Debugf("Identity.ReadToken: token info:\n%s\n", log.ToJSON(body))
// 			return body.Token, nil
// 		}

// 		err = FromResponse(res)
// 		log.Debugf("Identity.ReadToken: API call unsuccessful: %v", err)
// 		return nil, err
// 	}

// 	log.Errorf("Identity.ReadToken: error creating request: %v\n", err)
// 	return nil, err
// }

// type readTokenRequestQuery struct {
// 	NoCatalog    bool `url:"nocatalog,omitempty"`
// 	AllowExpired bool `url:"allow_expired,omitempty"`
// }

// type readTokenRequestHeaders map[string][]string

// type readTokenRequestBody struct{}

// type readTokenResponseBody struct {
// 	Token *Token `json:"token,omitempty"`
// }

// // initReadTokenRequestQuery creates the struct used to pass the request
// // options that go on the query string.
// func initReadTokenRequestQuery(opts *ReadTokenOpts) (interface{}, error) {
// 	return &readTokenRequestQuery{
// 		NoCatalog:    opts.NoCatalog,
// 		AllowExpired: opts.AllowExpired,
// 	}, nil
// }

// // initReadTokenRequestHeaders creates a map of header values to be
// // passed to the server along with the request.
// func initReadTokenRequestHeaders(token string, opts *ReadTokenOpts) (readTokenRequestHeaders, error) {
// 	return readTokenRequestHeaders{
// 		"X-Auth-Token": []string{
// 			token,
// 		},
// 		"X-Subject-Token": []string{
// 			opts.SubjectToken,
// 		},
// 	}, nil
// }

// // initReadTokenRequestBody creates the structure representing the request
// // entity; the struct will be automatically serialised to JSON by the client.
// func initReadTokenRequestBody(opts *ReadTokenOpts) (interface{}, error) {
// 	return nil, nil
// }

// /*
//  * CHECK TOKEN
//  */

// // CheckTokenOpts contains the set of parameters and options used to
// // perform the valudation of a token on the Identity server.
// type CheckTokenOpts struct {
// 	AllowExpired bool
// 	SubjectToken string
// }

// // CheckToken uses the provided parameters to check the given token for validity
// // on the Identity server; this API requires a valid admin token.
// func (api IdentityAPI) CheckToken(token string, opts *CheckTokenOpts) (bool, error) {
// 	query, _ := initCheckTokenRequestQuery(opts)

// 	headers, _ := initCheckTokenRequestHeaders(token, opts)

// 	// no entities in body!

// 	log.Debugf("Identity.CheckToken: checking subject token:\n%s\n", opts.SubjectToken)

// 	var err error
// 	sling := api.RequestFactory.New().Head("/identity/v3/auth/tokens").QueryStruct(query)
// 	for key, values := range headers {
// 		for _, value := range values {
// 			sling.Add(key, value)
// 		}
// 	}
// 	if req, err := sling.Request(); err == nil {
// 		res, err := api.Client.HTTPClient.Do(req)
// 		if err != nil {
// 			log.Errorf("Identity.CheckToken: error sending request: %v", err)
// 			return false, err
// 		}
// 		defer res.Body.Close()

// 		log.Debugf("Identity.CheckToken: X-Subject-Token header: %s\n", res.Header.Get("X-Subject-Token"))

// 		if log.IsDebug() {
// 			bytes, _ := ioutil.ReadAll(res.Body)
// 			body := string(bytes)
// 			log.Debugf("Identity.CheckToken: response is:\n%s\n", body)
// 		}

// 		if res.StatusCode == http.StatusOK {
// 			log.Debugln("Identity.CheckToken: token is still valid (200)")
// 			return true, nil
// 		} else if res.StatusCode == 204 {
// 			log.Debugln("Identity.CheckToken: token is not valid anymore (204)")
// 			return false, nil
// 		}

// 		log.Debugf("Identity.CheckToken: API call unsuccessful: %v", err)
// 		err = FromResponse(res)
// 		return false, err
// 	}

// 	log.Errorf("Identity.CheckToken: error creating request: %v\n", err)
// 	return false, err
// }

// type checkTokenRequestQuery struct {
// 	AllowExpired bool `url:"allow_expired,omitempty"`
// }

// type checkTokenRequestHeaders map[string][]string

// type checkTokenRequestBody struct{}

// type checkTokenResponseBody struct{}

// // initCheckTokenRequestQuery creates the struct used to pass the request
// // options that go on the query string.
// func initCheckTokenRequestQuery(opts *CheckTokenOpts) (interface{}, error) {
// 	return &readTokenRequestQuery{
// 		AllowExpired: opts.AllowExpired,
// 	}, nil
// }

// // initCheckTokenRequestHeaders creates a map of header values to be
// // passed to the server along with the request.
// func initCheckTokenRequestHeaders(token string, opts *CheckTokenOpts) (checkTokenRequestHeaders, error) {
// 	return checkTokenRequestHeaders{
// 		"X-Auth-Token": []string{
// 			token,
// 		},
// 		"X-Subject-Token": []string{
// 			opts.SubjectToken,
// 		},
// 	}, nil
// }

// // initCheckTokenRequestBody creates the structure representing the request
// // entity; the struct will be automatically serialised to JSON by the client.
// func initCheckTokenRequestBody(opts *CheckTokenOpts) (interface{}, error) {
// 	return nil, nil
// }

// /*
//  * DELETE TOKEN
//  */

// // DeleteTokenOpts contains the set of parameters and options used to revoke
// // a token on the Identity server.
// type DeleteTokenOpts struct {
// 	SubjectToken string
// }

// // DeleteToken uses the provided parameters to revoke the given token for validity
// // on the Identity server; this API requires a valid admin token.
// func (api IdentityAPI) DeleteToken(token string, opts *DeleteTokenOpts) (bool, error) {

// 	// no parameters in query

// 	headers, _ := initDeleteTokenRequestHeaders(token, opts)

// 	// no entities in body!

// 	log.Debugf("Identity.DeleteToken: checking subject token:\n%s\n", opts.SubjectToken)

// 	var err error
// 	sling := api.RequestFactory.New().Delete("/identity/v3/auth/tokens")
// 	for key, values := range headers {
// 		for _, value := range values {
// 			sling.Add(key, value)
// 		}
// 	}
// 	if req, err := sling.Request(); err == nil {
// 		res, err := api.Client.HTTPClient.Do(req)
// 		if err != nil {
// 			log.Errorf("Identity.DeleteToken: error sending request: %v", err)
// 			return false, err
// 		}
// 		defer res.Body.Close()

// 		log.Debugf("Identity.DeleteToken: X-Subject-Token header: %s\n", res.Header.Get("X-Subject-Token"))

// 		if res.StatusCode == http.StatusOK {
// 			log.Debugln("Identity.DeleteToken: token is still valid (200)")
// 			return true, nil
// 		} else if res.StatusCode == 204 {
// 			log.Debugln("Identity.DeleteToken: token is not valid anymore (204)")
// 			return false, nil
// 		}

// 		log.Debugf("Identity.DeleteToken: API call unsuccessful: %v", err)
// 		err = FromResponse(res)
// 		return false, err
// 	}

// 	log.Errorf("Identity.DeleteToken: error creating request: %v\n", err)
// 	return false, err
// }

// type deleteTokenRequestQuery struct{}

// type deleteTokenRequestHeaders map[string][]string

// type deleteTokenRequestBody struct{}

// type deleteTokenResponseBody struct{}

// // initDeleteTokenRequestQuery creates the struct used to pass the request
// // options that go on the query string.
// func initDeleteTokenRequestQuery(opts *CheckTokenOpts) (interface{}, error) {
// 	return nil, nil
// }

// // initDeleteTokenRequestHeaders creates a map of header values to be
// // passed to the server along with the request.
// func initDeleteTokenRequestHeaders(token string, opts *DeleteTokenOpts) (deleteTokenRequestHeaders, error) {
// 	return deleteTokenRequestHeaders{
// 		"X-Auth-Token": []string{
// 			token,
// 		},
// 		"X-Subject-Token": []string{
// 			opts.SubjectToken,
// 		},
// 	}, nil
// }

// // initDeleteTokenRequestBody creates the structure representing the request
// // entity; the struct will be automatically serialised to JSON by the client.
// func initDeleteTokenRequestBody(opts *DeleteTokenOpts) (interface{}, error) {
// 	return nil, nil
// }

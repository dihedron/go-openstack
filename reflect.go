package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dihedron/go-openstack/openstack"
)

type MyOpts struct {
	Query   *MyQuery   `request:"params,omitempty"`
	Headers *MyHeaders `request:"headers"`
	Entity  *MyEntity  `request:"entity,omitempty"`
	Ignore  *string    `json:"ignore"`
}

type MyQuery struct {
	Query1 *int    `url:"query1,omitempty"`
	Query2 *string `url:"query2,omitempty"`
	Query3 *string `url:"query3,omitempty"`
}

type MyHeaders struct {
	Header1 *string   `header:"X-Subject-Token,omitempty"`
	Header2 *[]string `header:"X-Auth-Token,omitempty"`
}

type MyEntity struct {
	Name    *string `json:"name,omitempty"`
	Surname *string `json:"surname,omitempty"`
}

func Convert(opts interface{}) {
	// TypeOf returns the reflection Type that represents the dynamic type of variable.
	// If variable is a nil interface value, TypeOf returns nil.
	t := reflect.TypeOf(opts).Elem()

	// Get the type and kind of our user variable
	fmt.Println("Type:", t.Name())
	fmt.Println("Kind:", t.Kind())

	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		// Get the field, returns https://golang.org/pkg/reflect/#StructField
		field := t.Field(i)

		// Get the field tag value
		tag := field.Tag.Get("request")
		if len(strings.TrimSpace(tag)) > 0 {
			values := strings.Split(tag, ",")
			fmt.Printf("%d. %v (%v), tag: '%v'\n", i+1, field.Name, field.Type.Name(), values)
		} else {
			fmt.Printf("%d. %v (%v), no tag\n", i+1, field.Name, field.Type.Name())
		}

	}
}

func test1() {
	opts := &MyOpts{
		Query: &MyQuery{
			Query1: openstack.Int(1),
			Query2: openstack.String("value2"),
			Query3: openstack.String("value3"),
		},
		Headers: &MyHeaders{
			Header1: openstack.String("subject_token_abccdefghijklmnopqrstuvwxyz"),
			Header2: &[]string{
				"header_value_1",
				"header_value_2",
			},
		},
		Entity: &MyEntity{
			Name:    openstack.String("name"),
			Surname: openstack.String("surname"),
		},
	}
	Convert(opts)
}

// CreateTokenOpts contains the set of parameters and options used to
// perform an authentication (create an authentication token).
type MyComplexOpts struct {
	NoCatalog        bool    `api:"query" url:"nocatalog,omitempty"`
	AllowExpired     bool    `api:"query" url:"allow_expired,omitempty"`
	AuthToken        *string `api:"header" header:"X-Auth-Token"`
	Method           string  `api:"entity" snl:""`
	UserID           *string
	UserName         *string
	UserDomainID     *string
	UserDomainName   *string
	UserPassword     *string
	TokenID          *string
	ScopeProjectID   *string
	ScopeProjectName *string
	ScopeDomainID    *string
	ScopeDomainName  *string
	UnscopedToken    *bool
}

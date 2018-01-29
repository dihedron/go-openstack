package openstack

import (
	"os"
	"testing"

	"github.com/dihedron/go-openstack/log"
)

func TestCreateTokenFromEnv(t *testing.T) {

	if os.Getenv("OS_USERNAME") == "" || os.Getenv("OS_PASSWORD") == "" || os.Getenv("OS_USER_DOMAIN_NAME") == "" {
		t.Errorf("Identity.TestCreateTokenFromEnv: OS_* environment variables must be set for this test to be run")
		t.FailNow()
	}

	log.SetLevel(log.ERR)
	log.SetStream(os.Stdout)
	log.SetTimeFormat("15:04:05.000")

	client := NewDefaultClient("")
	if client == nil {
		t.Errorf("Identity.TestCreateTokenFromEnv: error initialising client")
		t.FailNow()
	}
	client.IdentityV3().CreateTokenFromEnv()
}

func TestCreateTokenParam(t *testing.T) {
	endpoint := "http://192.168.56.101/identity/" // my shiny devstack :-)

	client := NewDefaultClient(endpoint)
	if client == nil {
		t.Errorf("Identity.TestCreateTokenFromEnv: error initialising client")
		t.FailNow()
	}

	tests := map[string]*CreateTokenOpts{
		"UserNameUserDomainPasswordImplicitlyUnscopedCatalog": &CreateTokenOpts{
			Method:         "password",
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
		},
		"UserNameUserDomainPasswordExplicitlyUnscopedCatalog": &CreateTokenOpts{
			Method:         "password",
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			UnscopedToken:  Bool(true),
		},
		"UserNameUserDomainPasswordScopedDomainIDCatalog": &CreateTokenOpts{
			Method:         "password",
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			ScopeDomainID:  String("default"),
		},
		"UserNameUserDomainPasswordScopedDomainNameCatalog": &CreateTokenOpts{
			Method:          "password",
			UserName:        String("admin"),
			UserDomainName:  String("Default"),
			UserPassword:    String("password"),
			ScopeDomainName: String("Default"),
		},
		"UserNameUserDomainPasswordScopedProjectIDCatalog": &CreateTokenOpts{
			Method:         "password",
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			ScopeProjectID: String("b5ca4b54c504463291d138f0c24e1a20"),
		},
		"UserNameUserDomainPasswordScopedProjectNameDomainNameCatalog": &CreateTokenOpts{
			Method:           "password",
			UserName:         String("admin"),
			UserDomainName:   String("Default"),
			UserPassword:     String("password"),
			ScopeProjectName: String("admin"),
			ScopeDomainName:  String("Default"),
		},
		"UserNameUserDomainPasswordScopedProjectNameDomainIDCatalog": &CreateTokenOpts{
			Method:           "password",
			UserName:         String("admin"),
			UserDomainName:   String("Default"),
			UserPassword:     String("password"),
			ScopeProjectName: String("admin"),
			ScopeDomainID:    String("default"),
		},
		"UserNameUserDomainPasswordImplicitlyUnscopedNoCatalog": &CreateTokenOpts{
			Method:         "password",
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			NoCatalog:      true,
		},
		"UserNameUserDomainPasswordExplicitlyUnscopedNoCatalog": &CreateTokenOpts{
			Method:         "password",
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			UnscopedToken:  Bool(true),
			NoCatalog:      true,
		},
		"UserNameUserDomainPasswordScopedDomainIDNoCatalog": &CreateTokenOpts{
			Method:         "password",
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			ScopeDomainID:  String("default"),
			NoCatalog:      true,
		},
		"UserNameUserDomainPasswordScopedDomainNameNoCatalog": &CreateTokenOpts{
			Method:          "password",
			UserName:        String("admin"),
			UserDomainName:  String("Default"),
			UserPassword:    String("password"),
			ScopeDomainName: String("Default"),
			NoCatalog:       true,
		},
		"UserNameUserDomainPasswordScopedProjectIDNoCatalog": &CreateTokenOpts{
			Method:         "password",
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			ScopeProjectID: String("b5ca4b54c504463291d138f0c24e1a20"),
			NoCatalog:      true,
		},
		"UserNameUserDomainPasswordScopedProjectNameDomainNameNoCatalog": &CreateTokenOpts{
			Method:           "password",
			UserName:         String("admin"),
			UserDomainName:   String("Default"),
			UserPassword:     String("password"),
			ScopeProjectName: String("admin"),
			ScopeDomainName:  String("Default"),
			NoCatalog:        true,
		},
		"UserNameUserDomainPasswordScopedProjectNameDomainIDNoCatalog": &CreateTokenOpts{
			Method:           "password",
			UserName:         String("admin"),
			UserDomainName:   String("Default"),
			UserPassword:     String("password"),
			ScopeProjectName: String("admin"),
			ScopeDomainID:    String("default"),
			NoCatalog:        true,
		},
	}
	for test, opts := range tests {
		token, _, _, _ := client.IdentityV3().CreateToken(opts)

		if len(token) == 0 {
			t.Errorf("Identity.TestCreateToken%s: no token returned", test)
			t.FailNow()
		}
	}
}

package openstack

import (
	"os"
	"testing"

	"github.com/dihedron/go-log/log"
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

	tests := map[string]*CreateTokenOptions{
		"UserNameUserDomainPasswordImplicitlyUnscopedCatalog": &CreateTokenOptions{
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
		},
		"UserNameUserDomainPasswordExplicitlyUnscopedCatalog": &CreateTokenOptions{
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			UnscopedToken:  Bool(true),
		},
		"UserNameUserDomainPasswordScopedDomainIDCatalog": &CreateTokenOptions{
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			ScopeDomainID:  String("default"),
		},
		"UserNameUserDomainPasswordScopedDomainNameCatalog": &CreateTokenOptions{
			UserName:        String("admin"),
			UserDomainName:  String("Default"),
			UserPassword:    String("password"),
			ScopeDomainName: String("Default"),
		},
		"UserNameUserDomainPasswordScopedProjectIDCatalog": &CreateTokenOptions{
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			ScopeProjectID: String("b5ca4b54c504463291d138f0c24e1a20"),
		},
		"UserNameUserDomainPasswordScopedProjectNameDomainNameCatalog": &CreateTokenOptions{
			UserName:         String("admin"),
			UserDomainName:   String("Default"),
			UserPassword:     String("password"),
			ScopeProjectName: String("admin"),
			ScopeDomainName:  String("Default"),
		},
		"UserNameUserDomainPasswordScopedProjectNameDomainIDCatalog": &CreateTokenOptions{
			UserName:         String("admin"),
			UserDomainName:   String("Default"),
			UserPassword:     String("password"),
			ScopeProjectName: String("admin"),
			ScopeDomainID:    String("default"),
		},
		"UserNameUserDomainPasswordImplicitlyUnscopedNoCatalog": &CreateTokenOptions{
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			NoCatalog:      true,
		},
		"UserNameUserDomainPasswordExplicitlyUnscopedNoCatalog": &CreateTokenOptions{
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			UnscopedToken:  Bool(true),
			NoCatalog:      true,
		},
		"UserNameUserDomainPasswordScopedDomainIDNoCatalog": &CreateTokenOptions{
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			ScopeDomainID:  String("default"),
			NoCatalog:      true,
		},
		"UserNameUserDomainPasswordScopedDomainNameNoCatalog": &CreateTokenOptions{
			UserName:        String("admin"),
			UserDomainName:  String("Default"),
			UserPassword:    String("password"),
			ScopeDomainName: String("Default"),
			NoCatalog:       true,
		},
		"UserNameUserDomainPasswordScopedProjectIDNoCatalog": &CreateTokenOptions{
			UserName:       String("admin"),
			UserDomainName: String("Default"),
			UserPassword:   String("password"),
			ScopeProjectID: String("b5ca4b54c504463291d138f0c24e1a20"),
			NoCatalog:      true,
		},
		"UserNameUserDomainPasswordScopedProjectNameDomainNameNoCatalog": &CreateTokenOptions{
			UserName:         String("admin"),
			UserDomainName:   String("Default"),
			UserPassword:     String("password"),
			ScopeProjectName: String("admin"),
			ScopeDomainName:  String("Default"),
			NoCatalog:        true,
		},
		"UserNameUserDomainPasswordScopedProjectNameDomainIDNoCatalog": &CreateTokenOptions{
			UserName:         String("admin"),
			UserDomainName:   String("Default"),
			UserPassword:     String("password"),
			ScopeProjectName: String("admin"),
			ScopeDomainID:    String("default"),
			NoCatalog:        true,
		},
	}
	for test, opts := range tests {
		token, _, _ := client.IdentityV3().CreateToken(opts)

		if token == nil {
			t.Errorf("Identity.TestCreateToken%s: no token returned", test)
			t.FailNow()
		}
	}
}

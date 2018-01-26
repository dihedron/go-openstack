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

	log.SetLevel(log.DBG)
	log.SetStream(os.Stdout)
	log.SetTimeFormat("15:04:05.000")

	client, err := NewDefaultClient("")
	if err != nil {
		t.Errorf("Identity.TestCreateTokenFromEnv: error initialising client: %v", err)
		t.FailNow()
	}
	client.Identity.CreateTokenFromEnv()
}

func TestCreateTokenUserNameUserDomainPasswordUnscoped(t *testing.T) {
	endpoint := "http://192.168.56.101" // my shiny devstack :-)

	client, err := NewDefaultClient(endpoint)
	if err != nil {
		t.Errorf("Identity.TestCreateTokenFromEnv: error initialising client: %v", err)
		t.FailNow()
	}

	opts := &CreateTokenOpts{
		Method:         "password",
		UserName:       String("admin"),
		UserDomainName: String("Default"),
		UserPassword:   String("password"),
	}
	token, _, _, _ := client.Identity.CreateToken(opts)

	if len(token) == 0 {
		t.Errorf("Identity.TestCreateTokenUnscopedPassword: no token returned")
		t.FailNow()
	}
}

func TestCreateTokenUserIDPasswordUnscoped(t *testing.T) {
	endpoint := "http://192.168.56.101" // my shiny devstack :-)

	client, err := NewDefaultClient(endpoint)
	if err != nil {
		t.Errorf("Identity.TestCreateTokenFromEnv: error initialising client: %v", err)
		t.FailNow()
	}

	opts := &CreateTokenOpts{
		/*
			Method: CreateTokenMethodToken,
			TokenID:         String("gAAAAABaZgmbPZtoEyuTzJXmggwMAyjLZSiknQJPeR4m1FQaL0dpv1nvvVZvd-B3PORQnRqXrR3OevmRKvMqrXwiam02xElVJXOQHKkExqpTK4kkBnttb-kZRxyS3AJLTLjOr7rxzGP2jw7OwGfOclzNxRIRZF00Ha88ApD0iNFKBczP9PBv4A8"),
			ScopeDomainName: String("Default"),
		*/

		Method: "password",
		//NoCatalog:      true,
		UserName:       String("admin"),
		UserDomainName: String("Default"),
		UserPassword:   String("password"),
		//ScopeProjectID: String("0877bbc0712043639e29f026cd56b9c7"),
		/*
			//ScopeProjectName: String("admin"),
			//ScopeDomainName:  String("Default"),
			//ScopeProjectName: String("demo"),
			//ScopeDomainID:    String("default"),
			//UnscopedToken: Bool(true),
		*/
	}
	token, object, _, _ := client.Identity.CreateToken(opts)

	log.Debugf("Identity.TestCreateTokenUserIDPasswordUnscoped: token is:\n%s\ntoken info:\n%s\n", token, log.ToJSON(object))

	if len(token) == 0 {
		t.Errorf("Identity.TestCreateTokenUserIDPasswordUnscoped: no token returned")
		t.FailNow()
	}
}

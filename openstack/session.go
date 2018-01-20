// Copyright 2017 Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"net/http"
	"net/url"
)

// Connection represents a connection to OpenStack.
type Connection struct {

	// HTTP client used to communicate with the API.
	Client *http.Client

	// CatalogueURL is the base URL for catalog API requests; from this URL, it should
	// be possible to reconstruct the whole set of enabled OpenStack services,
	// their versions and endpoints.
	CatalogueURL *url.URL

	/*
		// token type used to make authenticated API calls.
		tokenType tokenType

		// token used to make authenticated API calls.
		token string
	*/

	// Identity is the remote service providing user authentication and
	// authorization; for details see OpenStack Keystone:
	// https://developer.openstack.org/api-ref/identity/v3/
	//Identity *IdentityService
}

const (
	// LibraryVersion is the version of the current library.
	LibraryVersion string = "1.0.0"

	// DefaultUserAgent is the user agent used when communicating with the
	// OpenStack API when none is provided by the library user.
	DefaultUserAgent string = "go-openstack"
)

/*
// NewDefaultConnection initiates a new OpenStack Connection
// with the default parameters.
func NewDefaultConnection() *Connection {
	return &Connection{
		client: httpclient.Defaults(httpclient.Map{
			httpclient.OPT_USERAGENT: DefaultUserAgent + "/" + LibraryVersion,
			"Accept-Language":        "en-us",
		}),
	}

}
*/

/*
// NewConnection initiates a new OpenStack Connection.
func NewConnection(client *http.Client) *Connection {
	if client == nil {
		client = http.DefaultClient()
	}
	return &Connection{
		client: client,
	}
}
*/

/*
// NegotiateVersion queries the base endpoint of an API to choose the most
// recent non-experimental alternative from a service's published versions.
// It returns the highest-Priority Version among the alternatives that are
// provided, as well as its corresponding endpoint.
func (c *Client) NegotiateVersion(endpoint string, base string, supported []*Version) (*Version, string, error) {
	type response struct {
		Versions struct {
			Values []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Links  []struct {
					Href string `json:"href"`
					Rel  string `json:"rel"`
				} `json:"links"`
			} `json:"values"`
		} `json:"versions"`
	}

	// if a full endpoint is specified, the first attempt to match
	// a version is against the endpoint URL
	if endpoint != "" {
		log.Printf("[D] checking API endpoint: %s", endpoint)
		apiEndpoint := NormalizeURL(endpoint)
		for _, version := range supported {
			if strings.HasSuffix(apiEndpoint, version.Suffix) {
				log.Printf("[D] version found: %v", version)
				return version, apiEndpoint, nil
			}
		}
	}

	var resp response
	r, err := c.client.Get(base, nil)
	if err != nil {
		log.Printf("[E] error performing version negotiation: %v", err)
	}
	// parse the response body as JSON, if requested to do so.
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return nil, "", err
	}

	/*
		_, err := client.Request("GET", client.IdentityBase, &gophercloud.RequestOpts{
			JSONResponse: &resp,
			OkCodes:      []int{200, 300},
		})

		if err != nil {
			return nil, "", err
		}

		byID := make(map[string]*Version)
		for _, version := range recognized {
			byID[version.ID] = version
		}

		var highest *Version
		var endpoint string

		for _, value := range resp.Versions.Values {
			href := ""
			for _, link := range value.Links {
				if link.Rel == "self" {
					href = normalize(link.Href)
				}
			}

			if matching, ok := byID[value.ID]; ok {
				// Prefer a version that exactly matches the provided endpoint.
				if href == identityEndpoint {
					if href == "" {
						return nil, "", fmt.Errorf("Endpoint missing in version %s response from %s", value.ID, client.IdentityBase)
					}
					return matching, href, nil
				}

				// Otherwise, find the highest-priority version with a whitelisted status.
				if goodStatus[strings.ToLower(value.Status)] {
					if highest == nil || matching.Priority > highest.Priority {
						highest = matching
						endpoint = href
					}
				}
			}
		}

		if highest == nil {
			return nil, "", fmt.Errorf("No supported version available from endpoint %s", client.IdentityBase)
		}
		if endpoint == "" {
			return nil, "", fmt.Errorf("Endpoint missing in version %s response from %s", highest.ID, client.IdentityBase)
		}

		return highest, endpoint, nil

	return nil, "", nil
}
*/

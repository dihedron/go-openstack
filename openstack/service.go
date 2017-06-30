// Copyright 2017 Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package openstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dihedron/go-openstack/log"
)

// Service is the base struct for all OpenStack service providers.
type Service struct {
	// The base OpenStack session; this will contain a reference to
	// the HTTP client used to make the acctual requests.
	session *Connection
	// The token used to authenticate requests.
	token *string
	// The endpoint URL to connect to.
	endpoint string
	// The base URL of the endpoint
	base string
	// The version of the API service.
	version string
	// The set of support versions
	supportedVersions []Version
}

// Version represents an API version on any of the many
// services provided by openStack (Keystone, Nova, Horizon
// etc.).
type Version struct {
	ID         string      `json:"id,omitempty"`
	Status     string      `json:"status,omitempty"`
	Links      []Link      `json:"links,omitempty"`
	MediaTypes []MediaType `json:"media-types,omitempty"`
	Updated    *time.Time  `json:"updated,omitempty"`
}

// MediaType represents the media types supported by the API
// endpoint, e.g. "application/json".
type MediaType struct {
	Base string `json:"base,omitempty"`
	Type string `json:"type,omitempty"`
}

// Link represents an hypermedia link in the API.
type Link struct {
	Href string `json:"href,omitempty"`
	Type string `json:"type,omitempty"`
	Rel  string `json:"rel,omitempty"`
}

// VersionStatus represents the possible values for the status of
// an API version.
var VersionStatus = map[string]bool{
	"current":   true,
	"supported": true,
	"stable":    true,
}

// GetVersions retrieves the list of supported versions by
// the given service endpoint.
//
// Any OpenStack service enpoint (Nova, Keystone, Glance etc.)
// exposes severeal version of its management APIs; in order for the
// client to be able to negotiate the API version with the server
// prior to authenticating (as authencìtication itself can expose
// mutliple API versions), this method leverages the only API that
// is guaranteed to require no authentication on all services.
func (s *Service) GetVersions() ([]Version, error) {

	res, err := s.session.client.Get(s.base, nil)
	if err != nil {
		log.Printf("[E] error connecting to service %s: %v", s.base, err)
		return nil, err
	}
	defer res.Body.Close()

	buffer := new(bytes.Buffer)
	buffer.ReadFrom(res.Body)

	if log.IsDebug() {
		var pretty bytes.Buffer
		json.Indent(&pretty, buffer.Bytes(), "", "  ")
		log.Debugf("read:\n%s\n", pretty.String())
	}

	// parse the response body as JSON, if requested to do so.
	var response struct {
		Data struct {
			Versions []Version `json:"values"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(buffer).Decode(&response); err != nil {
		log.Printf("[E] error decoding response body into JSON: %v", err)
		return nil, err
	}
	log.Printf("[I] version retrieved: %v", response.Data.Versions)
	// register the set of supported versions into the Service
	s.supportedVersions = response.Data.Versions
	return response.Data.Versions, nil
}

const (
	v20 = "v2.0"
	v30 = "v3.0"
)

// RequestedVersion can be used by the client to negotiate an API version
// by comparing its supported version against those exposed by the API
// endpoint; the Priority field allows to specify the preferred API version.
type RequestedVersion struct {
	ID       string
	Suffix   string
	Priority int
}

// String returns a textual representation of a equestedVersion object.
func (v RequestedVersion) String() string {
	return fmt.Sprintf("%s { suffix: %q, priority: %d }", v.ID, v.Suffix, v.Priority)
}

// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/dihedron/go-openstack/log"
)

// Profile represents the list of service endpoints to use when connecting
// to the given URL as a Keystone instance, each represented as a Filter.
type Profile struct {
	AuthURL *string  `json:"url,omitempty"`
	Filters []Filter `json:"filters,omitempty"`
}

// Filter represents a single service instance and endpoint; when using
// the services exposed by a given Keystone instance through the catalog,
// only services matching one such filter will be employed, effectively
// masking out all those that do not match; a filter also expresses a
// preference in terms of micro-version (aka APIVersion) for those services
// that support it.
type Filter struct {
	// Type is the type of service, e.g. "compute".
	Type *string `json:"type,omitempty"`
	// Name os the name of the service, e.g. "nova".
	Name *string `json:"name,omitempty"`
	// Region is the OpenStack region to which the service applies.
	Region      *string `json:"region,omitempty"`
	Interface   *string `json:"interface,omitempty"`
	Version     *string `json:"version,omitempty"`
	APIVersion  *string `json:"microversion,omitempty"`
	EndpointURL *string `json:"url,omitempty"`
}

// InitProfile initialises the Client's profile using all the information
// available in the catalog as per the identity service; once initialised
// and saved to disk, the user can edit it and reload it so that the Client
// is configured with the right set of filters.
func (c *Client) InitProfile() error {
	if c.Profile == nil {
		c.Profile = &Profile{
			AuthURL: c.Authenticator.AuthURL,
		}
	}
	if c.Authenticator.AuthURL == nil || c.Authenticator.TokenInfo == nil || c.Authenticator.TokenInfo.Catalog == nil {
		log.Errorln("Client.InitProfile: to initi a profile, the client must be connected to the identity service")
		return fmt.Errorf("no connection to identity service yet")
	}
	for _, service := range *c.Authenticator.TokenInfo.Catalog {
		if service.Endpoints != nil {
			if c.Profile.Filters == nil {
				c.Profile.Filters = []Filter{}
			}
			for _, endpoint := range *service.Endpoints {
				filter := Filter{
					Type:        service.Type,
					Name:        service.Name,
					Region:      endpoint.Region,
					Interface:   endpoint.Interface,
					EndpointURL: endpoint.URL,
				}
				c.Profile.Filters = append(c.Profile.Filters, filter)
			}
		}
	}

	return nil
}

// SaveProfile saves the currently loaded profile to the given io.Writer.
func (c *Client) SaveProfile(writer io.Writer) error {
	if c.Profile == nil {
		log.Errorln("Client.SaveProfile: no valid profile available")
		return fmt.Errorf("no valid profile loaded")
	}

	data, err := json.MarshalIndent(c.Profile, "", "  ")
	if err != nil {
		log.Errorf("Client.SaveProfile: error marshalling profile to JSON: %v\n", err)
		return err
	}

	_, err = writer.Write(data)
	if err != nil {
		log.Errorf("Client.SaveProfile: error writing profile to file: %v\n", err)
		return err
	}
	return nil
}

// SaveProfileTo save the currently loaded profile to a file with the given path.
func (c *Client) SaveProfileTo(path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Errorf("Client.SaveProfileTo: error creating profile file: %v\n", err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()
	return c.SaveProfile(writer)
}

// LoadProfile loads the user's services profile reading data from the given reader.
func (c *Client) LoadProfile(reader io.Reader) error {
	if c.Profile != nil {
		log.Warnln("Client.LoadProfile: replacing existing profile")
		c.Profile = nil
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Errorf("Client.LoadProfile: error reading from input stream: %v\n", err)
		return err
	}

	profile := &Profile{}

	err = json.Unmarshal(data, profile)
	if err != nil {
		log.Errorf("Client.LoadProfile: error unmarshalling profile from JSON: %v\n", err)
		return err
	}

	c.Profile = profile

	log.Debugf("Client.LoadProfile: profile loaded:\n%s\n", log.ToJSON(c.Profile))
	return nil
}

// LoadProfileFrom loads the user's services provider reading data from the file
// at the given path.
func (c *Client) LoadProfileFrom(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Errorf("Client.LoadProfileFrom: error opening profile file: %v\n", err)
		return err
	}
	defer file.Close()

	log.Debugf("Client.LoadProfileFrom: loading profile from %q\n", path)
	reader := bufio.NewReader(file)
	return c.LoadProfile(reader)
}

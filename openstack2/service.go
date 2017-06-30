package openstack2

import (
	"github.com/dihedron/go-openstack/log"
	"github.com/dihedron/go-openstack/openstack"
)

const (
	// AllServicePreferences is a wildcard representing
	// all ServicePreferences.
	AllServicePreferences string = "*"
)

// ServiceType represents the type of Service; it is the
// key into the map of currently registered ServicePreferences
// in the Profile.
type ServiceType string

const (
	// Compute represents the identifier of the Compute (Nova) service.
	Compute ServiceType = "compute"
	// Identity represents the identifier of the Identity (Keystone) service.
	Identity = "identity"
	// Network represents the identifier of the Network (Neutron) service.
	Network = "network"
)

// ServicePreferences represents the preferences for
// a given Service (e.g. "compute" or "identity")
type ServicePreferences struct {
	Name       string
	Region     string
	Version    string
	APIVersion string
	Interface  string
}

// SetName sets the desired name for the given service.
func (s *ServicePreference) SetName(name string) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	s.Name = name
	return nil
}

// SetRegion sets the desired region for the given service.
func (s *ServicePreference) SetRegion(region string) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	s.Region = region
	return nil
}

// SetVersion sets the desired version for the given service.
func (s *ServicePreference) SetVersion(version string) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	s.Version = version
	return nil
}

// SetAPIVersion sets the desired API micro-version for the given service.
func (s *ServicePreference) SetAPIVersion(version string) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	s.APIVersion = version
	return nil
}

// SetInterface sets the desired interface for the given service.
func (s *ServicePreference) SetInterface(interf string) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	s.Interface = interf
	return nil
}

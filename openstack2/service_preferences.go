package openstack2

import (
	"strings"

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

// InterfaceType is used to restrict the values available for the parameter
// used to specify the elevel of exposure of an endpoint.
type InterfaceType string

const (
	// PublicInterface represents an enpoint with public exposure (the default).
	PublicInterface InterfaceType = "public"
	// InternalInterface reresents an endpoint with internal exposure.
	InternalInterface = "internal"
	// AdminInterface represents an endpoint with administrative exposure.
	AdminInterface = "admin"
)

// ServicePreferences represents the preferences for a given Service (e.g.
// "compute" or "identity"); it maps to the Python SDK's service filter.
type ServicePreferences struct {
	Name              *string // service name, e.g. "matrix"
	Region            *string // service region, e.g. "zion"
	Version           *string // service version, e.g. "v3"
	APIVersion        *string // API microversion, not part of URL
	Interface         *string // the exposure of the endpoint, should one of "PUBLIC" (default), "INTERNAL" or "ADMIN"
	RequiresProjectID bool    // true if the service's endpoint expects the project id to be included
}

// SetName sets the desired name for the given service.
func (s *ServicePreferences) SetName(name string) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	s.Name = openstack.String(name)
	return nil
}

// SetRegion sets the desired region for the given service.
func (s *ServicePreferences) SetRegion(region string) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	s.Region = openstack.String(region)
	return nil
}

// SetVersion sets the desired version for the given service.
func (s *ServicePreferences) SetVersion(version string) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	s.Version = openstack.String(version)
	return nil
}

// SetAPIVersion sets the desired API micro-version for the given service.
func (s *ServicePreferences) SetAPIVersion(version string) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	s.APIVersion = openstack.String(version)
	return nil
}

// SetInterface sets the desired interface for the given service.
func (s *ServicePreferences) SetInterface(interf string) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	switch InterfaceType(strings.ToLower(interf)) {
	case PublicInterface, InternalInterface, AdminInterface:
		s.Interface = openstack.String(interf)
	default:
		log.Errorf("invalid value for interface, should be one of %s, %s or %s\n", PublicInterface, InternalInterface, AdminInterface)
		return openstack.ErrorInvalidInput
	}
	return nil
}

// SetRequiresProjectID sets whether the service endpoint expects the project ID to be specified;
// this value is non-nullable and has a sensible default as "false"
func (s *ServicePreferences) SetRequiresProjectID(value bool) error {
	if s == nil {
		log.Errorln("invalid reference to ServicePreference")
		return openstack.ErrorInvalidReference
	}
	s.RequiresProjectID = value
	return nil
}

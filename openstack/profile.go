package openstack

import (
	"github.com/dihedron/go-openstack/log"
)

// Profile is the struct that is used to define the various preferences
// for different services. The preference values that are currently supported
// are: service name, region, (API micro-)version, exposure (interface), and
// whether the enspoint requires a project id to be specified. The Profile and
// the Connection structs are the most important client facing structs.
type Profile struct {
	services map[ServiceType]*ServicePreferences
}

// NewProfile returns a new uninitialised Profile.
func NewProfile() Profile {
	return Profile{
		services: map[ServiceType]*ServicePreferences{},
	}
}

// NewDefaultProfile returns a new Profile pre-initialised with a new
// ServicePreferences for each Service available in the predefined set.
func NewDefaultProfile() Profile {
	// TODO: add an entry in the map as we implement new services.
	return NewProfile()
}

// GetAllServices returns a list of all know services.
func (p Profile) GetAllServices() ([]ServiceType, error) {
	services := make([]ServiceType, len(p.services))
	for service := range p.services {
		services = append(services, service)
	}
	return services, nil
}

// GetServicePreferences returns a reference to the service preferences values; these
// can be modified directly if necessary.
func (p Profile) GetServicePreferences(service ServiceType) (ServiceType, *ServicePreferences, error) {
	preferences, ok := p.services[service]
	if !ok {
		log.Errorf("no preferences available for service %q\n", service)
		return service, nil, ErrorNotFound
	}
	return service, preferences, nil
}

const (
	// AllServicePreferences is a wildcard representing all ServicePreferences.
	AllServicePreferences string = "*"
)

// ServiceType represents the type of Service; it is the key into the map of currently
// registered ServicePreferences in the Profile.
type ServiceType string

const (
	// Compute represents the identifier of the Compute (Nova) service.
	Compute ServiceType = "compute"
	// Identity represents the identifier of the Identity (Keystone) service.
	Identity = "identity"
	// Network represents the identifier of the Network (Neutron) service.
	Network = "network"
	// TODO: add others...
)

const (
	// PublicInterface represents an enpoint with public exposure (the default).
	PublicInterface string = "PUBLIC"
	// InternalInterface reresents an endpoint with internal exposure.
	InternalInterface string = "INTERNAL"
	// AdminInterface represents an endpoint with administrative exposure.
	AdminInterface string = "ADMIN"
)

// ServicePreferences represents the preferences for a given Service (e.g.
// "compute" or "identity"); it maps to the Python SDK's service filter.
// Each value is a pointer because this way we can track unset values as nil
// pointers.
type ServicePreferences struct {
	Name              *string // service name, e.g. "matrix"
	Region            *string // service region, e.g. "zion"
	Version           *string // service version, e.g. "v3"
	APIVersion        *string // API microversion, not part of URL
	Interface         *string // the exposure of the endpoint, should one of "PUBLIC" (default), "INTERNAL" or "ADMIN"
	RequiresProjectID bool    // true if the service's endpoint expects the project id to be included
}

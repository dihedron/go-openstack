package openstack2

import (
	"github.com/dihedron/go-openstack/log"
	"github.com/dihedron/go-openstack/openstack"
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
func NewProfile() *Profile {
	return &Profile{
		services: map[ServiceType]*ServicePreferences{},
	}
}

// NewDefaultProfile returns a new Profile pre-initialised with a new
// ServicePreferences for each Service available in the predefined set.
func NewDefaultProfile() *Profile {
	// TODO: add an entry in the ma as we implement new services.
	return NewProfile()
}

// GetAllServices returns a list of all know services.
func (p *Profile) GetAllServices() ([]ServiceType, error) {
	if p == nil {
		log.Errorf("invaid input reference to Profile")
		return nil, openstack.ErrorInvalidReference
	}
	services := make([]ServiceType, len(p.services))
	for service := range p.services {
		services = append(services, service)
	}
	return services, nil
}

// GetServicePreferences returns the service preferences values.
func (p *Profile) GetServicePreferences(service ServiceType) (ServiceType, *ServicePreferences, error) {
	if p == nil {
		log.Errorf("invaid input reference to Profile")
		return service, nil, openstack.ErrorInvalidReference
	}
	preferences, ok := p.services[service]
	if !ok {
		log.Errorf("no preferences available for service %q\n", service)
		return service, nil, openstack.ErrorNotFound
	}
	return service, preferences, nil
}

package openstack2

import (
	"github.com/dihedron/go-openstack/log"
	"github.com/dihedron/go-openstack/openstack"
)

// Profile is the struct that is used to define the various
// preferences for different services. The preferences that
// are currently supported are service name, region, version
// and interface. The Profile and the Connection structs are
// the most important user facing structs.
type Profile struct {
	services map[ServiceType]*ServicePreferences
}

// NewProfile returns a new, uninitialised Profile.
func NewProfile() *Profile {
	return &Profile{
		services: map[ServiceType]*ServicePreferences{},
	}
}

// GetAllServices returns a list of all know services.
func (p *Profile) GetAllServices() ([]ServiceType, error) {
	if p == nil {
		log.Errorf("invaid input reference to Profile")
		return nil, openstack.ErrorInvalidReference
	}
	services := make([]ServiceType, len(p.services))
	for service := range p.services {
		services = append(servicesces, service)
	}
	return services, nil
}

// GetFilter returns a service preferences; if the service type
// is not in the profile yet, it is initialised automatically.
func (p *Profile) GetFilter(service ServiceType) (ServiceType, *ServicePreferences, error) {
	if p == nil {
		log.Errorf("invaid input reference to Profile")
		return service, nil, openstack.ErrorInvalidReference
	}
	var preferences *ServicePreferences
	if preferences, ok := p.services[service]; !ok {
		preferences = &ServicePreferences{}
		p.services[service] = preferences
	}
	return service, preferences, nil
}

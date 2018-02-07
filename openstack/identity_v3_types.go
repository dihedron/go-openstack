// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

/*
 * AUTHENTICATION AND TOKEN MANAGEMENT
 */

// Authentication contains the identity entity used to authenticate users and
// issue tokens against a Keystone instance; it can be scoped (when either
// Project or Domain is specified), implicitly unscoped (when neither is specified)
// or expicitly unscoped (when the "unscoped" flag is set); see Scope for details.
type Authentication struct {
	Identity *Identity   `json:"identity,omitempty"`
	Scope    interface{} `json:"scope,omitempty"`
}

// Domain is a container of users, roles, projects and resources; it is itself
// associated with a limited number of services when a token is scoped to it.
type Domain struct {
	ID          *string `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
	Links       *Links  `json:"links,omitempty"`
}

// Endpoint is the address of a specific interface to a service; the Interface
// field specifies the kind of usage to which the endpoint is devoted, i.e.
// "public" (everyone can access it, including subscribers), "admin" (only cloud
// administators have access to it) and "internal" (technical endpoint, devoted
// to inter-service communications).
type Endpoint struct {
	ID        *string `json:"id,omitempty"`
	Interface *string `json:"interface,omitempty"`
	Region    *string `json:"region,omitempty"`
	RegionID  *string `json:"region_id,omitempty"`
	URL       *string `json:"url,omitempty"`
}

// Identity represents an identity, as granted by the Identity service to a
// user providing the given password or token with the given authentication
// method.
type Identity struct {
	Methods  *[]string `json:"methods,omitempty"`
	Password *Password `json:"password,omitempty"`
	Token    *Token    `json:"token,omitempty"`
}

// Links represents the links to the resource itself and its immediate siblings
// if available; it is used for embedding inside other resources as a rudimentary
// support for HATEOAS.
type Links struct {
	Self     *string `json:"self,omitempty"`
	Previous *string `json:"previous,omitempty"`
	Next     *string `json:"next,omitempty"`
}

// Password identifies a user's credentials; see User for details.
type Password struct {
	User *User `json:"user,omitempty"`
}

// Project represents a container that groups or isolates resources or identity
// objects; depending on the service operator, a project might map to a customer,
// account, organization, or tenant. A Project can itself be a container of other
// Project and act as a Domain; most services are associated with tokens issued
// with Project scope.
type Project struct {
	ID          *string `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Domain      *Domain `json:"domain,omitempty"`
	DomainID    *string `json:"domain_id,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
	Description *string `json:"description,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
	Links       *Links  `json:"links,omitempty"`
}

// Role is a personality that a user assumes to perform a specific set of
// operations. A role includes a set of rights and privileges. A user assumes
// that role inherits those rights and privileges.
type Role struct {
	ID    *string `json:"id,omitempty"`
	Name  *string `json:"name,omitempty"`
	Links *Links  `json:"links,omitempty"`
}

// Scope represents the scope of a Token; a Token can be issued either at Project
// or at Domain scope, but not both (mutually exclusive).
type Scope struct {
	Project *Project `json:"project,omitempty"`
	Domain  *Domain  `json:"domain,omitempty"` // either one or the other: if both, BadRequest!
}

// Service represents an OpenStack service, such as Compute (nova), Object Storage
// (swift), or Image service (glance), that provides one or more endpoints through
// which users can access resources and perform operations.
type Service struct {
	ID        *string     `json:"id,omitempty"`
	Name      *string     `json:"name,omitempty"`
	Type      *string     `json:"type,omitempy"`
	Endpoints *[]Endpoint `json:"endpoints,omitempty"`
}

// System is a system that a token can be scoped to.
type System struct {
	All *bool `json:"all,omitempty"`
}

// Token represents An alpha-numeric text string that enables access to OpenStack
// APIs and resources and all associated metadata. A token may be revoked at any
// time and is valid for a finite duration. While OpenStack Identity supports
// token-based authentication in this release, it intends to support additional
// protocols in the future. OpenStack Identity is an integration service that does
// not aspire to be a full-fledged identity store and management solution.
// The "value" field is not part od the JSON entity and is used to store the
// actual token value as returned by the API in the X-Subject-Auth header, so
// that data and metadata are alla vailable in one place throughout the API; as
// an "unofficial" addition, it is not tagged for JSON (un-)marshalling.
type Token struct {
	ID           *string    `json:"id,omitempty"`
	IssuedAt     *string    `json:"issued_at,omitempty"`
	ExpiresAt    *string    `json:"expires_at,omitempty"`
	User         *User      `json:"user,omitempty"`
	Roles        *[]Role    `json:"roles,omitempty"`
	Methods      *[]string  `json:"methods,omitempty"`
	AuditIDs     *[]string  `json:"audit_ids,omitempty"`
	Project      *Project   `json:"project,omitempty"`
	IsDomain     *bool      `json:"is_domain,omitempty"`
	IsAdminToken *bool      `json:"is_admin_token,omitempty"`
	Catalog      *[]Service `json:"catalog,omitempty"`
	Value        *string    `json:"-"`
}

// User is a digital representation of a person, system, or service that uses
// OpenStack cloud services. The Identity service validates that incoming requests
// are made by the user who claims to be making the call. Users have a login and
// can access resources by using assigned tokens. Users can be directly assigned
// to a particular project and behave as if they are contained in that project.
type User struct {
	ID                *string `json:"id,omitempty"`
	Name              *string `json:"name,omitempty"`
	Domain            *Domain `json:"domain,omitempty"`
	Password          *string `json:"password,omitempty"`
	PasswordExpiresAt *string `json:"password_expires_at,omitempty"`
}

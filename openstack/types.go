// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

// Bool returns a pointer to a boolean value that is safe
// to be used in OpenStack API structs.
func Bool(value bool) *bool {
	return &value
}

// String returns a string pointer that is safe to be
// used in OpenStack API structs.
func String(value string) *string {
	return &value
}

// Int returns a pointer to an int value that is safe
// to be used in OpenStack API structs.
func Int(value int) *int {
	return &value
}

// Int8 returns a pointer to an int8 value that is safe
// to be used in OpenStack API structs.
func Int8(value int8) *int8 {
	return &value
}

// Int16 returns a pointer to an int16 value that is safe
// to be used in OpenStack API structs.
func Int16(value int16) *int16 {
	return &value
}

// Int32 returns a pointer to an int32 value that is safe
// to be used in OpenStack API structs.
func Int32(value int32) *int32 {
	return &value
}

// Int64 returns a pointer to an int64 value that is safe
// to be used in OpenStack API structs.
func Int64(value int64) *int64 {
	return &value
}

// UInt returns a pointer to an uint value that is safe
// to be used in OpenStack API structs.
func UInt(value uint) *uint {
	return &value
}

// UInt8 returns a pointer to an uint8 value that is safe
// to be used in OpenStack API structs.
func UInt8(value uint8) *uint8 {
	return &value
}

// UInt16 returns a pointer to an uint16 value that is safe
// to be used in OpenStack API structs.
func UInt16(value uint16) *uint16 {
	return &value
}

// UInt32 returns a pointer to an uint32 value that is safe
// to be used in OpenStack API structs.
func UInt32(value uint32) *uint32 {
	return &value
}

// UInt64 returns a pointer to an uint64 value that is safe
// to be used in OpenStack API structs.
func UInt64(value uint64) *uint64 {
	return &value
}

// UIntPtr returns a pointer to an uintptr value that is safe
// to be used in OpenStack API structs.
func UIntPtr(value uintptr) *uintptr {
	return &value
}

// Byte returns a pointer to a byte value that is safe
// to be used in OpenStack API structs.
func Byte(value byte) *byte {
	return &value
}

// Rune returns a pointer to a rune value that is safe
// to be used in OpenStack API structs.
func Rune(value rune) *rune {
	return &value
}

// Float32 returns a pointer to a float32 value that is safe
// to be used in OpenStack API structs.
func Float32(value float32) *float32 {
	return &value
}

// Float64 returns a pointer to a float64 value that is safe
// to be used in OpenStack API structs.
func Float64(value float64) *float64 {
	return &value
}

// Complex64 returns a pointer to a complex64 value that is safe
// to be used in OpenStack API structs.
func Complex64(value complex64) *complex64 {
	return &value
}

// Complex128 returns a pointer to a complex128 value that is safe
// to be used in OpenStack API structs.
func Complex128(value complex128) *complex128 {
	return &value
}

// StringSlice returns a pointer to a []string value that is safe
// to be used in OpenStack API structs.
func StringSlice(value []string) *[]string {
	return &value
}

/*
 * AUTHENTICATION AND TOKEN MANAGEMENT
 */

// Authentication contains the identity entity used to authenticate users
// and issue tokens against a Keystone instance; it can be scoped (when
// either Project or Domain is specified), implicitly unscoped ()
type Authentication struct {
	Identity *Identity   `json:"identity,omitempty"`
	Scope    interface{} `json:"scope,omitempty"`
}

type Domain struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type Endpoint struct {
	ID        *string `json:"id,omitempty"`
	Interface *string `json:"interface,omitempty"`
	Region    *string `json:"region,omitempty"`
	RegionID  *string `json:"region_id,omitempty"`
	URL       *string `json:"url,omitempty"`
}

type Identity struct {
	Methods  *[]string `json:"methods,omitempty"`
	Password *Password `json:"password,omitempty"`
	Token    *Token    `json:"token,omitempty"`
}

type Password struct {
	User *User `json:"user,omitempty"`
}

// Project represents a container that groups or isolates resources or identity
// objects; depending on the service operator, a project might map to a customer,
// account, organization, or tenant.
type Project struct {
	ID     *string `json:"id,omitempty"`
	Name   *string `json:"name,omitempty"`
	Domain *Domain `json:"domain,omitempty"`
}

type Role struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

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

// Token represents An alpha-numeric text string that enables access to OpenStack
// APIs and resources and all associated metadata. A token may be revoked at any
// time and is valid for a finite duration. While OpenStack Identity supports
// token-based authentication in this release, it intends to support additional
// protocols in the future. OpenStack Identity is an integration service that does
// not aspire to be a full-fledged identity store and management solution.
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

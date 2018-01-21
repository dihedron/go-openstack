// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package openstack

// Authentication contains the identity entity used to
// authenticate users against a Keystone instance.
type Authentication struct {
	Identity *Identity `json:"identity,omitempty"`
}

type Identity struct {
	Methods  *[]string `json:"methods,omitempty"`
	Password *Password `json:"password,omitempty"`
}

type Scope struct {
	Project *Project `json:"project,omitempty"`
	Domain  *Domain  `json:"domain,omitempty"` // either one or the other: if both, BadRequest!
}

type Project struct {
	ID     *string `json:"id,omitemty"`
	Name   *string `json:"name,omitempty"`
	Domain *Domain `json:"domain,omitempty"`
}
type Password struct {
	User User `json:"user,omitempty"`
}

type User struct {
	ID                *string `json:"id,omitempty"`
	Name              *string `json:"name,omitempty"`
	Domain            *Domain `json:"domain,omitempty"`
	Password          *string `json:"password,omitempty"`
	PasswordExpiresAt *string `json:"password_expires_at,omitempty"`
}

type Domain struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type Role struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type Token struct {
	IssuedAt  *string   `json:"issued_at,omitempty"`
	ExpiresAt *string   `json:"expires_at,omitempty"`
	User      *User     `json:"user,omitempty"`
	Roles     *[]Role   `json:"roles,omitempty"`
	Methods   *[]string `json:"methods,omitempty"`
	AuditIds  *[]string `json:"audit_ids,omitempty"`
}

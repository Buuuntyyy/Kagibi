// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package ldap provides LDAP/AD directory synchronisation for Kagibi organisations.
package ldap

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	goldap "github.com/go-ldap/ldap/v3"

	"kagibi/backend/pkg"
)

// LDAPUser is a user record returned from the directory.
type LDAPUser struct {
	DN          string
	UID         string
	Email       string
	DisplayName string
}

// LDAPGroup is a group record returned from the directory.
type LDAPGroup struct {
	DN      string
	Name    string
	Members []string // member DNs
}

// Client wraps a connection to an LDAP/AD server.
type Client struct {
	cfg *pkg.OrgLDAPConfig
}

// NewClient builds a client from the stored config. Password must be already decrypted.
func NewClient(cfg *pkg.OrgLDAPConfig) *Client {
	return &Client{cfg: cfg}
}

// Dial opens a connection to the LDAP server. The caller must Close() the returned connection.
func (c *Client) Dial() (*goldap.Conn, error) {
	tlsCfg := &tls.Config{InsecureSkipVerify: c.cfg.TLSSkipVerify} // #nosec G402 — admin-configurable

	var (
		conn *goldap.Conn
		err  error
	)
	if strings.HasPrefix(c.cfg.URL, "ldaps://") {
		conn, err = goldap.DialURL(c.cfg.URL, goldap.DialWithTLSConfig(tlsCfg))
	} else {
		conn, err = goldap.DialURL(c.cfg.URL)
		if err == nil {
			if tlsErr := conn.StartTLS(tlsCfg); tlsErr != nil {
				log.Printf("[ldap] StartTLS unavailable for %s: %v", c.cfg.URL, tlsErr)
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("ldap dial %s: %w", c.cfg.URL, err)
	}
	return conn, nil
}

// Bind authenticates with the stored bind DN and the provided (plaintext) password.
func (c *Client) Bind(conn *goldap.Conn, password string) error {
	return conn.Bind(c.cfg.BindDN, password)
}

// SearchUsers performs a subtree search under UserBaseDN and returns all matching users.
func (c *Client) SearchUsers(conn *goldap.Conn) ([]LDAPUser, error) {
	filter := c.cfg.UserFilter
	if filter == "" {
		filter = "(objectClass=person)"
	}
	req := goldap.NewSearchRequest(
		c.cfg.UserBaseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		0, 0, false,
		filter,
		[]string{"dn", c.cfg.AttrUID, c.cfg.AttrEmail, c.cfg.AttrDisplayName},
		nil,
	)
	sr, err := conn.Search(req)
	if err != nil {
		return nil, fmt.Errorf("ldap user search: %w", err)
	}

	users := make([]LDAPUser, 0, len(sr.Entries))
	for _, e := range sr.Entries {
		u := LDAPUser{
			DN:          e.DN,
			UID:         e.GetAttributeValue(c.cfg.AttrUID),
			Email:       strings.ToLower(strings.TrimSpace(e.GetAttributeValue(c.cfg.AttrEmail))),
			DisplayName: e.GetAttributeValue(c.cfg.AttrDisplayName),
		}
		if u.UID == "" || u.Email == "" {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

// SearchGroups performs a subtree search under GroupBaseDN and returns all matching groups.
// Returns nil if GroupBaseDN is empty (group sync disabled).
func (c *Client) SearchGroups(conn *goldap.Conn) ([]LDAPGroup, error) {
	if c.cfg.GroupBaseDN == "" {
		return nil, nil
	}
	filter := c.cfg.GroupFilter
	if filter == "" {
		filter = "(objectClass=groupOfNames)"
	}
	req := goldap.NewSearchRequest(
		c.cfg.GroupBaseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		0, 0, false,
		filter,
		[]string{"dn", "cn", "member"},
		nil,
	)
	sr, err := conn.Search(req)
	if err != nil {
		return nil, fmt.Errorf("ldap group search: %w", err)
	}

	groups := make([]LDAPGroup, 0, len(sr.Entries))
	for _, e := range sr.Entries {
		g := LDAPGroup{
			DN:      e.DN,
			Name:    e.GetAttributeValue("cn"),
			Members: e.GetAttributeValues("member"),
		}
		if g.Name == "" {
			continue
		}
		groups = append(groups, g)
	}
	return groups, nil
}

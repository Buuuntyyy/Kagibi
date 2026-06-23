// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"kagibi/backend/internal/ldap"
	"kagibi/backend/pkg"
	"kagibi/backend/pkg/emailcrypto"

	"github.com/gin-gonic/gin"
)

// ldapConfigResponse is returned by GET /ldap — omits the encrypted bind password.
type ldapConfigResponse struct {
	ID                  int64              `json:"id"`
	OrgID               int64              `json:"org_id"`
	Enabled             bool               `json:"enabled"`
	URL                 string             `json:"url"`
	BindDN              string             `json:"bind_dn"`
	BindPasswordSet     bool               `json:"bind_password_set"`
	UserBaseDN          string             `json:"user_base_dn"`
	UserFilter          string             `json:"user_filter"`
	GroupBaseDN         string             `json:"group_base_dn"`
	GroupFilter         string             `json:"group_filter"`
	AttrEmail           string             `json:"attr_email"`
	AttrDisplayName     string             `json:"attr_display_name"`
	AttrUID             string             `json:"attr_uid"`
	TLSSkipVerify       bool               `json:"tls_skip_verify"`
	SyncIntervalMinutes int                `json:"sync_interval_minutes"`
	AutoDeprovisionDays int                `json:"auto_deprovision_days"`
	MinExpectedUsers    int                `json:"min_expected_users"`
	LastSyncAt          *time.Time         `json:"last_sync_at,omitempty"`
	LastSyncError       string             `json:"last_sync_error,omitempty"`
	LastSyncStats       *pkg.LDAPSyncStats `json:"last_sync_stats,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

func configToResponse(c *pkg.OrgLDAPConfig) ldapConfigResponse {
	return ldapConfigResponse{
		ID:                  c.ID,
		OrgID:               c.OrgID,
		Enabled:             c.Enabled,
		URL:                 c.URL,
		BindDN:              c.BindDN,
		BindPasswordSet:     c.BindPasswordEnc != "",
		UserBaseDN:          c.UserBaseDN,
		UserFilter:          c.UserFilter,
		GroupBaseDN:         c.GroupBaseDN,
		GroupFilter:         c.GroupFilter,
		AttrEmail:           c.AttrEmail,
		AttrDisplayName:     c.AttrDisplayName,
		AttrUID:             c.AttrUID,
		TLSSkipVerify:       c.TLSSkipVerify,
		SyncIntervalMinutes: c.SyncIntervalMinutes,
		AutoDeprovisionDays: c.AutoDeprovisionDays,
		MinExpectedUsers:    c.MinExpectedUsers,
		LastSyncAt:          c.LastSyncAt,
		LastSyncError:       c.LastSyncError,
		LastSyncStats:       c.LastSyncStats,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}
}

// GetLDAPConfig returns the LDAP configuration for an organisation.
// GET /api/v1/orgs/:orgID/ldap
func (h *OrgHandler) GetLDAPConfig(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin or owner required"})
		return
	}

	var cfg pkg.OrgLDAPConfig
	if err := h.DB.NewSelect().Model(&cfg).Where("org_id = ?", orgID).Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			// Return an empty default config if not yet created.
			c.JSON(http.StatusOK, ldapConfigResponse{
				OrgID:               orgID,
				UserFilter:          "(objectClass=person)",
				GroupFilter:         "(objectClass=groupOfNames)",
				AttrEmail:           "mail",
				AttrDisplayName:     "cn",
				AttrUID:             "uid",
				SyncIntervalMinutes: 60,
				AutoDeprovisionDays: 30,
				MinExpectedUsers:    1,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load ldap config"})
		return
	}
	c.JSON(http.StatusOK, configToResponse(&cfg))
}

// SaveLDAPConfig creates or updates the LDAP configuration for an organisation.
// PUT /api/v1/orgs/:orgID/ldap
func (h *OrgHandler) SaveLDAPConfig(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req struct {
		Enabled             bool   `json:"enabled"`
		URL                 string `json:"url"`
		BindDN              string `json:"bind_dn"`
		BindPassword        string `json:"bind_password"` // plaintext, encrypted before storage
		UserBaseDN          string `json:"user_base_dn"`
		UserFilter          string `json:"user_filter"`
		GroupBaseDN         string `json:"group_base_dn"`
		GroupFilter         string `json:"group_filter"`
		AttrEmail           string `json:"attr_email"`
		AttrDisplayName     string `json:"attr_display_name"`
		AttrUID             string `json:"attr_uid"`
		TLSSkipVerify       bool   `json:"tls_skip_verify"`
		SyncIntervalMinutes int    `json:"sync_interval_minutes"`
		AutoDeprovisionDays int    `json:"auto_deprovision_days"`
		MinExpectedUsers    int    `json:"min_expected_users"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin or owner required"})
		return
	}

	if req.SyncIntervalMinutes < 5 {
		req.SyncIntervalMinutes = 60
	}
	if req.MinExpectedUsers < 1 {
		req.MinExpectedUsers = 1
	}

	// Load or create config.
	var cfg pkg.OrgLDAPConfig
	isNew := false
	if err := h.DB.NewSelect().Model(&cfg).Where("org_id = ?", orgID).Scan(ctx); err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load existing config"})
			return
		}
		isNew = true
		cfg.OrgID = orgID
	}

	cfg.Enabled = req.Enabled
	cfg.URL = req.URL
	cfg.BindDN = req.BindDN
	cfg.UserBaseDN = req.UserBaseDN
	cfg.UserFilter = req.UserFilter
	cfg.GroupBaseDN = req.GroupBaseDN
	cfg.GroupFilter = req.GroupFilter
	cfg.AttrEmail = req.AttrEmail
	cfg.AttrDisplayName = req.AttrDisplayName
	cfg.AttrUID = req.AttrUID
	cfg.TLSSkipVerify = req.TLSSkipVerify
	cfg.SyncIntervalMinutes = req.SyncIntervalMinutes
	cfg.AutoDeprovisionDays = req.AutoDeprovisionDays
	cfg.MinExpectedUsers = req.MinExpectedUsers
	cfg.UpdatedAt = time.Now()

	if req.BindPassword != "" {
		enc, err := emailcrypto.Encrypt(req.BindPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt bind password"})
			return
		}
		cfg.BindPasswordEnc = enc
	}

	if isNew {
		if _, err := h.DB.NewInsert().Model(&cfg).Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create ldap config"})
			return
		}
	} else {
		if _, err := h.DB.NewUpdate().Model(&cfg).Where("org_id = ?", orgID).Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update ldap config"})
			return
		}
	}

	// Reload the scheduler goroutine for this org.
	ldap.Reload(orgID)

	h.logAudit(ctx, orgID, callerID, "ldap_config_updated", strconv.FormatInt(orgID, 10), "organization", "")
	c.JSON(http.StatusOK, configToResponse(&cfg))
}

// TestLDAPConnection attempts to bind to the LDAP server with the current config.
// POST /api/v1/orgs/:orgID/ldap/test
func (h *OrgHandler) TestLDAPConnection(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin or owner required"})
		return
	}

	var cfg pkg.OrgLDAPConfig
	if err := h.DB.NewSelect().Model(&cfg).Where("org_id = ?", orgID).Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ldap not configured"})
		return
	}
	if cfg.BindPasswordEnc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bind password not set"})
		return
	}

	bindPassword, err := emailcrypto.Decrypt(cfg.BindPasswordEnc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt bind password"})
		return
	}

	client := ldap.NewClient(&cfg)
	conn, err := client.Dial()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "connection failed: " + err.Error()})
		return
	}
	defer conn.Close()

	if err := client.Bind(conn, bindPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bind failed: " + err.Error()})
		return
	}

	// Attempt a quick user search to validate the filter.
	users, err := client.SearchUsers(conn)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "user search failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"users_found": len(users),
	})
}

// TriggerLDAPSync runs an immediate LDAP sync for an organisation.
// POST /api/v1/orgs/:orgID/ldap/sync
func (h *OrgHandler) TriggerLDAPSync(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin or owner required"})
		return
	}

	var cfg pkg.OrgLDAPConfig
	if err := h.DB.NewSelect().Model(&cfg).Where("org_id = ?", orgID).Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ldap not configured"})
		return
	}
	if !cfg.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ldap sync is disabled for this organization"})
		return
	}

	stats, err := ldap.SyncNow(ctx, h.DB, &cfg)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	h.logAudit(ctx, orgID, callerID, "ldap_sync_triggered", strconv.FormatInt(orgID, 10), "organization", "manual")
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// ListSuspendedLDAPMembers returns org members currently suspended by the LDAP sync.
// GET /api/v1/orgs/:orgID/ldap/suspended
func (h *OrgHandler) ListSuspendedLDAPMembers(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin or owner required"})
		return
	}

	var members []pkg.OrgMember
	if err := h.DB.NewSelect().Model(&members).
		Where("org_id = ? AND source = 'ldap' AND suspended_at IS NOT NULL", orgID).
		OrderExpr("suspended_at ASC").
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list suspended members"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": members})
}

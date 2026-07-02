// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

// RequestFolderAccess lets any org member submit a request to access a restricted folder.
// A pending request for the same (org, user, folder) is upserted — the message is updated.
func (h *OrgHandler) RequestFolderAccess(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req struct {
		FolderPath string `json:"folder_path" binding:"required"`
		Message    string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	folderPath := normPath(req.FolderPath)

	// Verify the folder actually exists.
	var count int
	if err := h.DB.NewSelect().TableExpr("org_folders").
		ColumnExpr("COUNT(*)").
		Where("org_id = ? AND path = ?", orgID, folderPath).
		Scan(ctx, &count); err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		return
	}

	// Verify the caller really cannot access the folder (no point requesting if already allowed).
	perm, err := h.resolvePermission(ctx, orgID, userID, folderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if perm >= PermRead {
		c.JSON(http.StatusConflict, gin.H{"error": "you already have access to this folder"})
		return
	}

	ar := &pkg.OrgAccessRequest{
		OrgID:      orgID,
		UserID:     userID,
		FolderPath: folderPath,
		Message:    req.Message,
		Status:     "pending",
	}
	if _, err := h.DB.NewInsert().Model(ar).
		On("CONFLICT (org_id, user_id, folder_path) DO UPDATE").
		Set("message = EXCLUDED.message, status = 'pending', resolved_by = '', resolved_at = NULL").
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit access request"})
		return
	}

	h.logAudit(ctx, orgID, userID, "access_requested", folderPath, "folder", folderPath)
	c.JSON(http.StatusCreated, ar)
}

// ListAccessRequests returns access requests for the org.
// Admins/owners see all requests; members see only their own.
func (h *OrgHandler) ListAccessRequests(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	statusFilter := c.Query("status") // optional: pending | approved | denied

	var requests []pkg.OrgAccessRequest
	q := h.DB.NewSelect().Model(&requests).
		Where("oar.org_id = ?", orgID).
		OrderExpr("oar.created_at DESC")

	if !canManage(role) {
		q = q.Where("oar.user_id = ?", callerID)
	}
	if statusFilter != "" {
		q = q.Where("oar.status = ?", statusFilter)
	}

	if err := q.Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list access requests"})
		return
	}
	if requests == nil {
		requests = []pkg.OrgAccessRequest{}
	}
	c.JSON(http.StatusOK, requests)
}

// ResolveAccessRequest approves or denies a pending access request. Admin/owner only.
// On approval the caller must also set a group permission separately — this endpoint
// only changes the request status and logs the decision.
func (h *OrgHandler) ResolveAccessRequest(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	requestID, err := strconv.ParseInt(c.Param("requestID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"` // approved | denied
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status != "approved" && req.Status != "denied" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be approved or denied"})
		return
	}

	ctx := c.Request.Context()
	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var ar pkg.OrgAccessRequest
	if err := h.DB.NewSelect().Model(&ar).
		Where("id = ? AND org_id = ?", requestID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "access request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch request"})
		return
	}
	if ar.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "request has already been resolved"})
		return
	}

	now := time.Now().UTC()
	ar.Status = req.Status
	ar.ResolvedBy = callerID
	ar.ResolvedAt = &now

	if _, err := h.DB.NewUpdate().Model(&ar).
		Column("status", "resolved_by", "resolved_at").
		WherePK().Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve request"})
		return
	}

	action := "access_request_approved"
	if req.Status == "denied" {
		action = "access_request_denied"
	}
	h.logAudit(ctx, orgID, callerID, action, ar.UserID, "folder", ar.FolderPath)
	c.JSON(http.StatusOK, ar)
}

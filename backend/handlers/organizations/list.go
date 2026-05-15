// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type OrgResponse struct {
	pkg.Organization
	MyRole            string `json:"my_role"`
	MyEncryptedOrgKey string `json:"my_encrypted_org_key,omitempty"`
	IsGroupAdmin      bool   `json:"is_group_admin"`
	LogoURL           string `json:"logo_url,omitempty"`
}

// presignLogoURL generates a 1-hour presigned GET URL for the given S3 key.
// Returns "" if S3 is not configured or the key is empty.
func presignLogoURL(ctx context.Context, key string) string {
	if key == "" || s3storage.Client == nil {
		return ""
	}
	presigner := s3svc.NewPresignClient(s3storage.Client)
	req, err := presigner.PresignGetObject(ctx, &s3svc.GetObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(key),
	}, s3svc.WithPresignExpires(time.Hour))
	if err != nil {
		return ""
	}
	return req.URL
}

func (h *OrgHandler) ListOrgs(c *gin.Context) {
	userID := c.GetString("user_id")
	if !h.hasOrgAccess(userID) {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "organizations require a paid plan on the cloud"})
		return
	}
	ctx := c.Request.Context()

	var memberships []pkg.OrgMember
	if err := h.DB.NewSelect().Model(&memberships).
		Where("user_id = ?", userID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch memberships"})
		return
	}
	if len(memberships) == 0 {
		c.JSON(http.StatusOK, []OrgResponse{})
		return
	}

	orgIDs := make([]int64, len(memberships))
	roleByOrg := make(map[int64]string, len(memberships))
	keyByOrg := make(map[int64]string, len(memberships))
	for i, m := range memberships {
		orgIDs[i] = m.OrgID
		roleByOrg[m.OrgID] = m.Role
		keyByOrg[m.OrgID] = m.EncryptedOrgKey
	}

	var orgs []pkg.Organization
	if err := h.DB.NewSelect().Model(&orgs).
		Where("id IN (?)", bun.In(orgIDs)).
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organizations"})
		return
	}

	result := make([]OrgResponse, len(orgs))
	for i, org := range orgs {
		result[i] = OrgResponse{
			Organization:      org,
			MyRole:            roleByOrg[org.ID],
			MyEncryptedOrgKey: keyByOrg[org.ID],
			LogoURL:           presignLogoURL(ctx, org.LogoPath),
		}
	}
	c.JSON(http.StatusOK, result)
}

func (h *OrgHandler) GetOrg(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	var org pkg.Organization
	if err := h.DB.NewSelect().Model(&org).Where("id = ?", orgID).Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organization"})
		return
	}

	var membership pkg.OrgMember
	_ = h.DB.NewSelect().Model(&membership).
		Where("org_id = ? AND user_id = ?", orgID, userID).
		Scan(ctx)

	var adminGroupCount int
	_ = h.DB.NewSelect().
		TableExpr("org_group_members ogm").
		ColumnExpr("COUNT(*)").
		Join("JOIN org_groups og ON og.id = ogm.group_id").
		Where("og.org_id = ? AND ogm.user_id = ? AND ogm.role = 'admin'", orgID, userID).
		Scan(ctx, &adminGroupCount)

	c.JSON(http.StatusOK, OrgResponse{
		Organization:      org,
		MyRole:            role,
		MyEncryptedOrgKey: membership.EncryptedOrgKey,
		IsGroupAdmin:      adminGroupCount > 0,
		LogoURL:           presignLogoURL(ctx, org.LogoPath),
	})
}

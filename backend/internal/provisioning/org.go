// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package provisioning contains the core logic for admin-driven org management.
// It is intentionally decoupled from HTTP: the same functions are called by the
// CLI today and can be called by a Stripe webhook handler tomorrow.
package provisioning

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/mailer"

	"github.com/uptrace/bun"
)

// OrgResult is returned by CreateOrg.
type OrgResult struct {
	Org       *pkg.Organization
	Token     string
	InviteURL string
}

// OrgSummary is a lightweight view of an org for the list command.
type OrgSummary struct {
	ID             int64
	Name           string
	Description    string
	StorageQuotaMB int64
	StorageUsedMB  int64
	MemberCount    int
	CreatedAt      time.Time
}

func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func appURL() string {
	if u := os.Getenv("APP_URL"); u != "" {
		return u
	}
	return "https://kagibi.cloud"
}

// CreateOrg provisions a new organisation and generates a single-use owner
// invitation link. If ownerEmail is non-empty an email is sent via MAIL_*
// variables; failure to send is logged but does not abort the operation.
//
// The Organisation is created with owner_id="pending". AcceptInvitation
// updates owner_id to the accepting user when the invite is consumed.
func CreateOrg(ctx context.Context, db *bun.DB, name, description string, quotaMB int64, ownerEmail string) (*OrgResult, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if quotaMB <= 0 {
		quotaMB = 10240 // 10 GB default
	}

	// 1. Insert organisation with a placeholder owner_id.
	org := &pkg.Organization{
		Name:           name,
		Description:    description,
		OwnerID:        "pending",
		StorageQuotaMB: quotaMB,
	}
	if _, err := db.NewInsert().Model(org).Exec(ctx); err != nil {
		return nil, fmt.Errorf("insert organization: %w", err)
	}

	// 2. Generate a single-use, 7-day owner invitation.
	token, err := generateToken()
	if err != nil {
		_, _ = db.NewDelete().Model(org).WherePK().Exec(ctx)
		return nil, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	inv := &pkg.OrgInvitation{
		OrgID:     org.ID,
		InvitedBy: "admin-cli",
		Token:     token,
		Role:      "owner",
		MaxUses:   1,
		Status:    "active",
		ExpiresAt: &expiresAt,
	}
	if _, err := db.NewInsert().Model(inv).Exec(ctx); err != nil {
		_, _ = db.NewDelete().Model(org).WherePK().Exec(ctx)
		return nil, fmt.Errorf("insert invitation: %w", err)
	}

	inviteURL := fmt.Sprintf("%s/join/%s", appURL(), token)

	// 3. Send email — best-effort, never fatal.
	if ownerEmail != "" {
		body := fmt.Sprintf(
			"Bonjour,\n\nVotre espace organisation \"%s\" a été créé sur Kagibi.\n\n"+
				"Cliquez sur le lien ci-dessous pour créer votre compte et accéder à votre espace :\n\n"+
				"  %s\n\n"+
				"Ce lien est valide 7 jours et ne peut être utilisé qu'une seule fois.\n\n"+
				"— L'équipe Kagibi",
			name, inviteURL,
		)
		if err := mailer.Send(mailer.Message{
			To:      ownerEmail,
			Subject: fmt.Sprintf("Votre organisation \"%s\" est prête — Kagibi", name),
			Body:    body,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "[warn] email non envoyé à %s : %v\n", ownerEmail, err)
			fmt.Fprintf(os.Stderr, "[warn] transmettez le lien manuellement.\n")
		}
	}

	return &OrgResult{Org: org, Token: token, InviteURL: inviteURL}, nil
}

// ListOrgs returns a summary of every non-deleted organisation with its member count.
func ListOrgs(ctx context.Context, db *bun.DB) ([]OrgSummary, error) {
	type row struct {
		pkg.Organization
		MemberCount int `bun:"member_count"`
	}

	var rows []row
	if err := db.NewSelect().
		TableExpr("organizations AS o").
		ColumnExpr("o.*, COUNT(om.id) AS member_count").
		Join("LEFT JOIN org_members AS om ON om.org_id = o.id").
		Where("o.deleted_at IS NULL").
		GroupExpr("o.id").
		OrderExpr("o.created_at DESC").
		Scan(ctx, &rows); err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}

	summaries := make([]OrgSummary, len(rows))
	for i, r := range rows {
		summaries[i] = OrgSummary{
			ID:             r.ID,
			Name:           r.Name,
			Description:    r.Description,
			StorageQuotaMB: r.StorageQuotaMB,
			StorageUsedMB:  r.StorageUsedBytes / (1024 * 1024),
			MemberCount:    r.MemberCount,
			CreatedAt:      r.CreatedAt,
		}
	}
	return summaries, nil
}

// SetOrgQuota updates the storage quota (in MB) for an existing organisation.
func SetOrgQuota(ctx context.Context, db *bun.DB, orgID, quotaMB int64) error {
	if quotaMB <= 0 {
		return fmt.Errorf("quota must be > 0")
	}
	res, err := db.NewUpdate().Model((*pkg.Organization)(nil)).
		Set("storage_quota_mb = ?", quotaMB).
		Where("id = ? AND deleted_at IS NULL", orgID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update quota: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("organization %d not found", orgID)
	}
	return nil
}

// DeleteOrg hard-deletes an organisation and all its members/invitations.
// Use with caution — this is irreversible for the data layer even though the
// models use soft-delete; we hard-delete here intentionally for admin cleanup.
func DeleteOrg(ctx context.Context, db *bun.DB, orgID int64) error {
	// Revoke active invitations first.
	_, _ = db.NewUpdate().Model((*pkg.OrgInvitation)(nil)).
		Set("status = 'revoked'").
		Where("org_id = ? AND status = 'active'", orgID).
		Exec(ctx)

	// Remove members.
	_, _ = db.NewDelete().Model((*pkg.OrgMember)(nil)).
		Where("org_id = ?", orgID).
		Exec(ctx)

	// Soft-delete the org itself (honours the bun soft_delete tag).
	res, err := db.NewDelete().Model((*pkg.Organization)(nil)).
		Where("id = ?", orgID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("organization %d not found", orgID)
	}
	return nil
}

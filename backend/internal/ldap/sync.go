// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package ldap

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/emailcrypto"
	"kagibi/backend/pkg/mailer"

	"github.com/uptrace/bun"
)

// Syncer performs a full LDAP sync for one organisation.
type Syncer struct {
	db  *bun.DB
	cfg *pkg.OrgLDAPConfig
}

// NewSyncer builds a Syncer for the given organisation config.
func NewSyncer(db *bun.DB, cfg *pkg.OrgLDAPConfig) *Syncer {
	return &Syncer{db: db, cfg: cfg}
}

// Run executes a full sync cycle and persists the result back to org_ldap_configs.
func (s *Syncer) Run(ctx context.Context) {
	start := time.Now()
	stats, err := s.run(ctx)
	duration := int(time.Since(start).Milliseconds())
	if stats != nil {
		stats.DurationMs = duration
	}

	now := time.Now()
	errStr := ""
	if err != nil {
		errStr = err.Error()
		slog.Error("ldap.sync_error", "org_id", s.cfg.OrgID, "err", err)
	} else {
		slog.Info("ldap.sync_ok", "org_id", s.cfg.OrgID, "duration_ms", duration, "stats", stats)
	}

	update := map[string]any{
		"last_sync_at":    now,
		"last_sync_error": errStr,
		"updated_at":      now,
	}
	if stats != nil {
		update["last_sync_stats"] = stats
	}
	if _, dbErr := s.db.NewUpdate().
		TableExpr("org_ldap_configs").
		SetColumn("last_sync_at", "?", now).
		SetColumn("last_sync_error", "?", errStr).
		SetColumn("last_sync_stats", "?:type:jsonb", stats).
		SetColumn("updated_at", "?", now).
		Where("org_id = ?", s.cfg.OrgID).
		Exec(ctx); dbErr != nil {
		slog.Error("ldap.persist_result_failed", "org_id", s.cfg.OrgID, "err", dbErr)
	}
}

func (s *Syncer) run(ctx context.Context) (*pkg.LDAPSyncStats, error) {
	bindPassword, err := emailcrypto.Decrypt(s.cfg.BindPasswordEnc)
	if err != nil {
		return nil, fmt.Errorf("decrypt bind password: %w", err)
	}

	client := NewClient(s.cfg)
	conn, err := client.Dial()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := client.Bind(conn, bindPassword); err != nil {
		return nil, fmt.Errorf("ldap bind: %w", err)
	}

	ldapUsers, err := client.SearchUsers(conn)
	if err != nil {
		return nil, err
	}

	// Safety check 1: empty result most likely means a network or filter problem.
	if len(ldapUsers) < s.cfg.MinExpectedUsers {
		return nil, fmt.Errorf("ldap returned %d users, expected at least %d — aborting to protect existing members",
			len(ldapUsers), s.cfg.MinExpectedUsers)
	}

	ldapGroups, err := client.SearchGroups(conn)
	if err != nil {
		return nil, err
	}

	stats := &pkg.LDAPSyncStats{
		UsersFound: len(ldapUsers),
		GroupsFound: len(ldapGroups),
	}

	// Safety check 2: refuse if >20% of current LDAP members would be removed in one cycle.
	var currentLDAPCount int
	if err := s.db.NewSelect().TableExpr("org_members").
		ColumnExpr("COUNT(*)").
		Where("org_id = ? AND source = 'ldap' AND suspended_at IS NULL", s.cfg.OrgID).
		Scan(ctx, &currentLDAPCount); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("count existing ldap members: %w", err)
	}

	ldapEmailSet := make(map[string]LDAPUser, len(ldapUsers))
	ldapUIDSet := make(map[string]LDAPUser, len(ldapUsers))
	for _, u := range ldapUsers {
		ldapEmailSet[u.Email] = u
		ldapUIDSet[u.UID] = u
	}

	if currentLDAPCount > 0 {
		var existingLDAPMembers []struct {
			LdapUID string `bun:"ldap_uid"`
		}
		if err := s.db.NewSelect().TableExpr("org_members").
			ColumnExpr("ldap_uid").
			Where("org_id = ? AND source = 'ldap' AND suspended_at IS NULL AND ldap_uid != ''", s.cfg.OrgID).
			Scan(ctx, &existingLDAPMembers); err == nil {
			goneCount := 0
			for _, m := range existingLDAPMembers {
				if _, ok := ldapUIDSet[m.LdapUID]; !ok {
					goneCount++
				}
			}
			if currentLDAPCount > 5 && goneCount*100/currentLDAPCount > 20 {
				return nil, fmt.Errorf("ldap sync would remove %d/%d members (>20%%) — aborting (possible filter misconfiguration)",
					goneCount, currentLDAPCount)
			}
		}
	}

	// Process each LDAP user.
	for _, lu := range ldapUsers {
		if err := s.processUser(ctx, lu, stats); err != nil {
			slog.Warn("ldap.process_user_error", "org_id", s.cfg.OrgID, "email", lu.Email, "err", err)
			stats.UsersSkipped++
		}
	}

	// Deprovision members absent from the directory.
	if err := s.deprovisionAbsent(ctx, ldapUIDSet, stats); err != nil {
		slog.Error("ldap.deprovision_error", "org_id", s.cfg.OrgID, "err", err)
	}

	// Sync groups if configured.
	if len(ldapGroups) > 0 {
		// Build DN→UID map for group membership resolution.
		dnToUID := make(map[string]string, len(ldapUsers))
		for _, lu := range ldapUsers {
			dnToUID[lu.DN] = lu.UID
		}
		s.syncGroups(ctx, ldapGroups, dnToUID, stats)
	}

	return stats, nil
}

// processUser ensures a LDAP user is either an org member or has a pending invitation.
func (s *Syncer) processUser(ctx context.Context, lu LDAPUser, stats *pkg.LDAPSyncStats) error {
	// Find Kagibi account by email.
	var kagibiUser pkg.User
	if err := s.db.NewSelect().Model(&kagibiUser).
		Where("LOWER(email) = ?", lu.Email).
		Scan(ctx); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("lookup user by email: %w", err)
	}

	if kagibiUser.ID != "" {
		// User has a Kagibi account — check org membership.
		var member pkg.OrgMember
		err := s.db.NewSelect().Model(&member).
			Where("org_id = ? AND user_id = ?", s.cfg.OrgID, kagibiUser.ID).
			Scan(ctx)

		if err == sql.ErrNoRows {
			// Not yet a member — check for a pending invitation before creating one.
			pending, _ := s.db.NewSelect().TableExpr("org_invitations").
				ColumnExpr("id").
				Where("org_id = ? AND (target_user_id = ? OR notified_email_encrypted != '') AND status = 'active'", s.cfg.OrgID, kagibiUser.ID).
				Count(ctx)
			if pending > 0 {
				stats.UsersSkipped++
				return nil
			}
			if err := s.createInvitation(ctx, kagibiUser.ID, lu.Email, lu.DisplayName); err != nil {
				return err
			}
			stats.UsersInvited++
			return nil
		}
		if err != nil {
			return fmt.Errorf("check membership: %w", err)
		}

		// Already a member — tag as ldap source and unsuspend if needed.
		_, _ = s.db.NewUpdate().TableExpr("org_members").
			SetColumn("source", "?", "ldap").
			SetColumn("ldap_uid", "?", lu.UID).
			SetColumn("suspended_at", "NULL").
			Where("org_id = ? AND user_id = ?", s.cfg.OrgID, kagibiUser.ID).
			Exec(ctx)
		stats.UsersSkipped++ // already a member
		return nil
	}

	// No Kagibi account yet — create a link-type invitation.
	token, err := generateToken()
	if err != nil {
		return err
	}
	inv := &pkg.OrgInvitation{
		OrgID:     s.cfg.OrgID,
		InvitedBy: "ldap-sync",
		Token:     token,
		Role:      "member",
		MaxUses:   1,
		Status:    "active",
	}
	if _, err := s.db.NewInsert().Model(inv).Exec(ctx); err != nil {
		return fmt.Errorf("create invitation: %w", err)
	}

	// Send invite email asynchronously.
	joinURL := os.Getenv("APP_URL") + "/join/" + token
	go func() {
		if err := mailer.SendOrgInvite(lu.Email, "LDAP sync", getOrgName(s.db, s.cfg.OrgID), "member", joinURL, "fr"); err != nil {
			slog.Warn("ldap.invite_email_failed", "email", lu.Email, "err", err)
		}
	}()
	stats.UsersInvited++
	return nil
}

// createInvitation creates a direct invitation for an existing Kagibi user.
func (s *Syncer) createInvitation(ctx context.Context, targetUserID, email, displayName string) error {
	token, err := generateToken()
	if err != nil {
		return err
	}
	inv := &pkg.OrgInvitation{
		OrgID:        s.cfg.OrgID,
		InvitedBy:    "ldap-sync",
		Token:        token,
		TargetUserID: &targetUserID,
		Role:         "member",
		MaxUses:      1,
		Status:       "active",
	}
	if _, err := s.db.NewInsert().Model(inv).Exec(ctx); err != nil {
		return fmt.Errorf("create direct invitation: %w", err)
	}

	joinURL := os.Getenv("APP_URL") + "/join/" + token
	go func() {
		if err := mailer.SendOrgInvite(email, "LDAP sync", getOrgName(s.db, s.cfg.OrgID), "member", joinURL, "fr"); err != nil {
			slog.Warn("ldap.direct_invite_email_failed", "email", email, "err", err)
		}
	}()
	return nil
}

// deprovisionAbsent suspends LDAP members who no longer appear in the directory,
// and removes those past the auto-deprovision grace period.
func (s *Syncer) deprovisionAbsent(ctx context.Context, ldapUIDSet map[string]LDAPUser, stats *pkg.LDAPSyncStats) error {
	var ldapMembers []struct {
		UserID  string     `bun:"user_id"`
		LdapUID string     `bun:"ldap_uid"`
		SuspAt  *time.Time `bun:"suspended_at"`
	}
	if err := s.db.NewSelect().TableExpr("org_members").
		ColumnExpr("user_id, ldap_uid, suspended_at").
		Where("org_id = ? AND source = 'ldap'", s.cfg.OrgID).
		Scan(ctx, &ldapMembers); err != nil {
		return fmt.Errorf("list ldap members: %w", err)
	}

	now := time.Now()
	for _, m := range ldapMembers {
		if _, ok := ldapUIDSet[m.LdapUID]; ok {
			continue // still in directory
		}

		if m.SuspAt == nil {
			// First absence — suspend.
			if _, err := s.db.NewUpdate().TableExpr("org_members").
				SetColumn("suspended_at", "?", now).
				Where("org_id = ? AND user_id = ?", s.cfg.OrgID, m.UserID).
				Exec(ctx); err != nil {
				slog.Error("ldap.suspend_failed", "org_id", s.cfg.OrgID, "user_id", m.UserID, "err", err)
				continue
			}
			stats.UsersSuspended++
			slog.Info("ldap.user_suspended", "org_id", s.cfg.OrgID, "user_id", m.UserID)
			continue
		}

		// Already suspended — check grace period.
		if s.cfg.AutoDeprovisionDays <= 0 {
			continue // manual-only deprovisioning
		}
		deadline := m.SuspAt.Add(time.Duration(s.cfg.AutoDeprovisionDays) * 24 * time.Hour)
		if now.Before(deadline) {
			continue
		}

		// Grace period expired — remove member.
		if err := s.removeMember(ctx, m.UserID); err != nil {
			slog.Error("ldap.remove_failed", "org_id", s.cfg.OrgID, "user_id", m.UserID, "err", err)
			continue
		}
		stats.UsersDeleted++
		slog.Info("ldap.user_removed", "org_id", s.cfg.OrgID, "user_id", m.UserID, "reason", "grace_period_expired")
	}
	return nil
}

func (s *Syncer) removeMember(ctx context.Context, userID string) error {
	if _, err := s.db.NewDelete().TableExpr("org_folder_permissions").
		Where("org_id = ? AND user_id = ?", s.cfg.OrgID, userID).Exec(ctx); err != nil {
		return fmt.Errorf("remove folder permissions: %w", err)
	}
	// Delete group memberships for groups belonging to this org.
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM org_group_members ogm
		 USING org_groups og
		 WHERE og.id = ogm.group_id AND og.org_id = ? AND ogm.user_id = ?`,
		s.cfg.OrgID, userID,
	); err != nil {
		return fmt.Errorf("remove group memberships: %w", err)
	}
	if _, err := s.db.NewDelete().TableExpr("org_members").
		Where("org_id = ? AND user_id = ?", s.cfg.OrgID, userID).Exec(ctx); err != nil {
		return fmt.Errorf("remove org member: %w", err)
	}
	return nil
}

// syncGroups creates or updates org groups from LDAP groups.
func (s *Syncer) syncGroups(ctx context.Context, groups []LDAPGroup, dnToUID map[string]string, stats *pkg.LDAPSyncStats) {
	for _, lg := range groups {
		s.syncGroup(ctx, lg, dnToUID, stats)
	}
}

func (s *Syncer) syncGroup(ctx context.Context, lg LDAPGroup, dnToUID map[string]string, stats *pkg.LDAPSyncStats) {
	var group pkg.OrgGroup
	err := s.db.NewSelect().Model(&group).
		Where("org_id = ? AND ldap_dn = ?", s.cfg.OrgID, lg.DN).
		Scan(ctx)

	now := time.Now()
	if err == sql.ErrNoRows {
		group = pkg.OrgGroup{
			OrgID:        s.cfg.OrgID,
			Name:         lg.Name,
			Source:       "ldap",
			LdapDN:       lg.DN,
			LastSyncedAt: &now,
			CreatedBy:    "ldap-sync",
		}
		if _, err := s.db.NewInsert().Model(&group).Exec(ctx); err != nil {
			slog.Error("ldap.create_group_failed", "org_id", s.cfg.OrgID, "group", lg.Name, "err", err)
			return
		}
		stats.GroupsCreated++
	} else if err != nil {
		slog.Error("ldap.lookup_group_failed", "org_id", s.cfg.OrgID, "group", lg.Name, "err", err)
		return
	} else {
		// Update name and sync timestamp.
		if _, err := s.db.NewUpdate().Model(&group).
			SetColumn("name", "?", lg.Name).
			SetColumn("last_synced_at", "?", now).
			SetColumn("updated_at", "?", now).
			Where("id = ?", group.ID).
			Exec(ctx); err != nil {
			slog.Warn("ldap.update_group_failed", "org_id", s.cfg.OrgID, "group", lg.Name, "err", err)
		}
		stats.GroupsUpdated++
	}

	// Rebuild group membership: add new members, remove departed ones.
	// Only consider org members who have a Kagibi account.
	newMemberUIDs := make(map[string]bool)
	for _, memberDN := range lg.Members {
		uid, ok := dnToUID[memberDN]
		if !ok {
			continue
		}
		// Resolve LDAP UID → Kagibi user ID.
		var kagibiUserID string
		if err := s.db.NewSelect().TableExpr("org_members").
			ColumnExpr("user_id").
			Where("org_id = ? AND ldap_uid = ?", s.cfg.OrgID, uid).
			Scan(ctx, &kagibiUserID); err != nil {
			continue
		}
		newMemberUIDs[kagibiUserID] = true
		// Upsert group membership.
		gm := &pkg.OrgGroupMember{
			GroupID: group.ID,
			UserID:  kagibiUserID,
			Role:    "member",
		}
		_, _ = s.db.NewInsert().Model(gm).
			On("CONFLICT (group_id, user_id) DO NOTHING").
			Exec(ctx)
	}

	// Remove members no longer in the LDAP group.
	var existing []struct {
		UserID string `bun:"user_id"`
	}
	if err := s.db.NewSelect().TableExpr("org_group_members").
		ColumnExpr("user_id").
		Where("group_id = ?", group.ID).
		Scan(ctx, &existing); err == nil {
		for _, m := range existing {
			if !newMemberUIDs[m.UserID] {
				_, _ = s.db.NewDelete().TableExpr("org_group_members").
					Where("group_id = ? AND user_id = ?", group.ID, m.UserID).
					Exec(ctx)
			}
		}
	}
}

// getOrgName returns the organisation name for a given ID (for invitation emails).
func getOrgName(db *bun.DB, orgID int64) string {
	var org struct {
		Name string `bun:"name"`
	}
	_ = db.NewSelect().TableExpr("organizations").ColumnExpr("name").
		Where("id = ?", orgID).Scan(context.Background(), &org)
	return org.Name
}

func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// SyncNow is a convenience wrapper for one-off manual sync calls from HTTP handlers.
func SyncNow(ctx context.Context, db *bun.DB, cfg *pkg.OrgLDAPConfig) (*pkg.LDAPSyncStats, error) {
	s := NewSyncer(db, cfg)
	start := time.Now()
	stats, runErr := s.run(ctx)
	if stats != nil {
		stats.DurationMs = int(time.Since(start).Milliseconds())
	}
	return stats, runErr
}

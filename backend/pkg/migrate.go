// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// internal/migrate.go
package pkg

import (
	"context"
	"fmt"
	"log"

	"kagibi/backend/pkg/emailcrypto"

	"github.com/uptrace/bun"
)

func Migrate(db *bun.DB) error {
	ctx := context.Background()

	if err := migrateAuthUsers(ctx, db); err != nil {
		return err
	}
	if err := migrateCoreModels(ctx, db); err != nil {
		return err
	}
	if err := migrateSchemaAlterations(ctx, db); err != nil {
		return err
	}
	if err := migrateCoreIndices(ctx, db); err != nil {
		return err
	}
	if err := migrateBillingTables(ctx, db); err != nil {
		return err
	}
	if err := migrateUserSettings(ctx, db); err != nil {
		return err
	}
	if err := migrateEmailEncryption(ctx, db); err != nil {
		return err
	}
	if err := migrateOrganizationTables(ctx, db); err != nil {
		return err
	}
	if err := migrateOrgTags(ctx, db); err != nil {
		return err
	}
	if err := migrateOrgFavorites(ctx, db); err != nil {
		return err
	}
	if err := migrateOrgTrashColumns(ctx, db); err != nil {
		return err
	}
	if err := migrateNewFeatureColumns(ctx, db); err != nil {
		return err
	}
	migrateEnsurePartialOrgIndexes(ctx, db)
	if err := migrateFilesUniqueIndex(ctx, db); err != nil {
		return err
	}
	migrateChunkSizeColumns(ctx, db)

	return nil
}

func migrateAuthUsers(ctx context.Context, db *bun.DB) error {
	// auth_users table — used by LocalProvider (AUTH_PROVIDER=local)
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "auth_users" (
		"id"            TEXT PRIMARY KEY,
		"email"         VARCHAR(255) UNIQUE NOT NULL,
		"password_hash" VARCHAR(255) NOT NULL,
		"created_at"    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return fmt.Errorf("failed to create auth_users table: %w", err)
	}
	_, err = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_auth_users_email ON auth_users (email);`)
	if err != nil {
		log.Printf("Warning: failed to create idx_auth_users_email: %v", err)
	}

	// TOTP MFA columns (added after initial auth_users creation)
	for _, col := range []string{
		`ALTER TABLE "auth_users" ADD COLUMN IF NOT EXISTS "totp_secret"          VARCHAR`,
		`ALTER TABLE "auth_users" ADD COLUMN IF NOT EXISTS "totp_enabled"         BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE "auth_users" ADD COLUMN IF NOT EXISTS "totp_factor_id"       VARCHAR`,
		`ALTER TABLE "auth_users" ADD COLUMN IF NOT EXISTS "totp_friendly_name"   VARCHAR`,
		`ALTER TABLE "auth_users" ADD COLUMN IF NOT EXISTS "totp_last_code"       VARCHAR`,
		`ALTER TABLE "auth_users" ADD COLUMN IF NOT EXISTS "totp_last_code_at"    TIMESTAMPTZ`,
		`ALTER TABLE "auth_users" ADD COLUMN IF NOT EXISTS "totp_failed_attempts" INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE "auth_users" ADD COLUMN IF NOT EXISTS "totp_locked_until"    TIMESTAMPTZ`,
	} {
		if _, err := db.ExecContext(ctx, col); err != nil {
			log.Printf("Warning: failed to add TOTP column: %v", err)
		}
	}
	return nil
}

func migrateCoreModels(ctx context.Context, db *bun.DB) error {
	// Crée les tables si elles n'existent pas
	models := []any{
		(*User)(nil),
		(*Friendship)(nil),
		(*UserPlan)(nil),
		(*File)(nil),
		(*Folder)(nil),
		(*Tag)(nil),
		(*ShareLink)(nil),
		(*ShareFileKey)(nil),
		(*ImportedShare)(nil),
		(*FileShare)(nil),
		(*FolderShare)(nil),
		(*FolderFileKey)(nil),
		(*FolderFolderKey)(nil),
		(*RecentActivity)(nil),
		(*P2PSignal)(nil),
		(*P2PInvite)(nil),
		(*RealtimeEvent)(nil),
		(*Organization)(nil),
		(*OrgMember)(nil),
		(*OrgInvitation)(nil),
		(*OrgFolder)(nil),
		(*OrgFile)(nil),
		(*OrgFolderPermission)(nil),
		(*OrgAuditLog)(nil),
		(*OrgGroup)(nil),
		(*OrgGroupMember)(nil),
		(*OrgGroupPermission)(nil),
	}

	for _, model := range models {
		_, err := db.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	return nil
}

func migrateSchemaAlterations(ctx context.Context, db *bun.DB) error {
	// Manually add the 'path' column to 'share_links' if it doesn't exist
	_, err := db.ExecContext(ctx, `ALTER TABLE "share_links" ADD COLUMN IF NOT EXISTS "path" VARCHAR NOT NULL DEFAULT '';`)
	if err != nil {
		return fmt.Errorf("failed to add path column to share_links: %w", err)
	}

	// Add encrypted_key to files
	_, err = db.ExecContext(ctx, `ALTER TABLE "files" ADD COLUMN IF NOT EXISTS "encrypted_key" VARCHAR;`)
	if err != nil {
		return fmt.Errorf("failed to add encrypted_key column to files: %w", err)
	}

	// Add encrypted_key to folders
	_, err = db.ExecContext(ctx, `ALTER TABLE "folders" ADD COLUMN IF NOT EXISTS "encrypted_key" VARCHAR;`)
	if err != nil {
		return fmt.Errorf("failed to add encrypted_key column to folders: %w", err)
	}

	// Add encrypted_key to share_links
	_, err = db.ExecContext(ctx, `ALTER TABLE "share_links" ADD COLUMN IF NOT EXISTS "encrypted_key" VARCHAR;`)
	if err != nil {
		return fmt.Errorf("failed to add encrypted_key column to share_links: %w", err)
	}

	// Add encrypted_key to folder_shares
	_, err = db.ExecContext(ctx, `ALTER TABLE "folder_shares" ADD COLUMN IF NOT EXISTS "encrypted_key" VARCHAR;`)
	if err != nil {
		return fmt.Errorf("failed to add encrypted_key column to folder_shares: %w", err)
	}

	// Folder sizes table (with FK to folders)
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "folder_sizes" (
		"folder_id" BIGINT PRIMARY KEY REFERENCES "folders"("id") ON DELETE CASCADE,
		"user_id" VARCHAR NOT NULL,
		"size_bytes" BIGINT NOT NULL DEFAULT 0,
		"updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return fmt.Errorf("failed to create folder_sizes table: %w", err)
	}

	// Initialize folder sizes if table is empty
	if err := EnsureFolderSizesInitialized(ctx, db); err != nil {
		log.Printf("Warning: failed to initialize folder sizes: %v", err)
	}

	// Update existing rows to have a default path if it's empty
	_, err = db.ExecContext(ctx, `UPDATE "share_links" SET "path" = '' WHERE "path" IS NULL;`)
	if err != nil {
		return fmt.Errorf("failed to update existing paths in share_links: %w", err)
	}

	// Guest invite columns — added when the invite flow was extended to support
	// users without a Kagibi account (guest auto-auth via ephemeral JWT).
	for _, col := range []string{
		`ALTER TABLE "p2p_invites" ADD COLUMN IF NOT EXISTS "is_guest" BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE "p2p_invites" ALTER COLUMN "recipient_email" DROP NOT NULL`,
		// Single-use guard: set atomically on first guest-auth call to prevent token reuse.
		`ALTER TABLE "p2p_invites" ADD COLUMN IF NOT EXISTS "guest_authed_at" TIMESTAMPTZ`,
	} {
		if _, err := db.ExecContext(ctx, col); err != nil {
			log.Printf("Warning: failed to apply p2p_invites guest column: %v", err)
		}
	}

	// Org file public share: link a share_link back to its source org
	if _, err := db.ExecContext(ctx,
		`ALTER TABLE "share_links" ADD COLUMN IF NOT EXISTS "org_id" BIGINT`,
	); err != nil {
		log.Printf("Warning: failed to add org_id column to share_links: %v", err)
	}

	// Single-use link columns
	for _, col := range []string{
		`ALTER TABLE "share_links" ADD COLUMN IF NOT EXISTS "single_use" BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE "share_links" ADD COLUMN IF NOT EXISTS "used_at" TIMESTAMPTZ`,
	} {
		if _, err := db.ExecContext(ctx, col); err != nil {
			log.Printf("Warning: failed to add single_use/used_at column to share_links: %v", err)
		}
	}

	// restrict_to_groups column added after initial table creation
	if _, err := db.ExecContext(ctx,
		`ALTER TABLE "org_group_permissions" ADD COLUMN IF NOT EXISTS "restrict_to_groups" BOOLEAN NOT NULL DEFAULT false`,
	); err != nil {
		log.Printf("Warning: failed to add restrict_to_groups column: %v", err)
	}

	// role column for group members (admin | member)
	if _, err := db.ExecContext(ctx,
		`ALTER TABLE "org_group_members" ADD COLUMN IF NOT EXISTS "role" VARCHAR NOT NULL DEFAULT 'member'`,
	); err != nil {
		log.Printf("Warning: failed to add role column to org_group_members: %v", err)
	}

	// Share permissions columns
	for _, col := range []string{
		`ALTER TABLE "share_links"    ADD COLUMN IF NOT EXISTS "perm_download" BOOLEAN NOT NULL DEFAULT true`,
		`ALTER TABLE "share_links"    ADD COLUMN IF NOT EXISTS "perm_create"   BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE "share_links"    ADD COLUMN IF NOT EXISTS "perm_delete"   BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE "share_links"    ADD COLUMN IF NOT EXISTS "perm_move"     BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE "file_shares"    ADD COLUMN IF NOT EXISTS "perm_download" BOOLEAN NOT NULL DEFAULT true`,
		`ALTER TABLE "folder_shares"  ADD COLUMN IF NOT EXISTS "perm_download" BOOLEAN NOT NULL DEFAULT true`,
		`ALTER TABLE "folder_shares"  ADD COLUMN IF NOT EXISTS "perm_create"   BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE "folder_shares"  ADD COLUMN IF NOT EXISTS "perm_delete"   BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE "folder_shares"  ADD COLUMN IF NOT EXISTS "perm_move"     BOOLEAN NOT NULL DEFAULT false`,
	} {
		if _, err := db.ExecContext(ctx, col); err != nil {
			log.Printf("Warning: failed to add share permission column: %v", err)
		}
	}

	// Per-item access overrides within a shared folder
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "share_item_overrides" (
		"id"           BIGSERIAL PRIMARY KEY,
		"share_id"     BIGINT NOT NULL REFERENCES "share_links"("id") ON DELETE CASCADE,
		"item_path"    VARCHAR NOT NULL,
		"item_type"    VARCHAR NOT NULL CHECK (item_type IN ('file', 'folder')),
		"access_level" VARCHAR NOT NULL DEFAULT 'full' CHECK (access_level IN ('full', 'readonly', 'none')),
		"can_delete"   BOOLEAN NOT NULL DEFAULT true,
		"can_download" BOOLEAN NOT NULL DEFAULT true,
		UNIQUE("share_id", "item_path")
	);`)
	if err != nil {
		log.Printf("Warning: failed to create share_item_overrides table: %v", err)
	}

	// can_download added to share_item_overrides after initial table creation
	if _, err := db.ExecContext(ctx, `ALTER TABLE "share_item_overrides" ADD COLUMN IF NOT EXISTS "can_download" BOOLEAN NOT NULL DEFAULT true`); err != nil {
		log.Printf("Warning: failed to add can_download column to share_item_overrides: %v", err)
	}

	// Org E2E encryption columns — added when the organisation crypto layer was introduced.
	// Tables created before this migration will be missing these columns.
	for _, col := range []string{
		`ALTER TABLE "org_members"     ADD COLUMN IF NOT EXISTS "encrypted_org_key"        TEXT`,
		`ALTER TABLE "org_invitations" ADD COLUMN IF NOT EXISTS "encrypted_org_key"        TEXT`,
		`ALTER TABLE "org_invitations" ADD COLUMN IF NOT EXISTS "notified_email_encrypted" TEXT`,
	} {
		if _, err := db.ExecContext(ctx, col); err != nil {
			log.Printf("Warning: failed to add encrypted_org_key column: %v", err)
		}
	}

	// The original p2p_signals_signal_type_check constraint omitted 'reject' and 'p2p_ping'.
	// Drop it and replace with the full allowed set so reject signals can be stored.
	_, err = db.ExecContext(ctx, `ALTER TABLE "p2p_signals" DROP CONSTRAINT IF EXISTS "p2p_signals_signal_type_check"`)
	if err != nil {
		log.Printf("Warning: failed to drop p2p_signals_signal_type_check: %v", err)
	}
	_, err = db.ExecContext(ctx, `ALTER TABLE "p2p_signals" ADD CONSTRAINT "p2p_signals_signal_type_check"
		CHECK (signal_type IN ('offer', 'answer', 'candidate', 'reject', 'p2p_ping', 'invite_accepted'))`)
	if err != nil {
		log.Printf("Warning: failed to recreate p2p_signals_signal_type_check: %v", err)
	}

	return nil
}

func migrateCoreIndices(ctx context.Context, db *bun.DB) error {
	for _, idx := range []string{
		// Files
		`CREATE INDEX IF NOT EXISTS idx_files_user_path   ON files (user_id, path)`,
		`CREATE INDEX IF NOT EXISTS idx_files_is_preview  ON files (is_preview)`,
		// Folders
		`CREATE INDEX IF NOT EXISTS idx_folders_user_path ON folders (user_id, path)`,
		// Folder sizes
		`CREATE INDEX IF NOT EXISTS idx_folder_sizes_user ON folder_sizes (user_id)`,
		// Shares
		`CREATE INDEX IF NOT EXISTS idx_folder_shares_user    ON folder_shares (shared_with_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_file_shares_user      ON file_shares (shared_with_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_folder_file_keys_file ON folder_file_keys (file_id)`,
		// P2P signals
		`CREATE INDEX IF NOT EXISTS idx_p2p_signals_created ON p2p_signals (created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_p2p_signals_target  ON p2p_signals (target_id, consumed)`,
		// Profiles
		`CREATE INDEX IF NOT EXISTS idx_profiles_avatar_url ON profiles (avatar_url)`,
		// Realtime events
		`CREATE INDEX IF NOT EXISTS idx_realtime_events_created_at ON realtime_events (created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_realtime_events_user_id    ON realtime_events (user_id)`,
		// Recent activities
		`CREATE INDEX IF NOT EXISTS idx_recent_activities_user_date ON recent_activities (user_id, accessed_at DESC)`,
	} {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			log.Printf("Warning: failed to create index: %v", err)
		}
	}
	return nil
}

func migrateBillingTables(ctx context.Context, db *bun.DB) error {
	// Add synced column to files (desktop sync indicator)
	_, err := db.ExecContext(ctx, `ALTER TABLE "files" ADD COLUMN IF NOT EXISTS "synced" BOOLEAN NOT NULL DEFAULT false;`)
	if err != nil {
		log.Printf("Warning: failed to add synced column to files: %v", err)
	}

	// Lago billing invoices
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "billing_invoices" (
		"id"                 VARCHAR PRIMARY KEY,
		"lago_invoice_id"    VARCHAR UNIQUE NOT NULL,
		"user_id"            VARCHAR NOT NULL,
		"invoice_number"     VARCHAR NOT NULL,
		"status"             VARCHAR NOT NULL DEFAULT 'draft',
		"payment_status"     VARCHAR NOT NULL DEFAULT 'pending',
		"currency"           VARCHAR NOT NULL DEFAULT 'EUR',
		"total_amount_cents" BIGINT NOT NULL DEFAULT 0,
		"payment_link_id"    VARCHAR,
		"payment_link_url"   VARCHAR,
		"issuing_date"       TIMESTAMPTZ NOT NULL,
		"created_at"         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updated_at"         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		log.Printf("Warning: failed to create billing_invoices table: %v", err)
	}

	// Stripe customers
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "stripe_customers" (
		"id"                    UUID DEFAULT gen_random_uuid() PRIMARY KEY,
		"stripe_customer_id"    TEXT UNIQUE NOT NULL,
		"kagibi_user_id"    TEXT UNIQUE NOT NULL,
		"email"                 TEXT NOT NULL,
		"name"                  TEXT,
		"metadata"              JSONB DEFAULT '{}',
		"created_at"            TIMESTAMPTZ NOT NULL DEFAULT now(),
		"updated_at"            TIMESTAMPTZ NOT NULL DEFAULT now()
	);`)
	if err != nil {
		log.Printf("Warning: failed to create stripe_customers table: %v", err)
	}

	// Stripe invoices (FK to stripe_customers)
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "stripe_invoices" (
		"id"                  UUID DEFAULT gen_random_uuid() PRIMARY KEY,
		"stripe_customer_id"  TEXT NOT NULL REFERENCES "stripe_customers"("stripe_customer_id") ON DELETE CASCADE,
		"stripe_invoice_id"   TEXT UNIQUE NOT NULL,
		"number"              TEXT UNIQUE NOT NULL,
		"status"              TEXT NOT NULL,
		"amount_cents"        INTEGER NOT NULL,
		"currency"            TEXT NOT NULL DEFAULT 'EUR',
		"issued_at"           TIMESTAMPTZ,
		"paid_at"             TIMESTAMPTZ,
		"download_url"        TEXT,
		"payment_url"         TEXT,
		"created_at"          TIMESTAMPTZ NOT NULL DEFAULT now(),
		"updated_at"          TIMESTAMPTZ NOT NULL DEFAULT now()
	);`)
	if err != nil {
		log.Printf("Warning: failed to create stripe_invoices table: %v", err)
	}

	// Stripe subscriptions (FK to stripe_customers)
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "stripe_subscriptions" (
		"id"                       UUID DEFAULT gen_random_uuid() PRIMARY KEY,
		"stripe_customer_id"       TEXT NOT NULL REFERENCES "stripe_customers"("stripe_customer_id") ON DELETE CASCADE,
		"stripe_subscription_id"   TEXT UNIQUE NOT NULL,
		"plan_code"                TEXT NOT NULL,
		"status"                   TEXT NOT NULL DEFAULT 'active',
		"current_period_start"     TIMESTAMPTZ NOT NULL,
		"current_period_end"       TIMESTAMPTZ NOT NULL,
		"cancel_at_period_end"     BOOLEAN DEFAULT false,
		"canceled_at"              TIMESTAMPTZ,
		"created_at"               TIMESTAMPTZ NOT NULL DEFAULT now(),
		"updated_at"               TIMESTAMPTZ NOT NULL DEFAULT now()
	);`)
	if err != nil {
		log.Printf("Warning: failed to create stripe_subscriptions table: %v", err)
	}

	// Usage events (FK to stripe_customers)
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "usage_events" (
		"id"                  UUID DEFAULT gen_random_uuid() PRIMARY KEY,
		"stripe_customer_id"  TEXT NOT NULL REFERENCES "stripe_customers"("stripe_customer_id") ON DELETE CASCADE,
		"event_type"          TEXT NOT NULL,
		"bytes_delta"         BIGINT NOT NULL,
		"idempotency_key"     TEXT UNIQUE,
		"metadata"            JSONB DEFAULT '{}',
		"timestamp"           TIMESTAMPTZ NOT NULL DEFAULT now(),
		"created_at"          TIMESTAMPTZ NOT NULL DEFAULT now()
	);`)
	if err != nil {
		log.Printf("Warning: failed to create usage_events table: %v", err)
	}

	// Webhook events (Stripe webhook log)
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "webhook_events" (
		"id"                  UUID DEFAULT gen_random_uuid() PRIMARY KEY,
		"event_type"          TEXT NOT NULL,
		"stripe_customer_id"  TEXT,
		"kagibi_user_id"  TEXT,
		"payload"             JSONB NOT NULL,
		"relayed_at"          TIMESTAMPTZ,
		"relay_status"        TEXT,
		"relay_error"         TEXT,
		"created_at"          TIMESTAMPTZ NOT NULL DEFAULT now()
	);`)
	if err != nil {
		log.Printf("Warning: failed to create webhook_events table: %v", err)
	}

	// Colonnes ajoutées après la création initiale des tables — idempotentes via IF NOT EXISTS.
	// stripe_customers.kagibi_user_id : absent des déploiements antérieurs au schéma courant.
	// webhook_events.kagibi_user_id  : idem.
	for _, alteration := range []string{
		`ALTER TABLE "stripe_customers" ADD COLUMN IF NOT EXISTS "kagibi_user_id" TEXT`,
		`ALTER TABLE "webhook_events"   ADD COLUMN IF NOT EXISTS "kagibi_user_id" TEXT`,
	} {
		if _, err := db.ExecContext(ctx, alteration); err != nil {
			log.Printf("Warning: failed to alter billing table: %v", err)
		}
	}

	// Stripe-related indices
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS idx_billing_invoices_user           ON billing_invoices (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_billing_invoices_payment_status ON billing_invoices (payment_status)`,
		`CREATE INDEX IF NOT EXISTS idx_stripe_customers_stripe_id      ON stripe_customers (stripe_customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stripe_customers_user_id        ON stripe_customers (kagibi_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stripe_invoices_customer        ON stripe_invoices (stripe_customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stripe_invoices_status          ON stripe_invoices (status)`,
		`CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_customer   ON stripe_subscriptions (stripe_customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_status     ON stripe_subscriptions (status)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_customer_time             ON usage_events (stripe_customer_id, "timestamp")`,
		`CREATE INDEX IF NOT EXISTS idx_usage_idempotency               ON usage_events (idempotency_key)`,
		`CREATE INDEX IF NOT EXISTS idx_webhook_events_type             ON webhook_events (event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_webhook_events_user             ON webhook_events (kagibi_user_id)`,
	} {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			log.Printf("Warning: failed to create index: %v", err)
		}
	}
	return nil
}

func migrateUserSettings(ctx context.Context, db *bun.DB) error {
	// --- RGPD Article 17 - Droit à l'effacement ---

	// Add deleted_at column to profiles table for soft delete
	_, err := db.ExecContext(ctx, `ALTER TABLE "profiles" ADD COLUMN IF NOT EXISTS "deleted_at" TIMESTAMPTZ;`)
	if err != nil {
		log.Printf("Warning: failed to add deleted_at column to profiles: %v", err)
	}

	_, err = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_profiles_deleted_at ON profiles (deleted_at) WHERE deleted_at IS NULL;`)
	if err != nil {
		log.Printf("Warning: failed to create idx_profiles_deleted_at: %v", err)
	}

	// --- USER PLANS ---
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "user_plans" (
		"user_id"             VARCHAR PRIMARY KEY,
		"plan"                VARCHAR NOT NULL DEFAULT 'free',
		"storage_limit"       BIGINT NOT NULL DEFAULT 21474836480,
		"storage_used"        BIGINT NOT NULL DEFAULT 0,
		"p2p_max_exchanges"   INTEGER NOT NULL DEFAULT 5,
		"p2p_exchanges_used"  INTEGER NOT NULL DEFAULT 0,
		"created_at"          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updated_at"          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		log.Printf("Warning: failed to create user_plans table: %v", err)
	}

	_, err = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_user_plans_plan ON user_plans (plan);`)
	if err != nil {
		log.Printf("Warning: failed to create idx_user_plans_plan: %v", err)
	}

	// Backfill missing user_plans rows from profiles (best-effort; profiles may not have legacy plan columns)
	_, err = db.ExecContext(ctx, `INSERT INTO user_plans (user_id, plan, storage_limit, storage_used, p2p_max_exchanges, p2p_exchanges_used)
		SELECT p.id, 'free', 21474836480, 0, 5, 0
		FROM profiles p
		ON CONFLICT (user_id) DO NOTHING;`)
	if err != nil {
		log.Printf("Warning: failed to backfill user_plans: %v", err)
	}

	// --- USER SECURITY SETTINGS ---
	// Stores MFA preferences and state per user.
	// mfa_enabled / mfa_verified are kept in sync by the MFA handlers (local mode)
	// or by the frontend (Supabase mode). The require_mfa_* columns are user-controlled.
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "user_security_settings" (
		"user_id"                            VARCHAR PRIMARY KEY,
		"mfa_enabled"                        BOOLEAN NOT NULL DEFAULT false,
		"mfa_verified"                       BOOLEAN NOT NULL DEFAULT false,
		"require_mfa_on_login"               BOOLEAN NOT NULL DEFAULT false,
		"require_mfa_on_destructive_actions" BOOLEAN NOT NULL DEFAULT false,
		"require_mfa_on_downloads"           BOOLEAN NOT NULL DEFAULT false,
		"created_at"                         TIMESTAMPTZ NOT NULL DEFAULT now(),
		"updated_at"                         TIMESTAMPTZ NOT NULL DEFAULT now()
	);`)
	if err != nil {
		log.Printf("Warning: failed to create user_security_settings table: %v", err)
	}

	// Add created_at / updated_at to existing user_security_settings installs that predate this column
	for _, col := range []string{
		`ALTER TABLE "user_security_settings" DROP CONSTRAINT IF EXISTS "user_security_settings_user_id_fkey"`,
		`ALTER TABLE "user_plans" DROP CONSTRAINT IF EXISTS "user_plans_user_id_fkey"`,
		`ALTER TABLE "user_security_settings" ADD COLUMN IF NOT EXISTS "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()`,
		`ALTER TABLE "user_security_settings" ADD COLUMN IF NOT EXISTS "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now()`,
	} {
		if _, err := db.ExecContext(ctx, col); err != nil {
			log.Printf("Warning: failed to add timestamp column to user_security_settings: %v", err)
		}
	}

	_, err = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_security_settings_user_id ON user_security_settings (user_id);`)
	if err != nil {
		log.Printf("Warning: failed to create idx_security_settings_user_id: %v", err)
	}

	return nil
}

func migrateOrganizationTables(ctx context.Context, db *bun.DB) error {
	// storage_used_bytes added after initial org table creation (Phase 2)
	if _, err := db.ExecContext(ctx,
		`ALTER TABLE "organizations" ADD COLUMN IF NOT EXISTS "storage_used_bytes" BIGINT NOT NULL DEFAULT 0`,
	); err != nil {
		log.Printf("Warning: failed to add storage_used_bytes to organizations: %v", err)
	}

	// logo_path stores the S3 object key for the org's custom logo (empty = no logo)
	if _, err := db.ExecContext(ctx,
		`ALTER TABLE "organizations" ADD COLUMN IF NOT EXISTS "logo_path" TEXT NOT NULL DEFAULT ''`,
	); err != nil {
		log.Printf("Warning: failed to add logo_path to organizations: %v", err)
	}

	// Unique membership: one row per (org, user) pair
	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_org_members_org_user ON org_members (org_id, user_id)`,
	); err != nil {
		log.Printf("Warning: failed to create org_members unique index: %v", err)
	}

	// Unique file path per org (prevents duplicate paths in same org)
	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_org_files_org_path ON org_files (org_id, path) WHERE deleted_at IS NULL`,
	); err != nil {
		log.Printf("Warning: failed to create org_files unique path index: %v", err)
	}

	// Unique folder path per org
	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_org_folders_org_path ON org_folders (org_id, path) WHERE deleted_at IS NULL`,
	); err != nil {
		log.Printf("Warning: failed to create org_folders unique path index: %v", err)
	}

	// Unique permission override per (org, user, folder_path)
	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_org_permissions ON org_folder_permissions (org_id, user_id, folder_path)`,
	); err != nil {
		log.Printf("Warning: failed to create org_folder_permissions unique index: %v", err)
	}

	// Groups: unique name per org, unique LDAP GUID per org
	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_org_groups_name ON org_groups (org_id, name)`,
	); err != nil {
		log.Printf("Warning: failed to create uq_org_groups_name: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_org_groups_ldap_guid ON org_groups (org_id, ldap_guid) WHERE ldap_guid IS NOT NULL AND ldap_guid != ''`,
	); err != nil {
		log.Printf("Warning: failed to create uq_org_groups_ldap_guid: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_org_group_members ON org_group_members (group_id, user_id)`,
	); err != nil {
		log.Printf("Warning: failed to create uq_org_group_members: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_org_group_permissions ON org_group_permissions (org_id, group_id, folder_path)`,
	); err != nil {
		log.Printf("Warning: failed to create uq_org_group_permissions: %v", err)
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS idx_organizations_owner_id       ON organizations (owner_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_members_org_id           ON org_members (org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_members_user_id          ON org_members (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_invitations_org_id       ON org_invitations (org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_invitations_token        ON org_invitations (token)`,
		`CREATE INDEX IF NOT EXISTS idx_org_invitations_target       ON org_invitations (target_user_id) WHERE target_user_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_org_files_org_folder         ON org_files (org_id, folder_path)`,
		`CREATE INDEX IF NOT EXISTS idx_org_files_uploaded_by        ON org_files (uploaded_by)`,
		`CREATE INDEX IF NOT EXISTS idx_org_folders_org_parent       ON org_folders (org_id, parent_path)`,
		`CREATE INDEX IF NOT EXISTS idx_org_folder_perms_org_user    ON org_folder_permissions (org_id, user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_audit_logs_org_created    ON org_audit_logs (org_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_org_groups_org_id            ON org_groups (org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_group_members_group_id   ON org_group_members (group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_group_members_user_id    ON org_group_members (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_group_perms_org_group    ON org_group_permissions (org_id, group_id)`,
	} {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			log.Printf("Warning: failed to create org index: %v", err)
		}
	}
	return nil
}

// migrateEmailEncryption adds email_hash / email_encrypted columns, migrates existing
// plaintext emails from auth_users and profiles, adds UNIQUE indexes on the hash
// columns, then nullifies the legacy plaintext email columns.
//
// Idempotent: safe to run on every startup. Rows that already have email_hash set
// are skipped. The old email column is kept (nullable) for rollback convenience.
func migrateEmailEncryption(ctx context.Context, db *bun.DB) error {
	// 1. Add new columns (nullable so existing rows don't immediately violate NOT NULL).
	for _, stmt := range []string{
		`ALTER TABLE "auth_users" ADD COLUMN IF NOT EXISTS "email_hash"      VARCHAR(64)`,
		`ALTER TABLE "auth_users" ADD COLUMN IF NOT EXISTS "email_encrypted" TEXT`,
		`ALTER TABLE "profiles"   ADD COLUMN IF NOT EXISTS "email_hash"      VARCHAR(64)`,
		`ALTER TABLE "profiles"   ADD COLUMN IF NOT EXISTS "email_encrypted" TEXT`,
		`ALTER TABLE "p2p_invites" ADD COLUMN IF NOT EXISTS "recipient_email_encrypted" TEXT`,
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("Warning: migrateEmailEncryption add column: %v", err)
		}
	}

	// 2. Migrate existing plaintext emails to hash + encrypted form.
	migrateEmailRows(ctx, db, "auth_users")
	migrateEmailRows(ctx, db, "profiles")
	migrateP2PInviteEmails(ctx, db)

	// 3. Add partial UNIQUE indexes on email_hash (partial so NULL rows are not indexed).
	for _, stmt := range []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_users_email_hash ON auth_users (email_hash) WHERE email_hash IS NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_email_hash   ON profiles   (email_hash) WHERE email_hash IS NOT NULL`,
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("Warning: migrateEmailEncryption create index: %v", err)
		}
	}

	// 4. Relax legacy email column constraints so Go code can INSERT without the old column.
	//    These are best-effort: fail silently if the column/constraint doesn't exist (new installs).
	for _, stmt := range []string{
		`ALTER TABLE "auth_users" ALTER COLUMN "email" DROP NOT NULL`,
		`ALTER TABLE "auth_users" DROP CONSTRAINT IF EXISTS "auth_users_email_key"`,
		`DROP INDEX IF EXISTS idx_auth_users_email`,
		`ALTER TABLE "profiles"   ALTER COLUMN "email" DROP NOT NULL`,
		`ALTER TABLE "profiles"   DROP CONSTRAINT IF EXISTS "profiles_email_key"`,
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("Warning: migrateEmailEncryption relax constraints: %v", err)
		}
	}

	// 5. Nullify plaintext email data for rows that have been encrypted successfully.
	for _, stmt := range []string{
		`UPDATE "auth_users"  SET "email" = NULL WHERE "email_hash" IS NOT NULL AND "email" IS NOT NULL`,
		`UPDATE "profiles"    SET "email" = NULL WHERE "email_hash" IS NOT NULL AND "email" IS NOT NULL`,
		`UPDATE "p2p_invites" SET "recipient_email" = NULL WHERE "recipient_email_encrypted" IS NOT NULL AND "recipient_email" IS NOT NULL`,
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("Warning: migrateEmailEncryption nullify plaintext: %v", err)
		}
	}

	return nil
}

// migrateEmailRows encrypts plaintext emails in the given table (auth_users or profiles).
// Only rows where email_hash IS NULL and email IS NOT NULL are processed.
func migrateEmailRows(ctx context.Context, db *bun.DB, table string) {
	type row struct {
		ID    string `bun:"id"`
		Email string `bun:"email"`
	}
	var rows []row
	if err := db.NewSelect().TableExpr(table).
		Column("id", "email").
		Where("email IS NOT NULL AND email_hash IS NULL").
		Scan(ctx, &rows); err != nil {
		log.Printf("Warning: migrateEmailRows fetch %s: %v", table, err)
		return
	}
	if len(rows) == 0 {
		return
	}
	log.Printf("[migrate] Encrypting %d email(s) in %s…", len(rows), table)
	for _, r := range rows {
		h := emailcrypto.Hash(r.Email)
		enc, err := emailcrypto.Encrypt(r.Email)
		if err != nil {
			log.Printf("Warning: migrateEmailRows encrypt %s id=%s: %v", table, r.ID, err)
			continue
		}
		q := fmt.Sprintf(`UPDATE "%s" SET email_hash = ?, email_encrypted = ? WHERE id = ?`, table)
		if _, err := db.ExecContext(ctx, q, h, enc, r.ID); err != nil {
			log.Printf("Warning: migrateEmailRows update %s id=%s: %v", table, r.ID, err)
		}
	}
}

// migrateP2PInviteEmails encrypts plaintext recipient_email in p2p_invites.
func migrateOrgTags(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS org_tags (
			id             BIGSERIAL PRIMARY KEY,
			org_id         BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			encrypted_name TEXT NOT NULL,
			color          VARCHAR(7) NOT NULL DEFAULT '#6366f1',
			created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("failed to create org_tags table: %w", err)
	}
	if _, err := db.ExecContext(ctx,
		`CREATE INDEX IF NOT EXISTS idx_org_tags_org_id ON org_tags (org_id)`); err != nil {
		log.Printf("Warning: failed to create idx_org_tags_org_id: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`ALTER TABLE org_files ADD COLUMN IF NOT EXISTS tag_ids BIGINT[] NOT NULL DEFAULT '{}'`); err != nil {
		log.Printf("Warning: failed to add tag_ids to org_files: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`ALTER TABLE org_folders ADD COLUMN IF NOT EXISTS tag_ids BIGINT[] NOT NULL DEFAULT '{}'`); err != nil {
		log.Printf("Warning: failed to add tag_ids to org_folders: %v", err)
	}
	return nil
}

func migrateOrgFavorites(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS org_favorites (
			id         BIGSERIAL PRIMARY KEY,
			org_id     BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			user_id    TEXT NOT NULL,
			item_id    BIGINT NOT NULL,
			item_type  VARCHAR(10) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (org_id, user_id, item_id, item_type)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create org_favorites: %w", err)
	}
	_, _ = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_org_fav_user ON org_favorites (org_id, user_id)`)
	return nil
}

func migrateOrgTrashColumns(ctx context.Context, db *bun.DB) error {
	for _, stmt := range []string{
		`ALTER TABLE org_files   ADD COLUMN IF NOT EXISTS deleted_by  TEXT    NOT NULL DEFAULT ''`,
		`ALTER TABLE org_files   ADD COLUMN IF NOT EXISTS delete_root BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE org_folders ADD COLUMN IF NOT EXISTS deleted_by  TEXT    NOT NULL DEFAULT ''`,
		`ALTER TABLE org_folders ADD COLUMN IF NOT EXISTS delete_root BOOLEAN NOT NULL DEFAULT FALSE`,
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("Warning: migrateOrgTrashColumns: %v", err)
		}
	}
	_, _ = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_org_files_trash   ON org_files   (org_id, deleted_at) WHERE deleted_at IS NOT NULL AND delete_root = TRUE`)
	_, _ = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_org_folders_trash ON org_folders (org_id, deleted_at) WHERE deleted_at IS NOT NULL AND delete_root = TRUE`)
	return nil
}

// migrateNewFeatureColumns adds columns introduced by the low-effort feature batch:
// download counter on share links, MFA enforcement flag on organizations,
// and per-member storage quota on org_members.
func migrateNewFeatureColumns(ctx context.Context, db *bun.DB) error {
	for _, stmt := range []string{
		`ALTER TABLE "share_links"    ADD COLUMN IF NOT EXISTS "download_count" BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE "organizations"  ADD COLUMN IF NOT EXISTS "require_mfa"    BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE "org_members"    ADD COLUMN IF NOT EXISTS "quota_bytes"    BIGINT NOT NULL DEFAULT 0`,
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("Warning: migrateNewFeatureColumns: %v", err)
		}
	}
	return nil
}

// migrateEnsurePartialOrgIndexes drops and recreates the org folder/file unique
// path indexes as PARTIAL indexes (WHERE deleted_at IS NULL). If a previous run
// created them as full unique indexes the IF NOT EXISTS guard would have left them
// non-partial, causing every insert after a soft-delete to fail with a false 409.
func migrateEnsurePartialOrgIndexes(ctx context.Context, db *bun.DB) {
	for _, pair := range [][2]string{
		{"uq_org_folders_org_path", "CREATE UNIQUE INDEX uq_org_folders_org_path ON org_folders (org_id, path) WHERE deleted_at IS NULL"},
		{"uq_org_files_org_path", "CREATE UNIQUE INDEX uq_org_files_org_path ON org_files (org_id, path) WHERE deleted_at IS NULL"},
	} {
		name, stmt := pair[0], pair[1]
		// Check if the index is already a partial index; if so, skip.
		var isPartial bool
		_ = db.QueryRowContext(ctx,
			`SELECT indpred IS NOT NULL FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid WHERE c.relname = ?`, name,
		).Scan(&isPartial)
		if isPartial {
			continue
		}
		// Drop (possibly non-partial) and recreate as partial.
		if _, err := db.ExecContext(ctx, `DROP INDEX IF EXISTS `+name); err != nil {
			log.Printf("Warning: migrateEnsurePartialOrgIndexes drop %s: %v", name, err)
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("Warning: migrateEnsurePartialOrgIndexes create %s: %v", name, err)
		}
	}
}

// migrateChunkSizeColumns adds the chunk_size column to files (and org_files) with a
// DEFAULT of 10 485 760 (10 MB). PostgreSQL fills existing rows with the default value
// at query time without rewriting the table, so old files transparently keep 10 MB
// semantics and new files store their actual encryption chunk size.
func migrateChunkSizeColumns(ctx context.Context, db *bun.DB) {
	for _, stmt := range []string{
		`ALTER TABLE "files"     ADD COLUMN IF NOT EXISTS "chunk_size" BIGINT NOT NULL DEFAULT 10485760`,
		`ALTER TABLE "org_files" ADD COLUMN IF NOT EXISTS "chunk_size" BIGINT NOT NULL DEFAULT 10485760`,
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("Warning: migrateChunkSizeColumns: %v", err)
		}
	}
}

// migrateFilesUniqueIndex deduplicates any duplicate (user_id, path) rows in the
// files table created by the pre-fix non-atomic upsert, then adds a UNIQUE index
// so the new ON CONFLICT DO UPDATE upsert can operate safely.
func migrateFilesUniqueIndex(ctx context.Context, db *bun.DB) error {
	// Remove duplicate rows keeping the newest (highest id) for each (user_id, path).
	if _, err := db.ExecContext(ctx, `
		DELETE FROM files a
		USING files b
		WHERE a.id < b.id
		  AND a.user_id = b.user_id
		  AND a.path = b.path
	`); err != nil {
		log.Printf("Warning: migrateFilesUniqueIndex dedup: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_files_user_path ON files (user_id, path)`,
	); err != nil {
		return fmt.Errorf("migrateFilesUniqueIndex: %w", err)
	}
	return nil
}

func migrateP2PInviteEmails(ctx context.Context, db *bun.DB) {
	type row struct {
		ID             int64  `bun:"id"`
		RecipientEmail string `bun:"recipient_email"`
	}
	var rows []row
	if err := db.NewSelect().TableExpr("p2p_invites").
		Column("id", "recipient_email").
		Where("recipient_email IS NOT NULL AND recipient_email_encrypted IS NULL").
		Scan(ctx, &rows); err != nil {
		log.Printf("Warning: migrateP2PInviteEmails fetch: %v", err)
		return
	}
	if len(rows) == 0 {
		return
	}
	log.Printf("[migrate] Encrypting %d P2P invite recipient email(s)…", len(rows))
	for _, r := range rows {
		enc, err := emailcrypto.Encrypt(r.RecipientEmail)
		if err != nil {
			log.Printf("Warning: migrateP2PInviteEmails encrypt id=%d: %v", r.ID, err)
			continue
		}
		if _, err := db.ExecContext(ctx,
			`UPDATE "p2p_invites" SET recipient_email_encrypted = ? WHERE id = ?`, enc, r.ID,
		); err != nil {
			log.Printf("Warning: migrateP2PInviteEmails update id=%d: %v", r.ID, err)
		}
	}
}

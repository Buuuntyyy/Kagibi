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

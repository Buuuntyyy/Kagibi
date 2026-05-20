// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// internal/models.go
package pkg

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// WSHub is the subset of the WebSocket hub used by pkg to push events.
// It is set at startup by main.go to avoid an import cycle.
type WSHub interface {
	SendEventToUser(userID, eventType string, id int64, payload map[string]any)
	SendP2PSignalToUser(targetUserID, senderID, signalType string, signalID int64, payload map[string]any)
}

// wsHub is the global hub instance injected at startup.
var wsHub WSHub

// SetWSHub registers the WebSocket hub so that EmitRealtimeEvent can push live events.
func SetWSHub(h WSHub) { wsHub = h }

type User struct {
	bun.BaseModel `bun:"table:profiles,alias:p"`

	ID             string `bun:"id,pk"`
	Name           string `bun:"name,notnull"`
	EmailHash      string `bun:"email_hash,notnull" json:"-"`
	EmailEncrypted string `bun:"email_encrypted,notnull" json:"-"`
	Email          string `bun:"-" json:"email"` // virtual: populated by DecryptUserEmail after any DB load
	AvatarURL      string `bun:"avatar_url,notnull,default:'/avatars/default.png'" json:"avatar_url"`
	// PasswordHash removed as it is handled by Supabase
	Salt                       string     `bun:"salt,notnull" json:"salt"`
	EncryptedMasterKey         string     `bun:"encrypted_master_key,notnull" json:"encrypted_master_key"`
	EncryptedMasterKeyRecovery string     `bun:"encrypted_master_key_recovery,notnull" json:"encrypted_master_key_recovery"`
	RecoveryHash               string     `bun:"recovery_hash,notnull" json:"recovery_hash"`
	RecoverySalt               string     `bun:"recovery_salt,notnull" json:"recovery_salt"`
	FriendCode                 string     `bun:"friend_code,unique,notnull" json:"friend_code"`                    // Short unique code for friends
	PublicKey                  string     `bun:"public_key" json:"public_key"`                                     // RSA Public Key (Standard PEM format)
	EncryptedPrivateKey        string     `bun:"encrypted_private_key" json:"encrypted_private_key"`               // RSA Private Key (Encrypted with MasterKey)
	EncryptFilenames           bool       `bun:"encrypt_filenames,notnull,default:false" json:"encrypt_filenames"` // Client-side filename encryption opt-in
	CreatedAt                  time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt                  time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	DeletedAt                  *time.Time `bun:"deleted_at,soft_delete,nullzero" json:"deleted_at,omitempty"` // RGPD Article 17 - Soft delete
}

type Friendship struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	UserID1   string    `bun:"user_id_1,notnull"` // Initiator
	UserID2   string    `bun:"user_id_2,notnull"` // Recipient
	Status    string    `bun:"status,notnull"`    // 'pending', 'accepted'
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

type UserPlan struct {
	bun.BaseModel    `bun:"table:user_plans,alias:up"`
	UserID           string    `bun:"user_id,pk" json:"user_id"`
	Plan             string    `bun:"plan,notnull,default:'free'" json:"plan"`
	StorageLimit     int64     `bun:"storage_limit,notnull,default:21474836480" json:"storage_limit"`
	StorageUsed      int64     `bun:"storage_used,notnull,default:0" json:"storage_used"`
	P2PMaxExchanges  int       `bun:"p2p_max_exchanges,notnull,default:-1" json:"p2p_max_exchanges"`
	P2PExchangesUsed int       `bun:"p2p_exchanges_used,notnull,default:0" json:"p2p_exchanges_used"`
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

type File struct {
	ID           int64     `bun:"id,pk,autoincrement"`
	Name         string    `bun:"name,notnull"`
	Path         string    `bun:"path,notnull"`                                     // Chemin relatif (ex: "/dossier1/fichier.txt")
	Size         int64     `bun:"size,notnull"`                                     // Taille en octets
	MimeType     string    `bun:"mime_type,notnull"`                                // Ex: "application/pdf"
	UserID       string    `bun:"user_id,notnull"`                                  // Propriétaire du fichier
	EncryptedKey string    `bun:"encrypted_key"`                                    // Clé du fichier chiffrée avec la MasterKey (pour les nouveaux fichiers)
	Tags         []string  `bun:"tags,array"`                                       // Tags
	PreviewID    *int64    `bun:"preview_id" json:"preview_id"`                     // ID du fichier de prévisualisation (miniature/compressé)
	Preview      *File     `bun:"rel:belongs-to,join:preview_id=id" json:"preview"` // Metadata du fichier preview
	IsPreview    bool      `bun:"is_preview,default:false" json:"is_preview"`       // Indique si c'est un fichier de prévisualisation (masqué par défaut)
	Synced       bool      `bun:"synced,default:false" json:"synced"`               // true si uploadé via la sync desktop
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

// Champs non persistés utilisés pour l'API
func (File) TableName() string { return "files" }

// Champs ajoutés pour l'API (non mappés en base)
// `bun:"-"` indique à bun de ne pas essayer de mapper ces champs.
type FileWithShare struct {
	File
	Shared     bool       `bun:"-" json:"shared"`
	ShareToken *string    `bun:"-" json:"share_token,omitempty"`
	ShareID    *int64     `bun:"-" json:"share_id,omitempty"`
	ExpiresAt  *time.Time `bun:"-" json:"expires_at,omitempty"`
}

type FolderWithShare struct {
	Folder
	Shared     bool       `bun:"-" json:"shared"`
	ShareToken *string    `bun:"-" json:"share_token,omitempty"`
	ShareID    *int64     `bun:"-" json:"share_id,omitempty"`
	ExpiresAt  *time.Time `bun:"-" json:"expires_at,omitempty"`
}

type FileShare struct {
	ID               int64     `bun:"id,pk,autoincrement"`
	FileID           int64     `bun:"file_id,notnull"`
	SharedWithUserID string    `bun:"shared_with_user_id,notnull"`
	EncryptedKey     string    `bun:"encrypted_key,notnull"` // File Key Encrypted with recipient's Public Key
	Permission       string    `bun:"permission,notnull,default:'read'"`
	PermDownload     bool      `bun:"perm_download,notnull,default:true"`
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

type FolderShare struct {
	ID               int64     `bun:"id,pk,autoincrement"`
	FolderID         int64     `bun:"folder_id,notnull"`
	SharedWithUserID string    `bun:"shared_with_user_id,notnull"`
	EncryptedKey     string    `bun:"encrypted_key"` // FolderKey encrypted with Recipient's Public Key
	Permission       string    `bun:"permission,notnull,default:'read'"`
	PermDownload     bool      `bun:"perm_download,notnull,default:true"`
	PermCreate       bool      `bun:"perm_create,notnull,default:false"`
	PermDelete       bool      `bun:"perm_delete,notnull,default:false"`
	PermMove         bool      `bun:"perm_move,notnull,default:false"`
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

type FolderFileKey struct {
	FolderID     int64     `bun:"folder_id,pk"`
	FileID       int64     `bun:"file_id,pk"`
	EncryptedKey string    `bun:"encrypted_key,notnull"` // FileKey encrypted with FolderKey
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

type FolderFolderKey struct {
	ParentFolderID int64     `bun:"parent_folder_id,pk"`   // The Root Shared Folder
	SubFolderID    int64     `bun:"sub_folder_id,pk"`      // The Subfolder
	EncryptedKey   string    `bun:"encrypted_key,notnull"` // SubfolderKey encrypted with ParentFolderKey
	CreatedAt      time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

type ShareFileKey struct {
	ShareID      int64  `bun:"share_id,pk"`
	FileID       int64  `bun:"file_id,pk"`
	EncryptedKey string `bun:"encrypted_key,notnull"`
}

type Folder struct {
	ID           int64     `bun:"id,pk,autoincrement"`
	Name         string    `bun:"name,notnull"`
	Path         string    `bun:"path,notnull"` // Chemin relatif (ex: "/dossier1")
	UserID       string    `bun:"user_id,notnull"`
	EncryptedKey string    `bun:"encrypted_key"` // FolderKey encrypted with MasterKey
	Tags         []string  `bun:"tags,array"`    // Tags
	SizeBytes    int64     `bun:"size_bytes,scanonly" json:"size_bytes,omitempty"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

type FolderSize struct {
	bun.BaseModel `bun:"table:folder_sizes,alias:fs"`
	FolderID      int64     `bun:"folder_id,pk"`
	UserID        string    `bun:"user_id,notnull"`
	SizeBytes     int64     `bun:"size_bytes,notnull,default:0"`
	UpdatedAt     time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

type Tag struct {
	ID     int64  `bun:"id,pk,autoincrement" json:"id"`
	UserID string `bun:"user_id,notnull" json:"user_id"`
	Name   string `bun:"name,notnull" json:"name"`
	Color  string `bun:"color,notnull" json:"color"` // Code Hex (ex: #FF0000)
}

type ShareLink struct {
	bun.BaseModel `bun:"table:share_links,alias:sl"`

	ID           int64      `bun:"id,pk,autoincrement"`
	ResourceID   int64      `bun:"resource_id,notnull"`   // ID du fichier ou dossier
	ResourceType string     `bun:"resource_type,notnull"` // "file" | "folder" | "org_file"
	Path         string     `bun:"path"`                  // Base path of the shared resource
	OwnerID      string     `bun:"owner_id,notnull"`      // Créateur du lien
	Token        string     `bun:"token,unique,notnull"`  // Le code dans l'URL (ex: "xYz123")
	EncryptedKey string     `bun:"encrypted_key"`         // Clé du fichier chiffrée avec la ShareKey (pour les partages de fichiers)
	PasswordHash string     `bun:"password_hash"`         // Optionnel : mot de passe
	ExpiresAt    *time.Time `bun:"expires_at"`            // Optionnel : expiration
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	Views         int64      `bun:"views,default:0"`
	DownloadCount int64      `bun:"download_count,default:0" json:"download_count"`
	SingleUse    bool       `bun:"single_use,default:false"`   // Link is invalidated after first download
	UsedAt       *time.Time `bun:"used_at"`                    // Set when a single-use link is consumed
	PermDownload bool       `bun:"perm_download,default:true"` // Can download files
	PermCreate   bool       `bun:"perm_create,default:false"`  // Folder: can create files/dirs
	PermDelete   bool       `bun:"perm_delete,default:false"`  // Folder: can delete files/dirs
	PermMove     bool       `bun:"perm_move,default:false"`    // Folder: can move files/dirs
	OrgID        *int64     `bun:"org_id" json:"org_id,omitempty"` // set for org_file shares
}

// ShareItemOverride stores per-item access restrictions within a shared folder.
type ShareItemOverride struct {
	bun.BaseModel `bun:"table:share_item_overrides,alias:sio"`

	ID          int64  `bun:"id,pk,autoincrement" json:"id"`
	ShareID     int64  `bun:"share_id,notnull" json:"share_id"`
	ItemPath    string `bun:"item_path,notnull" json:"item_path"`
	ItemType    string `bun:"item_type,notnull" json:"item_type"`       // "file" | "folder"
	AccessLevel string `bun:"access_level,notnull" json:"access_level"` // "full" | "readonly" | "none"
	CanDelete   bool   `bun:"can_delete,notnull,default:true" json:"can_delete"`
	CanDownload bool   `bun:"can_download,notnull,default:true" json:"can_download"`
}

type ImportedShare struct {
	bun.BaseModel `bun:"table:imported_shares,alias:ish"`

	ID          int64      `bun:"id,pk,autoincrement"`
	UserID      string     `bun:"user_id,notnull"`       // L'utilisateur qui a importé le partage
	ShareLinkID int64      `bun:"share_link_id,notnull"` // Le partage importé
	ShareLink   *ShareLink `bun:"rel:belongs-to,join:share_link_id=id"`
	CreatedAt   time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

type RecentActivity struct {
	ID         int64     `bun:"id,pk,autoincrement"`
	UserID     string    `bun:"user_id,notnull"`
	FileID     *int64    `bun:"file_id"`   // Nullable, set if it's a file
	FolderID   *int64    `bun:"folder_id"` // Nullable, set if it's a folder
	File       *File     `bun:"rel:belongs-to,join:file_id=id"`
	Folder     *Folder   `bun:"rel:belongs-to,join:folder_id=id"`
	AccessedAt time.Time `bun:"accessed_at,nullzero,notnull,default:current_timestamp"`
}

// RealtimeEvent represents an event for Supabase Realtime
type RealtimeEvent struct {
	bun.BaseModel `bun:"table:realtime_events"`

	ID        int64          `bun:"id,pk,autoincrement"`
	UserID    string         `bun:"user_id,notnull"`
	EventType string         `bun:"event_type,notnull"`
	Payload   map[string]any `bun:"payload,type:jsonb"`
	CreatedAt time.Time      `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

// P2PInvite is a time-limited token that lets an authenticated user invite
// anyone (guest or registered) to a P2P transfer without both being online simultaneously.
// RecipientEmail is optional — used only for notification; RecipientID is a guest UUID for
// guest invites (IsGuest=true) or the registered user ID for account-based invites.
type P2PInvite struct {
	bun.BaseModel `bun:"table:p2p_invites"`

	ID                      int64      `bun:"id,pk,autoincrement" json:"id"`
	Token                   string     `bun:"token,unique,notnull" json:"token"`
	SenderID                string     `bun:"sender_id,notnull" json:"sender_id"`
	SenderName              string     `bun:"sender_name,notnull" json:"sender_name"`
	RecipientEmailEncrypted string     `bun:"recipient_email_encrypted" json:"-"` // AES-256-GCM encrypted, nullable
	RecipientEmail          string     `bun:"-" json:"recipient_email"`           // virtual: decrypted from RecipientEmailEncrypted
	RecipientID             string     `bun:"recipient_id,notnull" json:"recipient_id"`
	TransferID              string     `bun:"transfer_id,notnull" json:"transfer_id"`
	FileName                string     `bun:"file_name,notnull" json:"file_name"`
	FileSize                int64      `bun:"file_size,notnull" json:"file_size"`
	IsGuest                 bool       `bun:"is_guest,notnull,default:false" json:"is_guest"`
	ExpiresAt               time.Time  `bun:"expires_at,notnull" json:"expires_at"`
	GuestAuthedAt           *time.Time `bun:"guest_authed_at" json:"guest_authed_at,omitempty"`
	AcceptedAt              *time.Time `bun:"accepted_at" json:"accepted_at,omitempty"`
	CreatedAt               time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// P2PSignal represents a WebRTC signaling message
type P2PSignal struct {
	bun.BaseModel `bun:"table:p2p_signals"`

	ID         int64          `bun:"id,pk,autoincrement" json:"id"`
	SenderID   string         `bun:"sender_id,notnull" json:"sender_id"`
	TargetID   string         `bun:"target_id,notnull" json:"target_id"`
	SignalType string         `bun:"signal_type,notnull" json:"signal_type"`
	Payload    map[string]any `bun:"payload,type:jsonb" json:"payload"`
	CreatedAt  time.Time      `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	Consumed   bool           `bun:"consumed,notnull,default:false" json:"consumed"`
}

// Organization represents a group or professional entity with shared storage and members.
type Organization struct {
	bun.BaseModel `bun:"table:organizations,alias:org"`

	ID               int64      `bun:"id,pk,autoincrement" json:"id"`
	Name             string     `bun:"name,notnull" json:"name"`
	Description      string     `bun:"description" json:"description"`
	OwnerID          string     `bun:"owner_id,notnull" json:"owner_id"`
	LogoPath         string     `bun:"logo_path,notnull,default:''" json:"logo_path,omitempty"`
	StorageQuotaMB   int64      `bun:"storage_quota_mb,notnull,default:10240" json:"storage_quota_mb"` // 10 GB default
	StorageUsedBytes int64      `bun:"storage_used_bytes,notnull,default:0" json:"storage_used_bytes"`
	RequireMFA       bool       `bun:"require_mfa,default:false" json:"require_mfa"`
	CreatedAt        time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt        time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt        *time.Time `bun:"deleted_at,soft_delete,nullzero" json:"deleted_at,omitempty"`
}

// OrgMember represents a user's membership in an organization.
type OrgMember struct {
	bun.BaseModel `bun:"table:org_members,alias:om"`

	ID              int64     `bun:"id,pk,autoincrement" json:"id"`
	OrgID           int64     `bun:"org_id,notnull" json:"org_id"`
	UserID          string    `bun:"user_id,notnull" json:"user_id"`
	Role            string    `bun:"role,notnull,default:'member'" json:"role"` // owner | admin | member | viewer
	EncryptedOrgKey string    `bun:"encrypted_org_key" json:"encrypted_org_key,omitempty"` // org key encrypted with this member's RSA public key
	QuotaBytes      int64     `bun:"quota_bytes,default:0" json:"quota_bytes"` // 0 = use org-level default
	JoinedAt        time.Time `bun:"joined_at,nullzero,notnull,default:current_timestamp" json:"joined_at"`
}

// OrgInvitation is a token-based or direct invite to join an organization.
type OrgInvitation struct {
	bun.BaseModel `bun:"table:org_invitations,alias:oi"`

	ID                     int64      `bun:"id,pk,autoincrement" json:"id"`
	OrgID                  int64      `bun:"org_id,notnull" json:"org_id"`
	InvitedBy              string     `bun:"invited_by,notnull" json:"invited_by"`
	Token                  string     `bun:"token,unique,notnull" json:"token"`
	TargetUserID           *string    `bun:"target_user_id" json:"target_user_id,omitempty"`       // set for direct invites
	EncryptedOrgKey        string     `bun:"encrypted_org_key" json:"encrypted_org_key,omitempty"` // pre-encrypted for direct invites
	NotifiedEmailEncrypted string     `bun:"notified_email_encrypted" json:"-"`                    // AES-256-GCM encrypted recipient email
	Role                   string     `bun:"role,notnull,default:'member'" json:"role"`
	MaxUses                int        `bun:"max_uses,notnull,default:0" json:"max_uses"` // 0 = unlimited
	Uses                   int        `bun:"uses,notnull,default:0" json:"uses"`
	ExpiresAt              *time.Time `bun:"expires_at" json:"expires_at,omitempty"`
	Status                 string     `bun:"status,notnull,default:'active'" json:"status"` // active | revoked
	CreatedAt              time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// OrgTag is a colored label that can be applied to org files and folders.
type OrgTag struct {
	bun.BaseModel `bun:"table:org_tags,alias:ot"`

	ID            int64     `bun:"id,pk,autoincrement" json:"id"`
	OrgID         int64     `bun:"org_id,notnull" json:"org_id"`
	EncryptedName string    `bun:"encrypted_name,notnull" json:"encrypted_name"`
	Color         string    `bun:"color,notnull" json:"color"`
	CreatedAt     time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// OrgFavorite is a pinned file or folder for a specific org member.
type OrgFavorite struct {
	bun.BaseModel `bun:"table:org_favorites,alias:ofav"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	OrgID     int64     `bun:"org_id,notnull" json:"org_id"`
	UserID    string    `bun:"user_id,notnull" json:"user_id"`
	ItemID    int64     `bun:"item_id,notnull" json:"item_id"`
	ItemType  string    `bun:"item_type,notnull" json:"item_type"` // "file" | "folder"
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// OrgFolder is a directory inside an organization's shared storage.
type OrgFolder struct {
	bun.BaseModel `bun:"table:org_folders,alias:of"`

	ID           int64      `bun:"id,pk,autoincrement" json:"id"`
	OrgID        int64      `bun:"org_id,notnull" json:"org_id"`
	Name         string     `bun:"name,notnull" json:"name"`
	Path         string     `bun:"path,notnull" json:"path"`             // full virtual path, e.g. "/documents/contracts"
	ParentPath   string     `bun:"parent_path,notnull,default:'/'" json:"parent_path"` // path.Dir(path)
	CreatedBy    string     `bun:"created_by,notnull" json:"created_by"` // user_id
	EncryptedKey string     `bun:"encrypted_key" json:"encrypted_key,omitempty"` // folder key encrypted with org_key
	TagIDs       []int64    `bun:"tag_ids,array,nullzero,default:'{}'" json:"tag_ids"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt    *time.Time `bun:"deleted_at,soft_delete,nullzero" json:"deleted_at,omitempty"`
	DeletedBy    string     `bun:"deleted_by,notnull,default:''" json:"deleted_by,omitempty"`
	DeleteRoot   bool       `bun:"delete_root,notnull,default:false" json:"-"`
	TotalSize    int64      `bun:"-" json:"total_size,omitempty"` // computed on list, not stored
}

// OrgFile is a file inside an organization's shared storage.
type OrgFile struct {
	bun.BaseModel `bun:"table:org_files,alias:ofile"`

	ID           int64      `bun:"id,pk,autoincrement" json:"id"`
	OrgID        int64      `bun:"org_id,notnull" json:"org_id"`
	Name         string     `bun:"name,notnull" json:"name"`
	Path         string     `bun:"path,notnull" json:"path"`                 // full path including name, e.g. "/documents/report.pdf"
	FolderPath   string     `bun:"folder_path,notnull,default:'/'" json:"folder_path"` // path.Dir(path)
	Size         int64      `bun:"size,notnull,default:0" json:"size"`
	MimeType     string     `bun:"mime_type,notnull,default:''" json:"mime_type"`
	UploadedBy   string     `bun:"uploaded_by,notnull" json:"uploaded_by"` // user_id
	EncryptedKey string     `bun:"encrypted_key" json:"encrypted_key,omitempty"`
	TagIDs       []int64    `bun:"tag_ids,array,nullzero,default:'{}'" json:"tag_ids"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt    *time.Time `bun:"deleted_at,soft_delete,nullzero" json:"deleted_at,omitempty"`
	DeletedBy    string     `bun:"deleted_by,notnull,default:''" json:"deleted_by,omitempty"`
	DeleteRoot   bool       `bun:"delete_root,notnull,default:false" json:"-"`
}

// OrgFolderPermission stores per-user access overrides for a folder path within an org.
// Permissions are inherited: the most specific (deepest) path wins.
// A "none" level blocks access regardless of the member's role.
type OrgFolderPermission struct {
	bun.BaseModel `bun:"table:org_folder_permissions,alias:ofp"`

	ID           int64     `bun:"id,pk,autoincrement" json:"id"`
	OrgID        int64     `bun:"org_id,notnull" json:"org_id"`
	UserID       string    `bun:"user_id,notnull" json:"user_id"`
	FolderPath   string    `bun:"folder_path,notnull,default:'/'" json:"folder_path"` // "/" = org root
	Level        string    `bun:"level,notnull" json:"level"`                         // read | write | manage | none
	PermCreate   bool      `bun:"perm_create,notnull,default:false" json:"perm_create"`
	PermDelete   bool      `bun:"perm_delete,notnull,default:false" json:"perm_delete"`
	PermDownload bool      `bun:"perm_download,notnull,default:true" json:"perm_download"`
	PermMove     bool      `bun:"perm_move,notnull,default:false" json:"perm_move"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// OrgGroup is a named collection of org members used for bulk permission assignment.
// source = "internal" for manually managed groups; "ldap" for directory-synced groups.
type OrgGroup struct {
	bun.BaseModel `bun:"table:org_groups,alias:og"`

	ID          int64      `bun:"id,pk,autoincrement" json:"id"`
	OrgID       int64      `bun:"org_id,notnull" json:"org_id"`
	Name        string     `bun:"name,notnull" json:"name"`
	Description string     `bun:"description" json:"description"`
	CreatedBy   string     `bun:"created_by,notnull" json:"created_by"`

	// LDAP fields — populated only when source = "ldap"
	Source       string     `bun:"source,notnull,default:'internal'" json:"source"` // "internal" | "ldap"
	LdapDN       string     `bun:"ldap_dn" json:"ldap_dn,omitempty"`
	LdapGUID     string     `bun:"ldap_guid" json:"ldap_guid,omitempty"`
	LastSyncedAt *time.Time `bun:"last_synced_at" json:"last_synced_at,omitempty"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// OrgGroupMember records a user's membership in an org group.
// AddedBy is empty when the membership originates from an LDAP sync.
type OrgGroupMember struct {
	bun.BaseModel `bun:"table:org_group_members,alias:ogm"`

	ID       int64     `bun:"id,pk,autoincrement" json:"id"`
	GroupID  int64     `bun:"group_id,notnull" json:"group_id"`
	UserID   string    `bun:"user_id,notnull" json:"user_id"`
	Role     string    `bun:"role,notnull,default:'member'" json:"role"` // admin | member
	AddedBy  string    `bun:"added_by" json:"added_by,omitempty"`
	JoinedAt time.Time `bun:"joined_at,nullzero,notnull,default:current_timestamp" json:"joined_at"`
}

// OrgGroupPermission stores folder-level access overrides for a group.
// Resolution rule: direct user overrides beat group overrides; among multiple
// group overrides the most permissive wins ("none" at group level never blocks).
type OrgGroupPermission struct {
	bun.BaseModel `bun:"table:org_group_permissions,alias:ogp"`

	ID           int64     `bun:"id,pk,autoincrement" json:"id"`
	OrgID        int64     `bun:"org_id,notnull" json:"org_id"`
	GroupID      int64     `bun:"group_id,notnull" json:"group_id"`
	FolderPath       string    `bun:"folder_path,notnull,default:'/'" json:"folder_path"`
	Level            string    `bun:"level,notnull" json:"level"` // read | write | manage | none
	PermCreate       bool      `bun:"perm_create,notnull,default:false" json:"perm_create"`
	PermDelete       bool      `bun:"perm_delete,notnull,default:false" json:"perm_delete"`
	PermDownload     bool      `bun:"perm_download,notnull,default:true" json:"perm_download"`
	PermMove         bool      `bun:"perm_move,notnull,default:false" json:"perm_move"`
	// RestrictToGroups, when true, makes this path inaccessible to org members
	// who are not in any group with an explicit permission on this path.
	// Owners and admins are never affected.
	RestrictToGroups bool      `bun:"restrict_to_groups,notnull,default:false" json:"restrict_to_groups"`
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// OrgAuditLog records security-relevant events within an organization.
// The log is append-only — records are never updated or deleted.
type OrgAuditLog struct {
	bun.BaseModel `bun:"table:org_audit_logs,alias:oal"`

	ID         int64     `bun:"id,pk,autoincrement" json:"id"`
	OrgID      int64     `bun:"org_id,notnull" json:"org_id"`
	ActorID    string    `bun:"actor_id,notnull" json:"actor_id"`
	Action     string    `bun:"action,notnull" json:"action"` // member_joined | member_removed | role_changed | file_uploaded | file_downloaded | file_deleted | permission_set | permission_removed | invitation_created | invitation_revoked | key_rotated | key_provisioned
	TargetID   string    `bun:"target_id,notnull,default:''" json:"target_id,omitempty"`
	TargetType string    `bun:"target_type,notnull,default:''" json:"target_type,omitempty"`
	Detail     string    `bun:"detail,notnull,default:''" json:"detail,omitempty"`
	CreatedAt  time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// EmitRealtimeEvent inserts an event into the realtime_events table and
// pushes it immediately over WebSocket if the user is connected.
func EmitRealtimeEvent(ctx context.Context, db *bun.DB, userID, eventType string, payload map[string]any) error {
	event := &RealtimeEvent{
		UserID:    userID,
		EventType: eventType,
		Payload:   payload,
	}
	if _, err := db.NewInsert().Model(event).Exec(ctx); err != nil {
		return err
	}
	// Push over WebSocket (no-op if hub not set or user has no active connection)
	if wsHub != nil {
		wsHub.SendEventToUser(userID, eventType, event.ID, payload)
	}
	return nil
}

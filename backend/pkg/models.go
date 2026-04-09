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
	SendP2PSignalToUser(targetUserID, senderID, signalType string, payload map[string]any)
}

// wsHub is the global hub instance injected at startup.
var wsHub WSHub

// SetWSHub registers the WebSocket hub so that EmitRealtimeEvent can push live events.
func SetWSHub(h WSHub) { wsHub = h }

type User struct {
	bun.BaseModel `bun:"table:profiles,alias:p"`

	ID        string `bun:"id,pk"`
	Name      string `bun:"name,notnull"`
	Email     string `bun:"email,unique,notnull"`
	AvatarURL string `bun:"avatar_url,notnull,default:'/avatars/default.png'" json:"avatar_url"`
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
	P2PMaxExchanges  int       `bun:"p2p_max_exchanges,notnull,default:5" json:"p2p_max_exchanges"`
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
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

type FolderShare struct {
	ID               int64     `bun:"id,pk,autoincrement"`
	FolderID         int64     `bun:"folder_id,notnull"`
	SharedWithUserID string    `bun:"shared_with_user_id,notnull"`
	EncryptedKey     string    `bun:"encrypted_key"` // FolderKey encrypted with Recipient's Public Key
	Permission       string    `bun:"permission,notnull,default:'read'"`
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
	ResourceType string     `bun:"resource_type,notnull"` // "file" ou "folder"
	Path         string     `bun:"path"`                  // Base path of the shared resource
	OwnerID      string     `bun:"owner_id,notnull"`      // Créateur du lien
	Token        string     `bun:"token,unique,notnull"`  // Le code dans l'URL (ex: "xYz123")
	EncryptedKey string     `bun:"encrypted_key"`         // Clé du fichier chiffrée avec la ShareKey (pour les partages de fichiers)
	PasswordHash string     `bun:"password_hash"`         // Optionnel : mot de passe
	ExpiresAt    *time.Time `bun:"expires_at"`            // Optionnel : expiration
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	Views        int64      `bun:"views,default:0"`
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

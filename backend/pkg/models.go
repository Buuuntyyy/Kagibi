// internal/models.go
package pkg

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:profiles,alias:p"`

	ID                         string    `bun:"id,pk"`
	Name                       string    `bun:"name,notnull"`
	Email                      string    `bun:"email,unique,notnull"`
	// PasswordHash removed as it is handled by Supabase
	Salt                       string    `bun:"salt,notnull" json:"salt"`
	EncryptedMasterKey         string    `bun:"encrypted_master_key,notnull" json:"encrypted_master_key"`
	EncryptedMasterKeyRecovery string    `bun:"encrypted_master_key_recovery,notnull" json:"encrypted_master_key_recovery"`
	RecoveryHash               string    `bun:"recovery_hash,notnull" json:"recovery_hash"`
	RecoverySalt               string    `bun:"recovery_salt,notnull" json:"recovery_salt"`
	StorageUsed                int64     `bun:"storage_used,notnull,default:0" json:"storage_used"`
	StorageLimit               int64     `bun:"storage_limit,notnull,default:16106127360" json:"storage_limit"` // Default 15GB
	Plan                       string    `bun:"plan,notnull,default:'free'" json:"plan"`
	FriendCode                 string    `bun:"friend_code,unique,notnull" json:"friend_code"`      // Short unique code for friends
	PublicKey                  string    `bun:"public_key" json:"public_key"`                       // RSA Public Key (Standard PEM format)
	EncryptedPrivateKey        string    `bun:"encrypted_private_key" json:"encrypted_private_key"` // RSA Private Key (Encrypted with MasterKey)
	CreatedAt                  time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt                  time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

type Friendship struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	UserID1   string    `bun:"user_id_1,notnull"` // Initiator
	UserID2   string    `bun:"user_id_2,notnull"` // Recipient
	Status    string    `bun:"status,notnull"`    // 'pending', 'accepted'
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
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
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
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

// internal/models.go
package pkg

import "time"

type User struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	Email     string    `bun:"email,unique,notnull"`
	PasswordHash string `bun:"password_hash,notnull"`
	Salt	  string    `bun:"salt,notnull" json:"salt"`
	EncryptedMasterKey string `bun:"encrypted_master_key,notnull" json:"encrypted_master_key"`
	EncryptedMasterKeyRecovery string `bun:"encrypted_master_key_recovery,notnull" json:"encrypted_master_key_recovery"`
	RecoveryHash string `bun:"recovery_hash,notnull" json:"recovery_hash"`
	RecoverySalt string `bun:"recovery_salt,notnull" json:"recovery_salt"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

type File struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	Path      string    `bun:"path,notnull"`       // Chemin relatif (ex: "/dossier1/fichier.txt")
	Size      int64     `bun:"size,notnull"`       // Taille en octets
	MimeType  string    `bun:"mime_type,notnull"`  // Ex: "application/pdf"
	UserID    int64     `bun:"user_id,notnull"`    // Propriétaire du fichier
	Tags      []string  `bun:"tags,array"`         // Tags
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

type Folder struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	Path      string    `bun:"path,notnull"`       // Chemin relatif (ex: "/dossier1")
	UserID    int64     `bun:"user_id,notnull"`
	Tags      []string  `bun:"tags,array"`         // Tags
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

type Tag struct {
	ID     int64  `bun:"id,pk,autoincrement" json:"id"`
	UserID int64  `bun:"user_id,notnull" json:"user_id"`
	Name   string `bun:"name,notnull" json:"name"`
	Color  string `bun:"color,notnull" json:"color"` // Code Hex (ex: #FF0000)
}

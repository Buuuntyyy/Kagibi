// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// RGPD Article 20 - Droit à la portabilité des données
// Règlement (UE) 2016/679 - Article 20
// Loi Informatique et Libertés (loi n°78-17) modifiée - Article 55
//
// L'utilisateur a le droit de recevoir les données à caractère personnel
// le concernant, dans un format structuré, couramment utilisé et lisible
// par machine (JSON), et a le droit de transmettre ces données à un autre
// responsable du traitement sans que le responsable du traitement auquel
// les données ont été communiquées y fasse obstacle.
//
// Ce handler exporte :
//   - Profil utilisateur (données d'identité)
//   - Arborescence de fichiers et dossiers (métadonnées)
//   - Tags personnalisés
//   - Liens de partage créés
//   - Relations d'amitié
//   - Activité récente
//
// Les fichiers chiffrés eux-mêmes ne sont PAS inclus dans cet export JSON :
// ils sont téléchargeables via l'interface existante. Les clés de chiffrement
// sont incluses sous forme chiffrée uniquement (elles ne sont exploitables
// qu'avec la clé maître de l'utilisateur).

package users

import (
	"context"
	"fmt"
	"kagibi/backend/pkg"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// PortabilityExport représente l'intégralité des données personnelles
// d'un utilisateur, conformément à l'article 20 du RGPD.
// Le format JSON est un format structuré, couramment utilisé et lisible
// par machine, tel qu'exigé par le règlement.
type PortabilityExport struct {
	ExportMetadata ExportMetadata    `json:"export_metadata"`
	Profile        ExportProfile     `json:"profile"`
	Files          []ExportFile      `json:"files"`
	Folders        []ExportFolder    `json:"folders"`
	Tags           []ExportTag       `json:"tags"`
	ShareLinks     []ExportShareLink `json:"share_links"`
	Friends        []ExportFriend    `json:"friends"`
	RecentActivity []ExportActivity  `json:"recent_activity"`
}

type ExportMetadata struct {
	ExportDate    string `json:"export_date"`
	Format        string `json:"format"`
	FormatVersion string `json:"format_version"`
	LegalBasis    string `json:"legal_basis"`
	ServiceName   string `json:"service_name"`
	UserID        string `json:"user_id"`
}

type ExportProfile struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Email               string    `json:"email"`
	AvatarURL           string    `json:"avatar_url"`
	Plan                string    `json:"plan"`
	StorageUsed         int64     `json:"storage_used_bytes"`
	StorageLimit        int64     `json:"storage_limit_bytes"`
	FriendCode          string    `json:"friend_code"`
	PublicKey           string    `json:"public_key"`
	EncryptedPrivateKey string    `json:"encrypted_private_key"`
	EncryptedMasterKey  string    `json:"encrypted_master_key"`
	Salt                string    `json:"salt"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ExportFile struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size_bytes"`
	MimeType     string    `json:"mime_type"`
	EncryptedKey string    `json:"encrypted_key"`
	Tags         []string  `json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ExportFolder struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	EncryptedKey string    `json:"encrypted_key"`
	Tags         []string  `json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ExportTag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type ExportShareLink struct {
	ID           int64      `json:"id"`
	ResourceID   int64      `json:"resource_id"`
	ResourceType string     `json:"resource_type"`
	Path         string     `json:"path"`
	Token        string     `json:"token"`
	ExpiresAt    *time.Time `json:"expires_at"`
	Views        int64      `json:"views"`
	CreatedAt    time.Time  `json:"created_at"`
}

type ExportFriend struct {
	FriendID  string    `json:"friend_id"`
	Status    string    `json:"status"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type ExportActivity struct {
	FileID     *int64    `json:"file_id,omitempty"`
	FileName   string    `json:"file_name,omitempty"`
	FolderID   *int64    `json:"folder_id,omitempty"`
	FolderName string    `json:"folder_name,omitempty"`
	AccessedAt time.Time `json:"accessed_at"`
}

// ExportUserDataHandler exporte toutes les données personnelles de l'utilisateur
// au format JSON conformément au RGPD Article 20 (portabilité).
func ExportUserDataHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Contexte utilisateur non trouvé"})
		return
	}
	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Format d'ID utilisateur invalide"})
		return
	}

	ctx := c.Request.Context()

	profile, err := buildExportProfile(ctx, db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	exportFiles, err := fetchExportFiles(ctx, db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des fichiers"})
		return
	}

	exportFolders, err := fetchExportFolders(ctx, db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des dossiers"})
		return
	}

	exportTags, err := fetchExportTags(ctx, db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des tags"})
		return
	}

	exportShares, err := fetchExportShares(ctx, db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des partages"})
		return
	}

	exportFriends, err := fetchExportFriends(ctx, db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des amis"})
		return
	}

	exportActivity, err := fetchExportActivity(ctx, db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération de l'activité"})
		return
	}

	export := PortabilityExport{
		ExportMetadata: ExportMetadata{
			ExportDate:    time.Now().UTC().Format(time.RFC3339),
			Format:        "application/json",
			FormatVersion: "1.0",
			LegalBasis:    "RGPD Article 20 - Droit à la portabilité des données / Loi Informatique et Libertés art. 55",
			ServiceName:   "Kagibi",
			UserID:        userID,
		},
		Profile:        profile,
		Files:          exportFiles,
		Folders:        exportFolders,
		Tags:           exportTags,
		ShareLinks:     exportShares,
		Friends:        exportFriends,
		RecentActivity: exportActivity,
	}

	filename := fmt.Sprintf("kagibi-export-%s.json", time.Now().UTC().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.IndentedJSON(http.StatusOK, export)
}

func buildExportProfile(ctx context.Context, db *bun.DB, userID string) (ExportProfile, error) {
	var user pkg.User
	if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(ctx); err != nil {
		return ExportProfile{}, err
	}

	var plan pkg.UserPlan
	_ = db.NewSelect().Model(&plan).Where("user_id = ?", userID).Scan(ctx)

	return ExportProfile{
		ID:                  user.ID,
		Name:                user.Name,
		Email:               user.Email,
		AvatarURL:           user.AvatarURL,
		Plan:                plan.Plan,
		StorageUsed:         plan.StorageUsed,
		StorageLimit:        plan.StorageLimit,
		FriendCode:          user.FriendCode,
		PublicKey:           user.PublicKey,
		EncryptedPrivateKey: user.EncryptedPrivateKey,
		EncryptedMasterKey:  user.EncryptedMasterKey,
		Salt:                user.Salt,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
	}, nil
}

func fetchExportFiles(ctx context.Context, db *bun.DB, userID string) ([]ExportFile, error) {
	var files []pkg.File
	if err := db.NewSelect().Model(&files).
		Where("user_id = ?", userID).
		Where("is_preview = ?", false).
		OrderExpr("path ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	result := make([]ExportFile, len(files))
	for i, f := range files {
		result[i] = ExportFile{
			ID:           f.ID,
			Name:         f.Name,
			Path:         f.Path,
			Size:         f.Size,
			MimeType:     f.MimeType,
			EncryptedKey: f.EncryptedKey,
			Tags:         f.Tags,
			CreatedAt:    f.CreatedAt,
			UpdatedAt:    f.UpdatedAt,
		}
	}
	return result, nil
}

func fetchExportFolders(ctx context.Context, db *bun.DB, userID string) ([]ExportFolder, error) {
	var folders []pkg.Folder
	if err := db.NewSelect().Model(&folders).
		Where("user_id = ?", userID).
		OrderExpr("path ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	result := make([]ExportFolder, len(folders))
	for i, f := range folders {
		result[i] = ExportFolder{
			ID:           f.ID,
			Name:         f.Name,
			Path:         f.Path,
			EncryptedKey: f.EncryptedKey,
			Tags:         f.Tags,
			CreatedAt:    f.CreatedAt,
			UpdatedAt:    f.UpdatedAt,
		}
	}
	return result, nil
}

func fetchExportTags(ctx context.Context, db *bun.DB, userID string) ([]ExportTag, error) {
	var tags []pkg.Tag
	if err := db.NewSelect().Model(&tags).
		Where("user_id = ?", userID).
		OrderExpr("name ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	result := make([]ExportTag, len(tags))
	for i, t := range tags {
		result[i] = ExportTag{
			ID:    t.ID,
			Name:  t.Name,
			Color: t.Color,
		}
	}
	return result, nil
}

func fetchExportShares(ctx context.Context, db *bun.DB, userID string) ([]ExportShareLink, error) {
	var links []pkg.ShareLink
	if err := db.NewSelect().Model(&links).
		Where("owner_id = ?", userID).
		OrderExpr("created_at DESC").
		Scan(ctx); err != nil {
		return nil, err
	}

	result := make([]ExportShareLink, len(links))
	for i, l := range links {
		result[i] = ExportShareLink{
			ID:           l.ID,
			ResourceID:   l.ResourceID,
			ResourceType: l.ResourceType,
			Path:         l.Path,
			Token:        l.Token,
			ExpiresAt:    l.ExpiresAt,
			Views:        l.Views,
			CreatedAt:    l.CreatedAt,
		}
	}
	return result, nil
}

func fetchExportFriends(ctx context.Context, db *bun.DB, userID string) ([]ExportFriend, error) {
	var friendships []pkg.Friendship
	if err := db.NewSelect().Model(&friendships).
		Where("user_id_1 = ? OR user_id_2 = ?", userID, userID).
		OrderExpr("created_at DESC").
		Scan(ctx); err != nil {
		return nil, err
	}

	result := make([]ExportFriend, len(friendships))
	for i, f := range friendships {
		friendID := f.UserID2
		role := "initiator"
		if f.UserID2 == userID {
			friendID = f.UserID1
			role = "recipient"
		}
		result[i] = ExportFriend{
			FriendID:  friendID,
			Status:    f.Status,
			Role:      role,
			CreatedAt: f.CreatedAt,
		}
	}
	return result, nil
}

func fetchExportActivity(ctx context.Context, db *bun.DB, userID string) ([]ExportActivity, error) {
	var activities []pkg.RecentActivity
	if err := db.NewSelect().Model(&activities).
		Relation("File").
		Relation("Folder").
		Where("?TableAlias.user_id = ?", userID).
		OrderExpr("accessed_at DESC").
		Limit(500).
		Scan(ctx); err != nil {
		return nil, err
	}

	result := make([]ExportActivity, len(activities))
	for i, a := range activities {
		entry := ExportActivity{
			FileID:     a.FileID,
			FolderID:   a.FolderID,
			AccessedAt: a.AccessedAt,
		}
		if a.File != nil {
			entry.FileName = a.File.Name
		}
		if a.Folder != nil {
			entry.FolderName = a.Folder.Name
		}
		result[i] = entry
	}
	return result, nil
}

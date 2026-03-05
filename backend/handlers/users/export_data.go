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
	"fmt"
	"net/http"
	"safercloud/backend/pkg"
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

	// 1. Profil utilisateur
	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	planState, err := pkg.FindUserPlanByUserID(db, userID)
	if err != nil || planState == nil {
		planState = &pkg.UserPlan{
			UserID:       userID,
			Plan:         pkg.PlanFree,
			StorageLimit: pkg.StorageFree,
			StorageUsed:  0,
		}
	}

	profile := ExportProfile{
		ID:                  user.ID,
		Name:                user.Name,
		Email:               user.Email,
		AvatarURL:           user.AvatarURL,
		Plan:                planState.Plan,
		StorageUsed:         planState.StorageUsed,
		StorageLimit:        planState.StorageLimit,
		FriendCode:          user.FriendCode,
		PublicKey:           user.PublicKey,
		EncryptedPrivateKey: user.EncryptedPrivateKey,
		EncryptedMasterKey:  user.EncryptedMasterKey,
		Salt:                user.Salt,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
	}

	// 2. Fichiers (métadonnées uniquement, hors previews)
	var dbFiles []pkg.File
	err = db.NewSelect().Model(&dbFiles).
		Where("user_id = ? AND is_preview = false", userID).
		Order("path ASC", "name ASC").
		Scan(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des fichiers"})
		return
	}

	exportFiles := make([]ExportFile, 0, len(dbFiles))
	for _, f := range dbFiles {
		exportFiles = append(exportFiles, ExportFile{
			ID:           f.ID,
			Name:         f.Name,
			Path:         f.Path,
			Size:         f.Size,
			MimeType:     f.MimeType,
			EncryptedKey: f.EncryptedKey,
			Tags:         f.Tags,
			CreatedAt:    f.CreatedAt,
			UpdatedAt:    f.UpdatedAt,
		})
	}

	// 3. Dossiers
	var dbFolders []pkg.Folder
	err = db.NewSelect().Model(&dbFolders).
		Where("user_id = ?", userID).
		Order("path ASC").
		Scan(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des dossiers"})
		return
	}

	exportFolders := make([]ExportFolder, 0, len(dbFolders))
	for _, f := range dbFolders {
		exportFolders = append(exportFolders, ExportFolder{
			ID:           f.ID,
			Name:         f.Name,
			Path:         f.Path,
			EncryptedKey: f.EncryptedKey,
			Tags:         f.Tags,
			CreatedAt:    f.CreatedAt,
			UpdatedAt:    f.UpdatedAt,
		})
	}

	// 4. Tags
	var dbTags []pkg.Tag
	err = db.NewSelect().Model(&dbTags).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des tags"})
		return
	}

	exportTags := make([]ExportTag, 0, len(dbTags))
	for _, t := range dbTags {
		exportTags = append(exportTags, ExportTag{
			ID:    t.ID,
			Name:  t.Name,
			Color: t.Color,
		})
	}

	// 5. Liens de partage
	var dbShares []pkg.ShareLink
	err = db.NewSelect().Model(&dbShares).
		Where("owner_id = ?", userID).
		Scan(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des partages"})
		return
	}

	exportShares := make([]ExportShareLink, 0, len(dbShares))
	for _, s := range dbShares {
		exportShares = append(exportShares, ExportShareLink{
			ID:           s.ID,
			ResourceID:   s.ResourceID,
			ResourceType: s.ResourceType,
			Path:         s.Path,
			Token:        s.Token,
			ExpiresAt:    s.ExpiresAt,
			Views:        s.Views,
			CreatedAt:    s.CreatedAt,
		})
	}

	// 6. Amis
	var dbFriendships []pkg.Friendship
	err = db.NewSelect().Model(&dbFriendships).
		Where("user_id_1 = ? OR user_id_2 = ?", userID, userID).
		Scan(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des amis"})
		return
	}

	exportFriends := make([]ExportFriend, 0, len(dbFriendships))
	for _, f := range dbFriendships {
		friendID := f.UserID2
		role := "initiateur"
		if f.UserID1 != userID {
			friendID = f.UserID1
			role = "destinataire"
		}
		exportFriends = append(exportFriends, ExportFriend{
			FriendID:  friendID,
			Status:    f.Status,
			Role:      role,
			CreatedAt: f.CreatedAt,
		})
	}

	// 7. Activité récente
	var dbActivity []pkg.RecentActivity
	err = db.NewSelect().Model(&dbActivity).
		Where("?TableAlias.user_id = ?", userID).
		Relation("File").
		Relation("Folder").
		Order("accessed_at DESC").
		Scan(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération de l'activité"})
		return
	}

	exportActivity := make([]ExportActivity, 0, len(dbActivity))
	for _, a := range dbActivity {
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
		exportActivity = append(exportActivity, entry)
	}

	// Construction de l'export final
	export := PortabilityExport{
		ExportMetadata: ExportMetadata{
			ExportDate:    time.Now().UTC().Format(time.RFC3339),
			Format:        "application/json",
			FormatVersion: "1.0",
			LegalBasis:    "RGPD Article 20 - Droit à la portabilité des données / Loi Informatique et Libertés art. 55",
			ServiceName:   "SaferCloud",
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

	// Réponse en téléchargement JSON
	filename := fmt.Sprintf("safercloud-export-%s.json", time.Now().UTC().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.IndentedJSON(http.StatusOK, export)
}

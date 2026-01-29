package users

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"regexp"
	"safercloud/backend/pkg"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

// validateUsername validates and sanitizes the username
func validateUsername(name string) (string, error) {
	// 1. Length check
	if utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 100 {
		return "", fmt.Errorf("username must be between 1 and 100 characters")
	}

	// 2. Allowed characters (letters, numbers, spaces, hyphens, underscores, accented characters)
	validName := regexp.MustCompile(`^[a-zA-Z0-9\s\-_À-ÿ]+$`)
	if !validName.MatchString(name) {
		return "", fmt.Errorf("username contains invalid characters")
	}

	// 3. No control characters
	for _, r := range name {
		if r < 32 || r == 127 {
			return "", fmt.Errorf("username contains control characters")
		}
	}

	// 4. Sanitize HTML entities (defense in depth)
	sanitized := html.EscapeString(name)

	// 5. Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized, nil
}

// UpdateProfileHandler handles updating user profile (name)
func UpdateProfileHandler(c *gin.Context, db *bun.DB) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Custom validation
	sanitizedName, err := validateUsername(req.Name)
	if err != nil {
		log.Printf("SECURITY: Invalid username attempt - Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID, _ := userIDInterface.(string)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Fetch user
	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user name with sanitized value
	user.Name = sanitizedName

	_, err = db.NewUpdate().Model(user).Column("name").Where("id = ?", userID).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}

	log.Printf("INFO: Profile updated - UserID: %s, Name: %s", userID, sanitizedName)
	c.JSON(http.StatusOK, user)
}

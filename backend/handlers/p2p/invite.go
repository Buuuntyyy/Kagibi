// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package p2p

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/authprovider"
	"kagibi/backend/pkg/mailer"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// guestTokenIssuer is satisfied by LocalProvider only.
type guestTokenIssuer interface {
	GenerateGuestToken(guestUserID string, expiresAt time.Time) (string, error)
}

// generateUUID generates a random RFC 4122 v4 UUID.
func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// CreateInviteHandler creates a P2P invite for an authenticated user.
// The invite is always a guest invite — any user (with or without a Kagibi account)
// can follow the link without registering. The optional recipient_email is used only
// to send an email notification.
//
// POST /api/v1/p2p/invite
func CreateInviteHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		senderID, _ := c.Get("user_id")

		var req struct {
			RecipientEmail string `json:"recipient_email"`
			FileName       string `json:"file_name" binding:"required"`
			FileSize       int64  `json:"file_size" binding:"required"`
			TransferID     string `json:"transfer_id" binding:"required"`
			SendEmail      bool   `json:"send_email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate optional email format when provided
		if req.RecipientEmail != "" {
			req.RecipientEmail = strings.ToLower(strings.TrimSpace(req.RecipientEmail))
		}

		// Load sender profile for name + self-invite check
		var sender pkg.User
		if err := db.NewSelect().Model(&sender).Where("id = ?", senderID).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load sender profile"})
			return
		}
		if req.RecipientEmail != "" && strings.EqualFold(sender.Email, req.RecipientEmail) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot invite yourself"})
			return
		}

		// Generate opaque invite token (32 bytes → 43 URL-safe chars)
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		token := base64.RawURLEncoding.EncodeToString(raw)

		// Guest invite: auto-generate recipient UUID
		guestID, err := generateUUID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate guest ID"})
			return
		}

		invite := &pkg.P2PInvite{
			Token:          token,
			SenderID:       senderID.(string),
			SenderName:     sender.Name,
			RecipientEmail: req.RecipientEmail,
			RecipientID:    guestID,
			TransferID:     req.TransferID,
			FileName:       req.FileName,
			FileSize:       req.FileSize,
			IsGuest:        true,
			ExpiresAt:      time.Now().Add(24 * time.Hour),
		}

		if _, err := db.NewInsert().Model(invite).Exec(c.Request.Context()); err != nil {
			log.Printf("[P2P] Failed to create invite sender=%s: %v", senderID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invite"})
			return
		}

		if req.SendEmail && req.RecipientEmail != "" {
			go func() {
				if err := mailer.SendP2PInvite(req.RecipientEmail, sender.Name, req.FileName, req.FileSize, token); err != nil {
					log.Printf("[P2P] Invite email failed to %s: %v", req.RecipientEmail, err)
				}
			}()
		}

		c.JSON(http.StatusOK, gin.H{
			"token":          token,
			"recipient_id":   guestID,
			"recipient_name": "Guest",
			"expires_at":     invite.ExpiresAt,
		})
	}
}

// GuestAuthHandler issues a short-lived guest JWT for a P2P invite link recipient.
// No prior account is required. The guest JWT authorises only the WebSocket + P2P
// invite/accept endpoints for the duration of the invite.
//
// POST /api/v1/p2p/guest-auth  (public — no auth required)
func GuestAuthHandler(db *bun.DB, provider authprovider.AuthProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		issuer, ok := provider.(guestTokenIssuer)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Guest auth is only supported with the local auth provider"})
			return
		}

		var req struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var invite pkg.P2PInvite
		if err := db.NewSelect().Model(&invite).Where("token = ?", req.Token).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found"})
			return
		}
		if time.Now().After(invite.ExpiresAt) {
			c.JSON(http.StatusGone, gin.H{"error": "This invite has expired"})
			return
		}
		if !invite.IsGuest {
			c.JSON(http.StatusForbidden, gin.H{"error": "This invite is not a guest invite"})
			return
		}

		jwt, err := issuer.GenerateGuestToken(invite.RecipientID, invite.ExpiresAt)
		if err != nil {
			log.Printf("[P2P] Failed to generate guest token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate guest token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"jwt":           jwt,
			"guest_user_id": invite.RecipientID,
			"sender_name":   invite.SenderName,
			"file_name":     invite.FileName,
			"file_size":     invite.FileSize,
			"expires_at":    invite.ExpiresAt,
		})
	}
}

// GetInviteHandler retrieves an invite. Requires the caller to be the intended
// recipient (RecipientID == user_id from JWT — works for both regular and guest JWTs).
//
// GET /api/v1/p2p/invite/:token
func GetInviteHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		token := c.Param("token")

		var invite pkg.P2PInvite
		if err := db.NewSelect().Model(&invite).Where("token = ?", token).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found"})
			return
		}

		if time.Now().After(invite.ExpiresAt) {
			c.JSON(http.StatusGone, gin.H{"error": "This invite has expired"})
			return
		}

		if invite.RecipientID != userID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "This invite is not for your account"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":       invite.Token,
			"sender_id":   invite.SenderID,
			"sender_name": invite.SenderName,
			"file_name":   invite.FileName,
			"file_size":   invite.FileSize,
			"transfer_id": invite.TransferID,
			"expires_at":  invite.ExpiresAt,
			"accepted":    invite.AcceptedAt != nil,
			"is_guest":    invite.IsGuest,
		})
	}
}

// AcceptInviteHandler marks the invite as accepted and signals the sender via
// the P2P WebSocket channel so they can initiate the WebRTC handshake.
// For guest invites, the caller must provide their freshly-generated public_key
// so the sender can encrypt the file key for the recipient.
//
// POST /api/v1/p2p/invite/:token/accept
func AcceptInviteHandler(db *bun.DB, hub pkg.WSHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		token := c.Param("token")

		var body struct {
			PublicKey string `json:"public_key"`
		}
		// Best-effort body parse — public_key is optional for non-guest invites
		_ = c.ShouldBindJSON(&body)

		var invite pkg.P2PInvite
		if err := db.NewSelect().Model(&invite).Where("token = ?", token).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found"})
			return
		}

		if time.Now().After(invite.ExpiresAt) {
			c.JSON(http.StatusGone, gin.H{"error": "This invite has expired"})
			return
		}

		if invite.RecipientID != userID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "This invite is not for your account"})
			return
		}

		// Idempotent: already accepted invites are still OK
		if invite.AcceptedAt == nil {
			now := time.Now()
			invite.AcceptedAt = &now
			db.NewUpdate().Model(&invite).Column("accepted_at").Where("id = ?", invite.ID).Exec(c.Request.Context())
		}

		// Build signal payload — include public_key so sender can encrypt the file key
		signalPayload := map[string]any{
			"transfer_id":  invite.TransferID,
			"invite_token": token,
		}
		if body.PublicKey != "" {
			signalPayload["public_key"] = body.PublicKey
		}

		signal := &pkg.P2PSignal{
			SenderID:   userID.(string),
			TargetID:   invite.SenderID,
			SignalType: "invite_accepted",
			Payload:    signalPayload,
		}
		if _, err := db.NewInsert().Model(signal).Exec(c.Request.Context()); err != nil {
			log.Printf("[P2P] Failed to store invite_accepted signal: %v", err)
		}
		hub.SendP2PSignalToUser(invite.SenderID, userID.(string), "invite_accepted", signal.ID, signalPayload)

		c.JSON(http.StatusOK, gin.H{
			"status":      "accepted",
			"transfer_id": invite.TransferID,
			"sender_id":   invite.SenderID,
		})
	}
}

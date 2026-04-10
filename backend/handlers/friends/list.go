// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// internal/handlers/friends/list.go
package friends

import (
	"context"
	"kagibi/backend/pkg"
	"net/http"

	"github.com/uptrace/bun"

	"github.com/gin-gonic/gin"
)

type FriendResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`               // Maybe hide this if privacy concerned, but usually visible to friends
	Status    string `json:"status"`              // pending_sent, pending_received, accepted
	RequestID int64  `json:"requestId,omitempty"` // ID of the friendship row, useful for cancelling/accepting
	PublicKey string `json:"public_key"`          // NEW: Required for encrypted sharing
	Online    bool   `json:"online"`              // REALTIME STATUS - Now handled by Supabase Presence
}

type FriendHandler struct {
	DB       *bun.DB
	IsOnline func(userID string) bool // injected from WebSocket hub; nil-safe
}

func NewFriendHandler(db *bun.DB, isOnline func(string) bool) *FriendHandler {
	return &FriendHandler{DB: db, IsOnline: isOnline}
}

func (h *FriendHandler) ListFriends(c *gin.Context) {
	currentUserID := c.GetString("user_id")

	friendships, err := h.getUserFriendships(c.Request.Context(), currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des amis"})
		return
	}

	results := h.buildFriendResponses(c.Request.Context(), currentUserID, friendships)

	c.JSON(http.StatusOK, results)
}

func (h *FriendHandler) getUserFriendships(ctx context.Context, userID string) ([]pkg.Friendship, error) {
	var friendships []pkg.Friendship
	err := h.DB.NewSelect().
		Model(&friendships).
		Where("user_id_1 = ?", userID).
		WhereOr("user_id_2 = ?", userID).
		Scan(ctx)
	return friendships, err
}

func (h *FriendHandler) buildFriendResponses(ctx context.Context, currentUserID string, friendships []pkg.Friendship) []FriendResponse {
	var results []FriendResponse
	for _, f := range friendships {
		if resp := h.processSingleFriendship(ctx, currentUserID, f); resp != nil {
			results = append(results, *resp)
		}
	}
	return results
}

func (h *FriendHandler) processSingleFriendship(ctx context.Context, currentUserID string, f pkg.Friendship) *FriendResponse {
	otherUserID, status := determineFriendStatus(currentUserID, f)

	var otherUser pkg.User
	if err := h.DB.NewSelect().Model(&otherUser).Where("id = ?", otherUserID).Scan(ctx); err != nil {
		return nil
	}

	online := h.IsOnline != nil && h.IsOnline(otherUser.ID)

	return &FriendResponse{
		ID:        otherUser.ID,
		Name:      otherUser.Name,
		Email:     otherUser.Email,
		Status:    status,
		RequestID: f.ID,
		PublicKey: otherUser.PublicKey,
		Online:    online,
	}
}

func determineFriendStatus(currentUserID string, f pkg.Friendship) (string, string) {
	if f.Status == "accepted" {
		if f.UserID1 == currentUserID {
			return f.UserID2, "accepted"
		}
		return f.UserID1, "accepted"
	}

	if f.UserID1 == currentUserID {
		return f.UserID2, "pending_sent"
	}
	return f.UserID1, "pending_received"
}

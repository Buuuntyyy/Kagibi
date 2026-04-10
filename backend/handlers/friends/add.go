// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// internal/handlers/friends/add.go
package friends

import (
	"log"
	"net/http"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

const logFriendUpdateFailed = "Failed to emit friend_update event: %v"

type AddFriendRequest struct {
	FriendCode string `json:"friendCode"`
}

func (h *FriendHandler) AddFriend(c *gin.Context) {
	var req AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides"})
		return
	}

	currentUserID := c.GetString("user_id")

	// 1. Find user by friend code
	var targetUser pkg.User
	err := h.DB.NewSelect().Model(&targetUser).Where("friend_code = ?", req.FriendCode).Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Code ami introuvable"})
		return
	}

	if targetUser.ID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vous ne pouvez pas vous ajouter vous-même"})
		return
	}

	// 2. Check if friendship already exists
	exists, err := h.DB.NewSelect().Model((*pkg.Friendship)(nil)).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("user_id_1 = ? AND user_id_2 = ?", currentUserID, targetUser.ID).
				WhereOr("user_id_1 = ? AND user_id_2 = ?", targetUser.ID, currentUserID)
		}).Exists(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Vous êtes déjà amis ou une demande est en attente"})
		return
	}

	// 3. Create Friendship request (Pending)
	friendship := &pkg.Friendship{
		UserID1:   currentUserID,
		UserID2:   targetUser.ID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	_, err = h.DB.NewInsert().Model(friendship).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de créer la demande d'ami"})
		return
	}

	// Notify target user via Supabase Realtime
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), h.DB, targetUser.ID, "friend_update", map[string]interface{}{
		"action": "friend_request_received",
	}); err != nil {
		log.Printf(logFriendUpdateFailed, err)
	}
	// Notify sender
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), h.DB, currentUserID, "friend_update", map[string]interface{}{
		"action": "friend_request_sent",
	}); err != nil {
		log.Printf(logFriendUpdateFailed, err)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Demande d'ami envoyée"})
}

func (h *FriendHandler) AcceptFriend(c *gin.Context) {
	friendshipID := c.Param("id")
	currentUserID := c.GetString("user_id")

	var friendship pkg.Friendship
	// Verify the request exists and is for the current user
	res, err := h.DB.NewUpdate().
		Model(&friendship).
		Set("status = ?", "accepted").
		Where("id = ? AND user_id_2 = ? AND status = 'pending'", friendshipID, currentUserID).
		Returning("*").
		Exec(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'acceptation"})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Demande introuvable ou déjà traitée"})
		return
	}

	// Notify requester via Supabase Realtime
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), h.DB, friendship.UserID1, "friend_update", map[string]interface{}{
		"action": "friend_request_accepted",
	}); err != nil {
		log.Printf(logFriendUpdateFailed, err)
	}
	// Notify accepter
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), h.DB, currentUserID, "friend_update", map[string]interface{}{
		"action": "friend_request_accepted",
	}); err != nil {
		log.Printf(logFriendUpdateFailed, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ami accepté"})
}

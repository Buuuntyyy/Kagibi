// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"net/http"
	"strconv"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

func (h *OrgHandler) ListFavorites(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	userID := c.GetString("user_id")
	role, err := h.memberRole(c.Request.Context(), orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var favs []pkg.OrgFavorite
	if err := h.DB.NewSelect().Model(&favs).
		Where("ofav.org_id = ? AND ofav.user_id = ?", orgID, userID).
		OrderExpr("ofav.created_at ASC").
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if favs == nil {
		favs = []pkg.OrgFavorite{}
	}
	c.JSON(http.StatusOK, favs)
}

func (h *OrgHandler) AddFavorite(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	userID := c.GetString("user_id")
	role, err := h.memberRole(c.Request.Context(), orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var body struct {
		ItemID   int64  `json:"item_id" binding:"required"`
		ItemType string `json:"item_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || (body.ItemType != "file" && body.ItemType != "folder") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id and item_type (file|folder) required"})
		return
	}

	fav := &pkg.OrgFavorite{OrgID: orgID, UserID: userID, ItemID: body.ItemID, ItemType: body.ItemType}
	if _, err := h.DB.NewInsert().Model(fav).
		On("CONFLICT (org_id, user_id, item_id, item_type) DO NOTHING").
		Returning("*").
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fav)
}

func (h *OrgHandler) RemoveFavorite(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	userID := c.GetString("user_id")
	role, err := h.memberRole(c.Request.Context(), orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	itemType := c.Param("itemType")
	itemID, err := strconv.ParseInt(c.Param("itemID"), 10, 64)
	if err != nil || (itemType != "file" && itemType != "folder") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}

	if _, err := h.DB.NewDelete().Model((*pkg.OrgFavorite)(nil)).
		Where("org_id = ? AND user_id = ? AND item_id = ? AND item_type = ?", orgID, userID, itemID, itemType).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

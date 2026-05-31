// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package notifications

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"kagibi/backend/pkg"
)

// ListNotifications returns the 50 most recent notifications for the current user.
func ListNotifications(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	var notifs []pkg.Notification
	if err := db.NewSelect().Model(&notifs).
		Where("user_id = ?", userID.(string)).
		Order("created_at DESC").
		Limit(50).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"notifications": notifs})
}

// GetUnreadCount returns the number of unread notifications.
func GetUnreadCount(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	count, err := db.NewSelect().Model((*pkg.Notification)(nil)).
		Where("user_id = ? AND is_read = false", userID.(string)).
		Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkNotificationRead marks a single notification as read.
func MarkNotificationRead(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}
	if _, err := db.NewUpdate().Model((*pkg.Notification)(nil)).
		Set("is_read = true").
		Where("id = ? AND user_id = ?", id, userID.(string)).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notification"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "read"})
}

// MarkAllRead marks all notifications of the current user as read.
func MarkAllRead(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	if _, err := db.NewUpdate().Model((*pkg.Notification)(nil)).
		Set("is_read = true").
		Where("user_id = ?", userID.(string)).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "done"})
}

// DeleteNotification permanently removes a notification owned by the current user.
func DeleteNotification(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}
	if _, err := db.NewDelete().Model((*pkg.Notification)(nil)).
		Where("id = ? AND user_id = ?", id, userID.(string)).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete notification"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

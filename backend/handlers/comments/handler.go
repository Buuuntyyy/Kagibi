// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package comments

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"kagibi/backend/pkg"
)

// commentResponse adds read status and author display name on top of the stored model.
type commentResponse struct {
	pkg.FileComment
	AuthorName string `json:"author_name"`
	IsRead     bool   `json:"is_read"`
}

// buildResponses enriches a slice of comments with author names and per-user read state.
func buildResponses(ctx context.Context, db *bun.DB, comments []pkg.FileComment, userID string) []commentResponse {
	// Batch-fetch author names
	seenAuthors := map[string]bool{}
	for _, c := range comments {
		seenAuthors[c.AuthorID] = true
	}
	authorIDs := make([]string, 0, len(seenAuthors))
	for id := range seenAuthors {
		authorIDs = append(authorIDs, id)
	}
	nameMap := map[string]string{}
	if len(authorIDs) > 0 {
		var users []pkg.User
		_ = db.NewSelect().Model(&users).Where("id IN (?)", bun.In(authorIDs)).Scan(ctx)
		for _, u := range users {
			nameMap[u.ID] = u.Name
		}
	}

	// Batch-fetch read state for current user
	commentIDs := make([]int64, len(comments))
	for i, c := range comments {
		commentIDs[i] = c.ID
	}
	readMap := map[int64]bool{}
	if len(commentIDs) > 0 {
		var reads []pkg.FileCommentRead
		_ = db.NewSelect().Model(&reads).
			Where("comment_id IN (?) AND user_id = ?", bun.In(commentIDs), userID).
			Scan(ctx)
		for _, r := range reads {
			readMap[r.CommentID] = true
		}
	}

	result := make([]commentResponse, len(comments))
	for i, c := range comments {
		result[i] = commentResponse{FileComment: c, AuthorName: nameMap[c.AuthorID], IsRead: readMap[c.ID]}
	}
	return result
}

// ListFileComments returns all comments on a personal file visible to the caller.
func ListFileComments(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	uid := userID.(string)
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	if !canAccessFile(c.Request.Context(), db, fileID, uid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var comments []pkg.FileComment
	if err := db.NewSelect().Model(&comments).
		Where("file_id = ?", fileID).
		Order("created_at ASC").
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch comments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": buildResponses(c.Request.Context(), db, comments, uid)})
}

// AddFileComment creates a new comment on a personal file.
func AddFileComment(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	uid := userID.(string)
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	var req struct {
		Content  string `json:"content" binding:"required"`
		ParentID *int64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var file pkg.File
	if err := db.NewSelect().Model(&file).Where("id = ?", fileID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	if !canAccessFile(c.Request.Context(), db, fileID, uid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var author pkg.User
	_ = db.NewSelect().Model(&author).Where("id = ?", uid).Scan(c.Request.Context())

	comment := &pkg.FileComment{
		FileID:   &fileID,
		AuthorID: uid,
		Content:  req.Content,
		ParentID: req.ParentID,
	}
	if _, err := db.NewInsert().Model(comment).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	// Author auto-reads their own comment
	autoRead(c.Request.Context(), db, comment.ID, uid)

	go notifyFileComment(db, file, comment, author.Name)
	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

// ListOrgFileComments returns all comments on an org file for org members.
func ListOrgFileComments(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	uid := userID.(string)
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org ID"})
		return
	}
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	if !isOrgMember(c.Request.Context(), db, orgID, uid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var comments []pkg.FileComment
	if err := db.NewSelect().Model(&comments).
		Where("org_file_id = ?", fileID).
		Order("created_at ASC").
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch comments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": buildResponses(c.Request.Context(), db, comments, uid)})
}

// AddOrgFileComment creates a new comment on an org file.
func AddOrgFileComment(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	uid := userID.(string)
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org ID"})
		return
	}
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	var req struct {
		Content  string `json:"content" binding:"required"`
		ParentID *int64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isOrgMember(c.Request.Context(), db, orgID, uid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var file pkg.OrgFile
	if err := db.NewSelect().Model(&file).
		Where("id = ? AND org_id = ?", fileID, orgID).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var author pkg.User
	_ = db.NewSelect().Model(&author).Where("id = ?", uid).Scan(c.Request.Context())

	comment := &pkg.FileComment{
		OrgFileID: &fileID,
		OrgID:     &orgID,
		AuthorID:  uid,
		Content:   req.Content,
		ParentID:  req.ParentID,
	}
	if _, err := db.NewInsert().Model(comment).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	autoRead(c.Request.Context(), db, comment.ID, uid)
	go notifyOrgFileComment(db, file, orgID, comment, author.Name, uid)
	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

// EditComment updates the content of a comment (author only).
func EditComment(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	uid := userID.(string)
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var comment pkg.FileComment
	if err := db.NewSelect().Model(&comment).Where("id = ?", commentID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	if comment.AuthorID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	comment.Content = req.Content
	comment.UpdatedAt = time.Now()
	if _, err := db.NewUpdate().Model(&comment).Column("content", "updated_at").Where("id = ?", commentID).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update comment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

// DeleteComment removes a comment (author only).
func DeleteComment(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	uid := userID.(string)
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}

	var comment pkg.FileComment
	if err := db.NewSelect().Model(&comment).Where("id = ?", commentID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	if comment.AuthorID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if _, err := db.NewDelete().Model(&comment).Where("id = ?", commentID).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete comment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// MarkCommentRead records that the current user has read a comment.
func MarkCommentRead(c *gin.Context, db *bun.DB) {
	userID, _ := c.Get("user_id")
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}
	autoRead(c.Request.Context(), db, commentID, userID.(string))
	c.JSON(http.StatusOK, gin.H{"status": "read"})
}

// ResolveComment toggles the resolved state of a comment.
func ResolveComment(c *gin.Context, db *bun.DB) {
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}

	var req struct {
		IsResolved bool `json:"is_resolved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := db.NewUpdate().Model((*pkg.FileComment)(nil)).
		Set("is_resolved = ?, updated_at = ?", req.IsResolved, time.Now()).
		Where("id = ?", commentID).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update comment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// BatchCommentCounts returns, for each file ID, the total number of unresolved comments.
func BatchCommentCounts(c *gin.Context, db *bun.DB) {
	var req struct {
		FileIDs    []int64 `json:"file_ids"`
		OrgFileIDs []int64 `json:"org_file_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	type countRow struct {
		ID    int64 `bun:"id"`
		Count int   `bun:"count"`
	}

	fileCounts := map[int64]int{}
	if len(req.FileIDs) > 0 {
		var rows []countRow
		_ = db.NewSelect().
			TableExpr("file_comments fc").
			ColumnExpr("fc.file_id AS id, COUNT(*) AS count").
			Where("fc.file_id IN (?)", bun.In(req.FileIDs)).
			Where("fc.is_resolved = false").
			GroupExpr("fc.file_id").
			Scan(c.Request.Context(), &rows)
		for _, r := range rows {
			fileCounts[r.ID] = r.Count
		}
	}

	orgFileCounts := map[int64]int{}
	if len(req.OrgFileIDs) > 0 {
		var rows []countRow
		_ = db.NewSelect().
			TableExpr("file_comments fc").
			ColumnExpr("fc.org_file_id AS id, COUNT(*) AS count").
			Where("fc.org_file_id IN (?)", bun.In(req.OrgFileIDs)).
			Where("fc.is_resolved = false").
			GroupExpr("fc.org_file_id").
			Scan(c.Request.Context(), &rows)
		for _, r := range rows {
			orgFileCounts[r.ID] = r.Count
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"file_counts":     fileCounts,
		"org_file_counts": orgFileCounts,
	})
}

// ── helpers ──────────────────────────────────────────────────────────────────

func autoRead(ctx context.Context, db *bun.DB, commentID int64, userID string) {
	r := &pkg.FileCommentRead{CommentID: commentID, UserID: userID}
	_, _ = db.NewInsert().Model(r).On("CONFLICT DO NOTHING").Exec(ctx)
}

func canAccessFile(ctx context.Context, db *bun.DB, fileID int64, userID string) bool {
	var file pkg.File
	if err := db.NewSelect().Model(&file).Where("id = ?", fileID).Scan(ctx); err != nil {
		return false
	}
	// 1. Owner
	if file.UserID == userID {
		return true
	}
	// 2. Direct file share
	var fileShare pkg.FileShare
	if err := db.NewSelect().Model(&fileShare).
		Where("file_id = ? AND shared_with_user_id = ?", fileID, userID).
		Scan(ctx); err == nil {
		return true
	}
	// 3. Folder share: check any folder containing this file that is shared with the user
	var folderFileKeys []pkg.FolderFileKey
	if err := db.NewSelect().Model(&folderFileKeys).
		Where("file_id = ?", fileID).
		Scan(ctx); err == nil && len(folderFileKeys) > 0 {
		folderIDs := make([]int64, len(folderFileKeys))
		for i, k := range folderFileKeys {
			folderIDs[i] = k.FolderID
		}
		var folderShare pkg.FolderShare
		if err := db.NewSelect().Model(&folderShare).
			Where("folder_id IN (?) AND shared_with_user_id = ?", bun.In(folderIDs), userID).
			Scan(ctx); err == nil {
			return true
		}
	}
	return false
}

func isOrgMember(ctx context.Context, db *bun.DB, orgID int64, userID string) bool {
	var m pkg.OrgMember
	return db.NewSelect().Model(&m).
		Where("org_id = ? AND user_id = ?", orgID, userID).
		Scan(ctx) == nil
}

func notifyFileComment(db *bun.DB, file pkg.File, comment *pkg.FileComment, actorName string) {
	ctx := context.Background()

	// Folder path for navigation (strip filename from path)
	folderPath := parentPath(file.Path)

	// Collect recipients for "comment_added"
	recipients := map[string]bool{file.UserID: true}
	var fileShares []pkg.FileShare
	_ = db.NewSelect().Model(&fileShares).Where("file_id = ?", *comment.FileID).Scan(ctx)
	for _, s := range fileShares {
		recipients[s.SharedWithUserID] = true
	}
	delete(recipients, comment.AuthorID)

	for recipientID := range recipients {
		notif := &pkg.Notification{
			UserID: recipientID, ActorID: comment.AuthorID, ActorName: actorName,
			Type: "comment_added", ResourceID: file.ID, ResourceType: "file",
			ResourceName: file.Name, ResourcePath: folderPath, CommentID: &comment.ID,
		}
		_, _ = db.NewInsert().Model(notif).Exec(ctx)
		_ = pkg.EmitRealtimeEvent(ctx, db, recipientID, "notification_update", map[string]any{
			"type": "comment_added", "resource_name": file.Name, "actor_name": actorName,
		})
	}

	// If this is a reply, also notify the parent comment author specifically.
	if comment.ParentID != nil {
		var parent pkg.FileComment
		if err := db.NewSelect().Model(&parent).Where("id = ?", *comment.ParentID).Scan(ctx); err == nil {
			if parent.AuthorID != comment.AuthorID && !recipients[parent.AuthorID] {
				notif := &pkg.Notification{
					UserID: parent.AuthorID, ActorID: comment.AuthorID, ActorName: actorName,
					Type: "reply_added", ResourceID: file.ID, ResourceType: "file",
					ResourceName: file.Name, ResourcePath: folderPath, CommentID: &comment.ID,
				}
				_, _ = db.NewInsert().Model(notif).Exec(ctx)
				_ = pkg.EmitRealtimeEvent(ctx, db, parent.AuthorID, "notification_update", map[string]any{
					"type": "reply_added", "resource_name": file.Name, "actor_name": actorName,
				})
			}
		}
	}
}

func notifyOrgFileComment(db *bun.DB, file pkg.OrgFile, orgID int64, comment *pkg.FileComment, actorName, actorID string) {
	ctx := context.Background()

	folderPath := file.FolderPath

	var members []pkg.OrgMember
	_ = db.NewSelect().Model(&members).Where("org_id = ?", orgID).Scan(ctx)

	notified := map[string]bool{}
	for _, m := range members {
		if m.UserID == actorID {
			continue
		}
		notif := &pkg.Notification{
			UserID: m.UserID, ActorID: actorID, ActorName: actorName,
			Type: "comment_added", ResourceID: file.ID, ResourceType: "org_file",
			ResourceName: file.Name, ResourcePath: folderPath, OrgID: &orgID, CommentID: &comment.ID,
		}
		_, _ = db.NewInsert().Model(notif).Exec(ctx)
		_ = pkg.EmitRealtimeEvent(ctx, db, m.UserID, "notification_update", map[string]any{
			"type": "comment_added", "resource_name": file.Name, "actor_name": actorName,
		})
		notified[m.UserID] = true
	}

	// Notify parent comment author for replies (if not already an org member)
	if comment.ParentID != nil {
		var parent pkg.FileComment
		if err := db.NewSelect().Model(&parent).Where("id = ?", *comment.ParentID).Scan(ctx); err == nil {
			if parent.AuthorID != actorID && !notified[parent.AuthorID] {
				notif := &pkg.Notification{
					UserID: parent.AuthorID, ActorID: actorID, ActorName: actorName,
					Type: "reply_added", ResourceID: file.ID, ResourceType: "org_file",
					ResourceName: file.Name, ResourcePath: folderPath, OrgID: &orgID, CommentID: &comment.ID,
				}
				_, _ = db.NewInsert().Model(notif).Exec(ctx)
				_ = pkg.EmitRealtimeEvent(ctx, db, parent.AuthorID, "notification_update", map[string]any{
					"type": "reply_added", "resource_name": file.Name, "actor_name": actorName,
				})
			}
		}
	}
}

// parentPath returns the directory portion of a full file path.
func parentPath(filePath string) string {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' {
			if i == 0 {
				return "/"
			}
			return filePath[:i]
		}
	}
	return "/"
}

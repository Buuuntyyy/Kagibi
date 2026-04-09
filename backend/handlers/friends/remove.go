// internal/handlers/friends/remove.go
package friends

import (
	"kagibi/backend/pkg"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func (h *FriendHandler) RemoveFriend(c *gin.Context) {
	friendID := c.Param("id") // This is the User ID of the friend to remove, NOT the friendship ID
	currentUserID := c.GetString("user_id")

	// Delete relation where (u1=me AND u2=friend) OR (u1=friend AND u2=me)
	_, err := h.DB.NewDelete().
		Model((*pkg.Friendship)(nil)).
		WhereGroup(" AND ", func(q *bun.DeleteQuery) *bun.DeleteQuery {
			return q.Where("user_id_1 = ? AND user_id_2 = ?", currentUserID, friendID).
				WhereOr("user_id_1 = ? AND user_id_2 = ?", friendID, currentUserID)
		}).
		Exec(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de supprimer l'ami"})
		return
	}

	// Notify the removed friend via Supabase Realtime
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), h.DB, friendID, "friend_update", map[string]interface{}{
		"action": "friend_removed",
	}); err != nil {
		log.Printf(logFriendUpdateFailed, err)
	}
	// Notify self
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), h.DB, currentUserID, "friend_update", map[string]interface{}{
		"action": "friend_removed",
	}); err != nil {
		log.Printf(logFriendUpdateFailed, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ami supprimé"})
}

func (h *FriendHandler) RejectFriend(c *gin.Context) {
	friendshipID := c.Param("id")
	currentUserID := c.GetString("user_id")

	var friendship pkg.Friendship
	// Verify request exists and addressed to me
	err := h.DB.NewDelete().
		Model(&friendship).
		Where("id = ? AND user_id_2 = ? AND status = 'pending'", friendshipID, currentUserID).
		Returning("*").
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de rejeter la demande"})
		return
	}

	// Notify sender via Supabase Realtime
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), h.DB, friendship.UserID1, "friend_update", map[string]interface{}{
		"action": "friend_request_rejected",
	}); err != nil {
		log.Printf(logFriendUpdateFailed, err)
	}
	// Notify self
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), h.DB, currentUserID, "friend_update", map[string]interface{}{
		"action": "friend_request_rejected",
	}); err != nil {
		log.Printf(logFriendUpdateFailed, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Demande rejetée"})
}

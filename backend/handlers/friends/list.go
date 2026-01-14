// internal/handlers/friends/list.go
package friends

import (
	"net/http"
	"safercloud/backend/pkg"
	"safercloud/backend/pkg/ws"

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
	Online    bool   `json:"online"`              // REALTIME STATUS
}

type FriendHandler struct {
	DB *bun.DB
	WS *ws.Manager
}

func NewFriendHandler(db *bun.DB, ws *ws.Manager) *FriendHandler {
	return &FriendHandler{DB: db, WS: ws}
}

func (h *FriendHandler) ListFriends(c *gin.Context) {
	currentUserID := c.GetString("user_id")

	var friendships []pkg.Friendship
	// Fetch all friendships involving me
	err := h.DB.NewSelect().
		Model(&friendships).
		Where("user_id_1 = ?", currentUserID).
		WhereOr("user_id_2 = ?", currentUserID).
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des amis"})
		return
	}

	var results []FriendResponse

	for _, f := range friendships {
		var otherUserID string
		var status string
		var requestID int64 = f.ID

		if f.Status == "accepted" {
			status = "accepted"
			if f.UserID1 == currentUserID {
				otherUserID = f.UserID2
			} else {
				otherUserID = f.UserID1
			}
		} else {
			// Pending logic
			if f.UserID1 == currentUserID {
				status = "pending_sent"
				otherUserID = f.UserID2
			} else {
				status = "pending_received"
				otherUserID = f.UserID1
			}
		}

		// Fetch other user details (optimize this with a Join later if needed)
		var otherUser pkg.User
		err := h.DB.NewSelect().Model(&otherUser).Where("id = ?", otherUserID).Scan(c.Request.Context())
		if err == nil {
			isOnline := false
			if h.WS != nil {
				isOnline = h.WS.IsUserOnline(otherUser.ID)
			}

			results = append(results, FriendResponse{
				ID:        otherUser.ID,
				Name:      otherUser.Name,
				Email:     otherUser.Email,
				Status:    status,
				RequestID: requestID,
				PublicKey: otherUser.PublicKey,
				Online:    isOnline,
			})
		}
	}

	c.JSON(http.StatusOK, results)
}

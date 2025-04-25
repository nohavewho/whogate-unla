package handler

import (
	"github.com/mcp-ecosystem/mcp-gateway/internal/apiserver/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Chat struct {
	db database.Database
}

func NewChat(db database.Database) *Chat {
	return &Chat{db: db}
}

func (h *Chat) HandleGetChatSessions(c *gin.Context) {
	sessions, err := h.db.GetSessions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get chat sessions"})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

func (h *Chat) HandleGetChatMessages(c *gin.Context) {
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	// Get pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Get messages with pagination
	messages, err := h.db.GetMessagesWithPagination(c.Request.Context(), sessionId, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

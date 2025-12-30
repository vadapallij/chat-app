package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"chat-app/models"
	"chat-app/services"
	"chat-app/workflows"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatHandler handles chat-related HTTP requests
type ChatHandler struct {
	db          *sql.DB
	vllmService *services.VLLMService
	dbosCtx     dbos.DBOSContext
	workflows   *workflows.ChatWorkflows
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *sql.DB, vllmService *services.VLLMService, dbosCtx dbos.DBOSContext, wf *workflows.ChatWorkflows) *ChatHandler {
	return &ChatHandler{
		db:          db,
		vllmService: vllmService,
		dbosCtx:     dbosCtx,
		workflows:   wf,
	}
}

// CreateConversation creates a new conversation using DBOS workflow
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	// Run durable workflow
	handle, err := dbos.RunWorkflow(h.dbosCtx, h.workflows.CreateConversationWorkflow, "")
	if err != nil {
		log.Printf("Failed to start CreateConversation workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	conv, err := handle.GetResult()
	if err != nil {
		log.Printf("CreateConversation workflow failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	c.JSON(http.StatusCreated, conv)
}

// ListConversations lists all conversations
func (h *ChatHandler) ListConversations(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		"SELECT id, created_at FROM conversations ORDER BY created_at DESC")
	if err != nil {
		log.Printf("Database error listing conversations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list conversations"})
		return
	}
	defer rows.Close()

	conversations := []models.Conversation{}
	for rows.Next() {
		var conv models.Conversation
		if err := rows.Scan(&conv.ID, &conv.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan conversation"})
			return
		}
		conversations = append(conversations, conv)
	}

	c.JSON(http.StatusOK, conversations)
}

// GetConversation retrieves a conversation by ID
func (h *ChatHandler) GetConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var conv models.Conversation
	err = h.db.QueryRowContext(c.Request.Context(),
		"SELECT id, created_at FROM conversations WHERE id = $1", id).
		Scan(&conv.ID, &conv.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// DeleteConversation deletes a conversation using DBOS workflow
func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Run durable workflow
	handle, err := dbos.RunWorkflow(h.dbosCtx, h.workflows.DeleteConversationWorkflow, id)
	if err != nil {
		log.Printf("Failed to start DeleteConversation workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	_, err = handle.GetResult()
	if err != nil {
		log.Printf("DeleteConversation workflow failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted"})
}

// SendMessage sends a message and gets an AI response using DBOS workflow
func (h *ChatHandler) SendMessage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify conversation exists
	var exists bool
	err = h.db.QueryRowContext(c.Request.Context(),
		"SELECT EXISTS(SELECT 1 FROM conversations WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// Run durable workflow for message processing
	input := workflows.SendMessageInput{
		ConversationID: id,
		Content:        req.Content,
	}

	handle, err := dbos.RunWorkflow(h.dbosCtx, h.workflows.SendMessageWorkflow, input)
	if err != nil {
		log.Printf("Failed to start SendMessage workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	output, err := handle.GetResult()
	if err != nil {
		log.Printf("SendMessage workflow failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get AI response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		UserMessage:      output.UserMessage,
		AssistantMessage: output.AssistantMessage,
	})
}

// GetMessages retrieves all messages for a conversation
func (h *ChatHandler) GetMessages(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		"SELECT id, conversation_id, role, content, created_at FROM messages WHERE conversation_id = $1 ORDER BY created_at ASC",
		id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}
	defer rows.Close()

	messages := []models.Message{}
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan message"})
			return
		}
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, messages)
}

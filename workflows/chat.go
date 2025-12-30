package workflows

import (
	"context"
	"database/sql"
	"time"

	"chat-app/models"
	"chat-app/services"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
	"github.com/google/uuid"
)

// ChatWorkflows contains DBOS workflows for chat operations
type ChatWorkflows struct {
	db          *sql.DB
	vllmService *services.VLLMService
}

// NewChatWorkflows creates a new ChatWorkflows instance
func NewChatWorkflows(db *sql.DB, vllmService *services.VLLMService) *ChatWorkflows {
	return &ChatWorkflows{
		db:          db,
		vllmService: vllmService,
	}
}

// SendMessageInput contains the input for the SendMessage workflow
type SendMessageInput struct {
	ConversationID uuid.UUID
	Content        string
}

// SendMessageOutput contains the output of the SendMessage workflow
type SendMessageOutput struct {
	UserMessage      models.Message
	AssistantMessage models.Message
}

// SendMessageWorkflow is a durable workflow that handles sending a message and getting AI response
// If the workflow fails at any point, it will automatically resume from the last completed step
func (w *ChatWorkflows) SendMessageWorkflow(ctx dbos.DBOSContext, input SendMessageInput) (SendMessageOutput, error) {
	var output SendMessageOutput

	// Step 1: Get existing messages for context (durable step)
	messages, err := dbos.RunAsStep(ctx, func(stepCtx context.Context) ([]models.Message, error) {
		return w.getMessages(stepCtx, input.ConversationID)
	})
	if err != nil {
		return output, err
	}

	// Step 2: Save user message to database (durable step)
	userMsg, err := dbos.RunAsStep(ctx, func(stepCtx context.Context) (models.Message, error) {
		return w.saveMessage(stepCtx, input.ConversationID, "user", input.Content)
	})
	if err != nil {
		return output, err
	}
	output.UserMessage = userMsg

	// Step 3: Get AI response from vLLM (durable step - will retry on failure)
	aiResponse, err := dbos.RunAsStep(ctx, func(stepCtx context.Context) (string, error) {
		return w.vllmService.Chat(messages, input.Content)
	})
	if err != nil {
		return output, err
	}

	// Step 4: Save assistant message to database (durable step)
	assistantMsg, err := dbos.RunAsStep(ctx, func(stepCtx context.Context) (models.Message, error) {
		return w.saveMessage(stepCtx, input.ConversationID, "assistant", aiResponse)
	})
	if err != nil {
		return output, err
	}
	output.AssistantMessage = assistantMsg

	return output, nil
}

// getMessages retrieves all messages for a conversation
func (w *ChatWorkflows) getMessages(ctx context.Context, conversationID uuid.UUID) ([]models.Message, error) {
	rows, err := w.db.QueryContext(ctx,
		"SELECT id, conversation_id, role, content, created_at FROM messages WHERE conversation_id = $1 ORDER BY created_at ASC",
		conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// saveMessage saves a message to the database
func (w *ChatWorkflows) saveMessage(ctx context.Context, conversationID uuid.UUID, role, content string) (models.Message, error) {
	id := uuid.New()
	now := time.Now()

	_, err := w.db.ExecContext(ctx,
		"INSERT INTO messages (id, conversation_id, role, content, created_at) VALUES ($1, $2, $3, $4, $5)",
		id, conversationID, role, content, now)
	if err != nil {
		return models.Message{}, err
	}

	return models.Message{
		ID:             id,
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		CreatedAt:      now,
	}, nil
}

// CreateConversationWorkflow creates a new conversation durably
func (w *ChatWorkflows) CreateConversationWorkflow(ctx dbos.DBOSContext, _ string) (models.Conversation, error) {
	return dbos.RunAsStep(ctx, func(stepCtx context.Context) (models.Conversation, error) {
		id := uuid.New()
		now := time.Now()

		_, err := w.db.ExecContext(stepCtx,
			"INSERT INTO conversations (id, created_at) VALUES ($1, $2)",
			id, now)
		if err != nil {
			return models.Conversation{}, err
		}

		return models.Conversation{
			ID:        id,
			CreatedAt: now,
		}, nil
	})
}

// DeleteConversationWorkflow deletes a conversation and its messages durably
func (w *ChatWorkflows) DeleteConversationWorkflow(ctx dbos.DBOSContext, conversationID uuid.UUID) (bool, error) {
	// Step 1: Delete messages
	_, err := dbos.RunAsStep(ctx, func(stepCtx context.Context) (bool, error) {
		_, err := w.db.ExecContext(stepCtx, "DELETE FROM messages WHERE conversation_id = $1", conversationID)
		return err == nil, err
	})
	if err != nil {
		return false, err
	}

	// Step 2: Delete conversation
	return dbos.RunAsStep(ctx, func(stepCtx context.Context) (bool, error) {
		_, err := w.db.ExecContext(stepCtx, "DELETE FROM conversations WHERE id = $1", conversationID)
		return err == nil, err
	})
}

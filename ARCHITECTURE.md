# Architecture & Component Connections

This document explains how all components in the chat-app connect and work together.

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          USER'S BROWSER                                  │
│                     http://localhost:8080                                │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────┐            │
│  │  Frontend (static/index.html)                           │            │
│  │  - JavaScript for UI interactions                       │            │
│  │  - Sends HTTP requests via Fetch API                    │            │
│  │  - Renders chat messages                                │            │
│  └─────────────────────────────────────────────────────────┘            │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
                                │ HTTP REST API
                                │ (JSON)
                                ↓
┌─────────────────────────────────────────────────────────────────────────┐
│                     GO WEB SERVER (main.go)                              │
│                       Port 8080 (Gin Framework)                          │
│                                                                           │
│  ┌──────────────────────────────────────────────────────────┐           │
│  │  1. Router (Gin)                                         │           │
│  │     Routes: /api/conversations, /api/messages, etc.      │           │
│  └────────────────────┬─────────────────────────────────────┘           │
│                       │                                                   │
│                       ↓                                                   │
│  ┌──────────────────────────────────────────────────────────┐           │
│  │  2. Handlers (handlers/chat.go)                          │           │
│  │     - CreateConversation()                               │           │
│  │     - SendMessage() ← Main message flow                  │           │
│  │     - GetMessages()                                       │           │
│  └────────────────────┬─────────────────────────────────────┘           │
│                       │                                                   │
│                       │ Calls DBOS Workflows                             │
│                       ↓                                                   │
│  ┌──────────────────────────────────────────────────────────┐           │
│  │  3. DBOS Context (dbosCtx)                               │           │
│  │     - Workflow orchestrator                              │           │
│  │     - Tracks execution state                             │           │
│  │     - Enables crash recovery                             │           │
│  └────────────────────┬─────────────────────────────────────┘           │
│                       │                                                   │
│                       │ Executes                                         │
│                       ↓                                                   │
│  ┌──────────────────────────────────────────────────────────┐           │
│  │  4. Workflows (workflows/chat.go)                        │           │
│  │     SendMessageWorkflow:                                 │           │
│  │       Step 1: Get messages (DB) ──────────┐              │           │
│  │       Step 2: Save user message (DB) ─────┤              │           │
│  │       Step 3: Call AI (vLLM) ─────────────┤              │           │
│  │       Step 4: Save AI response (DB) ──────┤              │           │
│  └────────────────────┬──────────────────────┴──────────────┘           │
│                       │                      │                           │
└───────────────────────┼──────────────────────┼───────────────────────────┘
                        │                      │
                        │                      │
        ┌───────────────┴──────┐      ┌────────┴──────────────┐
        │                      │      │                        │
        ↓                      ↓      ↓                        │
┌──────────────────┐   ┌──────────────────┐                   │
│   PostgreSQL     │   │   vLLM Service   │                   │
│   Database       │   │  (services/      │                   │
│                  │   │   vllm.go)       │                   │
│  Port: 5432      │   │                  │                   │
│                  │   └────────┬─────────┘                   │
│  Tables:         │            │                             │
│  - conversations │            │ HTTP POST                   │
│  - messages      │            │ /v1/chat/completions        │
│                  │            │                             │
│  Stores:         │            ↓                             │
│  - Chat history  │   ┌──────────────────────┐              │
│  - User messages │   │   vLLM Server        │              │
│  - AI responses  │   │   (Python)           │              │
│  - DBOS state    │   │                      │              │
│                  │   │  Port: 5000          │              │
└──────────────────┘   │                      │              │
                       │  ┌────────────────┐  │              │
                       │  │ Llama 3.1 8B   │  │              │
                       │  │ Model          │  │              │
                       │  │ (HuggingFace)  │  │              │
                       │  └────────────────┘  │              │
                       │                      │              │
                       │  Running on:         │              │
                       │  RTX 5090 GPU        │              │
                       │  (~8-12GB VRAM)      │              │
                       └──────────────────────┘              │
                                                              │
        Environment Variables (.env)  ───────────────────────┘
        - DATABASE_URL=postgresql://jagadeesh@localhost:5432/chat_app
        - VLLM_BASE_URL=http://localhost:5000
        - PORT=8080
```

## Detailed Connection Flow

### 1. User Sends a Message

**Browser → Go Server**

```javascript
// static/index.html (line ~150)
fetch(`/api/conversations/${conversationId}/messages`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content: userMessage })
})
```

**What travels:** JSON `{ "content": "Hello AI!" }`

---

### 2. Gin Router Receives Request

**main.go (line 99)**

```go
api.POST("/conversations/:id/messages", chatHandler.SendMessage)
```

**Routing:** `POST /api/conversations/abc-123/messages` → `chatHandler.SendMessage()`

---

### 3. Handler Processes Request

**handlers/chat.go (line 126-172)**

```go
func (h *ChatHandler) SendMessage(c *gin.Context) {
    // Parse conversation ID
    id := c.Param("id")

    // Parse request body
    var req models.SendMessageRequest
    c.BindJSON(&req)  // Gets { "content": "Hello AI!" }

    // Create workflow input
    input := workflows.SendMessageInput{
        ConversationID: id,
        Content: req.Content,  // "Hello AI!"
    }

    // Start DBOS workflow
    handle := dbos.RunWorkflow(h.dbosCtx, h.workflows.SendMessageWorkflow, input)

    // Wait for workflow to complete
    output := handle.GetResult()

    // Return response
    c.JSON(200, output)
}
```

**Connection:** Handler → DBOS Workflow System

---

### 4. DBOS Orchestrates Workflow

**workflows/chat.go (line 43-80)**

```go
func (w *ChatWorkflows) SendMessageWorkflow(ctx dbos.DBOSContext, input SendMessageInput) {

    // STEP 1: Get conversation history
    messages := dbos.RunAsStep(ctx, func() {
        // SQL: SELECT * FROM messages WHERE conversation_id = input.ConversationID
        return w.getMessages(ctx, input.ConversationID)
    })
    // Returns: [{role: "user", content: "Hi"}, {role: "assistant", content: "Hello"}]

    // STEP 2: Save user message to database
    userMsg := dbos.RunAsStep(ctx, func() {
        // SQL: INSERT INTO messages (role, content, ...) VALUES ('user', 'Hello AI!', ...)
        return w.saveMessage(ctx, input.ConversationID, "user", input.Content)
    })

    // STEP 3: Call vLLM for AI response
    aiResponse := dbos.RunAsStep(ctx, func() {
        // HTTP POST to http://localhost:5000/v1/chat/completions
        return w.vllmService.Chat(messages, input.Content)
    })
    // Returns: "Hello! I'm an AI assistant. How can I help you?"

    // STEP 4: Save AI response to database
    assistantMsg := dbos.RunAsStep(ctx, func() {
        // SQL: INSERT INTO messages (role, content, ...) VALUES ('assistant', 'Hello! ...', ...)
        return w.saveMessage(ctx, input.ConversationID, "assistant", aiResponse)
    })

    return SendMessageOutput{
        UserMessage: userMsg,
        AssistantMessage: assistantMsg
    }
}
```

**Connections:**
- Workflow → PostgreSQL (Steps 1, 2, 4)
- Workflow → vLLM Service (Step 3)

---

### 5. vLLM Service Calls AI Model

**services/vllm.go (line 56-92)**

```go
func (s *VLLMService) Chat(messages []models.Message, userMessage string) (string, error) {

    // Convert to vLLM format
    vllmMessages := []VLLMMessage{
        {Role: "user", Content: "Hi"},
        {Role: "assistant", Content: "Hello"},
        {Role: "user", Content: "Hello AI!"},  // New message
    }

    // Prepare request
    reqBody := VLLMRequest{
        Model: "meta-llama/Meta-Llama-3.1-8B-Instruct",
        Messages: vllmMessages,
        MaxTokens: 4096,
        Temperature: 0.7,
    }

    // Call vLLM API
    resp := http.Post("http://localhost:5000/v1/chat/completions", reqBody)

    // Parse response
    return resp.Choices[0].Message.Content
    // Returns: "Hello! I'm an AI assistant. How can I help you?"
}
```

**Connection:** Go Service → vLLM HTTP Server (port 5000)

---

### 6. vLLM Server Runs Inference

**vLLM Server (Python process on port 5000)**

```
Receives: POST /v1/chat/completions
{
  "model": "meta-llama/Meta-Llama-3.1-8B-Instruct",
  "messages": [
    {"role": "user", "content": "Hi"},
    {"role": "assistant", "content": "Hello"},
    {"role": "user", "content": "Hello AI!"}
  ],
  "max_tokens": 4096,
  "temperature": 0.7
}

↓

Loads conversation into GPU memory
Runs Llama 3.1 8B model on RTX 5090
Generates tokens one by one
Returns complete response

↓

Response: {
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "Hello! I'm an AI assistant. How can I help you?"
    }
  }]
}
```

**Connection:** vLLM Server → RTX 5090 GPU → Llama 3.1 8B Model

---

### 7. PostgreSQL Stores Everything

**Database Tables**

```sql
-- conversations table
┌──────────┬─────────────────────┐
│ id       │ created_at          │
├──────────┼─────────────────────┤
│ abc-123  │ 2025-01-15 10:00:00 │
└──────────┴─────────────────────┘

-- messages table
┌──────────┬─────────────┬───────────┬──────────────────────┬─────────────────────┐
│ id       │ conv_id     │ role      │ content              │ created_at          │
├──────────┼─────────────┼───────────┼──────────────────────┼─────────────────────┤
│ msg-1    │ abc-123     │ user      │ Hi                   │ 2025-01-15 10:00:05 │
│ msg-2    │ abc-123     │ assistant │ Hello                │ 2025-01-15 10:00:06 │
│ msg-3    │ abc-123     │ user      │ Hello AI!            │ 2025-01-15 10:01:00 │
│ msg-4    │ abc-123     │ assistant │ Hello! I'm an AI...  │ 2025-01-15 10:01:02 │
└──────────┴─────────────┴───────────┴──────────────────────┴─────────────────────┘
```

**Connection:** Go Application → PostgreSQL (via DATABASE_URL)

---

### 8. Response Returns to User

**Go Server → Browser**

```go
// handlers/chat.go returns
c.JSON(200, ChatResponse{
    UserMessage: {
        ID: "msg-3",
        Role: "user",
        Content: "Hello AI!",
        CreatedAt: "2025-01-15T10:01:00Z"
    },
    AssistantMessage: {
        ID: "msg-4",
        Role: "assistant",
        Content: "Hello! I'm an AI assistant. How can I help you?",
        CreatedAt: "2025-01-15T10:01:02Z"
    }
})
```

**Browser receives and displays**

```javascript
// static/index.html
response.json().then(data => {
    displayMessage(data.user_message);      // Shows "Hello AI!"
    displayMessage(data.assistant_message);  // Shows "Hello! I'm an AI..."
});
```

---

## Environment Variables Connection

**How .env connects everything:**

```bash
# .env file
DATABASE_URL=postgresql://jagadeesh@localhost:5432/chat_app
VLLM_BASE_URL=http://localhost:5000
PORT=8080
```

**Usage in main.go:**

```go
// Line 22: Read database URL
dbURL := os.Getenv("DATABASE_URL")
// → "postgresql://jagadeesh@localhost:5432/chat_app"

// Line 28: Connect to PostgreSQL
db := sql.Open("postgres", dbURL)
// → Connects to PostgreSQL on port 5432

// Line 41: Read vLLM URL
vllmURL := os.Getenv("VLLM_BASE_URL")
// → "http://localhost:5000"

// Line 46: Create vLLM service
vllmService := services.NewVLLMService(vllmURL)
// → Will make HTTP calls to localhost:5000

// Line 115: Read server port
port := os.Getenv("PORT")
// → "8080"

// Line 120: Start web server
router.Run(":" + port)
// → Listens on http://localhost:8080
```

---

## Data Flow Summary

```
User types message in browser
    ↓
Browser sends HTTP POST to localhost:8080
    ↓
Gin router routes to SendMessage handler
    ↓
Handler starts DBOS workflow
    ↓
Workflow Step 1: Query PostgreSQL for conversation history
    ↓
Workflow Step 2: Save user message to PostgreSQL
    ↓
Workflow Step 3: Call vLLM service at localhost:5000
    ↓
vLLM service sends HTTP request to vLLM server
    ↓
vLLM server runs Llama 3.1 8B on RTX 5090 GPU
    ↓
GPU generates AI response
    ↓
vLLM server returns JSON response
    ↓
vLLM service extracts text from response
    ↓
Workflow Step 4: Save AI response to PostgreSQL
    ↓
Workflow returns both messages
    ↓
Handler sends JSON response to browser
    ↓
Browser displays both messages in chat UI
```

---

## Port Summary

| Component      | Port  | Protocol | Purpose                    |
|----------------|-------|----------|----------------------------|
| Go Web Server  | 8080  | HTTP     | Serves frontend & API      |
| vLLM Server    | 5000  | HTTP     | AI inference API           |
| PostgreSQL     | 5432  | TCP      | Database connections       |

---

## Key Dependencies

### Go Application Dependencies:
- `github.com/gin-gonic/gin` - Web framework
- `github.com/dbos-inc/dbos-transact-golang` - Workflow orchestration
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/google/uuid` - UUID generation

### External Services:
- **PostgreSQL** - Started with `systemctl start postgresql`
- **vLLM Server** - Started with `./start-vllm-rtx5090.sh`

### Configuration Files:
- `.env` - Environment variables (connects everything)
- `dbos.yaml` - DBOS configuration
- `migrations/001_init.sql` - Database schema

---

## What Happens on Startup

**Terminal 1: Start vLLM**
```bash
./start-vllm-rtx5090.sh
# → Loads Llama 3.1 8B into GPU memory (~8GB)
# → Starts HTTP server on port 5000
# → Ready to accept inference requests
```

**Terminal 2: Start Chat App**
```bash
go run main.go
# → Reads .env file
# → Connects to PostgreSQL (DATABASE_URL)
# → Creates vLLM service (VLLM_BASE_URL)
# → Initializes DBOS workflow system
# → Registers workflows
# → Starts Gin web server (PORT=8080)
# → Serves static files
# → Ready to accept HTTP requests
```

**Browser:**
```
http://localhost:8080
# → Downloads index.html from Go server
# → Loads JavaScript
# → Makes API call to /api/conversations
# → Displays conversation list
# → Ready for user input
```

---

## Error Handling & Recovery

**If vLLM crashes during inference:**
1. DBOS marks Step 3 as failed
2. User message is already saved (Step 2 completed)
3. When app restarts, DBOS automatically:
   - Skips Steps 1 & 2 (already done)
   - Retries Step 3 (call vLLM again)
   - Completes Step 4 (save response)

**If database connection fails:**
- Go app won't start (fails at line 35: db.Ping())
- User sees clear error message

**If Go app crashes:**
- DBOS recovery system activates on restart
- Incomplete workflows resume from last successful step
- No data loss

---

This architecture provides:
- ✅ **Reliability** (DBOS workflows)
- ✅ **Performance** (Local GPU inference)
- ✅ **Privacy** (All data stays local)
- ✅ **Scalability** (Can handle 512+ concurrent requests with RTX 5090)

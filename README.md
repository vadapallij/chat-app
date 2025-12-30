# Chat Application with vLLM & Llama 3.1 8B

A durable chat application powered by vLLM and Meta's Llama 3.1 8B model, built with Go, DBOS, and PostgreSQL. Optimized for NVIDIA RTX 5090.

## Features

- Real-time chat interface with AI responses
- Durable workflow execution using DBOS
- Conversation management (create, list, delete)
- Message persistence in PostgreSQL
- Local AI inference using vLLM and Llama 3.1 8B
- Clean, ChatGPT-like UI
- GPU-accelerated inference (150-250+ tokens/second on RTX 5090)

## Architecture

### Backend
- **Language**: Go 1.23
- **Framework**: Gin (HTTP server)
- **Durability**: DBOS Transact (durable workflows)
- **Database**: PostgreSQL
- **AI Engine**: vLLM with Llama 3.1 8B Instruct

### Frontend
- **Technology**: Vanilla JavaScript
- **Styling**: Custom CSS with dark theme
- **Communication**: REST API

## Prerequisites

- Go 1.23 or higher
- PostgreSQL 12+
- Python 3.8+ (for vLLM)
- NVIDIA GPU (RTX 5090 recommended, 32GB VRAM)
- CUDA 12.x and NVIDIA drivers
- HuggingFace account with Llama 3.1 access

## Quick Start

### Automated Setup (Recommended)

Run the automated setup script:

```bash
cd /home/jagadeesh/Desktop/chat-app
./setup.sh
```

This will guide you through installing all dependencies and configuring the application.

### Manual Setup

For detailed manual setup instructions, see:
- **[QUICK_START.md](QUICK_START.md)** - Step-by-step setup guide
- **[POSTGRESQL_SETUP.md](POSTGRESQL_SETUP.md)** - PostgreSQL installation and configuration
- **[VLLM_SETUP.md](VLLM_SETUP.md)** - vLLM setup optimized for RTX 5090

## Quick Start (Manual)

### 1. Install vLLM

See [VLLM_SETUP.md](VLLM_SETUP.md) for detailed instructions.

Quick install:
```bash
pipx install vllm
pipx run huggingface-cli login
```

### 2. Set Up Database

```bash
# Create PostgreSQL database
createdb chat_app

# Run migrations
psql chat_app < migrations/001_init.sql
```

### 3. Configure Environment

```bash
cp .env.example .env
# Edit .env with your database URL and vLLM URL
```

### 4. Start vLLM Server (Terminal 1)

Use the optimized script for RTX 5090:

```bash
./start-vllm-rtx5090.sh
```

Or manually:

```bash
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --host 0.0.0.0 \
  --port 5000 \
  --gpu-memory-utilization 0.95 \
  --max-model-len 32768 \
  --dtype bfloat16 \
  --max-num-seqs 512 \
  --enable-prefix-caching
```

Keep this terminal running.

### 5. Start Chat Application (Terminal 2)

```bash
# In a new terminal
go run main.go
```

### 6. Access the Application

Open your browser and navigate to:
```
http://localhost:8080
```

## Documentation

- **[QUICK_START.md](QUICK_START.md)** - Complete setup walkthrough
- **[VLLM_SETUP.md](VLLM_SETUP.md)** - vLLM configuration for RTX 5090
- **[POSTGRESQL_SETUP.md](POSTGRESQL_SETUP.md)** - Database setup and troubleshooting

## API Endpoints

### Conversations
- `POST /api/conversations` - Create new conversation
- `GET /api/conversations` - List all conversations
- `GET /api/conversations/:id` - Get conversation details
- `DELETE /api/conversations/:id` - Delete conversation

### Messages
- `POST /api/conversations/:id/messages` - Send message and get AI response
- `GET /api/conversations/:id/messages` - Get conversation history

### Health Check
- `GET /health` - Server health status

## Environment Variables

```bash
DATABASE_URL=postgresql://user:password@localhost:5432/chat_app
VLLM_BASE_URL=http://localhost:5000
PORT=8080
```

## Project Structure

```
chat-app/
├── main.go              # Application entry point
├── handlers/
│   └── chat.go          # HTTP request handlers
├── services/
│   ├── anthropic.go     # (Legacy) Anthropic service
│   └── vllm.go          # vLLM service for Llama 3.1
├── workflows/
│   └── chat.go          # DBOS durable workflows
├── models/
│   └── models.go        # Data structures
├── migrations/
│   └── 001_init.sql     # Database schema
├── static/
│   └── index.html       # Frontend application
├── .env.example         # Environment variables template
├── VLLM_SETUP.md        # vLLM installation guide
└── README.md            # This file
```

## How It Works

1. **Durable Workflows**: Uses DBOS to ensure message processing is resilient to failures
2. **Message Flow**:
   - User sends message via frontend
   - Backend saves message to PostgreSQL
   - Retrieves conversation history
   - Sends to vLLM for AI response
   - Saves AI response to database
   - Returns both messages to frontend

3. **Workflow Recovery**: If any step fails, DBOS automatically resumes from the last successful step

## Customization

### Change AI Model

Edit [services/vllm.go](services/vllm.go:54) to change the model:

```go
Model: "meta-llama/Meta-Llama-3.1-70B-Instruct",
```

### Adjust Response Parameters

Modify temperature and max tokens in [services/vllm.go](services/vllm.go:55-56):

```go
MaxTokens:   2048,
Temperature: 0.5,
```

### Customize UI

Edit [static/index.html](static/index.html) to modify the chat interface.

## Troubleshooting

### vLLM Connection Error

Check if vLLM is running:
```bash
curl http://localhost:5000/v1/models
```

### Database Connection Error

Verify PostgreSQL is running and credentials are correct:
```bash
psql $DATABASE_URL -c "SELECT 1;"
```

### Out of Memory

Reduce max model length in vLLM:
```bash
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --max-model-len 4096
```

## Migration from Anthropic

This application was originally built with Anthropic's Claude API. The integration has been updated to use vLLM with Llama 3.1 8B for local inference. The old Anthropic service is still available in [services/anthropic.go](services/anthropic.go) for reference.

## License

MIT

## Contributing

Contributions welcome! Please open an issue or submit a pull request.

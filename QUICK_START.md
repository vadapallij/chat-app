# Quick Start Guide - Chat App with vLLM on RTX 5090

Get your AI chat app running in just a few steps!

## Prerequisites Check

```bash
# Verify GPU
nvidia-smi

# Verify PostgreSQL
psql --version
# If not installed, see POSTGRESQL_SETUP.md

# Verify vLLM installation
vllm --version
# If not installed, continue to step 1 below
```

## Step-by-Step Setup

### 1. Install PostgreSQL (if not already installed)

**If PostgreSQL is not installed**, follow the detailed guide:
- See **[POSTGRESQL_SETUP.md](POSTGRESQL_SETUP.md)** for complete installation instructions

**Quick install for Ubuntu/Debian:**
```bash
sudo apt update && sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

### 2. Install vLLM (if not already installed)

**See [VLLM_SETUP.md](VLLM_SETUP.md) for detailed instructions**, or quick install:

```bash
# Remove old installation
pipx uninstall vllm

# Install fresh
pipx install vllm

# Login to HuggingFace
pipx run huggingface-cli login
```

Get your HuggingFace token from: https://huggingface.co/settings/tokens

### 3. Setup Database

```bash
# Create database
createdb chat_app

# Run migrations
psql chat_app < migrations/001_init.sql

# Verify tables were created
psql chat_app -c "\dt"
```

**Troubleshooting:** If you get permission or connection errors, see [POSTGRESQL_SETUP.md](POSTGRESQL_SETUP.md)

### 4. Configure Environment

```bash
# Copy example config
cp .env.example .env

# Edit .env file with your settings
nano .env
```

Set these values in `.env`:
```
DATABASE_URL=postgresql://YOUR_USERNAME@localhost:5432/chat_app
VLLM_BASE_URL=http://localhost:5000
PORT=8080
```

### 5. Start vLLM Server (Terminal 1)

Use the optimized startup script:

```bash
./start-vllm-rtx5090.sh
```

**OR** manually:

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

Wait for the message: "Uvicorn running on http://0.0.0.0:5000"

### 6. Start Chat Application (Terminal 2)

```bash
go run main.go
```

Wait for: "Starting server on port 8080"

### 7. Open Your Browser

Navigate to: **http://localhost:8080**

## Usage

1. Click "New Conversation" to start chatting
2. Type your message and press Enter or click Send
3. Watch the AI respond using your local Llama 3.1 8B model!

## Monitoring

Keep an eye on GPU usage in a third terminal:

```bash
watch -n 1 nvidia-smi
```

## Troubleshooting

### vLLM won't start

```bash
# Check if port 5000 is already in use
lsof -i :5000

# If something is using it, kill it or use a different port
```

### Chat app can't connect to vLLM

```bash
# Test vLLM is running
curl http://localhost:5000/v1/models

# Should return JSON with model info
```

### Database connection error

```bash
# Check PostgreSQL is running
pg_isready

# Test connection
psql chat_app -c "SELECT 1;"
```

## Performance Expectations

With your RTX 5090:
- **Response latency**: < 100ms for first token
- **Generation speed**: 150-250+ tokens/second
- **Context capacity**: Up to 32K tokens
- **VRAM usage**: ~8-12GB for the 8B model

## Stop the Application

1. In Terminal 2 (chat app): Press `Ctrl+C`
2. In Terminal 1 (vLLM): Press `Ctrl+C`

## Configuration Files

- **[VLLM_SETUP.md](VLLM_SETUP.md)**: Detailed vLLM setup guide
- **[README.md](README.md)**: Full project documentation
- **[.env.example](.env.example)**: Environment variables template

## Common Commands

```bash
# Check vLLM status
curl http://localhost:5000/health

# Check chat app status
curl http://localhost:8080/health

# List conversations (after creating some)
curl http://localhost:8080/api/conversations

# Monitor GPU
nvidia-smi dmon
```

## Next Steps

- Customize the UI in [static/index.html](static/index.html)
- Adjust model parameters in [services/vllm.go](services/vllm.go)
- Try different Llama models (70B, etc.) if you want more capabilities

Enjoy your local AI-powered chat application! 🚀

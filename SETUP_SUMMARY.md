# Setup Summary - Chat-App with vLLM on RTX 5090

## What Was Done

Your chat application has been successfully configured to use vLLM with Llama 3.1 8B instead of the Anthropic API. Here's what changed:

### Code Changes

1. **New vLLM Service** ([services/vllm.go](services/vllm.go))
   - Created Go service to communicate with vLLM's OpenAI-compatible API
   - Supports full conversation history
   - Configured for Llama 3.1 8B Instruct model

2. **Updated Application Files**
   - [main.go](main.go) - Replaced Anthropic service with vLLM service
   - [workflows/chat.go](workflows/chat.go) - Updated workflow to use vLLM
   - [handlers/chat.go](handlers/chat.go) - Updated handler references

3. **Configuration Files**
   - [.env.example](.env.example) - Environment template
   - `.env` - Your local configuration (create this)

### Documentation Created

1. **[README.md](README.md)** - Main project documentation
2. **[QUICK_START.md](QUICK_START.md)** - Step-by-step setup guide
3. **[VLLM_SETUP.md](VLLM_SETUP.md)** - vLLM installation optimized for RTX 5090
4. **[POSTGRESQL_SETUP.md](POSTGRESQL_SETUP.md)** - PostgreSQL installation and configuration
5. **[setup.sh](setup.sh)** - Automated setup script
6. **[start-vllm-rtx5090.sh](start-vllm-rtx5090.sh)** - Optimized vLLM startup script

## Next Steps to Get Running

### Option 1: Automated Setup (Easiest)

```bash
./setup.sh
```

This script will:
- Check all prerequisites
- Install PostgreSQL if needed
- Install vLLM if needed
- Create database and run migrations
- Generate `.env` file
- Guide you through HuggingFace login

### Option 2: Manual Setup

Follow these steps in order:

#### 1. Install PostgreSQL (if not installed)

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql

# Verify
psql --version
```

See [POSTGRESQL_SETUP.md](POSTGRESQL_SETUP.md) for detailed instructions.

#### 2. Install vLLM (if not installed)

```bash
# Uninstall broken installation
pipx uninstall vllm

# Fresh install
pipx install vllm

# Verify
vllm --version
```

#### 3. Setup Database

```bash
# Create database
createdb chat_app

# Run migrations
psql chat_app < migrations/001_init.sql

# Verify
psql chat_app -c "\dt"
```

#### 4. Configure Environment

```bash
# Create .env file
cp .env.example .env

# Edit with your username
nano .env
```

Set:
```
DATABASE_URL=postgresql://jagadeesh@localhost:5432/chat_app
VLLM_BASE_URL=http://localhost:5000
PORT=8080
```

#### 5. Login to HuggingFace

```bash
pipx run huggingface-cli login
```

Get token from: https://huggingface.co/settings/tokens

## Running the Application

### Terminal 1: Start vLLM Server

```bash
./start-vllm-rtx5090.sh
```

**Wait for:** "Uvicorn running on http://0.0.0.0:5000"

**This uses your RTX 5090 with:**
- 95% GPU memory utilization (30+ GB)
- 32K token context window
- bfloat16 precision
- Prefix caching enabled
- Expected: 150-250+ tokens/second

### Terminal 2: Start Chat Application

```bash
go run main.go
```

**Wait for:** "Starting server on port 8080"

### Terminal 3: Monitor GPU (Optional)

```bash
watch -n 1 nvidia-smi
```

### Browser: Access Application

Open: **http://localhost:8080**

## What to Expect

### Performance (RTX 5090)

- **First token latency**: ~50-100ms
- **Generation speed**: 150-250+ tokens/second
- **Context capacity**: Up to 32,768 tokens
- **VRAM usage**: ~8-12GB for the model
- **Concurrent users**: Can handle 512+ simultaneous sequences

### First Run

1. vLLM will download Llama 3.1 8B (~16GB) on first run
2. This may take 10-30 minutes depending on internet speed
3. Model is cached in `~/.cache/huggingface/hub/`
4. Subsequent runs are instant

### Testing

1. Click "New Conversation"
2. Type a message: "Hello! Tell me about yourself."
3. Watch the AI respond using your local model
4. Response should appear in 1-2 seconds

## Troubleshooting Quick Reference

### vLLM won't start

```bash
# Check if port is in use
lsof -i :5000

# Check HuggingFace login
pipx run huggingface-cli whoami

# Check GPU
nvidia-smi
```

### Database connection error

```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Test connection
psql chat_app -c "SELECT 1;"

# Check .env file
cat .env
```

### Chat app can't connect to vLLM

```bash
# Test vLLM endpoint
curl http://localhost:5000/v1/models

# Should return JSON with model info
```

### Out of GPU memory

Unlikely with RTX 5090 and 8B model, but if it happens:

```bash
# Use balanced settings
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --host 0.0.0.0 \
  --port 5000 \
  --gpu-memory-utilization 0.8 \
  --max-model-len 16384
```

## File Structure

```
chat-app/
├── services/
│   ├── vllm.go              ← New: vLLM integration
│   └── anthropic.go         ← Old: kept for reference
├── main.go                  ← Updated: uses vLLM
├── handlers/chat.go         ← Updated: uses vLLM
├── workflows/chat.go        ← Updated: uses vLLM
├── .env.example             ← New: environment template
├── .env                     ← Create: your config
├── setup.sh                 ← New: automated setup
├── start-vllm-rtx5090.sh   ← New: optimized vLLM launcher
├── README.md                ← Updated: complete docs
├── QUICK_START.md           ← New: setup guide
├── VLLM_SETUP.md            ← New: vLLM guide
├── POSTGRESQL_SETUP.md      ← New: database guide
└── SETUP_SUMMARY.md         ← This file
```

## Useful Commands

```bash
# Check vLLM status
curl http://localhost:5000/health

# Check app status
curl http://localhost:8080/health

# List conversations
curl http://localhost:8080/api/conversations

# View database
psql chat_app -c "SELECT * FROM conversations;"

# Monitor GPU
nvidia-smi dmon

# View vLLM logs
# (Check Terminal 1 where vLLM is running)

# Restart everything
# Ctrl+C in both terminals, then restart
```

## Configuration Options

### vLLM Performance Tuning

Edit [start-vllm-rtx5090.sh](start-vllm-rtx5090.sh) to adjust:

- `--gpu-memory-utilization`: 0.7-0.95 (default: 0.95)
- `--max-model-len`: 4096-32768 (default: 32768)
- `--max-num-seqs`: 64-1024 (default: 512)

### Application Settings

Edit [services/vllm.go](services/vllm.go:55-56) to adjust:

- `MaxTokens`: Response length (default: 4096)
- `Temperature`: Creativity 0.0-2.0 (default: 0.7)

## Additional Features to Try

### Switch to a Larger Model (if you have VRAM)

Edit [services/vllm.go](services/vllm.go:54):

```go
Model: "meta-llama/Meta-Llama-3.1-70B-Instruct",
```

Your RTX 5090 can handle 70B with quantization!

### Add System Prompts

Modify the chat service to add system prompts for specific behaviors.

### Streaming Responses

vLLM supports streaming for real-time token generation.

## Support & Resources

- **Project Documentation**: See README.md and linked guides
- **vLLM Documentation**: https://docs.vllm.ai/
- **Llama 3.1 Model Card**: https://huggingface.co/meta-llama/Meta-Llama-3.1-8B-Instruct
- **PostgreSQL Docs**: https://www.postgresql.org/docs/
- **DBOS Documentation**: https://docs.dbos.dev/

## Success Checklist

- [ ] PostgreSQL installed and running
- [ ] vLLM installed via pipx
- [ ] HuggingFace account with Llama 3.1 access
- [ ] Database `chat_app` created
- [ ] Migrations run successfully
- [ ] `.env` file configured
- [ ] vLLM server running on port 5000
- [ ] Chat app running on port 8080
- [ ] Browser can access http://localhost:8080
- [ ] First conversation works

Once all checked, you're ready to go! 🚀

## What's Different from Anthropic

| Feature | Anthropic (Before) | vLLM (Now) |
|---------|-------------------|------------|
| API | Cloud (api.anthropic.com) | Local (localhost:5000) |
| Cost | Pay per token | Free (after setup) |
| Privacy | Data sent to Anthropic | 100% local |
| Speed | Network dependent | GPU limited (~200 tok/s) |
| Model | Claude Sonnet 4 | Llama 3.1 8B |
| Context | 200K tokens | 32K tokens (configurable) |
| Setup | API key only | Full local setup |

Both work great, but vLLM gives you complete control and privacy!

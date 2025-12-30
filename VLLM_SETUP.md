# vLLM Setup Guide for Chat-App with NVIDIA RTX 5090

This guide will help you set up vLLM with Llama 3.1 8B model optimized for your NVIDIA RTX 5090 GPU.

## Prerequisites

- Python 3.8 or higher
- NVIDIA RTX 5090 GPU (32GB VRAM) - You have this! 🚀
- CUDA 12.x and compatible NVIDIA drivers
- HuggingFace account with access to Llama 3.1 models

## GPU Setup

### Verify NVIDIA Drivers and CUDA

```bash
# Check GPU
nvidia-smi

# Check CUDA version
nvcc --version
```

You should see your RTX 5090 listed with ~32GB VRAM available.

## Installation Steps

### 1. Fix Existing pipx Installation (if needed)

Since you have vLLM with missing metadata, first uninstall it:

```bash
pipx uninstall vllm
```

### 2. Install vLLM using pipx

```bash
pipx install vllm
```

**OR** if you prefer using pip in a virtual environment:

```bash
# Create a virtual environment
python3 -m venv vllm-env
source vllm-env/bin/activate

# Install vLLM
pip install vllm
```

### 3. Login to HuggingFace

You need access to Meta-Llama models:

```bash
# If using pipx
pipx run huggingface-cli login

# OR if using venv
huggingface-cli login
```

Follow the prompts and enter your HuggingFace token. You can get a token from https://huggingface.co/settings/tokens

### 4. Request Access to Llama 3.1

Visit https://huggingface.co/meta-llama/Meta-Llama-3.1-8B-Instruct and click "Request Access" if you haven't already.

## Running vLLM Server - OPTIMIZED FOR RTX 5090

With your RTX 5090's 32GB VRAM, you can run Llama 3.1 8B with maximum performance settings!

### Recommended Configuration for RTX 5090

This configuration maximizes throughput and allows for longer context:

```bash
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --host 0.0.0.0 \
  --port 5000 \
  --gpu-memory-utilization 0.95 \
  --max-model-len 32768 \
  --dtype bfloat16 \
  --max-num-seqs 512 \
  --enable-prefix-caching \
  --disable-log-requests
```

**What these flags do:**
- `--gpu-memory-utilization 0.95`: Use 95% of your 32GB VRAM (you have plenty!)
- `--max-model-len 32768`: Support up to 32K tokens context (Llama 3.1's full capacity)
- `--dtype bfloat16`: Use bfloat16 precision (optimal for RTX 5090)
- `--max-num-seqs 512`: Handle up to 512 concurrent sequences
- `--enable-prefix-caching`: Cache common prompts for faster responses
- `--disable-log-requests`: Reduce console clutter

### Alternative: Balanced Configuration

If you want to save some VRAM for other tasks:

```bash
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --host 0.0.0.0 \
  --port 5000 \
  --gpu-memory-utilization 0.8 \
  --max-model-len 16384 \
  --dtype bfloat16 \
  --max-num-seqs 256 \
  --enable-prefix-caching
```

### Quick Start (Basic)

For testing or if you prefer simpler settings:

```bash
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --host 0.0.0.0 \
  --port 5000 \
  --max-model-len 8192 \
  --dtype auto
```

## Configuration

### Environment Variables

Create a `.env` file in the chat-app directory:

```bash
cp .env.example .env
```

Edit `.env` and set:

```
DATABASE_URL=postgresql://user:password@localhost:5432/chat_app
VLLM_BASE_URL=http://localhost:5000
PORT=8080
```

## Starting the Chat Application

Once vLLM server is running:

1. In one terminal, keep vLLM server running
2. In another terminal, start your chat-app:

```bash
cd /home/jagadeesh/Desktop/chat-app
go run main.go
```

3. Open your browser and navigate to http://localhost:8080

## Troubleshooting

### Monitor GPU Usage

Keep an eye on GPU utilization while vLLM is running:

```bash
# In another terminal
watch -n 1 nvidia-smi
```

You should see vLLM using your RTX 5090 with high GPU utilization during inference.

### CUDA Out of Memory

If you somehow run out of VRAM (unlikely with 32GB for 8B model), reduce settings:

```bash
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --host 0.0.0.0 \
  --port 5000 \
  --max-model-len 8192 \
  --gpu-memory-utilization 0.7 \
  --dtype float16
```

### Slow First Request

The first request after starting vLLM may be slow due to model loading and compilation. This is normal. Subsequent requests will be much faster.

### Check vLLM Server Status

Test if vLLM is running:

```bash
curl http://localhost:5000/v1/models
```

You should see a JSON response with the model information.

### Check Chat-App Connection

Test the health endpoint:

```bash
curl http://localhost:8080/health
```

## Performance Tips for RTX 5090

### Expected Performance

With RTX 5090 and Llama 3.1 8B, you should expect:
- **First token latency**: ~50-100ms
- **Generation speed**: 150-250+ tokens/second
- **Context processing**: Extremely fast even with 32K tokens

### Advanced Optimizations

#### 1. Enable Flash Attention (if not auto-enabled)

```bash
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --host 0.0.0.0 \
  --port 5000 \
  --max-model-len 32768 \
  --dtype bfloat16 \
  --gpu-memory-utilization 0.95 \
  --enable-prefix-caching \
  --enable-chunked-prefill \
  --max-num-batched-tokens 8192
```

#### 2. For Maximum Throughput (Multiple Users)

```bash
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --host 0.0.0.0 \
  --port 5000 \
  --max-model-len 16384 \
  --dtype bfloat16 \
  --gpu-memory-utilization 0.95 \
  --max-num-seqs 1024 \
  --enable-prefix-caching
```

#### 3. For Lowest Latency (Single User Focus)

```bash
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --host 0.0.0.0 \
  --port 5000 \
  --max-model-len 32768 \
  --dtype bfloat16 \
  --gpu-memory-utilization 0.95 \
  --max-num-seqs 64 \
  --enable-prefix-caching
```

### Benchmark Your Setup

Test the inference speed:

```bash
# After vLLM server is running
curl http://localhost:5000/v1/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "meta-llama/Meta-Llama-3.1-8B-Instruct",
    "prompt": "Explain quantum computing in simple terms:",
    "max_tokens": 200
  }'
```

## Next Steps

After setting up vLLM:

1. Ensure PostgreSQL is running and the database is created
2. Run database migrations if needed
3. Start vLLM server (keep it running)
4. Start the chat-app Go server
5. Access the chat interface at http://localhost:8080

Enjoy your local AI chat application powered by Llama 3.1 8B!

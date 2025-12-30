#!/bin/bash

# Optimized vLLM startup script for NVIDIA RTX 5090
# This configuration maximizes the 32GB VRAM for best performance

echo "================================================"
echo "Starting vLLM with Llama 3.1 8B Instruct"
echo "Optimized for NVIDIA RTX 5090 (32GB VRAM)"
echo "================================================"
echo ""
echo "Server Configuration:"
echo "  - Port: 5000"
echo "  - Max Context Length: 32K tokens"
echo "  - GPU Memory: 95% utilization"
echo "  - Precision: bfloat16"
echo "  - Prefix Caching: Enabled"
echo ""
echo "Note: First request may be slow due to model loading"
echo "Expected performance: 150-250+ tokens/second"
echo ""
echo "Press Ctrl+C to stop the server"
echo "================================================"
echo ""

# Check if GPU is available
if ! command -v nvidia-smi &> /dev/null; then
    echo "ERROR: nvidia-smi not found. Please install NVIDIA drivers."
    exit 1
fi

# Display GPU info
echo "GPU Information:"
nvidia-smi --query-gpu=name,memory.total,driver_version --format=csv,noheader
echo ""

# Start vLLM server with optimized settings for RTX 5090
vllm serve meta-llama/Meta-Llama-3.1-8B-Instruct \
  --host 0.0.0.0 \
  --port 5000 \
  --gpu-memory-utilization 0.95 \
  --max-model-len 32768 \
  --dtype bfloat16 \
  --max-num-seqs 512 \
  --enable-prefix-caching \
  --enable-chunked-prefill \
  --max-num-batched-tokens 8192 \
  --disable-log-requests

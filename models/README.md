# Local AI Models Setup Guide

This guide explains how to set up local AI models for use with the AI Job application.

## Models Currently Supported

- **Qwen3-7B**: A powerful Chinese-English bilingual model from Alibaba.
- **DeepSeek Coder v3-7B**: A specialized coding model with strong programming abilities.

## Directory Structure

Models should be placed in the following directories:

```
models/
├── qwen/
│   └── Qwen3-7B/
│       └── (model files here)
└── deepseek/
    └── deepseek-coder-v3-7b/
        └── (model files here)
```

## Downloading Models

### Qwen3-7B

You can download Qwen3-7B from Hugging Face:

```bash
# Make sure you're in the project root directory
cd models/qwen
git lfs install
git clone https://huggingface.co/Qwen/Qwen3-7B-Instruct

# Or download using Hugging Face CLI
huggingface-cli download Qwen/Qwen3-7B-Instruct --local-dir Qwen3-7B
```

### DeepSeek Coder v3-7B

You can download DeepSeek Coder v3-7B from Hugging Face:

```bash
# Make sure you're in the project root directory
cd models/deepseek
git lfs install
git clone https://huggingface.co/deepseek-ai/deepseek-coder-v3-7b

# Or download using Hugging Face CLI
huggingface-cli download deepseek-ai/deepseek-coder-v3-7b --local-dir deepseek-coder-v3-7b
```

## Model Configuration

The models are configured in the `docker-compose.yml` file to be loaded automatically when the LLM server starts. 

## Hardware Requirements

- **GPU**: NVIDIA GPU with at least 12GB VRAM is recommended for both models.
- **RAM**: At least 16GB of system RAM.
- **Storage**: At least 30GB of free disk space for both models.

## Troubleshooting

- If you encounter CUDA out-of-memory errors, try setting `PYTORCH_CUDA_ALLOC_CONF=max_split_size_mb=128` as an environment variable.
- For quantized versions of these models, use the 4-bit or 8-bit variants if available to reduce memory usage.

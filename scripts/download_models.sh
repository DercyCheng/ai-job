#!/bin/bash
# Script to download Qwen3-7B and DeepSeek Coder v3-7B models

# Set up directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODELS_DIR="${SCRIPT_DIR}/../models"
QWEN_DIR="${MODELS_DIR}/qwen/Qwen3-7B"
DEEPSEEK_DIR="${MODELS_DIR}/deepseek/deepseek-coder-v3-7b"

# Check if git-lfs is installed
if ! command -v git-lfs &> /dev/null; then
    echo "git-lfs is not installed. Please install it first:"
    echo "  - macOS: brew install git-lfs"
    echo "  - Ubuntu/Debian: apt-get install git-lfs"
    echo "  - Other: https://git-lfs.github.com/"
    exit 1
fi

# Check if huggingface_hub is installed
if ! python3 -c "import huggingface_hub" &> /dev/null; then
    echo "huggingface_hub is not installed. Installing..."
    pip install huggingface_hub
fi

# Function to download model
download_model() {
    local model_name=$1
    local repo_id=$2
    local target_dir=$3
    
    echo "====================================="
    echo "Downloading ${model_name} to ${target_dir}"
    echo "====================================="
    
    if [ -d "${target_dir}" ] && [ "$(ls -A "${target_dir}")" ]; then
        echo "${model_name} directory exists and is not empty."
        read -p "Do you want to overwrite it? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Skipping ${model_name} download."
            return
        fi
    fi
    
    mkdir -p "${target_dir}"
    
    echo "Using huggingface_hub to download ${model_name}..."
    python3 -c "from huggingface_hub import snapshot_download; snapshot_download(repo_id='${repo_id}', local_dir='${target_dir}', local_dir_use_symlinks=False)"
    
    if [ $? -eq 0 ]; then
        echo "${model_name} downloaded successfully!"
    else
        echo "Failed to download ${model_name} using huggingface_hub."
        echo "Trying alternative method with git clone..."
        
        # Remove directory contents if it exists
        if [ -d "${target_dir}" ]; then
            rm -rf "${target_dir}"/*
        fi
        
        git lfs install
        git clone "https://huggingface.co/${repo_id}" "${target_dir}"
        
        if [ $? -eq 0 ]; then
            echo "${model_name} downloaded successfully using git clone!"
        else
            echo "Failed to download ${model_name}. Please download it manually."
            echo "Visit: https://huggingface.co/${repo_id}"
        fi
    fi
}

# Main execution
echo "This script will download Qwen3-7B and DeepSeek Coder v3-7B models."
echo "Note: These models are large (7-15GB each) and require a good internet connection."
echo ""

read -p "Download Qwen3-7B? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    download_model "Qwen3-7B" "Qwen/Qwen3-7B" "${QWEN_DIR}"
fi

echo ""
read -p "Download DeepSeek Coder v3-7B? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    download_model "DeepSeek Coder v3-7B" "deepseek-ai/deepseek-coder-v3-7b" "${DEEPSEEK_DIR}"
fi

echo ""
echo "====================================="
echo "Download process completed!"
echo "====================================="
echo ""
echo "To start the services with these models, run:"
echo "docker-compose up -d"

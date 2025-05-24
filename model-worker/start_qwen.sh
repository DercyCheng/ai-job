#!/bin/bash

# 检查Python环境
if ! command -v python3 &> /dev/null; then
    echo "Python3 未安装，请先安装Python3"
    exit 1
fi

# 创建虚拟环境
if [ ! -d "venv" ]; then
    echo "正在创建Python虚拟环境..."
    python3 -m venv venv
fi

# 激活虚拟环境
echo "激活虚拟环境..."
source venv/bin/activate

# 安装依赖
echo "安装依赖包..."
pip install -r requirements.txt

# 检查CUDA可用性
echo "检查CUDA可用性..."
python -c "import torch; print(f'CUDA可用: {torch.cuda.is_available()}')"

# 启动服务
echo "启动Qwen3模型服务..."
python qwen_app.py --model_path "Qwen/Qwen1.5-7B-Chat" --port 5001

# 注意: 该脚本会一直运行服务，需要使用Ctrl+C终止

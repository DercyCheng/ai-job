# Python模型工作节点

这个目录包含了基于Python实现的大语言模型服务，用于处理AI Gateway系统中的LLM请求。目前支持DeepSeek V3 7B和Qwen3模型。

## 功能特性

- 加载和部署多种大语言模型
  - DeepSeek V3 7B
  - Qwen3 7B
- 提供标准化的文本生成API
- 兼容OpenAI聊天完成API格式
- 支持健康检查

## 快速开始

### 安装依赖

```bash
# 创建虚拟环境
python -m venv venv

# 激活虚拟环境
source venv/bin/activate  # Linux/Mac
venv\Scripts\activate     # Windows

# 安装依赖
pip install -r requirements.txt
```

### 启动DeepSeek V3服务

```bash
# 使用默认配置启动
python app.py

# 或使用自定义配置
python app.py --model_path "deepseek-ai/deepseek-v3-7b" --port 5000 --device cuda
```

### 启动Qwen3服务

```bash
# 使用默认配置启动
python qwen_app.py

# 或使用自定义配置
python qwen_app.py --model_path "Qwen/Qwen1.5-7B-Chat" --port 5001 --device cuda
```

### 也可以使用启动脚本

```bash
# 启动DeepSeek V3
chmod +x start.sh
./start.sh

# 启动Qwen3
chmod +x start_qwen.sh
./start_qwen.sh
```

## API接口

### 健康检查

```
GET /health
```

### 文本生成

```
POST /v1/generate
Content-Type: application/json

{
  "prompt": "用户输入的提示文本",
  "max_length": 2048,
  "temperature": 0.7,
  "top_p": 0.95,
  "stop": ["停止序列1", "停止序列2"]
}
```

### 聊天完成 (OpenAI兼容)

```
POST /v1/chat/completions
Content-Type: application/json

{
  "messages": [
    {"role": "system", "content": "你是一个有帮助的AI助手。"},
    {"role": "user", "content": "你好，请问你能做什么？"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7,
  "top_p": 0.95,
  "stop": ["停止序列1", "停止序列2"],
  "stream": false
}
```

### 流式聊天完成 (OpenAI兼容)

```
POST /v1/chat/completions
Content-Type: application/json

{
  "messages": [
    {"role": "system", "content": "你是一个有帮助的AI助手。"},
    {"role": "user", "content": "你好，请问你能做什么？"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7,
  "top_p": 0.95,
  "stop": ["停止序列1", "停止序列2"],
  "stream": true
}
```

响应将作为Server-Sent Events (SSE)流返回，格式与OpenAI流式API兼容。

## 配置参数

### DeepSeek V3服务

| 参数名 | 说明 | 默认值 |
|--------|------|--------|
| model_path | 模型路径或Hugging Face模型ID | deepseek-ai/deepseek-v3-7b |
| port | 服务端口 | 5000 |
| host | 服务地址 | 0.0.0.0 |
| device | 运行设备 (cuda或cpu) | cuda |

### Qwen3服务

| 参数名 | 说明 | 默认值 |
|--------|------|--------|
| model_path | 模型路径或Hugging Face模型ID | Qwen/Qwen1.5-7B-Chat |
| port | 服务端口 | 5001 |
| host | 服务地址 | 0.0.0.0 |
| device | 运行设备 (cuda或cpu) | cuda |

## 与Gateway集成

该服务设计为AI网关的模型工作节点，接收来自MCP服务的请求，并返回模型推理结果。

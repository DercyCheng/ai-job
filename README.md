# AI Gateway

基于装饰器模式的API网关和MCP服务框架，集成多种大语言模型

## 功能特性

- 基于装饰器模式的可扩展架构
- 支持动态配置加载
- 内置请求日志记录和认证
- 支持反向代理和智能路由
- 模块化设计
- 集成多种大语言模型
  - DeepSeek V3 7B
  - Qwen3 7B
- 兼容OpenAI API格式

## 系统架构

```
                ┌──────────────┐     ┌──────────────────┐
                │              │     │                  │
客户端 ──────→ │   API网关    │ ──→ │    MCP服务       │
                │              │     │                  │
                └──────────────┘     └──────────────────┘
                                             │
                          ┌──────────────────┴───────────────────┐
                          │                                      │
                ┌─────────▼────────┐                ┌────────────▼─────────┐
                │                  │                │                      │
                │ DeepSeek模型节点 │                │   Qwen3模型节点      │
                │                  │                │                      │
                └──────────────────┘                └──────────────────────┘
```

## 快速开始

### 安装依赖

```bash
# 安装Go依赖
make deps
```

### 启动服务

1. 启动DeepSeek模型工作节点
```bash
cd model-worker && ./start.sh
```

2. 启动Qwen3模型工作节点
```bash
cd model-worker && ./start_qwen.sh
```

3. 启动认证服务
```bash
make run-auth
```

4. 启动MCP服务
```bash
make run-mcp
```

5. 启动API网关
```bash
make run-gateway
```

或者使用开发模式一键启动所有服务（需要多个终端）:
```bash
make dev
```

### 构建项目

```bash
make build
```

## 配置说明

编辑`configs/config.yaml`文件配置各服务参数:

```yaml
# MCP服务配置
mcp:
  port: 8080
  log_level: info
  workers:
    - name: "deepseek-worker"
      url: "http://localhost:5000"
      model: "deepseek-v3-7b"
      # 更多配置...

# API网关配置  
gateway:
  port: 8081
  log_level: info
  target_url: "http://localhost:8080"
  routes:
    - path: "/v1/chat"
      target: "http://localhost:8080/mcp/v1/chat"
      auth_required: true
    # 更多路由...

# 认证服务配置
auth:
  port: 8082
  log_level: info
  jwt_secret: "your-secret-key"
  token_expiry: 86400  # 24小时

# 模型配置
models:
  deepseek-v3-7b:
    name: "DeepSeek V3 7B"
    description: "DeepSeek V3是一个强大的中文大语言模型"
    context_length: 4096
    # 更多配置...
```

## API使用示例

### 认证

```bash
curl -X POST http://localhost:8082/auth/token \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 聊天完成

```bash
curl -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "deepseek-v3-7b",
    "messages": [
      {"role": "system", "content": "你是一个有帮助的AI助手。"},
      {"role": "user", "content": "你好，请问你能做什么?"}
    ],
    "temperature": 0.7
  }'
```

或者使用Qwen3模型：

```bash
curl -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "qwen3-7b",
    "messages": [
      {"role": "system", "content": "你是一个有帮助的AI助手。"},
      {"role": "user", "content": "你好，请问你能做什么?"}
    ],
    "temperature": 0.7
  }'
```

### 流式聊天完成

```bash
curl -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "deepseek-v3-7b",
    "messages": [
      {"role": "system", "content": "你是一个有帮助的AI助手。"},
      {"role": "user", "content": "你好，请问你能做什么?"}
    ],
    "temperature": 0.7,
    "stream": true
  }'
```

也可以使用Qwen3模型的流式聊天：

```bash
curl -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "qwen3-7b",
    "messages": [
      {"role": "system", "content": "你是一个有帮助的AI助手。"},
      {"role": "user", "content": "你好，请问你能做什么?"}
    ],
    "temperature": 0.7,
    "stream": true
  }'
```

### 获取可用模型

```bash
curl http://localhost:8081/v1/models \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 目录结构

```
ai-gateway/
├── cmd/                  # 命令行入口
│   ├── gateway/          # API网关入口
│   ├── mcp/              # MCP服务入口
│   └── auth/             # 认证服务入口
├── configs/              # 配置文件
├── docs/                 # 文档
├── internal/             # 内部实现
│   ├── gateway/          # 网关实现
│   └── mcp/              # MCP服务实现
├── model-worker/         # Python模型工作节点
└── pkg/                  # 公共包
    └── utils/            # 工具函数
```

## 开发指南

- 使用装饰器模式扩展功能
- 通过实现Gateway/MCP接口添加新功能
- 遵循模块化设计原则
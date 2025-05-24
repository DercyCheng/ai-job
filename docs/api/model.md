# 模型 API

## 协议支持

系统同时支持 HTTP/REST API 和 gRPC API 两种方式访问模型服务。以下文档主要描述 REST API，gRPC 接口定义见文档末尾。

## 获取可用模型列表

**请求方式**: GET

**URL**: `/models`

**请求头**:

```
Authorization: Bearer {token}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "models": [
      {
        "id": "gpt-4",
        "name": "GPT-4",
        "description": "OpenAI GPT-4 大型语言模型",
        "version": "1.0",
        "capabilities": ["chat", "completion", "embedding"],
        "provider": "openai",
        "status": "active"
      },
      {
        "id": "claude-3",
        "name": "Claude 3",
        "description": "Anthropic Claude 3 大型语言模型",
        "version": "1.0",
        "capabilities": ["chat", "completion"],
        "provider": "anthropic",
        "status": "active"
      },
      {
        "id": "deepseek-v3-14b",
        "name": "DeepSeek V3 14B",
        "description": "本地部署的 DeepSeek V3 14B 大型语言模型",
        "version": "3.0",
        "capabilities": ["chat", "completion", "code"],
        "provider": "local",
        "status": "active",
        "deployment": {
          "type": "local",
          "hardware": "NVIDIA A100",
          "quantization": "int8"
        }
      },
      {
        "id": "qwen3-7b",
        "name": "Qwen3 7B",
        "description": "本地部署的 Qwen3 7B 大型语言模型",
        "version": "3.0",
        "capabilities": ["chat", "completion", "tool_use"],
        "provider": "local",
        "status": "active",
        "deployment": {
          "type": "local",
          "hardware": "NVIDIA A10",
          "quantization": "int8"
        }
      }
    ]
  },
  "request_id": "uuid"
}
```

## 获取模型详细信息

**请求方式**: GET

**URL**: `/models/{model_id}`

**请求头**:

```
Authorization: Bearer {token}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "gpt-4",
    "name": "GPT-4",
    "description": "OpenAI GPT-4 大型语言模型",
    "version": "1.0",
    "capabilities": ["chat", "completion", "embedding"],
    "provider": "openai",
    "status": "active",
    "parameters": {
      "max_tokens": 8192,
      "temperature": {
        "min": 0.0,
        "max": 2.0,
        "default": 1.0
      },
      "top_p": {
        "min": 0.0,
        "max": 1.0,
        "default": 1.0
      }
    },
    "pricing": {
      "input_per_1k_tokens": 0.03,
      "output_per_1k_tokens": 0.06,
      "currency": "USD"
    },
    "stats": {
      "average_latency_ms": 500,
      "availability": 0.995
    }
  },
  "request_id": "uuid"
}
```

## 文本补全 API

**请求方式**: POST

**URL**: `/models/{model_id}/completions`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```json
{
  "prompt": "string",
  "max_tokens": 100,
  "temperature": 1.0,
  "top_p": 1.0,
  "n": 1,
  "stream": false,
  "stop": ["\n", "user:"]
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "cmpl-123456",
    "object": "text_completion",
    "created": 1677858242,
    "model": "gpt-4",
    "choices": [
      {
        "text": "这是模型生成的文本",
        "index": 0,
        "finish_reason": "stop"
      }
    ],
    "usage": {
      "prompt_tokens": 10,
      "completion_tokens": 20,
      "total_tokens": 30
    }
  },
  "request_id": "uuid"
}
```

## 聊天对话 API

**请求方式**: POST

**URL**: `/models/{model_id}/chat/completions`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```json
{
  "messages": [
    {
      "role": "system",
      "content": "你是一个有用的AI助手。"
    },
    {
      "role": "user",
      "content": "你好，请介绍一下自己。"
    }
  ],
  "max_tokens": 500,
  "temperature": 0.7,
  "stream": false
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "chatcmpl-123456",
    "object": "chat.completion",
    "created": 1677858242,
    "model": "gpt-4",
    "choices": [
      {
        "message": {
          "role": "assistant",
          "content": "你好！我是一个AI助手，基于大型语言模型构建..."
        },
        "finish_reason": "stop",
        "index": 0
      }
    ],
    "usage": {
      "prompt_tokens": 25,
      "completion_tokens": 40,
      "total_tokens": 65
    }
  },
  "request_id": "uuid"
}
```

## 向量嵌入 API

**请求方式**: POST

**URL**: `/models/{model_id}/embeddings`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```json
{
  "input": ["这是第一段文本", "这是第二段文本"],
  "encoding_format": "float"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "object": "list",
    "data": [
      {
        "object": "embedding",
        "embedding": [0.002, -0.003, 0.67, ...],
        "index": 0
      },
      {
        "object": "embedding",
        "embedding": [0.003, -0.002, 0.78, ...],
        "index": 1
      }
    ],
    "model": "text-embedding-ada-002",
    "usage": {
      "prompt_tokens": 20,
      "total_tokens": 20
    }
  },
  "request_id": "uuid"
}
```

## 本地模型管理 API

### 获取本地模型列表

**请求方式**: GET

**URL**: `/models/local`

**请求头**:

```
Authorization: Bearer {token}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "models": [
      {
        "id": "deepseek-v3-14b",
        "name": "DeepSeek V3 14B",
        "version": "3.0",
        "status": "active",
        "deployment": {
          "node_id": "worker-01",
          "hardware": "NVIDIA A100",
          "memory_usage": "28GB",
          "quantization": "int8",
          "loaded_at": "2025-05-22T10:00:00Z",
          "health": "healthy"
        },
        "performance": {
          "tokens_per_second": 80,
          "avg_latency_ms": 250,
          "max_batch_size": 8
        }
      },
      {
        "id": "qwen3-7b",
        "name": "Qwen3 7B",
        "version": "3.0",
        "status": "active",
        "deployment": {
          "node_id": "worker-02",
          "hardware": "NVIDIA A10",
          "memory_usage": "14GB",
          "quantization": "int8",
          "loaded_at": "2025-05-23T15:30:00Z",
          "health": "healthy"
        },
        "performance": {
          "tokens_per_second": 65,
          "avg_latency_ms": 180,
          "max_batch_size": 12
        }
      }
    ]
  },
  "request_id": "uuid"
}
```

### 部署新本地模型

**请求方式**: POST

**URL**: `/models/local/deploy`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```json
{
  "model_type": "deepseek-v3",
  "model_size": "7b",
  "quantization": "int8",
  "target_nodes": ["worker-03"],
  "deployment_config": {
    "max_batch_size": 8,
    "use_flash_attention": true,
    "use_kv_cache": true,
    "tensor_parallel_size": 1,
    "cpu_offload": false
  }
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "deployment_id": "deploy-123456",
    "status": "in_progress",
    "model_id": "deepseek-v3-7b",
    "estimated_time": "5m",
    "progress_url": "/models/local/deployments/deploy-123456"
  },
  "request_id": "uuid"
}
```

### 获取本地模型部署状态

**请求方式**: GET

**URL**: `/models/local/deployments/{deployment_id}`

**请求头**:

```
Authorization: Bearer {token}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "deployment_id": "deploy-123456",
    "model_id": "deepseek-v3-7b",
    "status": "in_progress",
    "progress": 65,
    "current_stage": "downloading_weights",
    "stages": [
      {"name": "preparation", "status": "completed", "duration": "10s"},
      {"name": "downloading_weights", "status": "in_progress", "progress": 65},
      {"name": "model_initialization", "status": "pending"},
      {"name": "model_loading", "status": "pending"},
      {"name": "service_registration", "status": "pending"}
    ],
    "logs": [
      {"timestamp": "2025-05-24T10:15:00Z", "level": "info", "message": "开始部署模型"},
      {"timestamp": "2025-05-24T10:15:10Z", "level": "info", "message": "准备工作完成"},
      {"timestamp": "2025-05-24T10:15:15Z", "level": "info", "message": "开始下载模型权重"}
    ],
    "estimated_completion": "2025-05-24T10:30:00Z"
  },
  "request_id": "uuid"
}
```

### 卸载本地模型

**请求方式**: DELETE

**URL**: `/models/local/{model_id}`

**请求头**:

```
Authorization: Bearer {token}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "model_id": "deepseek-v3-7b",
    "status": "unloading",
    "nodes": ["worker-03"]
  },
  "request_id": "uuid"
}
```

## gRPC API 接口定义

系统同时支持 gRPC 协议，以下是主要 Protocol Buffers 定义：

### 模型服务定义

```protobuf
syntax = "proto3";

package model;

service ModelService {
  // 获取可用模型列表
  rpc ListModels(ListModelsRequest) returns (ListModelsResponse);
  
  // 获取模型详情
  rpc GetModel(GetModelRequest) returns (GetModelResponse);
  
  // 文本补全
  rpc CreateCompletion(CompletionRequest) returns (CompletionResponse);
  
  // 流式文本补全
  rpc CreateCompletionStream(CompletionRequest) returns (stream CompletionChunk);
  
  // 聊天对话
  rpc CreateChatCompletion(ChatCompletionRequest) returns (ChatCompletionResponse);
  
  // 流式聊天对话
  rpc CreateChatCompletionStream(ChatCompletionRequest) returns (stream ChatCompletionChunk);
  
  // 向量嵌入
  rpc CreateEmbeddings(EmbeddingsRequest) returns (EmbeddingsResponse);
  
  // 本地模型管理
  rpc ListLocalModels(ListLocalModelsRequest) returns (ListLocalModelsResponse);
  rpc DeployLocalModel(DeployLocalModelRequest) returns (DeployLocalModelResponse);
  rpc GetDeploymentStatus(GetDeploymentStatusRequest) returns (GetDeploymentStatusResponse);
  rpc UnloadLocalModel(UnloadLocalModelRequest) returns (UnloadLocalModelResponse);
}

// 请求/响应定义省略，与REST API的JSON结构对应
```

### 编译与使用

1. 编译 Protocol Buffers

```bash
protoc --go_out=. --go-grpc_out=. model.proto
```

2. 客户端使用示例 (Go)

```go
package main

import (
    "context"
    "log"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    pb "example.com/ai-gateway/model"
)

func main() {
    // 建立安全连接
    creds, err := credentials.NewClientTLSFromFile("ca.pem", "")
    if err != nil {
        log.Fatalf("Failed to load credentials: %v", err)
    }
    
    conn, err := grpc.Dial("api.example.com:50051", grpc.WithTransportCredentials(creds))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()
    
    // 创建客户端
    client := pb.NewModelServiceClient(conn)
    
    // 调用聊天API
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
    defer cancel()
    
    resp, err := client.CreateChatCompletion(ctx, &pb.ChatCompletionRequest{
        ModelId: "deepseek-v3-14b",
        Messages: []*pb.Message{
            {Role: "system", Content: "你是一个有用的AI助手。"},
            {Role: "user", Content: "你好，请介绍一下自己。"},
        },
        MaxTokens: 500,
        Temperature: 0.7,
    })
    
    if err != nil {
        log.Fatalf("Request failed: %v", err)
    }
    
    log.Printf("Response: %s", resp.Choices[0].Message.Content)
}
```

3. 流式响应示例 (Go)

```go
// 流式聊天
stream, err := client.CreateChatCompletionStream(ctx, &pb.ChatCompletionRequest{
    ModelId: "qwen3-7b",
    Messages: []*pb.Message{
        {Role: "user", Content: "写一首关于春天的诗。"},
    },
    MaxTokens: 200,
    Temperature: 0.8,
    Stream: true,
})

if err != nil {
    log.Fatalf("Stream request failed: %v", err)
}

for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatalf("Stream receive error: %v", err)
    }
    
    // 处理流式响应
    fmt.Print(chunk.Choices[0].Delta.Content)
}
```

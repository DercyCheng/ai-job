# 双协议实现指南

本文档详细说明了 AI Gateway MCP Server 的 HTTP/gRPC 双协议实现方案，包括设计原则、代码结构、接口定义和使用示例。

## 设计原则

双协议实现遵循以下设计原则：

1. **统一的业务逻辑**：HTTP 和 gRPC 接口共享相同的核心业务逻辑
2. **协议无关的内部表示**：内部数据结构与传输协议解耦
3. **一致的认证机制**：两种协议支持相同的认证流程
4. **等价的功能集**：确保功能在两种协议上保持一致
5. **独立的性能优化**：针对不同协议特点进行单独优化

## 架构设计

```
┌───────────────────────────┐     ┌────────────────────────┐
│                           │     │                        │
│    HTTP/REST API 服务     │     │      gRPC API 服务     │
│    (Gin Framework)        │     │                        │
│                           │     │                        │
└─────────────┬─────────────┘     └──────────┬─────────────┘
              │                               │
              │                               │
              ▼                               ▼
┌─────────────────────────────────────────────────────────┐
│                                                         │
│                    适配器层                             │
│     (协议转换、请求/响应映射、错误处理)                 │
│                                                         │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                                                         │
│                   核心服务层                            │
│      (业务逻辑、模型调用、认证授权、限流)               │
│                                                         │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                                                         │
│                   数据访问层                            │
│         (数据库交互、缓存、外部服务调用)                │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## 代码结构

```
api-gateway/
├── cmd/
│   ├── server/
│   │   └── main.go          # 主入口，启动HTTP和gRPC服务
├── internal/
│   ├── adapter/
│   │   ├── http/            # HTTP适配器
│   │   │   ├── handler/
│   │   │   ├── middleware/
│   │   │   └── router.go
│   │   └── grpc/            # gRPC适配器
│   │       ├── handler/
│   │       ├── interceptor/
│   │       └── server.go
│   ├── core/                # 核心业务逻辑
│   │   ├── auth/
│   │   ├── model/
│   │   └── service/
│   └── proto/               # Protocol Buffers定义
│       ├── auth.proto
│       ├── model.proto
│       └── task.proto
└── pkg/
    ├── http/
    │   └── response/        # HTTP响应工具
    └── grpc/
        └── error/           # gRPC错误处理
```

## Protocol Buffers 定义

Protocol Buffers 是双协议实现的核心，定义了 gRPC 服务接口和消息结构：

### model.proto 示例

```protobuf
syntax = "proto3";

package mcp.model.v1;

option go_package = "github.com/example/ai-gateway/proto/model/v1;modelv1";

import "google/protobuf/timestamp.proto";

// 模型服务定义
service ModelService {
  // 获取模型列表
  rpc ListModels(ListModelsRequest) returns (ListModelsResponse);
  
  // 获取模型详情
  rpc GetModel(GetModelRequest) returns (GetModelResponse);
  
  // 创建聊天完成
  rpc CreateChatCompletion(ChatCompletionRequest) returns (ChatCompletionResponse);
  
  // 流式聊天完成
  rpc CreateChatCompletionStream(ChatCompletionRequest) returns (stream ChatCompletionChunk);
  
  // 创建向量嵌入
  rpc CreateEmbeddings(EmbeddingsRequest) returns (EmbeddingsResponse);
  
  // 部署本地模型
  rpc DeployLocalModel(DeployLocalModelRequest) returns (DeployLocalModelResponse);
  
  // 获取本地模型列表
  rpc ListLocalModels(ListLocalModelsRequest) returns (ListLocalModelsResponse);
}

// 请求和响应消息定义
message ListModelsRequest {
  // 过滤条件
  optional string provider = 1;
  optional bool include_local = 2;
}

message ListModelsResponse {
  repeated Model models = 1;
}

message Model {
  string id = 1;
  string name = 2;
  string provider = 3;
  string version = 4;
  string description = 5;
  repeated string capabilities = 6;
  string status = 7;
  
  // 本地模型特定信息
  optional LocalModelInfo local_info = 8;
}

message LocalModelInfo {
  string deployment_id = 1;
  string node_id = 2;
  string hardware = 3;
  string quantization = 4;
  google.protobuf.Timestamp loaded_at = 5;
}

// 更多消息定义...
```

## 代码实现

### 主入口 (main.go)

```go
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/ai-gateway/internal/adapter/grpc"
	"github.com/example/ai-gateway/internal/adapter/http"
	"github.com/example/ai-gateway/internal/config"
	"github.com/example/ai-gateway/internal/core/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化服务
	services := service.NewServices(cfg)

	// 启动 HTTP 服务
	httpServer := http.NewServer(cfg, services)
	go func() {
		log.Printf("Starting HTTP server on :%d", cfg.Server.HTTP.Port)
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 启动 gRPC 服务
	grpcServer := grpc.NewServer(cfg, services)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPC.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	go func() {
		log.Printf("Starting gRPC server on :%d", cfg.Server.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// 关闭 HTTP 服务
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	// 关闭 gRPC 服务
	grpcServer.GracefulStop()

	log.Println("Servers exited properly")
}
```

### HTTP 适配器示例

```go
// internal/adapter/http/handler/model_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/example/ai-gateway/internal/core/service"
	"github.com/example/ai-gateway/pkg/http/response"
)

type ModelHandler struct {
	modelService service.ModelService
}

func NewModelHandler(modelService service.ModelService) *ModelHandler {
	return &ModelHandler{
		modelService: modelService,
	}
}

// ListModels 获取模型列表
func (h *ModelHandler) ListModels(c *gin.Context) {
	provider := c.Query("provider")
	includeLocal := c.Query("include_local") == "true"

	models, err := h.modelService.ListModels(c.Request.Context(), provider, includeLocal)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"models": models,
	})
}

// 更多HTTP处理方法...
```

### gRPC 适配器示例

```go
// internal/adapter/grpc/handler/model_handler.go
package handler

import (
	"context"

	"github.com/example/ai-gateway/internal/core/service"
	pb "github.com/example/ai-gateway/internal/proto/model/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ModelHandler struct {
	pb.UnimplementedModelServiceServer
	modelService service.ModelService
}

func NewModelHandler(modelService service.ModelService) *ModelHandler {
	return &ModelHandler{
		modelService: modelService,
	}
}

// ListModels 获取模型列表
func (h *ModelHandler) ListModels(ctx context.Context, req *pb.ListModelsRequest) (*pb.ListModelsResponse, error) {
	provider := ""
	includeLocal := false
	
	if req.Provider != nil {
		provider = *req.Provider
	}
	
	if req.IncludeLocal != nil {
		includeLocal = *req.IncludeLocal
	}
	
	models, err := h.modelService.ListModels(ctx, provider, includeLocal)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	pbModels := make([]*pb.Model, len(models))
	for i, m := range models {
		pbModels[i] = mapModelToPb(m)
	}
	
	return &pb.ListModelsResponse{
		Models: pbModels,
	}, nil
}

// 更多gRPC处理方法和辅助函数...
```

### 服务层示例

```go
// internal/core/service/model_service.go
package service

import (
	"context"

	"github.com/example/ai-gateway/internal/core/model"
	"github.com/example/ai-gateway/internal/core/repository"
)

type ModelService interface {
	ListModels(ctx context.Context, provider string, includeLocal bool) ([]model.Model, error)
	GetModel(ctx context.Context, id string) (*model.Model, error)
	CreateChatCompletion(ctx context.Context, req model.ChatCompletionRequest) (*model.ChatCompletionResponse, error)
	CreateEmbeddings(ctx context.Context, req model.EmbeddingsRequest) (*model.EmbeddingsResponse, error)
	DeployLocalModel(ctx context.Context, req model.DeployRequest) (*model.DeployResponse, error)
	// 更多方法...
}

type modelService struct {
	modelRepo  repository.ModelRepository
	workerRepo repository.WorkerRepository
}

func NewModelService(modelRepo repository.ModelRepository, workerRepo repository.WorkerRepository) ModelService {
	return &modelService{
		modelRepo:  modelRepo,
		workerRepo: workerRepo,
	}
}

// 实现ModelService接口的方法...
```

## 客户端使用示例

### HTTP 客户端示例 (Go)

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	client := &http.Client{}
	
	// 准备请求
	reqBody, err := json.Marshal(map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "你好，请介绍一下自己"},
		},
		"max_tokens": 500,
		"temperature": 0.7,
	})
	if err != nil {
		panic(err)
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", "http://localhost:8080/models/deepseek-v3-14b/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer your-api-key")
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	// 处理响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Response:", string(body))
}
```

### gRPC 客户端示例 (Go)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "github.com/example/ai-gateway/proto/model/v1"
)

func main() {
	// 建立连接
	creds, err := credentials.NewClientTLSFromFile("ca.pem", "")
	if err != nil {
		log.Fatalf("Failed to load credentials: %v", err)
	}
	
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	
	// 创建客户端
	client := pb.NewModelServiceClient(conn)
	
	// 设置上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// 准备请求
	req := &pb.ChatCompletionRequest{
		ModelId: "deepseek-v3-14b",
		Messages: []*pb.Message{
			{Role: "user", Content: "你好，请介绍一下自己"},
		},
		MaxTokens: 500,
		Temperature: 0.7,
	}
	
	// 发送请求
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	
	// 处理响应
	fmt.Println("Model:", resp.Model)
	fmt.Println("Content:", resp.Choices[0].Message.Content)
}
```

## Python 客户端示例

### HTTP 客户端 (Python)

```python
import requests

# API配置
api_url = "http://localhost:8080"
api_key = "your-api-key"

# 创建聊天完成
def chat_completion(model_id, messages, max_tokens=500, temperature=0.7):
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {api_key}"
    }
    
    payload = {
        "messages": messages,
        "max_tokens": max_tokens,
        "temperature": temperature
    }
    
    response = requests.post(
        f"{api_url}/models/{model_id}/chat/completions",
        headers=headers,
        json=payload
    )
    
    if response.status_code == 200:
        return response.json()["data"]
    else:
        raise Exception(f"API请求失败: {response.status_code} - {response.text}")

# 使用示例
result = chat_completion(
    "deepseek-v3-14b",
    [{"role": "user", "content": "你好，请介绍一下自己"}]
)

print(f"模型回复: {result['choices'][0]['message']['content']}")
```

### gRPC 客户端 (Python)

```python
import grpc
import model_pb2
import model_pb2_grpc

# 建立连接
channel = grpc.secure_channel(
    'localhost:50051',
    grpc.ssl_channel_credentials(open('ca.pem', 'rb').read())
)

# 创建客户端
stub = model_pb2_grpc.ModelServiceStub(channel)

# 准备请求
request = model_pb2.ChatCompletionRequest(
    model_id="deepseek-v3-14b",
    messages=[
        model_pb2.Message(role="user", content="你好，请介绍一下自己")
    ],
    max_tokens=500,
    temperature=0.7
)

# 发送请求
try:
    response = stub.CreateChatCompletion(request)
    print(f"模型回复: {response.choices[0].message.content}")
except grpc.RpcError as e:
    print(f"RPC错误: {e.details()}")
finally:
    channel.close()
```

## 认证与授权

双协议系统使用统一的认证和授权机制：

### HTTP 认证中间件

```go
// internal/adapter/http/middleware/auth.go
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/example/ai-gateway/internal/core/auth"
	"github.com/example/ai-gateway/pkg/http/response"
)

func AuthMiddleware(authService auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "未提供认证令牌")
			c.Abort()
			return
		}
		
		// 解析令牌
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, 401, "无效的认证格式")
			c.Abort()
			return
		}
		
		token := parts[1]
		
		// 验证令牌
		userID, err := authService.ValidateToken(c, token)
		if err != nil {
			response.Error(c, 401, "无效的认证令牌")
			c.Abort()
			return
		}
		
		// 将用户ID保存到上下文
		c.Set("user_id", userID)
		c.Next()
	}
}
```

### gRPC 认证拦截器

```go
// internal/adapter/grpc/interceptor/auth.go
package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	
	"github.com/example/ai-gateway/internal/core/auth"
)

func AuthInterceptor(authService auth.Service) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 从元数据获取令牌
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "缺少元数据")
		}
		
		// 获取认证信息
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "未提供认证令牌")
		}
		
		// 解析令牌
		parts := strings.Split(authHeaders[0], " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return nil, status.Error(codes.Unauthenticated, "无效的认证格式")
		}
		
		token := parts[1]
		
		// 验证令牌
		userID, err := authService.ValidateToken(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "无效的认证令牌")
		}
		
		// 创建带用户ID的新上下文
		newCtx := context.WithValue(ctx, "user_id", userID)
		
		// 继续处理请求
		return handler(newCtx, req)
	}
}
```

## 性能优化

### HTTP 性能优化

1. **连接池管理**

```go
// 使用自定义Transport优化HTTP客户端连接池
transport := &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   100,
	MaxConnsPerHost:       100,
}

client := &http.Client{
	Transport: transport,
	Timeout:   30 * time.Second,
}
```

2. **HTTP/2 支持**

```go
// 启用HTTP/2支持
server := &http.Server{
	Addr:    fmt.Sprintf(":%d", cfg.Server.HTTP.Port),
	Handler: router,
	TLSConfig: &tls.Config{
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{"h2", "http/1.1"},
	},
}
```

### gRPC 性能优化

1. **流控制**

```go
// 设置流控制选项
opts := []grpc.ServerOption{
	grpc.InitialWindowSize(65536),     // 初始窗口大小
	grpc.InitialConnWindowSize(65536), // 初始连接窗口大小
	grpc.MaxConcurrentStreams(1000),   // 每个连接的最大并发流数
}

server := grpc.NewServer(opts...)
```

2. **消息大小限制**

```go
// 设置消息大小限制
opts := []grpc.ServerOption{
	grpc.MaxRecvMsgSize(4 * 1024 * 1024), // 4MB
	grpc.MaxSendMsgSize(4 * 1024 * 1024), // 4MB
}

server := grpc.NewServer(opts...)
```

3. **连接复用**

```go
// 客户端启用连接复用
conn, err := grpc.Dial(
	address,
	grpc.WithTransportCredentials(credentials),
	grpc.WithBlock(),
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second, // 每10秒ping一次服务器
		Timeout:             3 * time.Second,  // 3秒超时
		PermitWithoutStream: true,             // 允许在没有活动流的情况下发送ping
	}),
)
```

## 监控与指标

双协议系统需要分别为 HTTP 和 gRPC 提供监控指标：

### HTTP 指标中间件

```go
// internal/adapter/http/middleware/metrics.go
package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 处理请求
		c.Next()
		
		// 记录指标
		status := strconv.Itoa(c.Writer.Status())
		elapsed := time.Since(start).Seconds()
		
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(elapsed)
	}
}
```

### gRPC 指标拦截器

```go
// internal/adapter/grpc/interceptor/metrics.go
package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	grpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)
	
	grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(grpcRequestsTotal)
	prometheus.MustRegister(grpcRequestDuration)
}

func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		
		// 处理请求
		resp, err := handler(ctx, req)
		
		// 记录指标
		st, _ := status.FromError(err)
		elapsed := time.Since(start).Seconds()
		
		grpcRequestsTotal.WithLabelValues(info.FullMethod, st.Code().String()).Inc()
		grpcRequestDuration.WithLabelValues(info.FullMethod).Observe(elapsed)
		
		return resp, err
	}
}
```

## 开发与测试

### HTTP 测试

```go
// test/api/model_test.go
package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/example/ai-gateway/internal/adapter/http"
	"github.com/example/ai-gateway/internal/core/service"
	"github.com/example/ai-gateway/internal/mock"
)

func TestListModels(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	mockModelService := mock.NewMockModelService()
	
	// 创建测试服务器
	router := http.SetupRouter(mockModelService)
	
	// 发送测试请求
	req := httptest.NewRequest("GET", "/models", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, float64(200), response["code"])
	assert.NotNil(t, response["data"])
}
```

### gRPC 测试

```go
// test/grpc/model_test.go
package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	
	"github.com/example/ai-gateway/internal/adapter/grpc/handler"
	"github.com/example/ai-gateway/internal/mock"
	pb "github.com/example/ai-gateway/internal/proto/model/v1"
)

func TestListModels(t *testing.T) {
	// 设置模拟服务
	mockModelService := mock.NewMockModelService()
	handler := handler.NewModelHandler(mockModelService)
	
	// 创建内存中的gRPC服务器
	listener := bufconn.Listen(1024)
	server := grpc.NewServer()
	pb.RegisterModelServiceServer(server, handler)
	
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Fatalf("Server error: %v", err)
		}
	}()
	defer server.Stop()
	
	// 创建客户端连接
	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
	
	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithInsecure(),
	)
	assert.NoError(t, err)
	defer conn.Close()
	
	// 创建客户端
	client := pb.NewModelServiceClient(conn)
	
	// 执行测试
	resp, err := client.ListModels(context.Background(), &pb.ListModelsRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp.Models)
	assert.Greater(t, len(resp.Models), 0)
}
```

## 双协议切换策略

在实际生产环境中，可以根据不同场景选择合适的协议：

1. **内部服务通信**: 优先使用 gRPC，获得更好的性能和类型安全
2. **外部客户端集成**: 提供 HTTP/REST API，适配更广泛的客户端
3. **浏览器直接访问**: 使用 HTTP/REST API
4. **高并发流式响应**: 优先使用 gRPC 流式 API
5. **微服务间通信**: 使用 gRPC 获得更好的性能

## 总结

双协议实现为 AI Gateway MCP Server 提供了更广泛的适用性和更高的性能。HTTP/REST API 便于集成和调试，而 gRPC 则提供更好的性能和类型安全性。通过共享核心业务逻辑，我们确保了功能的一致性，同时针对不同协议特点进行了优化。

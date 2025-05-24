package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ModelWorker 表示一个模型工作节点
type ModelWorker struct {
	Name      string
	URL       string
	Model     string
	Priority  int
	MaxTokens int
	Timeout   int
	Streaming bool
}

// ChatMessage 表示聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 表示聊天请求
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatResponse 表示聊天响应
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ModelInfo 表示模型信息
type ModelInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	ContextLength int      `json:"context_length"`
	Capabilities  []string `json:"capabilities"`
}

// ModelService 处理模型相关请求的服务
type ModelService struct {
	Workers []ModelWorker
	Models  map[string]ModelInfo
}

// NewModelService 创建模型服务
func NewModelService(workers []ModelWorker, models map[string]ModelInfo) *ModelService {
	return &ModelService{
		Workers: workers,
		Models:  models,
	}
}

// findWorker 根据模型名称查找对应的工作节点
func (s *ModelService) findWorker(modelName string) *ModelWorker {
	for _, worker := range s.Workers {
		if worker.Model == modelName {
			return &worker
		}
	}
	return nil
}

// HandleChatRequest 处理聊天请求
func (s *ModelService) HandleChatRequest(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var request ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 查找对应的模型工作节点
	worker := s.findWorker(request.Model)
	if worker == nil {
		http.Error(w, fmt.Sprintf("Model %s not found", request.Model), http.StatusNotFound)
		return
	}

	// 准备转发请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
		return
	}

	// 设置超时
	client := &http.Client{
		Timeout: time.Duration(worker.Timeout) * time.Second,
	}

	// 创建新请求
	req, err := http.NewRequest("POST", worker.URL+"/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to model worker: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 转发响应头
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 设置响应状态码
	w.WriteHeader(resp.StatusCode)

	// 转发响应体
	io.Copy(w, resp.Body)
}

// HandleListModels 处理列出模型请求
func (s *ModelService) HandleListModels(w http.ResponseWriter, r *http.Request) {
	// 准备响应
	type ListModelsResponse struct {
		Object string      `json:"object"`
		Data   []ModelInfo `json:"data"`
	}

	// 构建模型列表
	var modelsList []ModelInfo
	for _, model := range s.Models {
		modelsList = append(modelsList, model)
	}

	// 返回响应
	response := ListModelsResponse{
		Object: "list",
		Data:   modelsList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// modelServiceDecorator 模型服务装饰器
type modelServiceDecorator struct {
	service Service
	model   *ModelService
}

// WithModelService 添加模型服务功能的装饰器
func WithModelService(service Service, model *ModelService) Service {
	return &modelServiceDecorator{
		service: service,
		model:   model,
	}
}

// HandleRequest 处理请求并根据路径分发
func (d *modelServiceDecorator) HandleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 根据路径分发到对应的处理函数
	switch {
	case path == "/mcp/v1/chat/completions" || path == "/mcp/v1/chat":
		d.model.HandleChatRequest(w, r)
	case path == "/mcp/v1/models":
		d.model.HandleListModels(w, r)
	case path == "/health":
		// 健康检查
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	default:
		// 默认处理
		d.service.HandleRequest(w, r)
	}
}

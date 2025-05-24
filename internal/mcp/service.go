package mcp

import (
	"fmt"
	"log"
	"net/http"
)

// Service 定义MCP服务接口
type Service interface {
	HandleRequest(w http.ResponseWriter, r *http.Request)
}

// BaseService 基础MCP服务实现
type BaseService struct{}

// NewBaseService 创建基础MCP服务
func NewBaseService() *BaseService {
	return &BaseService{}
}

// HandleRequest 处理MCP请求
func (s *BaseService) HandleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "MCP Base Service Response")
}

// loggingDecorator 日志装饰器
type loggingDecorator struct {
	service Service
}

// WithLogging 添加日志功能的装饰器
func WithLogging(service Service) Service {
	return &loggingDecorator{service: service}
}

func (d *loggingDecorator) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// 记录请求信息
	log.Printf("MCP request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// 调用实际处理
	d.service.HandleRequest(w, r)

	// 记录响应信息
	log.Printf("MCP response: %s %s", r.Method, r.URL.Path)
}

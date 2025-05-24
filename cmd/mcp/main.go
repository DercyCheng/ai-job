package main

import (
	"fmt"
	"log"
	"net/http"

	"ai-gatway/internal/mcp"
	"ai-gatway/pkg/utils"
)

func main() {
	// 加载配置
	port, logLevel, workers := utils.GetMCPConfig()
	models := utils.GetModelsConfig()

	// 转换工作节点格式
	var modelWorkers []mcp.ModelWorker
	for _, worker := range workers {
		modelWorkers = append(modelWorkers, mcp.ModelWorker{
			Name:      worker.Name,
			URL:       worker.URL,
			Model:     worker.Model,
			Priority:  worker.Priority,
			MaxTokens: worker.MaxTokens,
			Timeout:   worker.Timeout,
			Streaming: worker.Streaming,
		})
	}

	// 转换模型信息格式
	modelInfoMap := make(map[string]mcp.ModelInfo)
	for id, info := range models {
		modelInfoMap[id] = mcp.ModelInfo{
			ID:            id,
			Name:          info.Name,
			Description:   info.Description,
			ContextLength: info.ContextLength,
			Capabilities:  info.Capabilities,
		}
	}

	// 创建模型服务
	modelService := mcp.NewModelService(modelWorkers, modelInfoMap)

	// 创建基础MCP服务
	baseService := mcp.NewBaseService()

	// 使用装饰器模式添加功能
	service := mcp.WithLogging(baseService)
	service = mcp.WithModelService(service, modelService)

	// 设置HTTP路由
	http.HandleFunc("/mcp", service.HandleRequest)
	http.HandleFunc("/mcp/v1/chat/completions", service.HandleRequest)
	http.HandleFunc("/mcp/v1/chat", service.HandleRequest)
	http.HandleFunc("/mcp/v1/models", service.HandleRequest)
	http.HandleFunc("/health", service.HandleRequest)

	// 启动服务
	addr := fmt.Sprintf(":%d", port)
	log.Printf("MCP Server starting on %s with log level %s...\n", addr, logLevel)
	log.Printf("Loaded %d model workers and %d model definitions\n", len(modelWorkers), len(modelInfoMap))
	log.Fatal(http.ListenAndServe(addr, nil))
}

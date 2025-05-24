package utils

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	config     *viper.Viper
	configOnce sync.Once
)

// LoadConfig 加载并返回配置实例
func LoadConfig() (*viper.Viper, error) {
	var err error
	configOnce.Do(func() {
		config = viper.New()
		config.SetConfigName("config")
		config.SetConfigType("yaml")
		config.AddConfigPath("./configs")
		config.AddConfigPath("../configs")
		config.AddConfigPath("../../configs")

		if err = config.ReadInConfig(); err != nil {
			err = fmt.Errorf("failed to read config: %v", err)
			return
		}
	})

	return config, err
}

// Worker 表示模型工作节点配置
type Worker struct {
	Name      string
	URL       string
	Model     string
	Priority  int
	MaxTokens int
	Timeout   int
	Streaming bool
}

// ModelInfo 模型信息
type ModelInfo struct {
	Name          string
	Description   string
	ContextLength int
	Capabilities  []string
}

// Route 路由信息
type Route struct {
	Path         string
	Target       string
	AuthRequired bool
}

// ConsulConfig Consul配置
type ConsulConfig struct {
	Host     string
	Port     int
	Service  string
	CheckURL string
	Tags     []string
}

// GetConsulConfig 获取Consul配置
func GetConsulConfig() (host string, port int, service string, checkURL string, tags []string) {
	config, _ := LoadConfig()
	return config.GetString("consul.host"),
		config.GetInt("consul.port"),
		config.GetString("consul.service"),
		config.GetString("consul.check_url"),
		config.GetStringSlice("consul.tags")
}

// GetMCPConfig 获取MCP服务配置
func GetMCPConfig() (port int, logLevel string, workers []Worker) {
	config, _ := LoadConfig()

	// 解析工作节点配置
	var workerConfigs []map[string]interface{}
	if err := config.UnmarshalKey("mcp.workers", &workerConfigs); err == nil {
		for _, wc := range workerConfigs {
			worker := Worker{
				Name:      wc["name"].(string),
				URL:       wc["url"].(string),
				Model:     wc["model"].(string),
				Priority:  int(wc["priority"].(int64)),
				MaxTokens: int(wc["max_tokens"].(int64)),
				Timeout:   int(wc["timeout"].(int64)),
				Streaming: wc["streaming"].(bool),
			}
			workers = append(workers, worker)
		}
	}

	return config.GetInt("mcp.port"), config.GetString("mcp.log_level"), workers
}

// GetGatewayConfig 获取网关配置
func GetGatewayConfig() (port int, logLevel, targetURL string, routes []Route) {
	config, _ := LoadConfig()

	// 解析路由配置
	var routeConfigs []map[string]interface{}
	if err := config.UnmarshalKey("gateway.routes", &routeConfigs); err == nil {
		for _, rc := range routeConfigs {
			route := Route{
				Path:         rc["path"].(string),
				Target:       rc["target"].(string),
				AuthRequired: rc["auth_required"].(bool),
			}
			routes = append(routes, route)
		}
	}

	return config.GetInt("gateway.port"),
		config.GetString("gateway.log_level"),
		config.GetString("gateway.target_url"),
		routes
}

// GetAuthConfig 获取认证服务配置
func GetAuthConfig() (port int, logLevel, jwtSecret string, tokenExpiry int) {
	config, _ := LoadConfig()
	return config.GetInt("auth.port"),
		config.GetString("auth.log_level"),
		config.GetString("auth.jwt_secret"),
		config.GetInt("auth.token_expiry")
}

// GetModelsConfig 获取模型配置
func GetModelsConfig() map[string]ModelInfo {
	config, _ := LoadConfig()

	models := make(map[string]ModelInfo)
	modelsMap := config.GetStringMap("models")

	for modelID, modelData := range modelsMap {
		modelMap := modelData.(map[string]interface{})

		var capabilities []string
		if caps, ok := modelMap["capabilities"].([]interface{}); ok {
			for _, cap := range caps {
				capabilities = append(capabilities, cap.(string))
			}
		}

		models[modelID] = ModelInfo{
			Name:          modelMap["name"].(string),
			Description:   modelMap["description"].(string),
			ContextLength: int(modelMap["context_length"].(int64)),
			Capabilities:  capabilities,
		}
	}

	return models
}

package gateway

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RouteDecorator 路由装饰器
type RouteDecorator struct {
	gateway Gateway
	routes  map[string]string
}

// WithRouting 添加路由功能的装饰器
func WithRouting(gateway Gateway, routes map[string]string) Gateway {
	return &RouteDecorator{
		gateway: gateway,
		routes:  routes,
	}
}

// HandleRequest 处理请求并进行路由
func (d *RouteDecorator) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// 获取请求路径
	path := r.URL.Path

	// 检查是否有匹配的路由规则
	for pattern, targetURL := range d.routes {
		if strings.HasPrefix(path, pattern) {
			// 创建新的请求目标
			target, err := url.Parse(targetURL)
			if err != nil {
				http.Error(w, "Internal routing error", http.StatusInternalServerError)
				return
			}

			// 更新请求目标
			r.URL.Path = strings.Replace(path, pattern, "", 1)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}

			// 创建临时的反向代理
			proxy := NewBaseGatewayWithTarget(target)
			proxy.HandleRequest(w, r)
			return
		}
	}

	// 没有匹配的路由规则，使用默认处理
	d.gateway.HandleRequest(w, r)
}

// AuthDecorator 认证装饰器
type AuthDecorator struct {
	gateway        Gateway
	authRoutes     map[string]bool
	authServiceURL string
}

// WithAuth 添加认证功能的装饰器
func WithAuth(gateway Gateway, authRoutes map[string]bool, authServiceURL string) Gateway {
	return &AuthDecorator{
		gateway:        gateway,
		authRoutes:     authRoutes,
		authServiceURL: authServiceURL,
	}
}

// HandleRequest 处理请求并进行认证
func (d *AuthDecorator) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// 获取请求路径
	path := r.URL.Path

	// 检查是否需要认证
	requiresAuth := false
	for pattern, auth := range d.authRoutes {
		if strings.HasPrefix(path, pattern) {
			requiresAuth = auth
			break
		}
	}

	if requiresAuth {
		// 获取认证令牌
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
			return
		}

		// 在实际实现中，这里应该调用认证服务验证令牌
		// 简化起见，这里只检查令牌格式
		if !strings.HasPrefix(token, "Bearer ") {
			http.Error(w, "Unauthorized: Invalid token format", http.StatusUnauthorized)
			return
		}

		// 令牌验证通过，继续处理请求
	}

	// 继续处理请求
	d.gateway.HandleRequest(w, r)
}

// ModelRoutingDecorator 模型路由装饰器
type ModelRoutingDecorator struct {
	gateway      Gateway
	modelWorkers map[string]string
}

// WithModelRouting 添加模型路由功能的装饰器
func WithModelRouting(gateway Gateway, modelWorkers map[string]string) Gateway {
	return &ModelRoutingDecorator{
		gateway:      gateway,
		modelWorkers: modelWorkers,
	}
}

// HandleRequest 处理请求并进行模型路由
func (d *ModelRoutingDecorator) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// 检查请求是否包含模型参数（URL查询参数）
	modelName := r.URL.Query().Get("model")

	// 如果URL中没有模型参数，且为POST请求，尝试从请求体中获取模型信息
	if modelName == "" && r.Method == "POST" &&
		(strings.Contains(r.URL.Path, "/chat/completions") ||
			strings.Contains(r.URL.Path, "/completions")) {

		// 尝试读取请求体以查找模型名称
		var requestData map[string]interface{}

		// 保存请求体内容，因为读取后正文将被消费
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil {
			// 重新设置请求体，使其可以被后续处理
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// 解析JSON请求体
			err = json.Unmarshal(bodyBytes, &requestData)
			if err == nil && requestData["model"] != nil {
				// 从请求体中提取模型名称
				if modelStr, ok := requestData["model"].(string); ok {
					modelName = modelStr
				}
			}

			// 恢复请求体以供后续处理
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	// 根据模型名称选择对应的worker服务
	if modelName != "" && d.modelWorkers[modelName] != "" {
		// 找到对应的模型worker
		workerURL := d.modelWorkers[modelName]
		target, err := url.Parse(workerURL)
		if err != nil {
			http.Error(w, "Internal routing error", http.StatusInternalServerError)
			return
		}

		// 创建临时的反向代理
		proxy := NewBaseGatewayWithTarget(target)
		proxy.HandleRequest(w, r)
		return
	}

	// 没有找到对应的模型worker，使用默认处理
	d.gateway.HandleRequest(w, r)
}

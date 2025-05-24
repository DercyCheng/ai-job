package gateway

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Gateway 定义网关接口
type Gateway interface {
	HandleRequest(w http.ResponseWriter, r *http.Request)
}

// BaseGateway 基础网关实现
type BaseGateway struct {
	proxy *httputil.ReverseProxy
}

// NewBaseGateway 创建基础网关服务(使用默认目标)
func NewBaseGateway() *BaseGateway {
	// 默认转发到MCP服务
	target, _ := url.Parse("http://localhost:8080")
	return NewBaseGatewayWithTarget(target)
}

// NewBaseGatewayWithTarget 创建基础网关服务(指定目标URL)
func NewBaseGatewayWithTarget(target *url.URL) *BaseGateway {
	return &BaseGateway{
		proxy: httputil.NewSingleHostReverseProxy(target),
	}
}

// HandleRequest 处理网关请求
func (g *BaseGateway) HandleRequest(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r)
}

// loggingDecorator 日志装饰器
type loggingDecorator struct {
	gateway Gateway
}

// WithLogging 添加日志功能的装饰器
func WithLogging(gateway Gateway) Gateway {
	return &loggingDecorator{gateway: gateway}
}

func (d *loggingDecorator) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// 记录请求信息
	log.Printf("Incoming request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// 调用实际处理
	d.gateway.HandleRequest(w, r)

	// 记录响应信息
	log.Printf("Completed request: %s %s", r.Method, r.URL.Path)
}

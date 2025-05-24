package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-gatway/internal/gateway"
	"ai-gatway/pkg/utils"

	"github.com/hashicorp/consul/api"
)

// registerService 注册服务到Consul
func registerService(consul *api.Client, serviceID string, port int) error {
	host, _, serviceName, checkURL, tags := utils.GetConsulConfig() // Removed unused consulPort

	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Port:    port,
		Address: host, // Assuming host is the address of *this* service
		Tags:    tags,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d%s", host, port, checkURL),
			Interval: "10s",
			Timeout:  "5s",
		},
	}

	return consul.Agent().ServiceRegister(registration)
}

func main() {
	// 初始化Consul客户端
	consulHost, consulPortVal, _, _, _ := utils.GetConsulConfig()
	consulConfig := api.DefaultConfig()
	consulConfig.Address = fmt.Sprintf("%s:%d", consulHost, consulPortVal)
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Printf("Failed to initialize Consul client: %v", err)
	}

	// 获取网关配置
	port, _, targetURL, routes := utils.GetGatewayConfig() // Removed unused logLevel
	// Get Auth service configuration for the auth decorator
	authServicePort, _, _, _ := utils.GetAuthConfig()
	authServiceURL := fmt.Sprintf("http://localhost:%d", authServicePort) // Assuming auth service is on localhost

	// 注册服务到Consul
	serviceID := fmt.Sprintf("gateway-%d", port)
	if consulClient != nil {
		if err := registerService(consulClient, serviceID, port); err != nil {
			log.Printf("Failed to register service with Consul: %v", err)
		} else {
			log.Printf("Successfully registered service %s with Consul", serviceID)
		}
	} else {
		log.Printf("Skipping Consul registration as client failed to initialize.")
	}

	// 创建目标URL
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}

	// 创建反向代理 (Base Gateway)
	baseProxy := gateway.NewBaseGatewayWithTarget(target)

	// 设置路由
	for _, route := range routes {
		var currentGateway gateway.Gateway = baseProxy

		// Wrap with Auth decorator if required
		if route.AuthRequired {
			// For WithAuth, authRoutes is map[string]bool. We apply it per specific path.
			authMap := make(map[string]bool)
			authMap[route.Path] = true // Or use a more specific sub-path if needed
			currentGateway = gateway.WithAuth(currentGateway, authMap, authServiceURL)
		}

		// Wrap with Logging decorator
		loggedGateway := gateway.WithLogging(currentGateway)

		// http.Handle expects an http.Handler. We adapt our gateway.Gateway.
		http.Handle(route.Path, http.HandlerFunc(loggedGateway.HandleRequest))
	}

	// 添加健康检查端点
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 启动服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting gateway server on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start gateway server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gateway server...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Gateway server forced to shutdown: %v", err)
	}

	// Deregister from Consul
	if consulClient != nil {
		log.Printf("Deregistering service %s from Consul", serviceID)
		if err := consulClient.Agent().ServiceDeregister(serviceID); err != nil {
			log.Printf("Failed to deregister service %s from Consul: %v", serviceID, err)
		} else {
			log.Printf("Successfully deregistered service %s from Consul", serviceID)
		}
	}

	log.Println("Gateway server exiting.")
}

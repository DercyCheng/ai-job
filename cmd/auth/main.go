package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"ai-gatway/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
)

// TokenRequest 表示一个令牌请求
type TokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TokenResponse 表示一个令牌响应
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// 简单用户数据库
var users = map[string]string{
	"admin": "admin123",
	"user1": "password1",
}

func main() {
	// 加载配置
	port, logLevel, jwtSecret, tokenExpiry := utils.GetAuthConfig()

	// 设置路由
	http.HandleFunc("/auth/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 验证用户名和密码
		password, ok := users[req.Username]
		if !ok || password != req.Password {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// 创建JWT令牌
		expiresAt := time.Now().Add(time.Duration(tokenExpiry) * time.Second)
		claims := jwt.MapClaims{
			"sub": req.Username,
			"exp": expiresAt.Unix(),
			"iat": time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		// 返回令牌
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			Token:     tokenString,
			ExpiresAt: expiresAt.Unix(),
		})
	})

	http.HandleFunc("/auth/validate", func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取令牌
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[7:]

		// 解析和验证令牌
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// 令牌有效
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"valid": true})
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// 启动服务
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Auth Service starting on %s with log level %s...\n", addr, logLevel)
	log.Fatal(http.ListenAndServe(addr, nil))
}

#!/bin/bash

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 保存进程ID
MCP_PID=""
GATEWAY_PID=""
AUTH_PID=""
MODEL_PID=""

# 清理函数，确保在脚本退出时停止所有服务
cleanup() {
    echo -e "\n${YELLOW}停止所有服务...${NC}"
    [ ! -z "$GATEWAY_PID" ] && kill $GATEWAY_PID
    [ ! -z "$MCP_PID" ] && kill $MCP_PID
    [ ! -z "$AUTH_PID" ] && kill $AUTH_PID
    [ ! -z "$DEEPSEEK_PID" ] && kill $DEEPSEEK_PID
    [ ! -z "$QWEN_PID" ] && kill $QWEN_PID
    echo -e "${GREEN}所有服务已停止${NC}"
    exit
}

# 设置信号处理器
trap cleanup SIGINT SIGTERM

# 启动服务
start_services() {
    echo -e "${YELLOW}启动所有服务...${NC}"
    
    # 启动DeepSeek模型服务
    cd model-worker && python app.py &
    DEEPSEEK_PID=$!
    echo -e "${GREEN}Started DeepSeek Model service (PID: $DEEPSEEK_PID)${NC}"
    
    # 启动Qwen3模型服务
    python qwen_app.py &
    QWEN_PID=$!
    cd ..
    echo -e "${GREEN}Started Qwen3 Model service (PID: $QWEN_PID)${NC}"
    
    # 启动认证服务
    go run cmd/auth/main.go &
    AUTH_PID=$!
    echo -e "${GREEN}Started Auth service (PID: $AUTH_PID)${NC}"
    
    # 启动MCP服务
    go run cmd/mcp/main.go &
    MCP_PID=$!
    echo -e "${GREEN}Started MCP service (PID: $MCP_PID)${NC}"
    
    # 启动API网关
    go run cmd/gateway/main.go &
    GATEWAY_PID=$!
    echo -e "${GREEN}Started API Gateway (PID: $GATEWAY_PID)${NC}"
    
    # 等待服务启动
    echo -e "${YELLOW}等待服务启动...${NC}"
    sleep 10
}

# 检查服务状态
check_service() {
    local service_name=$1
    local url=$2
    local expected_status=$3
    
    echo -e "${YELLOW}检查${service_name}服务...${NC}"
    
    # 尝试连接服务
    local status_code=$(curl -s -o /dev/null -w "%{http_code}" $url)
    
    if [ "$status_code" == "$expected_status" ]; then
        echo -e "${GREEN}✓ ${service_name}服务正常 (状态码: ${status_code})${NC}"
        return 0
    else
        echo -e "${RED}✗ ${service_name}服务异常 (状态码: ${status_code}, 期望: ${expected_status})${NC}"
        return 1
    fi
}

# 测试认证服务
test_auth() {
    echo -e "\n${YELLOW}测试认证服务...${NC}"
    
    # 获取认证令牌
    echo -e "${YELLOW}获取认证令牌...${NC}"
    local token_response=$(curl -s -X POST http://localhost:8082/auth/token \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}')
    
    # 检查是否成功获取令牌
    if echo "$token_response" | grep -q "token"; then
        local token=$(echo "$token_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
        echo -e "${GREEN}✓ 成功获取认证令牌${NC}"
        echo "$token"
        return 0
    else
        echo -e "${RED}✗ 获取认证令牌失败: $token_response${NC}"
        return 1
    fi
}

# 测试模型列表
test_models() {
    local token=$1
    echo -e "\n${YELLOW}测试获取模型列表...${NC}"
    
    local models_response=$(curl -s http://localhost:8081/v1/models \
        -H "Authorization: Bearer $token")
    
    # 检查响应是否包含DeepSeek模型信息
    if echo "$models_response" | grep -q "deepseek-v3-7b"; then
        echo -e "${GREEN}✓ 成功获取DeepSeek模型信息${NC}"
        echo "$models_response" | grep -o '"name":"DeepSeek V3 7B"'
    else
        echo -e "${RED}✗ 获取DeepSeek模型信息失败${NC}"
        return 1
    fi
    
    # 检查响应是否包含Qwen3模型信息
    if echo "$models_response" | grep -q "qwen3-7b"; then
        echo -e "${GREEN}✓ 成功获取Qwen3模型信息${NC}"
        echo "$models_response" | grep -o '"name":"Qwen3 7B"'
        return 0
    else
        echo -e "${RED}✗ 获取Qwen3模型信息失败${NC}"
        return 1
    fi
}

# 测试聊天完成
test_chat() {
    local token=$1
    echo -e "\n${YELLOW}测试DeepSeek模型聊天完成API...${NC}"
    
    local chat_response=$(curl -s -X POST http://localhost:8081/v1/chat/completions \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d '{
            "model": "deepseek-v3-7b",
            "messages": [
                {"role": "system", "content": "你是一个有帮助的AI助手。"},
                {"role": "user", "content": "你好，请简单介绍一下你自己。"}
            ],
            "temperature": 0.7
        }')
    
    # 检查响应是否包含生成的文本
    if echo "$chat_response" | grep -q "content"; then
        echo -e "${GREEN}✓ DeepSeek模型聊天完成API测试成功${NC}"
        echo "$chat_response" | grep -o '"content":"[^"]*' | head -1
    else
        echo -e "${RED}✗ DeepSeek模型聊天完成API测试失败: $chat_response${NC}"
        return 1
    fi
    
    echo -e "\n${YELLOW}测试Qwen3模型聊天完成API...${NC}"
    
    local qwen_response=$(curl -s -X POST http://localhost:8081/v1/chat/completions \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d '{
            "model": "qwen3-7b",
            "messages": [
                {"role": "system", "content": "你是一个有帮助的AI助手。"},
                {"role": "user", "content": "你好，请简单介绍一下你自己。"}
            ],
            "temperature": 0.7
        }')
    
    # 检查响应是否包含生成的文本
    if echo "$qwen_response" | grep -q "content"; then
        echo -e "${GREEN}✓ Qwen3模型聊天完成API测试成功${NC}"
        echo "$qwen_response" | grep -o '"content":"[^"]*' | head -1
        return 0
    else
        echo -e "${RED}✗ Qwen3模型聊天完成API测试失败: $qwen_response${NC}"
        return 1
    fi
}

# 测试流式聊天完成
test_streaming_chat() {
    local token=$1
    echo -e "\n${YELLOW}测试DeepSeek模型流式聊天API...${NC}"
    
    # 使用curl进行流式请求 (需要安装jq来解析JSON)
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}需要安装jq工具来解析流式响应${NC}"
        echo -e "${YELLOW}可以使用以下命令安装: brew install jq (macOS) 或 apt-get install jq (Linux)${NC}"
        return 1
    fi
    
    # 测试DeepSeek模型的流式响应
    echo -e "${YELLOW}开始测试DeepSeek模型的流式响应...${NC}"
    curl -s -N -X POST http://localhost:8081/v1/chat/completions \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d '{
            "model": "deepseek-v3-7b",
            "messages": [
                {"role": "system", "content": "你是一个有帮助的AI助手。"},
                {"role": "user", "content": "用几个词语简单介绍下大模型技术"}
            ],
            "temperature": 0.7,
            "stream": true
        }' | head -n 10
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ DeepSeek模型流式聊天API测试成功${NC}"
    else
        echo -e "${RED}✗ DeepSeek模型流式聊天API测试失败${NC}"
        return 1
    fi
    
    # 测试Qwen3模型的流式响应
    echo -e "\n${YELLOW}测试Qwen3模型流式聊天API...${NC}"
    curl -s -N -X POST http://localhost:8081/v1/chat/completions \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d '{
            "model": "qwen3-7b",
            "messages": [
                {"role": "system", "content": "你是一个有帮助的AI助手。"},
                {"role": "user", "content": "用几个词语简单介绍下大模型技术"}
            ],
            "temperature": 0.7,
            "stream": true
        }' | head -n 10
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Qwen3模型流式聊天API测试成功${NC}"
        return 0
    else
        echo -e "${RED}✗ Qwen3模型流式聊天API测试失败${NC}"
        return 1
    fi
}

# 主测试流程
main() {
    echo -e "${YELLOW}开始测试AI Gateway系统...${NC}"
    
    # 启动所有服务
    start_services
    
    # 检查各服务状态
    check_service "Auth" "http://localhost:8082/health" "200" || { cleanup; return 1; }
    check_service "MCP" "http://localhost:8080/health" "200" || { cleanup; return 1; }
    check_service "Gateway" "http://localhost:8081/health" "200" || { cleanup; return 1; }
    check_service "DeepSeek Worker" "http://localhost:5000/health" "200" || { cleanup; return 1; }
    check_service "Qwen Worker" "http://localhost:5001/health" "200" || { cleanup; return 1; }
    
    # 测试认证服务
    token=$(test_auth)
    if [ $? -ne 0 ]; then
        cleanup
        return 1
    fi
    
    # 测试模型列表
    test_models "$token" || { cleanup; return 1; }
    
    # 测试聊天完成
    test_chat "$token" || { cleanup; return 1; }
    
    # 测试流式聊天完成
    test_streaming_chat "$token" || { cleanup; return 1; }
    
    echo -e "\n${GREEN}所有测试通过!${NC}"
    
    # 清理
    cleanup
}

# 执行主测试流程
main
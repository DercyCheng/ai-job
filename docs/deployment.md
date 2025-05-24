# 部署指南

本文档提供了分布式 AI 网关 MCP Server 的部署流程和配置说明。

## 系统要求

### 硬件推荐配置

**生产环境**:
- **API Gateway & MCP Service**:
  - CPU: 8+ 核
  - 内存: 16+ GB
  - 存储: 100+ GB SSD
- **Model Workers**:
  - CPU: 16+ 核
  - 内存: 64+ GB
  - GPU: NVIDIA A100/H100 (推荐)或同等性能
  - 存储: 200+ GB SSD
- **本地大型语言模型部署**:
  - CPU: 32+ 核
  - 内存: 128+ GB
  - GPU: NVIDIA A100 80GB (推荐) 或 4 x NVIDIA A10/RTX 4090
  - 存储: 500+ GB NVMe SSD
  - 网络: 10Gbps+

**开发/测试环境**:
- CPU: 4+ 核
- 内存: 8+ GB
- 存储: 50+ GB SSD
- 本地模型测试: 至少 1 x NVIDIA GPU (16GB+ 显存)

### 软件要求

- Docker 20.10+
- Docker Compose 2.0+
- Kubernetes 1.22+ (生产环境)
- Go 1.21+
- Python 3.10+
- MySQL 8.0+
- Redis 7.0+
- NVIDIA 驱动 525+ 和 CUDA 12.1+
- NVIDIA Container Toolkit (nvidia-docker2)

## 部署方式

### 1. Docker Compose 部署 (开发/测试环境)

#### 步骤 1: 克隆代码库

```bash
git clone https://github.com/example/ai-gateway.git
cd ai-gateway
```

#### 步骤 2: 环境配置

创建并编辑 `.env` 文件:

```bash
cp .env.example .env
```

根据需要修改环境变量。

#### 步骤 3: 启动服务

```bash
docker-compose up -d
```

### 2. Kubernetes 部署 (生产环境)

#### 步骤 1: 准备 Kubernetes 集群

确保你有一个可用的 Kubernetes 集群，并配置了 `kubectl`。

#### 步骤 2: 配置 Kubernetes Secret

```bash
kubectl create namespace ai-gateway

# 创建包含配置信息的 Secret
kubectl create secret generic ai-gateway-config \
  --from-file=config.yaml=./k8s/config.yaml \
  --namespace ai-gateway
```

#### 步骤 3: 部署服务

```bash
kubectl apply -f k8s/manifests/ --namespace ai-gateway
```

## 配置说明

### 主配置文件 (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  request_timeout: 30s
  max_request_size: "10MB"

database:
  driver: "mysql"
  host: "mysql"
  port: 3306
  user: "aigateway"
  password: "${DB_PASSWORD}"
  database: "aigateway"
  max_idle_connections: 10
  max_open_connections: 100
  connection_lifetime: "5m"
  
redis:
  host: "redis"
  port: 6379
  password: "${REDIS_PASSWORD}"
  db: 0
  
auth:
  jwt_secret: "${JWT_SECRET}"
  token_expiry: "24h"
  refresh_token_expiry: "168h"
  
models:
  providers:
    - name: "openai"
      enabled: true
      api_key: "${OPENAI_API_KEY}"
      base_url: "https://api.openai.com/v1"
      models:
        - "gpt-4"
        - "gpt-3.5-turbo"
    - name: "anthropic"
      enabled: true
      api_key: "${ANTHROPIC_API_KEY}"
      base_url: "https://api.anthropic.com"
      models:
        - "claude-3-opus"
        - "claude-3-sonnet"
        
workers:
  min_count: 5
  max_count: 20
  queue_size: 1000
  autoscaling:
    enabled: true
    scale_up_threshold: 70
    scale_down_threshold: 30
    scale_up_factor: 2
    scale_down_factor: 1
    cooldown_period: "5m"
    
logging:
  level: "info"
  format: "json"
  output: "stdout"
  file_path: "/var/log/ai-gateway.log"
  
monitoring:
  metrics:
    enabled: true
    port: 9090
  tracing:
    enabled: true
    provider: "jaeger"
    endpoint: "http://jaeger:14268/api/traces"
```

### 环境变量

创建 `.env` 文件，包含以下环境变量:

```
# 数据库配置
DB_PASSWORD=your_strong_password

# Redis配置
REDIS_PASSWORD=your_redis_password

# 认证配置
JWT_SECRET=your_jwt_secret_key

# 模型提供商 API 密钥
OPENAI_API_KEY=your_openai_api_key
ANTHROPIC_API_KEY=your_anthropic_api_key
```

## 扩展与伸缩

### 水平扩展

1. API Gateway & MCP Service:

```bash
# Docker Compose
docker-compose up -d --scale api-gateway=3 --scale mcp-service=5

# Kubernetes
kubectl scale deployment api-gateway --replicas=3 -n ai-gateway
kubectl scale deployment mcp-service --replicas=5 -n ai-gateway
```

2. Model Workers:

Model Workers 支持自动伸缩，可通过 `workers.autoscaling` 配置调整。

### 负载均衡

生产环境中，推荐使用:
- Kubernetes Ingress 或 Istio 进行流量管理
- 外部负载均衡器 (如 Nginx, HAProxy, 或云服务提供商的负载均衡服务)

## 监控与日志

### Prometheus & Grafana

系统暴露 Prometheus 指标，可通过以下方式访问:
- `http://<host>:9090/metrics`

### 日志收集

推荐使用 ELK Stack 或 Loki 进行日志收集与分析。

### 追踪

系统集成了 Jaeger 分布式追踪，可以跟踪请求在各个服务间的流转。

## 备份与恢复

### 数据库备份

```bash
# 创建备份
kubectl exec -it $(kubectl get pods -l app=mysql -n ai-gateway -o jsonpath="{.items[0].metadata.name}") -n ai-gateway -- \
  mysqldump -u root -p aigateway > backup_$(date +%Y%m%d).sql

# 恢复备份
kubectl exec -it $(kubectl get pods -l app=mysql -n ai-gateway -o jsonpath="{.items[0].metadata.name}") -n ai-gateway -- \
  mysql -u root -p aigateway < backup_20230101.sql
```

## 故障排查

### 常见问题

1. **服务无法启动**
   - 检查配置文件格式是否正确
   - 确认环境变量已正确设置
   - 检查数据库连接信息

2. **请求超时**
   - 检查模型服务是否正常运行
   - 检查网络连接
   - 调整 `server.request_timeout` 配置

3. **性能问题**
   - 增加服务实例数
   - 优化数据库查询
   - 调整缓存策略

### 获取帮助

- 查看 [GitHub Issues](https://github.com/example/ai-gateway/issues)
- 参与 [社区讨论](https://github.com/example/ai-gateway/discussions)
- 联系 [技术支持](mailto:support@example.com)

## 本地模型部署指南

本章节提供了如何部署和配置本地大型语言模型（如 DeepSeek V3 和 Qwen3）的详细说明。

### 硬件规划

根据模型大小，需要不同的硬件配置：

| 模型 | 参数量 | 最小GPU要求 | 推荐GPU配置 | 最小显存(INT8量化) |
|------|---------|-------------|--------------|---------------------|
| DeepSeek V3 7B | 7B | 1 x NVIDIA A10 | 1 x NVIDIA A100 | 16GB |
| DeepSeek V3 14B | 14B | 1 x NVIDIA A100 | 1 x NVIDIA A100 | 28GB |
| DeepSeek V3 42B | 42B | 2 x NVIDIA A100 | 4 x NVIDIA A100 | 80GB |
| Qwen3 0.5B | 0.5B | 1 x NVIDIA T4 | 1 x NVIDIA A10 | 4GB |
| Qwen3 1.5B | 1.5B | 1 x NVIDIA T4 | 1 x NVIDIA A10 | 8GB |
| Qwen3 7B | 7B | 1 x NVIDIA A10 | 1 x NVIDIA A100 | 16GB |
| Qwen3 14B | 14B | 1 x NVIDIA A100 | 1 x NVIDIA A100 | 28GB |
| Qwen3 72B | 72B | 2 x NVIDIA A100 | 4 x NVIDIA A100 | 80GB+ |

### 模型工作节点准备

#### 1. 安装NVIDIA驱动和CUDA

```bash
# 安装NVIDIA驱动
sudo apt-get update
sudo apt-get install -y nvidia-driver-525

# 安装CUDA Toolkit
wget https://developer.download.nvidia.com/compute/cuda/12.1.0/local_installers/cuda_12.1.0_530.30.02_linux.run
sudo sh cuda_12.1.0_530.30.02_linux.run

# 验证安装
nvidia-smi
nvcc --version
```

#### 2. 安装NVIDIA Container Toolkit

```bash
# 添加NVIDIA源
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | sudo tee /etc/apt/sources.list.d/nvidia-docker.list

# 安装nvidia-docker2
sudo apt-get update
sudo apt-get install -y nvidia-docker2

# 重启Docker
sudo systemctl restart docker
```

### 本地模型部署 - Docker方式

#### 步骤 1: 创建模型部署配置

创建 `model-deploy-config.yaml` 文件:

```yaml
models:
  - name: "deepseek-v3-14b"
    provider: "deepseek"
    version: "3.0"
    size: "14b"
    model_path: "/models/deepseek-v3-14b"
    quantization: "int8"  # 可选: none, int8, int4
    tensor_parallel: 1   # GPU并行数量
    max_batch_size: 8
    max_input_length: 4096
    max_output_length: 4096
    use_flash_attention: true
    runtime_options:
      kv_cache_fp16: true
      trust_remote_code: true
      cpu_offload: false
  
  - name: "qwen3-7b"
    provider: "qwen"
    version: "3.0"
    size: "7b"
    model_path: "/models/qwen3-7b"
    quantization: "int8"
    tensor_parallel: 1
    max_batch_size: 12
    max_input_length: 4096
    max_output_length: 2048
    use_flash_attention: true
    runtime_options:
      kv_cache_fp16: true
      trust_remote_code: true
      cpu_offload: false

runtime:
  device_map: "auto"
  compute_precision: "float16"  # 可选: float32, float16, bfloat16
  enable_cuda_graph: true
  max_memory: null  # 自动设置或指定如 {"cuda:0": "24GiB"}
  cache_dir: "/cache"
  log_level: "info"
```

#### 步骤 2: 部署模型工作节点

使用Docker Compose:

```yaml
# docker-compose.model-worker.yml
version: '3.8'

services:
  model-worker:
    image: aigateway/model-worker:latest
    runtime: nvidia
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: all
              capabilities: [gpu]
    volumes:
      - ./model-deploy-config.yaml:/app/config/model-deploy-config.yaml
      - ./models:/models
      - ./cache:/cache
    environment:
      - NVIDIA_VISIBLE_DEVICES=all
      - WORKER_ID=worker-01
      - MCP_SERVICE_URL=http://mcp-service:8080
      - LOG_LEVEL=info
      - MAX_CONCURRENT_REQUESTS=8
    ports:
      - "8090:8090"  # 模型工作节点API端口
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

启动服务:

```bash
docker-compose -f docker-compose.model-worker.yml up -d
```

### 本地模型部署 - Kubernetes方式

#### 步骤 1: 创建ConfigMap和PersistentVolumeClaim

```yaml
# model-worker-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: model-deploy-config
  namespace: ai-gateway
data:
  model-deploy-config.yaml: |
    models:
      - name: "deepseek-v3-14b"
        provider: "deepseek"
        version: "3.0"
        size: "14b"
        model_path: "/models/deepseek-v3-14b"
        quantization: "int8"
        tensor_parallel: 1
        max_batch_size: 8
        use_flash_attention: true
        # 其他配置...
      
      - name: "qwen3-7b"
        provider: "qwen"
        version: "3.0"
        size: "7b"
        model_path: "/models/qwen3-7b"
        quantization: "int8"
        tensor_parallel: 1
        max_batch_size: 12
        use_flash_attention: true
        # 其他配置...
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: models-storage
  namespace: ai-gateway
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 500Gi
  storageClassName: local-ssd  # 使用高性能存储
```

应用配置:

```bash
kubectl apply -f model-worker-config.yaml
```

#### 步骤 2: 部署模型工作节点

```yaml
# model-worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: model-worker
  namespace: ai-gateway
spec:
  replicas: 1  # 根据GPU节点数量调整
  selector:
    matchLabels:
      app: model-worker
  template:
    metadata:
      labels:
        app: model-worker
    spec:
      nodeSelector:
        gpu: "true"  # 标记有GPU的节点
      containers:
      - name: model-worker
        image: aigateway/model-worker:latest
        resources:
          limits:
            nvidia.com/gpu: 1  # 需要的GPU数量
            memory: "120Gi"
            cpu: "16"
          requests:
            nvidia.com/gpu: 1
            memory: "100Gi"
            cpu: "8"
        env:
        - name: WORKER_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: MCP_SERVICE_URL
          value: "http://mcp-service:8080"
        - name: LOG_LEVEL
          value: "info"
        - name: MAX_CONCURRENT_REQUESTS
          value: "8"
        volumeMounts:
        - name: config-volume
          mountPath: /app/config
        - name: models-storage
          mountPath: /models
        - name: cache-storage
          mountPath: /cache
        ports:
        - containerPort: 8090
        readinessProbe:
          httpGet:
            path: /health
            port: 8090
          initialDelaySeconds: 120
          periodSeconds: 30
      volumes:
      - name: config-volume
        configMap:
          name: model-deploy-config
      - name: models-storage
        persistentVolumeClaim:
          claimName: models-storage
      - name: cache-storage
        emptyDir: {}
```

应用部署:

```bash
kubectl apply -f model-worker-deployment.yaml
```

### 模型权重下载与准备

#### DeepSeek V3 模型

可以从Hugging Face下载DeepSeek V3模型:

```bash
# 安装Hugging Face CLI
pip install huggingface_hub

# 下载DeepSeek V3 14B模型
mkdir -p /models/deepseek-v3-14b
huggingface-cli download deepseek-ai/deepseek-v3-14b-base --local-dir /models/deepseek-v3-14b
```

#### Qwen3 模型

可以从Hugging Face下载Qwen3模型:

```bash
# 下载Qwen3 7B模型
mkdir -p /models/qwen3-7b
huggingface-cli download Qwen/Qwen3-7B-base --local-dir /models/qwen3-7b
```

#### 模型量化(可选)

使用GPTQ或AWQ进行量化可以显著减少显存需求:

```bash
# 使用AutoGPTQ量化DeepSeek V3模型
pip install auto-gptq optimum

# 量化示例
python -m auto_gptq.quantize \
  --model_name_or_path /models/deepseek-v3-14b \
  --output_dir /models/deepseek-v3-14b-int8 \
  --bits 8 \
  --group_size 128 \
  --desc_act
```

### 性能优化与监控

1. **性能监控**

```bash
# 安装监控工具
pip install py3nvml nvidia-ml-py

# 添加到配置中
monitoring:
  enabled: true
  metrics_port: 8091
  export_gpu_metrics: true
  profiling: false  # 生产环境建议关闭
```

2. **高级配置**

针对大型模型的优化选项:

```yaml
# 添加到模型配置中
advanced_options:
  kernel_injection: true           # 使用优化的CUDA内核
  sequence_parallelism: true       # 序列并行
  paged_attention: true            # 分页注意力机制
  sdp_backend: "flash_attention"   # 使用Flash Attention
  rope_scaling:                    # RoPE扩展
    type: "dynamic"
    factor: 2.0
```

### 双协议支持配置

配置同时支持HTTP和gRPC协议:

```yaml
# 在主配置文件中添加
server:
  http:
    enabled: true
    port: 8080
    max_request_size: "10MB"
  grpc:
    enabled: true
    port: 50051
    max_message_size: "10MB"
  protocol_buffers:
    proto_path: "./proto"
```

配置gRPC TLS加密:

```yaml
# 在主配置文件中添加
grpc_tls:
  enabled: true
  cert_file: "/certs/server.crt"
  key_file: "/certs/server.key"
  ca_file: "/certs/ca.crt"  # 如果需要客户端验证
```

### 本地模型测试

部署完成后，可以使用以下命令测试本地模型:

```bash
# 使用HTTP API测试
curl -X POST http://localhost:8080/models/deepseek-v3-14b/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "messages": [
      {"role": "user", "content": "请介绍一下你自己"}
    ],
    "max_tokens": 100,
    "temperature": 0.7
  }'

# 使用gRPC测试工具(grpcurl)
grpcurl -plaintext -d '{"model_id": "deepseek-v3-14b", "messages": [{"role": "user", "content": "请介绍一下你自己"}], "max_tokens": 100}' localhost:50051 model.ModelService/CreateChatCompletion
```

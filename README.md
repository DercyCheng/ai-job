# AI Job Scheduler

[![Go Report Card](https://goreportcard.com/badge/github.com/yourorg/ai-job)](https://goreportcard.com/report/github.com/yourorg/ai-job)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

AI Job Scheduler是一个分布式任务调度系统，专为AI/ML工作负载设计。它提供了：

- 任务调度和执行
- 资源管理
- 分布式工作节点
- 全面的监控和日志系统
- 可扩展的架构

## 功能特性

- **任务调度**：支持定时任务和即时任务
- **工作节点**：自动扩展的工作节点池
- **监控**：集成Prometheus和Grafana
- **日志**：集中式日志收集(Loki)
- **API**：RESTful API管理接口

## 快速开始

### 前提条件

- Docker 20.10+
- Docker Compose 1.29+
- Go 1.20+ (仅开发需要)

### 启动服务

```bash
# 克隆项目
git clone https://github.com/yourorg/ai-job.git
cd ai-job

# 启动服务
docker-compose up -d
```

服务启动后可以访问：

- API: http://localhost:8080
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090

## 配置

编辑 `config/config.yaml`文件配置系统参数：

```yaml
server:
  address: ":8080"
  timeout: 30s

database:
  host: "postgres"
  port: 5432
  user: "postgres"
  password: "password"
  name: "ai_job"

logging:
  level: "info"
  format: "json"
  outputs: ["stdout", "file"]
```

## 许可证

MIT License

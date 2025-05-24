# 开发指南

本文档提供 AI 网关 MCP Server 项目的开发指南，包括环境搭建、代码规范、架构说明以及功能扩展指导。

## 开发环境设置

### 前提条件

- Go 1.21+
- Python 3.10+
- Node.js 18+
- Docker & Docker Compose
- MySQL 8.0+
- Redis 7.0+
- IDE 推荐: Visual Studio Code 或 GoLand

### 本地开发环境设置

1. **克隆代码库**

```bash
git clone https://github.com/example/ai-gateway.git
cd ai-gateway
```

2. **设置开发环境**

```bash
# 安装后端依赖
go mod download

# 安装前端依赖
cd web
npm install
cd ..

# 启动开发数据库和 Redis
docker-compose -f docker-compose.dev.yml up -d
```

3. **配置开发环境变量**

创建 `.env.dev` 文件:

```bash
cp .env.example .env.dev
```

编辑 `.env.dev` 文件，配置必要的环境变量。

4. **运行开发服务器**

```bash
# 后端 API 服务 (Go)
go run cmd/api/main.go --config=configs/dev.yaml

# MCP 服务 (Go)
go run cmd/mcp/main.go --config=configs/dev.yaml

# 模型 Worker (Python)
cd worker
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install -r requirements.txt
python worker.py --config=../configs/worker.dev.yaml

# 前端开发服务器 (Vue3 + Vite)
cd web
npm run dev
```

## 项目结构

```
ai-gateway/
├── api/              # API 定义
│   ├── proto/        # Protocol Buffer 定义
│   └── openapi/      # OpenAPI 规范
├── cmd/              # 命令行入口
│   ├── api/          # API 服务入口
│   ├── mcp/          # MCP 服务入口
│   └── admin/        # 管理工具入口
├── configs/          # 配置文件
├── docs/             # 文档
├── internal/         # 内部包
│   ├── auth/         # 认证相关
│   ├── db/           # 数据库访问
│   ├── models/       # 数据模型
│   ├── handlers/     # API 处理器
│   ├── mcp/          # MCP 服务实现
│   └── utils/        # 工具函数
├── pkg/              # 可重用的公共包
│   ├── logger/       # 日志工具
│   ├── metrics/      # 指标收集
│   └── queue/        # 队列实现
├── scripts/          # 脚本工具
├── web/              # 前端代码 (Vue3 + TypeScript)
│   ├── src/          # 源代码
│   ├── public/       # 静态资源
│   └── vite.config.ts # Vite 配置
├── worker/           # 模型 Worker 实现 (Python)
├── go.mod            # Go 依赖管理
├── go.sum            # Go 依赖校验和
├── .env.example      # 环境变量示例
└── docker-compose.yml # Docker 编排配置
```

## 代码规范

### Go 代码规范

- 遵循 [Effective Go](https://golang.org/doc/effective_go)
- 使用 `gofmt` 或 `goimports` 进行代码格式化
- 使用 `golint` 和 `staticcheck` 进行代码检查
- 注释格式应符合 `godoc` 规范

### Python 代码规范

- 遵循 [PEP 8](https://www.python.org/dev/peps/pep-0008/)
- 使用 `black` 进行代码格式化
- 使用 `pylint` 或 `flake8` 进行代码检查
- 使用类型注解 (Type Hints)

### TypeScript/Vue 代码规范

- 遵循 Vue 官方风格指南
- 使用 ESLint 和 Prettier 进行代码格式化和检查
- 使用类型声明，避免 `any` 类型

## 开发流程

### Git 工作流

- 主分支: `main` - 稳定发布版本
- 开发分支: `develop` - 开发中的功能
- 功能分支: `feature/xxx` - 新功能开发
- 修复分支: `bugfix/xxx` - 问题修复
- 发布分支: `release/vX.Y.Z` - 版本发布准备

### 提交规范

使用语义化提交消息:

```
<type>(<scope>): <subject>

<body>

<footer>
```

类型 (`type`):
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档变更
- `style`: 代码风格变更 (格式化等)
- `refactor`: 代码重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建过程或辅助工具变更

### 代码审查

- 所有代码变更通过 Pull Request 提交
- 至少需要一名维护者的审查和批准
- CI 测试必须通过
- 遵循代码审查清单

## 架构说明

### 组件交互

```
┌────────────┐     ┌────────────┐     ┌────────────┐
│            │     │            │     │            │
│  客户端    │────▶│ API Gateway│────▶│ Auth Service│
│            │     │            │     │            │
└────────────┘     └──────┬─────┘     └────────────┘
                          │
                          ▼
                   ┌────────────┐
                   │            │
                   │ MCP Service│
                   │            │
                   └──────┬─────┘
                          │
                          ▼
         ┌────────────────┬────────────────┐
         │                │                │
┌────────▼─────┐  ┌───────▼──────┐  ┌──────▼───────┐
│              │  │              │  │              │
│ Model Worker │  │ Model Worker │  │ Model Worker │
│              │  │              │  │              │
└──────────────┘  └──────────────┘  └──────────────┘
```

### 数据流

1. 客户端发送请求到 API Gateway
2. API Gateway 验证请求并路由到相应服务
3. 如需身份验证，请求会被转发到 Auth Service
4. 验证通过后，请求被转发到 MCP Service
5. MCP Service 处理请求并根据负载分配到合适的 Model Worker
6. Worker 执行模型推理并返回结果
7. 结果通过原路径返回给客户端

## 核心模块开发指南

### API Gateway 开发

API Gateway 负责请求路由、认证、限流等功能。

**关键文件**:
- `cmd/api/main.go`: 入口点
- `internal/handlers/`: 请求处理器
- `internal/middleware/`: 中间件实现

**添加新端点**:

1. 在 `api/openapi/` 中定义新端点规范
2. 在 `internal/handlers/` 中实现处理器
3. 在 `internal/router/` 中注册路由

### MCP Service 开发

MCP Service 实现 Model Context Protocol，管理模型调用和上下文处理。

**关键文件**:
- `cmd/mcp/main.go`: 入口点
- `internal/mcp/`: MCP 实现
- `internal/mcp/protocol.go`: 协议定义

**扩展 MCP 功能**:

1. 在 `internal/mcp/protocol.go` 中定义新消息类型
2. 在 `internal/mcp/handler.go` 中实现处理逻辑
3. 在 `internal/mcp/server.go` 中注册处理器

### Model Worker 开发

Model Worker 负责执行模型推理，是系统的计算核心。

**关键文件**:
- `worker/worker.py`: Worker 入口点
- `worker/models/`: 模型实现
- `worker/utils/`: 工具函数

**添加新模型支持**:

1. 在 `worker/models/` 中实现新模型适配器
2. 在 `worker/models/__init__.py` 中注册模型
3. 更新配置文件以支持新模型

### 前端开发 (Vue3 + TS)

前端提供用户界面，使用 Vue3 和 TypeScript 开发。

**关键文件**:
- `web/src/main.ts`: 入口点
- `web/src/views/`: 页面组件
- `web/src/api/`: API 调用

**添加新页面**:

1. 在 `web/src/views/` 中创建新组件
2. 在 `web/src/router/index.ts` 中添加路由
3. 更新导航菜单

## 测试指南

### 单元测试

- Go: 使用标准 `testing` 包
- Python: 使用 `pytest`
- TypeScript: 使用 `jest` 和 `vue-test-utils`

**运行测试**:

```bash
# Go 测试
go test ./...

# Python 测试
cd worker
pytest

# 前端测试
cd web
npm test
```

### 集成测试

使用 `docker-compose` 启动完整环境进行集成测试:

```bash
docker-compose -f docker-compose.test.yml up
go test ./... -tags=integration
```

### 负载测试

使用 `k6` 进行负载测试:

```bash
k6 run scripts/load-test.js
```

## 调试技巧

### Go 服务调试

- 使用 [Delve](https://github.com/go-delve/delve) 进行调试
- 启用详细日志: `--log-level=debug`
- 使用 pprof 进行性能分析: `import _ "net/http/pprof"`

### Python Worker 调试

- 使用 `pdb` 或 IDE 调试器
- 启用详细日志: `--log-level=DEBUG`
- 使用 `cProfile` 进行性能分析

## 贡献指南

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 技术参考

- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [Vue3 文档](https://vuejs.org/guide/introduction.html)
- [Vite 文档](https://vitejs.dev/guide/)
- [Anthropic API 文档](https://docs.anthropic.com/)
- [OpenAI API 文档](https://platform.openai.com/docs/)

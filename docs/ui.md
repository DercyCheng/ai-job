# 前端UI开发指南

本文档提供 AI 网关 MCP Server 项目的前端 UI 开发指南，详细说明基于 Vite 构建的 Vue3 前端应用。

## 技术栈概述

AI 网关前端采用以下技术栈:

- **Vue3**: 核心框架，采用 Composition API
- **TypeScript**: 类型安全的 JavaScript 超集
- **Vite**: 下一代前端构建工具，提供极速的开发体验
- **Element Plus**: UI 组件库，提供丰富的组件和样式
- **Pinia**: Vue3 的状态管理方案
- **Vue Router**: 官方路由管理
- **Axios**: HTTP 请求客户端
- **ECharts**: 数据可视化图表库

## Vite 开发环境

### 为什么选择 Vite

Vite 相比传统构建工具有以下优势:

1. **极速的开发服务器启动**: 基于原生 ES 模块，无需打包整个应用
2. **即时的热模块替换 (HMR)**: 修改代码后立即在浏览器中反映，无需刷新页面
3. **按需编译**: 只编译当前页面需要的组件，开发体验流畅
4. **优化的构建**: 使用 Rollup 进行生产构建，产物更小更快
5. **内置支持**: 对 TypeScript、JSX、CSS 预处理器的开箱即用支持

### Vite 配置

项目的 Vite 配置位于 `web/vite.config.ts`，主要配置包括:

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    minify: 'terser',
    sourcemap: false,
  },
})
```

## 项目结构

```
web/
├── public/              # 静态资源
├── src/
│   ├── api/             # API 调用
│   ├── assets/          # 项目资源 (图片、样式等)
│   ├── components/      # 共享组件
│   ├── composables/     # 可复用的组合式函数
│   ├── layouts/         # 布局组件
│   ├── router/          # 路由配置
│   ├── stores/          # Pinia 状态管理
│   ├── types/           # TypeScript 类型定义
│   ├── utils/           # 工具函数
│   ├── views/           # 页面组件
│   ├── App.vue          # 根组件
│   ├── main.ts          # 入口文件
│   ├── env.d.ts         # 环境变量类型声明
│   └── shims-vue.d.ts   # Vue 类型声明
├── .eslintrc.js         # ESLint 配置
├── .prettierrc          # Prettier 配置
├── index.html           # HTML 模板
├── package.json         # 项目依赖
├── tsconfig.json        # TypeScript 配置
└── vite.config.ts       # Vite 配置
```

## 开发指南

### 初始设置

```bash
cd web
npm install
```

### 启动开发服务器

```bash
npm run dev
```

开发服务器默认运行在 `http://localhost:3000`。

### 构建生产版本

```bash
npm run build
```

构建产物位于 `web/dist` 目录。

### 预览生产版本

```bash
npm run preview
```

## 主要功能模块

### 身份验证 (Authentication)

身份验证模块管理用户登录、注册和权限控制。

**关键文件**:

- `src/stores/auth.ts`: 认证状态管理
- `src/api/auth.ts`: 认证相关 API 调用
- `src/views/auth/Login.vue`: 登录页面
- `src/router/guards.ts`: 路由守卫

### 模型管理 (Model Management)

模型管理模块提供 AI 模型的查看、配置和部署。

**关键文件**:

- `src/views/models/ModelList.vue`: 模型列表
- `src/views/models/ModelDetail.vue`: 模型详情
- `src/views/models/ModelDeploy.vue`: 模型部署
- `src/api/models.ts`: 模型相关 API

### 系统监控 (System Monitoring)

系统监控模块展示系统指标、日志和警报。

**关键文件**:

- `src/views/monitor/Dashboard.vue`: 监控仪表盘
- `src/views/monitor/Metrics.vue`: 系统指标
- `src/views/monitor/Logs.vue`: 系统日志
- `src/api/monitor.ts`: 监控相关 API

### 任务管理 (Task Management)

任务管理模块处理 AI 任务的创建、查询和管理。

**关键文件**:

- `src/views/tasks/TaskList.vue`: 任务列表
- `src/views/tasks/TaskDetail.vue`: 任务详情
- `src/views/tasks/TaskCreate.vue`: 创建任务
- `src/api/tasks.ts`: 任务相关 API

## 组件开发规范

### 命名约定

- 组件文件使用 PascalCase 命名，如 `ModelList.vue`
- 组件名称与文件名一致，也使用 PascalCase
- 页面组件放在 `views` 目录，共享组件放在 `components` 目录

### 组件结构

推荐使用 `<script setup>` 语法:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useModelStore } from '@/stores/model'

// 状态和逻辑
const models = ref([])
const loading = ref(false)

// 获取数据
const fetchModels = async () => {
  loading.value = true
  try {
    const modelStore = useModelStore()
    models.value = await modelStore.fetchModels()
  } catch (error) {
    console.error('Failed to fetch models:', error)
  } finally {
    loading.value = false
  }
}

// 生命周期
onMounted(() => {
  fetchModels()
})
</script>

<template>
  <div class="model-list">
    <h1>AI 模型列表</h1>
    <el-table v-loading="loading" :data="models">
      <!-- 表格内容 -->
    </el-table>
  </div>
</template>

<style scoped>
.model-list {
  padding: 20px;
}
</style>
```

## API 调用

使用 Axios 进行 API 调用，推荐创建一个 API 客户端实例:

```typescript
// src/utils/api.ts
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const apiClient = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 添加请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 添加响应拦截器
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default apiClient
```

## 状态管理

使用 Pinia 进行状态管理:

```typescript
// src/stores/model.ts
import { defineStore } from 'pinia'
import apiClient from '@/utils/api'

export const useModelStore = defineStore('model', {
  state: () => ({
    models: [],
    currentModel: null,
    loading: false,
  }),
  
  actions: {
    async fetchModels() {
      this.loading = true
      try {
        const response = await apiClient.get('/models')
        this.models = response.data
        return this.models
      } finally {
        this.loading = false
      }
    },
  
    async getModel(id: string) {
      this.loading = true
      try {
        const response = await apiClient.get(`/models/${id}`)
        this.currentModel = response.data
        return this.currentModel
      } finally {
        this.loading = false
      }
    },
  },
})
```

## 路由配置

使用 Vue Router 配置路由:

```typescript
// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import { setupRouterGuards } from './guards'

// 布局
import DefaultLayout from '@/layouts/DefaultLayout.vue'

// 认证相关页面
import Login from '@/views/auth/Login.vue'
import Register from '@/views/auth/Register.vue'

// 仪表盘
import Dashboard from '@/views/Dashboard.vue'

// 模型管理
import ModelList from '@/views/models/ModelList.vue'
import ModelDetail from '@/views/models/ModelDetail.vue'
import ModelDeploy from '@/views/models/ModelDeploy.vue'

// 任务管理
import TaskList from '@/views/tasks/TaskList.vue'
import TaskDetail from '@/views/tasks/TaskDetail.vue'
import TaskCreate from '@/views/tasks/TaskCreate.vue'

const routes = [
  {
    path: '/auth',
    children: [
      { path: 'login', component: Login },
      { path: 'register', component: Register },
    ],
  },
  {
    path: '/',
    component: DefaultLayout,
    meta: { requiresAuth: true },
    children: [
      { path: '', component: Dashboard },
      { path: 'models', component: ModelList },
      { path: 'models/:id', component: ModelDetail },
      { path: 'models/deploy', component: ModelDeploy },
      { path: 'tasks', component: TaskList },
      { path: 'tasks/:id', component: TaskDetail },
      { path: 'tasks/create', component: TaskCreate },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 设置路由守卫
setupRouterGuards(router)

export default router
```

## 样式指南

项目使用 Element Plus 作为 UI 组件库，配合自定义样式:

```scss
// src/assets/styles/variables.scss

// 主题色
$primary-color: #3498db;
$secondary-color: #2c3e50;
$success-color: #2ecc71;
$warning-color: #f39c12;
$danger-color: #e74c3c;
$info-color: #1abc9c;

// 文本颜色
$text-primary: #333333;
$text-secondary: #666666;
$text-muted: #999999;

// 背景色
$bg-primary: #ffffff;
$bg-secondary: #f5f7fa;
$bg-dark: #1d2025;

// 间距
$spacing-xs: 4px;
$spacing-sm: 8px;
$spacing-md: 16px;
$spacing-lg: 24px;
$spacing-xl: 32px;

// 字体
$font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial, sans-serif;
$font-size-base: 14px;
$font-size-small: 12px;
$font-size-large: 16px;
$font-size-xl: 20px;
$font-size-xxl: 24px;

// 边框
$border-radius: 4px;
$border-color: #dcdfe6;
```

## 国际化 (i18n)

使用 Vue I18n 实现多语言支持:

```typescript
// src/i18n/index.ts
import { createI18n } from 'vue-i18n'
import zh from './locales/zh.json'
import en from './locales/en.json'

const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: 'zh', // 默认语言
  fallbackLocale: 'en', // 回退语言
  messages: {
    zh,
    en,
  },
})

export default i18n
```

## 可访问性

遵循以下可访问性原则:

1. 使用语义化 HTML 元素 (`<button>`, `<a>`, `<nav>` 等)
2. 为图片提供 `alt` 属性
3. 确保表单字段有关联的标签
4. 确保颜色对比度符合 WCAG 标准
5. 支持键盘导航

## 性能优化

1. **组件懒加载**:

```typescript
// 路由懒加载
const TaskList = () => import('@/views/tasks/TaskList.vue')
```

2. **虚拟滚动**:

对长列表使用虚拟滚动组件 (`vue-virtual-scroller`)。

3. **资源优化**:

- 使用 WebP 格式图片
- 压缩和合并静态资源
- 使用 HTTP/2

4. **代码分割**:

Vite 会自动进行基于路由的代码分割。

## 部署

### 生产构建

```bash
npm run build
```

### Nginx 配置示例

```nginx
server {
    listen 80;
    server_name your-domain.com;

    root /path/to/web/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend-server:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Docker 部署

```dockerfile
# web/Dockerfile
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 测试

### 单元测试

使用 Vitest 进行单元测试:

```typescript
// src/components/__tests__/ModelCard.spec.ts
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ModelCard from '../ModelCard.vue'

describe('ModelCard', () => {
  it('renders correctly with props', () => {
    const model = {
      id: '1',
      name: 'GPT-4',
      description: 'Advanced language model',
      status: 'active',
    }
  
    const wrapper = mount(ModelCard, {
      props: { model },
    })
  
    expect(wrapper.text()).toContain('GPT-4')
    expect(wrapper.text()).toContain('Advanced language model')
    expect(wrapper.find('.status-badge').classes()).toContain('active')
  })
})
```

### 端到端测试

使用 Cypress 进行端到端测试:

```javascript
// cypress/e2e/login.cy.js
describe('Login Page', () => {
  it('should login with valid credentials', () => {
    cy.visit('/auth/login')
    cy.get('input[name="username"]').type('admin')
    cy.get('input[name="password"]').type('password123')
    cy.get('button[type="submit"]').click()
    cy.url().should('eq', Cypress.config().baseUrl + '/')
    cy.get('.user-profile').should('exist')
  })
})
```

## 常见问题与解决方案

### 1. Vite 开发服务器启动失败

**可能原因**: 端口被占用或依赖问题

**解决方案**:

- 检查端口占用: `lsof -i :3000`
- 尝试更改端口: 修改 `vite.config.ts` 中的 `server.port`
- 清理 `node_modules` 并重新安装: `rm -rf node_modules && npm install`

### 2. 路由跳转问题

**可能原因**: 路由配置错误或路由守卫问题

**解决方案**:

- 检查路由定义
- 检查路由守卫逻辑
- 使用 Vue DevTools 调试

### 3. API 请求失败

**可能原因**: 跨域问题或后端服务未启动

**解决方案**:

- 确保开发代理配置正确
- 检查后端服务状态
- 验证请求参数和认证信息

## 贡献指南

1. 遵循项目的代码风格和规范
2. 为新功能或修复编写单元测试
3. 确保提交前运行 ESLint 和测试
4. 使用语义化提交消息格式

## 参考资源

- [Vue3 官方文档](https://vuejs.org/)
- [Vite 官方文档](https://vitejs.dev/)
- [Element Plus 文档](https://element-plus.org/)
- [Pinia 文档](https://pinia.vuejs.org/)
- [Vue Router 文档](https://router.vuejs.org/)

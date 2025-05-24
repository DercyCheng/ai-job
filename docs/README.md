# AI Gateway MCP Server

分布式 AI 网关 MCP Server 项目文档

## 项目概述

本项目是一个分布式 AI 网关服务，基于 MCP (Model Context Protocol) 实现，用于统一管理和调度多个 AI 模型服务。项目采用微服务架构，实现了高可用、高并发、低延迟的 AI 调用网关。

## 技术栈

- **后端**:
  - Go (主要服务框架)
  - Python (AI 模型服务)
  - Gin (Web 框架)
  - GORM (ORM 框架)
  - MySQL (数据存储)
  
- **前端**:
  - Vue3
  - TypeScript
  - Vite (构建工具)

## 文档索引

- [系统架构](./architecture.md)
- [API 接口文档](./api/README.md)
- [部署指南](./deployment.md)
- [开发指南](./development.md)
- [数据库设计](./database.md)

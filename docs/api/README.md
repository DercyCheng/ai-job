# API 接口文档

## 接口规范

所有 API 遵循 RESTful 设计原则，使用 JSON 作为数据交换格式。

### 基础 URL

```
https://api.example.com/v1
```

### 认证方式

API 使用 Bearer Token 认证：

```
Authorization: Bearer {token}
```

### 响应格式

所有 API 响应遵循以下格式：

```json
{
  "code": 200,          // 状态码
  "message": "success", // 状态信息
  "data": {},           // 响应数据
  "request_id": "uuid"  // 请求唯一标识
}
```

### 错误码定义

| 错误码 | 描述 |
|--------|------|
| 200    | 成功 |
| 400    | 请求参数错误 |
| 401    | 未授权 |
| 403    | 禁止访问 |
| 404    | 资源不存在 |
| 500    | 服务器内部错误 |

## API 目录

- [认证 API](./auth.md)
- [模型 API](./model.md)
- [任务 API](./task.md)
- [系统 API](./system.md)

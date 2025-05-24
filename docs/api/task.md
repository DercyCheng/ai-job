# 任务 API

## 创建任务

**请求方式**: POST

**URL**: `/tasks`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```json
{
  "name": "string",
  "description": "string",
  "model_id": "string",
  "task_type": "completion|chat|embedding|finetune",
  "priority": "high|medium|low",
  "parameters": {
    // 根据任务类型不同而变化
    "max_tokens": 100,
    "temperature": 0.7
  },
  "inputs": [
    {
      "content": "string",
      "content_type": "text|image|audio",
      "role": "system|user"
    }
  ],
  "callback_url": "string"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "string",
    "status": "pending",
    "created_at": "datetime",
    "estimated_completion": "datetime"
  },
  "request_id": "uuid"
}
```

## 获取任务状态

**请求方式**: GET

**URL**: `/tasks/{task_id}`

**请求头**:

```
Authorization: Bearer {token}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "string",
    "name": "string",
    "description": "string",
    "model_id": "string",
    "task_type": "completion",
    "status": "running|completed|failed|canceled",
    "progress": 65,
    "created_at": "datetime",
    "started_at": "datetime",
    "completed_at": "datetime",
    "error": "string",
    "result_url": "string"
  },
  "request_id": "uuid"
}
```

## 获取任务结果

**请求方式**: GET

**URL**: `/tasks/{task_id}/result`

**请求头**:

```
Authorization: Bearer {token}
```

**响应示例** (对于聊天任务):

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "string",
    "model_id": "string",
    "task_type": "chat",
    "status": "completed",
    "result": {
      "choices": [
        {
          "message": {
            "role": "assistant",
            "content": "这是助手的回复"
          },
          "finish_reason": "stop"
        }
      ],
      "usage": {
        "prompt_tokens": 10,
        "completion_tokens": 20,
        "total_tokens": 30
      }
    },
    "created_at": "datetime",
    "completed_at": "datetime"
  },
  "request_id": "uuid"
}
```

## 取消任务

**请求方式**: POST

**URL**: `/tasks/{task_id}/cancel`

**请求头**:

```
Authorization: Bearer {token}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "string",
    "status": "canceled",
    "canceled_at": "datetime"
  },
  "request_id": "uuid"
}
```

## 批量创建任务

**请求方式**: POST

**URL**: `/tasks/batch`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```json
{
  "tasks": [
    {
      "name": "Task 1",
      "model_id": "gpt-4",
      "task_type": "completion",
      "parameters": { "max_tokens": 100 },
      "inputs": [{ "content": "Hello", "role": "user" }]
    },
    {
      "name": "Task 2",
      "model_id": "claude-3",
      "task_type": "chat",
      "parameters": { "temperature": 0.5 },
      "inputs": [{ "content": "Hi there", "role": "user" }]
    }
  ],
  "batch_name": "Example Batch",
  "callback_url": "https://example.com/callback"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "batch_id": "string",
    "task_ids": ["string", "string"],
    "created_at": "datetime"
  },
  "request_id": "uuid"
}
```

## 获取任务历史

**请求方式**: GET

**URL**: `/tasks/history`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```
?status=completed&model_id=gpt-4&start_date=2023-01-01&end_date=2023-01-31&page=1&per_page=20
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "tasks": [
      {
        "task_id": "string",
        "name": "string",
        "model_id": "string",
        "task_type": "completion",
        "status": "completed",
        "created_at": "datetime",
        "completed_at": "datetime"
      }
    ],
    "pagination": {
      "total": 100,
      "page": 1,
      "per_page": 20,
      "total_pages": 5
    }
  },
  "request_id": "uuid"
}
```

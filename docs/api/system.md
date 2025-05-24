# 系统管理 API

## 系统状态

**请求方式**: GET

**URL**: `/system/status`

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
    "status": "healthy",
    "version": "1.0.0",
    "uptime": "10d 4h 30m",
    "node_count": 5,
    "active_workers": 12,
    "queue_status": {
      "pending_tasks": 5,
      "processing_tasks": 8,
      "average_wait_time": "2.5s"
    },
    "system_load": {
      "cpu": 45.2,
      "memory": 60.8,
      "disk": 32.5
    }
  },
  "request_id": "uuid"
}
```

## 系统指标

**请求方式**: GET

**URL**: `/system/metrics`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```
?timeframe=hour&start_time=2023-01-01T00:00:00Z&end_time=2023-01-01T01:00:00Z
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "timeframe": "hour",
    "start_time": "2023-01-01T00:00:00Z",
    "end_time": "2023-01-01T01:00:00Z",
    "metrics": {
      "requests_per_minute": [
        {"timestamp": "2023-01-01T00:00:00Z", "value": 120},
        {"timestamp": "2023-01-01T00:01:00Z", "value": 125},
        {"timestamp": "2023-01-01T00:02:00Z", "value": 118}
      ],
      "average_response_time": [
        {"timestamp": "2023-01-01T00:00:00Z", "value": 235},
        {"timestamp": "2023-01-01T00:01:00Z", "value": 245},
        {"timestamp": "2023-01-01T00:02:00Z", "value": 228}
      ],
      "error_rate": [
        {"timestamp": "2023-01-01T00:00:00Z", "value": 0.5},
        {"timestamp": "2023-01-01T00:01:00Z", "value": 0.8},
        {"timestamp": "2023-01-01T00:02:00Z", "value": 0.3}
      ],
      "resource_utilization": {
        "cpu": [
          {"timestamp": "2023-01-01T00:00:00Z", "value": 42.5},
          {"timestamp": "2023-01-01T00:01:00Z", "value": 45.2},
          {"timestamp": "2023-01-01T00:02:00Z", "value": 40.8}
        ],
        "memory": [
          {"timestamp": "2023-01-01T00:00:00Z", "value": 58.3},
          {"timestamp": "2023-01-01T00:01:00Z", "value": 60.8},
          {"timestamp": "2023-01-01T00:02:00Z", "value": 59.5}
        ]
      }
    }
  },
  "request_id": "uuid"
}
```

## 模型使用统计

**请求方式**: GET

**URL**: `/system/stats/models`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```
?timeframe=day&start_date=2023-01-01&end_date=2023-01-31
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "timeframe": "day",
    "start_date": "2023-01-01",
    "end_date": "2023-01-31",
    "models": [
      {
        "model_id": "gpt-4",
        "total_requests": 12500,
        "total_tokens": 7850000,
        "average_response_time": 380,
        "success_rate": 99.2,
        "daily_stats": [
          {"date": "2023-01-01", "requests": 350, "tokens": 220000},
          {"date": "2023-01-02", "requests": 420, "tokens": 265000}
        ]
      },
      {
        "model_id": "claude-3",
        "total_requests": 8300,
        "total_tokens": 5200000,
        "average_response_time": 420,
        "success_rate": 98.7,
        "daily_stats": [
          {"date": "2023-01-01", "requests": 280, "tokens": 175000},
          {"date": "2023-01-02", "requests": 310, "tokens": 195000}
        ]
      }
    ]
  },
  "request_id": "uuid"
}
```

## 用户使用统计

**请求方式**: GET

**URL**: `/system/stats/users`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```
?timeframe=month&year=2023&month=1
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "timeframe": "month",
    "year": 2023,
    "month": 1,
    "total_users": 350,
    "active_users": 280,
    "new_users": 45,
    "total_requests": 85600,
    "total_tokens": 52300000,
    "top_users": [
      {
        "user_id": "user-1",
        "requests": 3200,
        "tokens": 1950000
      },
      {
        "user_id": "user-2",
        "requests": 2800,
        "tokens": 1750000
      }
    ],
    "usage_distribution": {
      "by_model": {
        "gpt-4": 58.3,
        "claude-3": 32.5,
        "other": 9.2
      },
      "by_endpoint": {
        "chat": 72.5,
        "completion": 18.3,
        "embedding": 9.2
      }
    }
  },
  "request_id": "uuid"
}
```

## 系统配置

**请求方式**: GET

**URL**: `/system/config`

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
    "rate_limits": {
      "global": {
        "requests_per_minute": 1000,
        "tokens_per_minute": 500000
      },
      "per_user": {
        "default": {
          "requests_per_minute": 20,
          "tokens_per_minute": 10000
        },
        "premium": {
          "requests_per_minute": 100,
          "tokens_per_minute": 50000
        }
      }
    },
    "models": {
      "enabled_models": ["gpt-4", "claude-3", "llama-3", "gemini-pro"],
      "default_model": "gpt-4"
    },
    "workers": {
      "min_instances": 5,
      "max_instances": 20,
      "scaling_rules": {
        "scale_up_threshold": 70,
        "scale_down_threshold": 30
      }
    },
    "queue": {
      "max_queue_size": 1000,
      "max_wait_time": 30
    },
    "logging": {
      "level": "info",
      "retention_days": 30
    }
  },
  "request_id": "uuid"
}
```

## 更新系统配置

**请求方式**: PATCH

**URL**: `/system/config`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```json
{
  "rate_limits": {
    "global": {
      "requests_per_minute": 1200
    },
    "per_user": {
      "premium": {
        "tokens_per_minute": 60000
      }
    }
  },
  "models": {
    "enabled_models": ["gpt-4", "claude-3", "llama-3", "gemini-pro", "mistral-large"]
  },
  "workers": {
    "max_instances": 25
  }
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "updated_fields": ["rate_limits.global.requests_per_minute", "rate_limits.per_user.premium.tokens_per_minute", "models.enabled_models", "workers.max_instances"],
    "applied_at": "datetime"
  },
  "request_id": "uuid"
}
```

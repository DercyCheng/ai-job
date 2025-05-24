# 认证 API

## 用户注册

**请求方式**: POST

**URL**: `/auth/register`

**请求参数**:

```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "string",
    "username": "string",
    "email": "string",
    "created_at": "datetime"
  },
  "request_id": "uuid"
}
```

## 用户登录

**请求方式**: POST

**URL**: `/auth/login`

**请求参数**:

```json
{
  "username": "string",
  "password": "string"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "string",
    "expires_in": 3600,
    "user_info": {
      "user_id": "string",
      "username": "string",
      "role": "string"
    }
  },
  "request_id": "uuid"
}
```

## 令牌刷新

**请求方式**: POST

**URL**: `/auth/refresh`

**请求头**:

```
Authorization: Bearer {refresh_token}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "string",
    "expires_in": 3600
  },
  "request_id": "uuid"
}
```

## 注销登录

**请求方式**: POST

**URL**: `/auth/logout`

**请求头**:

```
Authorization: Bearer {token}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": null,
  "request_id": "uuid"
}
```

## API 密钥生成

**请求方式**: POST

**URL**: `/auth/api-keys`

**请求头**:

```
Authorization: Bearer {token}
```

**请求参数**:

```json
{
  "name": "string",
  "permissions": ["read", "write"],
  "expires_in": 2592000
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "key_id": "string",
    "api_key": "string",
    "name": "string",
    "permissions": ["read", "write"],
    "created_at": "datetime",
    "expires_at": "datetime"
  },
  "request_id": "uuid"
}
```

# 数据库设计

## 概述

本文档描述 AI 网关 MCP Server 项目的数据库设计，包括表结构、索引、关系以及优化策略。系统使用 MySQL 8.0 作为主要数据库，结合 GORM 作为 ORM 框架。

## 数据库架构

### 数据库图表

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│                 │       │                 │       │                 │
│     users       │───┐   │     models      │───┐   │     tasks       │
│                 │   │   │                 │   │   │                 │
└─────────────────┘   │   └─────────────────┘   │   └─────────────────┘
                      │                         │
                      │   ┌─────────────────┐   │
                      └───│                 │───┘
                          │   task_logs     │
                          │                 │
                          └─────────────────┘
```

## 表结构设计

### 用户管理

#### users 表

存储用户信息和认证数据。

```sql
CREATE TABLE `users` (
  `id` varchar(36) NOT NULL,
  `username` varchar(50) NOT NULL,
  `email` varchar(100) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `role` enum('admin', 'user', 'api') NOT NULL DEFAULT 'user',
  `status` enum('active', 'inactive', 'suspended') NOT NULL DEFAULT 'active',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `last_login_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`),
  UNIQUE KEY `idx_users_email` (`email`),
  KEY `idx_users_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### api_keys 表

存储 API 密钥信息。

```sql
CREATE TABLE `api_keys` (
  `id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `name` varchar(100) NOT NULL,
  `key_hash` varchar(255) NOT NULL,
  `permissions` json NOT NULL,
  `last_used_at` datetime DEFAULT NULL,
  `expires_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_api_keys_user_id` (`user_id`),
  CONSTRAINT `fk_api_keys_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 模型管理

#### models 表

存储可用的 AI 模型信息。

```sql
CREATE TABLE `models` (
  `id` varchar(36) NOT NULL,
  `name` varchar(100) NOT NULL,
  `provider` varchar(50) NOT NULL,
  `version` varchar(20) NOT NULL,
  `description` text,
  `capabilities` json NOT NULL,
  `parameters` json NOT NULL,
  `pricing` json,
  `status` enum('active', 'inactive', 'deprecated') NOT NULL DEFAULT 'active',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_models_name_provider_version` (`name`, `provider`, `version`),
  KEY `idx_models_provider` (`provider`),
  KEY `idx_models_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### model_stats 表

存储模型使用统计信息。

```sql
CREATE TABLE `model_stats` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `model_id` varchar(36) NOT NULL,
  `date` date NOT NULL,
  `hour` tinyint(2) DEFAULT NULL,
  `request_count` int(11) NOT NULL DEFAULT '0',
  `token_count` bigint(20) NOT NULL DEFAULT '0',
  `error_count` int(11) NOT NULL DEFAULT '0',
  `avg_latency_ms` int(11) DEFAULT NULL,
  `p95_latency_ms` int(11) DEFAULT NULL,
  `p99_latency_ms` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_model_stats_model_id_date_hour` (`model_id`, `date`, `hour`),
  KEY `idx_model_stats_date` (`date`),
  CONSTRAINT `fk_model_stats_model_id` FOREIGN KEY (`model_id`) REFERENCES `models` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 任务管理

#### tasks 表

存储 AI 任务信息。

```sql
CREATE TABLE `tasks` (
  `id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `model_id` varchar(36) NOT NULL,
  `batch_id` varchar(36) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `description` text,
  `task_type` enum('completion', 'chat', 'embedding', 'finetune') NOT NULL,
  `status` enum('pending', 'running', 'completed', 'failed', 'canceled') NOT NULL DEFAULT 'pending',
  `priority` tinyint(4) NOT NULL DEFAULT '5',
  `parameters` json NOT NULL,
  `inputs` json NOT NULL,
  `result_location` varchar(255) DEFAULT NULL,
  `error` text,
  `callback_url` varchar(255) DEFAULT NULL,
  `callback_status` enum('pending', 'sent', 'failed') DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `started_at` datetime DEFAULT NULL,
  `completed_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_tasks_user_id` (`user_id`),
  KEY `idx_tasks_model_id` (`model_id`),
  KEY `idx_tasks_batch_id` (`batch_id`),
  KEY `idx_tasks_status` (`status`),
  KEY `idx_tasks_created_at` (`created_at`),
  CONSTRAINT `fk_tasks_model_id` FOREIGN KEY (`model_id`) REFERENCES `models` (`id`),
  CONSTRAINT `fk_tasks_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### task_logs 表

存储任务执行日志。

```sql
CREATE TABLE `task_logs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `task_id` varchar(36) NOT NULL,
  `worker_id` varchar(36) NOT NULL,
  `timestamp` datetime NOT NULL,
  `log_level` enum('debug', 'info', 'warning', 'error') NOT NULL,
  `message` text NOT NULL,
  `metadata` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_task_logs_task_id` (`task_id`),
  KEY `idx_task_logs_timestamp` (`timestamp`),
  CONSTRAINT `fk_task_logs_task_id` FOREIGN KEY (`task_id`) REFERENCES `tasks` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### task_results 表

存储较大的任务结果数据，与 tasks 表分离以提高性能。

```sql
CREATE TABLE `task_results` (
  `task_id` varchar(36) NOT NULL,
  `result_data` longtext NOT NULL,
  `result_format` enum('json', 'text', 'binary') NOT NULL DEFAULT 'json',
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`task_id`),
  CONSTRAINT `fk_task_results_task_id` FOREIGN KEY (`task_id`) REFERENCES `tasks` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 系统管理

#### workers 表

存储模型工作节点信息。

```sql
CREATE TABLE `workers` (
  `id` varchar(36) NOT NULL,
  `hostname` varchar(255) NOT NULL,
  `ip_address` varchar(45) NOT NULL,
  `status` enum('online', 'offline', 'busy', 'error') NOT NULL,
  `capabilities` json NOT NULL,
  `current_load` float DEFAULT NULL,
  `max_capacity` int(11) NOT NULL,
  `last_heartbeat` datetime NOT NULL,
  `registered_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workers_status` (`status`),
  KEY `idx_workers_last_heartbeat` (`last_heartbeat`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### system_metrics 表

存储系统性能指标。

```sql
CREATE TABLE `system_metrics` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `timestamp` datetime NOT NULL,
  `metric_name` varchar(100) NOT NULL,
  `metric_value` float NOT NULL,
  `dimensions` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_system_metrics_timestamp` (`timestamp`),
  KEY `idx_system_metrics_name` (`metric_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### audit_logs 表

存储系统审计日志。

```sql
CREATE TABLE `audit_logs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(36) DEFAULT NULL,
  `action` varchar(100) NOT NULL,
  `resource_type` varchar(50) NOT NULL,
  `resource_id` varchar(36) DEFAULT NULL,
  `ip_address` varchar(45) DEFAULT NULL,
  `user_agent` varchar(255) DEFAULT NULL,
  `request_id` varchar(36) DEFAULT NULL,
  `details` json DEFAULT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_audit_logs_user_id` (`user_id`),
  KEY `idx_audit_logs_action` (`action`),
  KEY `idx_audit_logs_resource` (`resource_type`, `resource_id`),
  KEY `idx_audit_logs_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

## 索引策略

### 主键索引
- 使用 UUID 或自增 ID 作为主键
- 自增 ID 用于高写入量的表 (如 `task_logs`, `system_metrics`)
- UUID 用于需要全局唯一标识的表 (如 `users`, `tasks`)

### 外键索引
- 所有外键关系都有相应索引

### 复合索引
- 针对常见查询模式创建复合索引
- 示例: `idx_model_stats_model_id_date_hour` 优化按模型和时间段查询统计数据

### 全文索引
- 对需要全文搜索的字段使用全文索引
- MySQL 8.0+ 支持更好的全文搜索能力

## 分区策略

### 按时间分区
- 大型日志和指标表使用按时间范围分区
- 示例: `task_logs` 表可按月分区，简化历史数据归档

```sql
ALTER TABLE task_logs PARTITION BY RANGE (TO_DAYS(timestamp)) (
    PARTITION p202301 VALUES LESS THAN (TO_DAYS('2023-02-01')),
    PARTITION p202302 VALUES LESS THAN (TO_DAYS('2023-03-01')),
    PARTITION p202303 VALUES LESS THAN (TO_DAYS('2023-04-01')),
    PARTITION future VALUES LESS THAN MAXVALUE
);
```

## 数据类型选择

- 使用 `varchar` 而非 `char` 存储变长字符串
- 使用 `json` 类型存储结构灵活的数据
- 使用 `enum` 类型代替字符串存储固定选项
- 使用 `datetime` 而非 `timestamp` 避免时区问题

## 数据迁移和版本控制

使用 GORM 的迁移工具或独立迁移工具 (如 golang-migrate) 管理数据库版本。

示例迁移文件:

```go
// 创建用户表
func MigrateCreateUserTable(db *gorm.DB) error {
    return db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id varchar(36) NOT NULL,
            username varchar(50) NOT NULL,
            email varchar(100) NOT NULL,
            password_hash varchar(255) NOT NULL,
            role enum('admin', 'user', 'api') NOT NULL DEFAULT 'user',
            status enum('active', 'inactive', 'suspended') NOT NULL DEFAULT 'active',
            created_at datetime NOT NULL,
            updated_at datetime NOT NULL,
            last_login_at datetime DEFAULT NULL,
            PRIMARY KEY (id),
            UNIQUE KEY idx_users_username (username),
            UNIQUE KEY idx_users_email (email),
            KEY idx_users_status (status)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
    `).Error
}
```

## GORM 模型定义

示例 GORM 模型:

```go
// User 模型
type User struct {
    ID           string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
    Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email        string    `gorm:"uniqueIndex;size:100;not null" json:"email"`
    PasswordHash string    `gorm:"size:255;not null" json:"-"`
    Role         string    `gorm:"type:enum('admin','user','api');default:user;not null" json:"role"`
    Status       string    `gorm:"type:enum('active','inactive','suspended');default:active;not null" json:"status"`
    CreatedAt    time.Time `gorm:"not null" json:"created_at"`
    UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
    LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
    
    // 关联
    APIKeys      []APIKey  `gorm:"foreignKey:UserID" json:"api_keys,omitempty"`
    Tasks        []Task    `gorm:"foreignKey:UserID" json:"tasks,omitempty"`
}

// BeforeCreate 钩子设置 UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = uuid.New().String()
    }
    return nil
}
```

## 数据库优化策略

### 查询优化

1. **使用适当的索引**
   - 对频繁查询的字段创建索引
   - 使用 `EXPLAIN` 分析查询执行计划

2. **避免全表扫描**
   - 总是在 WHERE 子句中使用索引列
   - 避免在索引列上使用函数

3. **分页查询优化**
   - 使用键集分页而非 OFFSET/LIMIT
   ```sql
   -- 优化前
   SELECT * FROM tasks ORDER BY created_at DESC LIMIT 10 OFFSET 100
   
   -- 优化后
   SELECT * FROM tasks WHERE created_at < ? ORDER BY created_at DESC LIMIT 10
   ```

### 连接池配置

```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
sqlDB, err := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### 读写分离

对于高负载系统，考虑配置读写分离:

```go
db, err := gorm.Open(mysql.New(mysql.Config{
    DSN: "write_dsn",
    Replicas: []string{"read_dsn_1", "read_dsn_2"},
    ReadTimeout: time.Second * 2,
    WriteTimeout: time.Second * 2,
}), &gorm.Config{})
```

## 监控与告警

关键监控指标:

1. **查询性能**
   - 慢查询次数和耗时
   - 索引使用率

2. **连接状态**
   - 活跃连接数
   - 等待连接数
   - 连接错误率

3. **资源使用**
   - CPU 使用率
   - 内存使用情况
   - 磁盘 I/O

## 数据备份策略

1. **定时备份**
   - 每日全量备份
   - 实时 binlog 备份

2. **备份验证**
   - 定期验证备份的完整性和可恢复性

3. **备份脚本示例**

```bash
#!/bin/bash
DATE=$(date +%Y%m%d)
BACKUP_DIR="/var/backups/mysql"

# 全量备份
mysqldump -u root -p --all-databases > "$BACKUP_DIR/full_backup_$DATE.sql"

# 压缩
gzip "$BACKUP_DIR/full_backup_$DATE.sql"

# 清理旧备份 (保留30天)
find "$BACKUP_DIR" -name "full_backup_*.sql.gz" -mtime +30 -delete
```

## 扩展建议

### 分库分表

随着数据量增长，考虑以下策略:

1. **垂直分表**
   - 将大表拆分为多个表，如将 `tasks` 表中的大字段移至 `task_details` 表

2. **水平分表**
   - 按用户 ID 或时间范围分片
   - 使用中间件如 ShardingSphere 管理分片

### 缓存策略

1. **缓存层**
   - 使用 Redis 缓存热点数据
   - 实现缓存预热和定时刷新

2. **缓存示例代码**

```go
func GetModelByID(id string) (*Model, error) {
    // 尝试从缓存获取
    cacheKey := "model:" + id
    if cachedModel, found := redisClient.Get(cacheKey); found {
        var model Model
        err := json.Unmarshal([]byte(cachedModel), &model)
        if err == nil {
            return &model, nil
        }
    }
    
    // 从数据库获取
    var model Model
    if err := db.First(&model, "id = ?", id).Error; err != nil {
        return nil, err
    }
    
    // 存入缓存
    modelJSON, _ := json.Marshal(model)
    redisClient.Set(cacheKey, modelJSON, time.Hour)
    
    return &model, nil
}
```

## 数据安全

1. **敏感数据加密**
   - 密码哈希存储
   - API 密钥加密
   - 个人信息加密

2. **访问控制**
   - 最小权限原则
   - 数据访问审计

3. **防 SQL 注入**
   - 使用参数化查询
   - ORM 框架自动防护

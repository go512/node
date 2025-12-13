# qlog - 高性能日志包

## 📋 功能特性

- ✅ **多种输出方式**: 控制台、文件、自动日志轮转
- ✅ **灵活格式化**: Text 和 JSON 格式
- ✅ **日志轮转**: 基于文件大小、时间和备份数量
- ✅ **结构化日志**: 支持字段和上下文信息
- ✅ **生产就绪**: 压缩、错误降级、资源管理
- ✅ **便捷方法**: 组件标识、请求追踪

## 🚀 快速开始

### 基础使用

```go
package main

import (
    "node/pkg/qlog"
)

func main() {
    // 使用默认配置
    logger := qlog.New()
    
    logger.Info("应用启动成功")
    logger.WithField("user_id", 12345).Info("用户登录")
}
```

### 自定义配置

```go
option := &qlog.Option{
    Level:             "debug",
    Output:            "file",
    OutputFilePath:    "logs/app.log",
    OutputFileMaxSize: 100, // 100MB
    Formatter:         "json",
    EnableCaller:      true,
    MaxBackups:        5,
    MaxAge:            30, // 30天
    Compress:          true,
}

logger := qlog.NewWithOption(option)
defer logger.Close() // 记得关闭文件
```

## 📖 配置选项详解

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Level` | string | "info" | 日志级别：debug/info/warn/error/fatal/panic |
| `Output` | string | "stdout" | 输出位置：stdout/stderr/file |
| `OutputFilePath` | string | "logs/app.log" | 日志文件路径 |
| `OutputFileMaxSize` | int64 | 100 | 文件最大大小（MB） |
| `Formatter` | string | "text" | 格式：text/json |
| `EnableCaller` | bool | true | 是否显示调用者信息 |
| `TimestampFormat` | string | "2006-01-02 15:04:05" | 时间格式 |
| `MaxBackups` | int | 3 | 最大备份文件数 |
| `MaxAge` | int | 30 | 日志保存天数 |
| `Compress` | bool | true | 是否压缩旧日志 |

## 💡 使用示例

### 1. 基础日志记录

```go
logger := qlog.New()

// 不同级别的日志
logger.Debug("调试信息")
logger.Info("普通信息")
logger.Warn("警告信息")
logger.Error("错误信息")

// 格式化日志
logger.Infof("用户 %s 在 %s 登录", "张三", "2024-01-01")
```

### 2. 结构化日志

```go
// 单字段
logger.WithField("user_id", 12345).Info("用户登录")

// 多字段
logger.WithFields(map[string]interface{}{
    "method": "POST",
    "path":   "/api/users",
    "status": 200,
    "duration": "150ms",
}).Info("HTTP 请求完成")

// 错误记录
err := errors.New("数据库连接失败")
logger.WithError(err).WithField("retry_count", 3).Error("操作失败")
```

### 3. 便捷方法

```go
// 组件标识
logger.WithComponent("auth-service").Info("认证服务启动")

// 请求追踪
logger.WithRequestID("req-123456").
    WithComponent("payment").
    Info("支付处理开始")

// 组合使用
logger.WithComponent("kafka-consumer").
    WithRequestID("msg-789012").
    WithFields(map[string]interface{}{
        "topic":     "user-events",
        "partition": 0,
        "offset":    123456,
    }).Info("消息处理成功")
```

### 4. 不同环境配置

#### 开发环境
```go
devOption := &qlog.Option{
    Level:        "debug",
    Output:       "stdout",
    Formatter:    "text",
    EnableCaller: true,
}
```

#### 测试环境
```go
testOption := &qlog.Option{
    Level:             "info",
    Output:            "file",
    OutputFilePath:    "logs/test.log",
    OutputFileMaxSize: 50,
    Formatter:         "json",
    MaxBackups:        5,
    Compress:          true,
}
```

#### 生产环境
```go
prodOption := &qlog.Option{
    Level:             "warn",
    Output:            "file",
    OutputFilePath:    "/var/log/app/app.log",
    OutputFileMaxSize: 500,
    Formatter:         "json",
    EnableCaller:      false, // 关闭调用者信息提升性能
    MaxBackups:        20,
    MaxAge:            90,
    Compress:          true,
}
```

## 🔧 在 Kafka 项目中集成

### 修改 kafka_consumer.go

```go
package kafkaPkg

import (
    "node/pkg/qlog"
)

// 在全局或结构体中定义 logger
var consumerLogger qlog.Logger

func init() {
    // 初始化消费者专用日志
    option := &qlog.Option{
        Level:             "info",
        Output:            "file",
        OutputFilePath:    "logs/kafka-consumer.log",
        OutputFileMaxSize: 100,
        Formatter:         "json",
        EnableCaller:      true,
        MaxBackups:        10,
        MaxAge:            30,
        Compress:          true,
    }
    consumerLogger = qlog.NewWithOption(option)
}

func Subscribe() {
    cfg := initConfig()
    
    consumerLogger.WithComponent("kafka-consumer").
        WithFields(map[string]interface{}{
            "topic":      cfg.Topic,
            "group_id":   cfg.GroupID,
            "brokers":    cfg.Brokers,
            "max_workers": cfg.MaxWorkers,
        }).Info("消费者订阅开始")

    partitionIds, err := getTopi(&cfg, cfg.Topic)
    if err != nil {
        consumerLogger.WithComponent("kafka-consumer").
            WithError(err).
            Error("获取分区信息失败")
        return
    }

    consumerLogger.WithComponent("kafka-consumer").
        WithField("partition_count", len(partitionIds)).
        Info("分区信息获取成功")
}
```

## 🎯 最佳实践

### 1. 日志级别使用

```go
// DEBUG: 详细的调试信息
logger.WithField("offset", 123456).Debug("处理消息详情")

// INFO: 一般信息，记录关键流程
logger.WithComponent("service").Info("服务启动完成")

// WARN: 警告信息，可能的问题但不影响运行
logger.WithField("memory_usage", "85%").Warn("内存使用率较高")

// ERROR: 错误信息，需要关注
logger.WithError(err).Error("数据库查询失败")
```

### 2. 结构化字段设计

```go
// 推荐的字段命名
logger.WithFields(map[string]interface{}{
    "event_type": "user_action",
    "user_id":    12345,
    "action":     "login",
    "ip_address": "192.168.1.100",
    "timestamp":  time.Now().Unix(),
}).Info("用户行为记录")
```

### 3. 性能考虑

```go
// 生产环境关闭调用者信息
if os.Getenv("ENV") == "production" {
    option.EnableCaller = false
}

// 使用条件日志避免不必要的格式化
if logger.IsLevelEnabled(logrus.DebugLevel) {
    logger.WithField("complex_data", expensiveOperation()).Debug()
}
```

### 4. 资源管理

```go
func main() {
    logger := qlog.NewWithOption(fileOption)
    defer logger.Close() // 确保文件句柄正确关闭
    
    // 应用逻辑...
}
```

## 🚨 注意事项

1. **性能**: 生产环境建议关闭 `EnableCaller`
2. **安全**: 避免在日志中记录敏感信息（密码、token等）
3. **存储**: 合理设置日志轮转，避免磁盘空间不足
4. **监控**: 建议集成日志监控系统（如 ELK Stack）

## 📦 依赖管理

确保 `go.mod` 包含以下依赖：

```go
require (
    github.com/sirupsen/logrus v1.9.3
    gopkg.in/natefinch/lumberjack.v2 v2.2.1
)
```

使用命令添加：
```bash
go get github.com/sirupsen/logrus
go get gopkg.in/natefinch/lumberjack.v2
```
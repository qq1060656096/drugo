# Log 包使用文档

## 简介

`log` 包是一个基于 [uber-go/zap](https://github.com/uber-go/zap) 和 [lumberjack](https://github.com/natefinch/lumberjack) 的高性能日志管理库，专为多业务模块场景设计，提供了完善的日志管理能力。

### 主要特性

- 🚀 **高性能**：基于 zap 构建，性能优异
- 📦 **多业务支持**：为不同业务模块提供独立的日志实例管理
- 🔄 **动态级别调整**：支持运行时动态修改日志级别
- 📁 **自动归档**：支持日志文件自动切分、压缩和清理
- 🎨 **多种格式**：支持 JSON、Console、Text 三种日志格式
- 🖥️ **控制台输出**：可同时输出到文件和控制台
- 🔒 **并发安全**：所有操作都是并发安全的

## 安装

```bash
go get github.com/qq1060656096/drugo/log
```

## 快速开始

### 1. 基本使用

```go
package main

import (
    "github.com/qq1060656096/drugo/log"
    "go.uber.org/zap"
)

func main() {
    // 创建日志配置
    cfg := log.Config{
        Dir:        "./logs",      // 日志目录
        Level:      "info",         // 日志级别
        Format:     "json",         // 日志格式
        MaxSize:    100,            // 单个文件最大 100MB
        MaxBackups: 30,             // 保留 30 个备份文件
        MaxAge:     7,              // 保留 7 天
        Compress:   true,           // 压缩旧日志
        Console:    true,           // 同时输出到控制台
    }

    // 创建日志管理器
    manager, err := log.NewManager(cfg)
    if err != nil {
        panic(err)
    }
    defer manager.Close()

    // 获取业务日志实例
    logger, err := manager.Get("app")
    if err != nil {
        panic(err)
    }

    // 使用日志
    logger.Info("应用启动成功",
        zap.String("version", "1.0.0"),
        zap.Int("port", 8080),
    )
}
```

### 2. 使用全局默认 Manager

```go
package main

import (
    "github.com/qq1060656096/drugo/log"
    "go.uber.org/zap"
)

func main() {
    // 初始化全局默认 Manager（只会初始化一次）
    log.Init(log.Config{
        Dir:     "./logs",
        Level:   "info",
        Format:  "json",
        Console: true,
    })

    // 在任何地方获取全局 Manager
    manager := log.Default()
    logger := manager.MustGet("app")

    logger.Info("使用全局 Manager")
}
```

## 配置说明

### Config 结构体

```go
type Config struct {
    Dir        string `yaml:"dir"`         // 日志目录（必填）
    Level      string `yaml:"level"`       // 日志级别: debug, info, warn, error
    Format     string `yaml:"format"`      // 日志格式: json, console, text(或standard)
    MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小(MB)
    MaxBackups int    `yaml:"max_backups"` // 保留的旧日志文件数量
    MaxAge     int    `yaml:"max_age"`     // 保留旧日志的最大天数
    Compress   bool   `yaml:"compress"`    // 是否压缩旧日志文件
    Console    bool   `yaml:"console"`     // 是否同时输出到控制台
}
```

### 配置项详解

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| `dir` | string | ✅ | 日志文件存储目录 | - |
| `level` | string | ❌ | 日志级别：`debug`、`info`、`warn`、`error` | `info` |
| `format` | string | ❌ | 日志格式：`json`、`console`、`text`/`standard` | `json` |
| `max_size` | int | ❌ | 单个日志文件最大大小（MB），超过后自动切分 | `100` |
| `max_backups` | int | ❌ | 保留的旧日志文件数量，0 表示保留所有 | `0` |
| `max_age` | int | ❌ | 保留旧日志的最大天数，0 表示不删除 | `0` |
| `compress` | bool | ❌ | 是否压缩旧日志文件为 `.gz` 格式 | `false` |
| `console` | bool | ❌ | 是否同时输出到控制台（stdout） | `false` |

### YAML 配置文件示例

```yaml
log:
  # 日志目录
  dir: "./runtime/logs"
  
  # 日志等级，可选值: debug, info, warn, error
  level: "info"
  
  # 日志格式，可选值: json, console, text
  format: "json"
  
  # 日志切分策略: 单个文件最大尺寸，单位 MB
  max_size: 100
  
  # 日志切分策略: 保留旧日志文件数量
  max_backups: 30
  
  # 日志切分策略: 保留旧日志最大天数
  max_age: 7
  
  # 是否压缩旧日志文件
  compress: true
  
  # 是否同时输出到控制台
  console: true
```

## 核心功能

### 1. 日志管理器

#### 创建管理器

```go
// 方式1：正常创建，返回错误
manager, err := log.NewManager(cfg)
if err != nil {
    // 处理错误
}

// 方式2：创建失败时 panic
manager := log.MustNewManager(cfg)

// 方式3：使用全局默认 Manager
log.Init(cfg)
manager := log.Default()
```

#### 获取日志实例

```go
// 方式1：正常获取，返回错误
logger, err := manager.Get("app")
if err != nil {
    // 处理错误
}

// 方式2：获取失败时 panic
logger := manager.MustGet("app")
```

**注意**：
- 每个业务名称（bizName）会创建独立的日志文件：`{dir}/{bizName}.log`
- 日志实例会被缓存，重复调用 `Get()` 返回相同的实例
- 业务名称不能为空

### 2. 日志记录

```go
logger, _ := manager.Get("app")

// 不同级别的日志
logger.Debug("调试信息", zap.String("key", "value"))
logger.Info("普通信息", zap.Int("count", 10))
logger.Warn("警告信息", zap.Error(err))
logger.Error("错误信息", zap.Stack("stacktrace"))

// 使用结构化字段
logger.Info("用户登录",
    zap.String("username", "john"),
    zap.String("ip", "192.168.1.1"),
    zap.Duration("elapsed", time.Second),
)

// 使用便捷函数记录任意数据
logger.Info("数据记录", log.Data(map[string]interface{}{
    "key1": "value1",
    "key2": 123,
}))
```

### 3. 动态级别调整

```go
// 设置日志级别
err := manager.SetLevel("app", "debug")
if err != nil {
    // 处理错误
}

// 获取当前日志级别
level, err := manager.GetLevel("app")
if err != nil {
    // 处理错误
}
fmt.Printf("当前日志级别: %s\n", level)
```

**支持的日志级别**（从低到高）：
- `debug`：调试信息
- `info`：普通信息
- `warn`：警告信息
- `error`：错误信息

### 4. 日志实例管理

```go
// 列出所有日志实例
names := manager.List()
fmt.Printf("已创建的日志实例: %v\n", names)

// 移除指定日志实例
err := manager.Remove("app")
if err != nil {
    // 处理错误
}

// 同步所有日志（刷新缓冲区到磁盘）
err := manager.Sync()
if err != nil {
    // 处理错误
}

// 关闭管理器（同步并清理所有日志实例）
err := manager.Close()
if err != nil {
    // 处理错误
}
```

## 日志格式

### 1. JSON 格式（推荐生产环境）

```json
{
  "level": "info",
  "ts": "2024-01-10T13:55:36+08:00",
  "caller": "main.go:25",
  "msg": "用户登录",
  "biz": "app",
  "username": "john",
  "ip": "192.168.1.1"
}
```

**特点**：
- 易于机器解析
- 适合日志采集和分析
- 推荐用于生产环境

### 2. Console 格式（适合开发环境）

```
2024-01-10 13:55:36  INFO  main.go:25  用户登录  biz=app username=john ip=192.168.1.1
```

**特点**：
- 彩色输出，易于阅读
- 适合开发调试
- 控制台友好

### 3. Text/Standard 格式（标准格式）

```
2024-01-10 13:55:36 [INFO] 用户登录 biz=app username=john ip=192.168.1.1
```

**特点**：
- 标准应用日志格式
- 可读性好
- 适合传统日志习惯

## 最佳实践

### 1. 项目结构示例

```
myapp/
├── conf/
│   └── log.yaml          # 日志配置文件
├── main.go
└── runtime/
    └── logs/             # 日志目录
        ├── app.log       # 应用日志
        ├── api.log       # API 日志
        └── db.log        # 数据库日志
```

### 2. 完整示例

```go
package main

import (
    "os"
    "os/signal"
    "syscall"

    "github.com/qq1060656096/drugo/log"
    "go.uber.org/zap"
)

func main() {
    // 初始化日志管理器
    cfg := log.Config{
        Dir:        "./runtime/logs",
        Level:      "info",
        Format:     "json",
        MaxSize:    100,
        MaxBackups: 30,
        MaxAge:     7,
        Compress:   true,
        Console:    true,
    }
    
    log.Init(cfg)
    manager := log.Default()
    
    // 确保程序退出前同步日志
    defer func() {
        if err := manager.Sync(); err != nil {
            // 处理同步错误
        }
    }()

    // 获取不同业务的日志实例
    appLogger := manager.MustGet("app")
    apiLogger := manager.MustGet("api")
    dbLogger := manager.MustGet("db")

    // 应用启动日志
    appLogger.Info("应用启动",
        zap.String("version", "1.0.0"),
        zap.Int("pid", os.Getpid()),
    )

    // API 请求日志
    apiLogger.Info("收到请求",
        zap.String("method", "GET"),
        zap.String("path", "/api/users"),
        zap.Int("status", 200),
    )

    // 数据库日志
    dbLogger.Info("数据库连接成功",
        zap.String("host", "localhost"),
        zap.Int("port", 3306),
    )

    // 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    appLogger.Info("应用关闭")
}
```

### 3. 与配置文件集成

```go
package main

import (
    "github.com/qq1060656096/drugo/log"
    "github.com/spf13/viper"
)

func main() {
    // 读取配置文件
    v := viper.New()
    v.SetConfigFile("conf/log.yaml")
    if err := v.ReadInConfig(); err != nil {
        panic(err)
    }

    // 解析日志配置
    var cfg log.Config
    if err := v.UnmarshalKey("log", &cfg); err != nil {
        panic(err)
    }

    // 初始化日志管理器
    log.Init(cfg)
    manager := log.Default()
    
    logger := manager.MustGet("app")
    logger.Info("日志初始化成功")
}
```

### 4. HTTP 服务中使用

```go
package main

import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/qq1060656096/drugo/log"
    "go.uber.org/zap"
)

func main() {
    // 初始化日志
    log.Init(log.Config{
        Dir:     "./logs",
        Level:   "info",
        Format:  "json",
        Console: true,
    })
    
    manager := log.Default()
    defer manager.Sync()

    // 获取日志实例
    accessLogger := manager.MustGet("access")
    appLogger := manager.MustGet("app")

    // 创建 Gin 路由
    r := gin.New()

    // 日志中间件
    r.Use(func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method

        c.Next()

        // 记录访问日志
        accessLogger.Info("HTTP请求",
            zap.String("method", method),
            zap.String("path", path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("latency", time.Since(start)),
            zap.String("ip", c.ClientIP()),
        )
    })

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    appLogger.Info("服务启动", zap.Int("port", 8080))
    r.Run(":8080")
}
```

### 5. 动态日志级别调整（适用于线上调试）

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/qq1060656096/drugo/log"
)

func main() {
    log.Init(log.Config{
        Dir:    "./logs",
        Level:  "info",
        Format: "json",
    })
    
    manager := log.Default()
    r := gin.Default()

    // 获取日志级别
    r.GET("/admin/log/level/:bizName", func(c *gin.Context) {
        bizName := c.Param("bizName")
        level, err := manager.GetLevel(bizName)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"bizName": bizName, "level": level})
    })

    // 设置日志级别
    r.PUT("/admin/log/level/:bizName", func(c *gin.Context) {
        bizName := c.Param("bizName")
        var req struct {
            Level string `json:"level" binding:"required"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        if err := manager.SetLevel(bizName, req.Level); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "message": "日志级别已更新",
            "bizName": bizName,
            "level":   req.Level,
        })
    })

    // 列出所有日志实例
    r.GET("/admin/log/list", func(c *gin.Context) {
        names := manager.List()
        c.JSON(http.StatusOK, gin.H{"loggers": names})
    })

    r.Run(":8080")
}
```

**使用示例**：

```bash
# 查看 app 日志的当前级别
curl http://localhost:8080/admin/log/level/app

# 临时开启 debug 日志
curl -X PUT http://localhost:8080/admin/log/level/app \
  -H "Content-Type: application/json" \
  -d '{"level":"debug"}'

# 调试完成后恢复 info 级别
curl -X PUT http://localhost:8080/admin/log/level/app \
  -H "Content-Type: application/json" \
  -d '{"level":"info"}'

# 查看所有日志实例
curl http://localhost:8080/admin/log/list
```

## 错误处理

### 错误类型

```go
// 预定义的哨兵错误
var (
    ErrInvalidLogLevel    = errors.New("invalid log level")
    ErrInvalidLogFormat   = errors.New("invalid log format")
    ErrEmptyBizName       = errors.New("bizName cannot be empty")
    ErrEmptyLogDir        = errors.New("log directory cannot be empty")
    ErrInvalidConfigValue = errors.New("invalid config value")
    ErrLoggerNotFound     = errors.New("logger not found")
)
```

### 错误检查函数

```go
// 检查具体的错误类型
if log.IsInvalidLogLevel(err) {
    // 处理无效日志级别错误
}

if log.IsInvalidLogFormat(err) {
    // 处理无效日志格式错误
}

if log.IsEmptyBizName(err) {
    // 处理空业务名称错误
}

if log.IsEmptyLogDir(err) {
    // 处理空日志目录错误
}

if log.IsInvalidConfigValue(err) {
    // 处理无效配置值错误
}

if log.IsLoggerNotFound(err) {
    // 处理 logger 不存在错误
}
```

### 错误处理示例

```go
logger, err := manager.Get("app")
if err != nil {
    switch {
    case log.IsEmptyBizName(err):
        fmt.Println("业务名称不能为空")
    case log.IsLoggerNotFound(err):
        fmt.Println("日志实例不存在")
    default:
        fmt.Printf("获取日志实例失败: %v\n", err)
    }
    return
}

// 设置日志级别时的错误处理
err = manager.SetLevel("app", "invalid_level")
if err != nil {
    if log.IsInvalidLogLevel(err) {
        fmt.Println("日志级别必须是: debug, info, warn, error 之一")
    } else if log.IsLoggerNotFound(err) {
        fmt.Println("日志实例不存在，请先调用 Get() 创建")
    }
}
```

## 性能优化建议

### 1. 合理选择日志级别

- **生产环境**：建议使用 `info` 或 `warn` 级别
- **开发环境**：可以使用 `debug` 级别
- **线上调试**：使用动态级别调整功能，临时开启 `debug`

### 2. 避免过度日志

```go
// ❌ 不推荐：在高频循环中记录 debug 日志
for i := 0; i < 10000; i++ {
    logger.Debug("处理项目", zap.Int("index", i))
}

// ✅ 推荐：记录关键信息或定期采样
logger.Info("开始批量处理", zap.Int("total", 10000))
for i := 0; i < 10000; i++ {
    // 处理逻辑
    if i%1000 == 0 {
        logger.Debug("处理进度", zap.Int("processed", i))
    }
}
logger.Info("批量处理完成")
```

### 3. 使用结构化字段

```go
// ❌ 不推荐：字符串拼接
logger.Info(fmt.Sprintf("用户 %s 从 %s 登录", username, ip))

// ✅ 推荐：使用结构化字段
logger.Info("用户登录",
    zap.String("username", username),
    zap.String("ip", ip),
)
```

### 4. 合理配置日志切分

```go
cfg := log.Config{
    MaxSize:    100,  // 100MB 切分一次，避免单个文件过大
    MaxBackups: 30,   // 保留 30 个备份，根据磁盘空间调整
    MaxAge:     7,    // 保留 7 天，过期自动删除
    Compress:   true, // 启用压缩，节省磁盘空间
}
```

## 常见问题

### Q1: 日志文件在哪里？

**A**: 日志文件存储在配置的 `Dir` 目录下，文件名格式为 `{bizName}.log`。例如：
- `./logs/app.log` - 应用日志
- `./logs/api.log` - API 日志
- `./logs/db.log` - 数据库日志

### Q2: 如何禁用控制台输出？

**A**: 将配置中的 `Console` 设置为 `false`：

```go
cfg := log.Config{
    Dir:     "./logs",
    Console: false,  // 不输出到控制台
}
```

### Q3: 日志文件什么时候会切分？

**A**: 当日志文件大小超过 `MaxSize`（MB）时自动切分，切分后的文件名格式为：
- `app.log` - 当前日志文件
- `app-2024-01-10T13-55-36.123.log` - 旧日志文件（带时间戳）
- `app-2024-01-10T13-55-36.123.log.gz` - 压缩后的旧日志文件

### Q4: 如何查看已压缩的日志？

**A**: 使用 `zcat` 或 `gzip -d` 命令：

```bash
# 查看压缩日志内容
zcat app-2024-01-10T13-55-36.123.log.gz

# 或者解压缩
gzip -d app-2024-01-10T13-55-36.123.log.gz
```

### Q5: 为什么 Sync() 返回错误？

**A**: 在某些操作系统上，对 `stdout`/`stderr` 调用 `Sync()` 可能会返回错误，这是正常的。Manager 的 `Sync()` 和 `Close()` 方法已经自动忽略这类错误。

### Q6: 多个业务模块的日志会混在一起吗？

**A**: 不会。每个业务模块（bizName）都有独立的日志文件，不会混在一起。

### Q7: 如何在单元测试中使用日志？

**A**: 在测试中可以使用临时目录：

```go
func TestMyFunction(t *testing.T) {
    // 创建临时日志目录
    tmpDir := t.TempDir()
    
    cfg := log.Config{
        Dir:     tmpDir,
        Level:   "debug",
        Console: false,  // 测试时不输出到控制台
    }
    
    manager, err := log.NewManager(cfg)
    require.NoError(t, err)
    defer manager.Close()
    
    logger := manager.MustGet("test")
    // 使用 logger 进行测试
}
```

### Q8: 如何集成到现有的框架中？

**A**: 参考"最佳实践"章节中的 HTTP 服务示例，可以很容易地集成到 Gin、Echo 等框架中。

## API 参考

### Manager 方法

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `NewManager(cfg)` | 创建新的日志管理器 | `(*Manager, error)` |
| `MustNewManager(cfg)` | 创建管理器，失败时 panic | `*Manager` |
| `Init(cfg)` | 初始化全局默认 Manager | - |
| `Default()` | 获取全局默认 Manager | `*Manager` |
| `Get(bizName)` | 获取日志实例 | `(*zap.Logger, error)` |
| `MustGet(bizName)` | 获取日志实例，失败时 panic | `*zap.Logger` |
| `SetLevel(bizName, level)` | 设置日志级别 | `error` |
| `GetLevel(bizName)` | 获取日志级别 | `(string, error)` |
| `List()` | 列出所有日志实例 | `[]string` |
| `Remove(bizName)` | 移除日志实例 | `error` |
| `Sync()` | 同步所有日志 | `error` |
| `Close()` | 关闭管理器 | `error` |

### 工具函数

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `Data(x)` | 创建数据字段 | `zap.Field` |
| `NewZapLogger(cfg, bizName)` | 创建 zap 日志实例 | `(*zap.Logger, zap.AtomicLevel, error)` |

### Config 验证

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `Validate()` | 验证配置有效性 | `error` |

## 相关链接

- [Zap 官方文档](https://pkg.go.dev/go.uber.org/zap)
- [Lumberjack 文档](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2)
- [项目主页](https://github.com/qq1060656096/drugo)

## License

本项目使用 [LICENSE](../LICENSE) 许可证。





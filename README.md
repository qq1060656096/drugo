# Drugo

<p align="center">
  <strong>一个轻量级、模块化的 Go 应用程序框架</strong>
</p>

<p align="center">
  <a href="#特性">特性</a> •
  <a href="#安装">安装</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#核心概念">核心概念</a> •
  <a href="#架构设计">架构设计</a> •
  <a href="#内置服务">内置服务</a> •
  <a href="#示例项目">示例项目</a>
</p>

---

## 简介

Drugo 是一个基于 Go 语言的轻量级应用程序框架，专注于提供**服务容器**、**生命周期管理**、**配置管理**和**日志管理**等核心能力。它遵循"约定优于配置"的原则，让开发者能够快速构建结构清晰、易于维护的应用程序。

## 特性

- 🚀 **服务容器** - 基于泛型的依赖注入容器，支持服务绑定与获取
- 📦 **生命周期管理** - 完整的 Boot → Run → Shutdown 服务生命周期
- ⚡ **优雅停机** - 内置信号监听，支持可配置的超时时间
- 🔧 **配置管理** - 基于 Viper，支持 YAML 配置、多业务配置、热加载
- 📝 **日志管理** - 基于 Zap，支持多业务日志、动态级别调整、日志切分
- 🔌 **可扩展** - 通过实现 `Service` 或 `Runner` 接口轻松扩展
- 🛡️ **类型安全** - 利用 Go 泛型提供类型安全的服务获取

## 安装

### 安装框架库

```bash
go get github.com/qq1060656096/drugo
```

### 快速使用（安装 CLI 工具）

```bash
go install github.com/qq1060656096/drugo/cmd/drugo@latest
```

安装完成后，可以使用 `drugo` 命令快速创建项目：

```bash
# 创建新项目
drugo new myapp

# 进入项目目录
cd myapp

# 安装依赖
go mod tidy

# 运行项目
make run

# 创建新模块 (在项目根目录下)
drugo module new user

# 创建新的 API 结构 (在模块目录下)
drugo module new-api user address
```

**要求**：Go 1.25.0 或更高版本

## 快速开始

### 1. 项目结构

推荐的项目结构：

```
myapp/
├── cmd/
│   └── app/
│       └── main.go           # 应用入口
├── conf/
│   ├── gin.yaml              # Gin 服务配置
│   └── log.yaml              # 日志配置
├── internal/
│   └── user/                 # 业务模块
│       ├── api/              # API 层
│       ├── biz/              # 业务逻辑层
│       ├── data/             # 数据访问层
│       └── service/          # 服务层
├── runtime/
│   └── logs/                 # 日志目录
├── go.mod
└── go.sum
```

### 2. 最小示例

```go
package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo/drugo"
	"github.com/qq1060656096/drugo/pkg/gomod"
	"github.com/qq1060656096/drugo/pkg/router"
	"github.com/qq1060656096/drugo-provider/ginsrv"
	"go.uber.org/zap"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	root := gomod.GmodRoot(wd)
	ctx := context.Background()

	// 创建应用
	app := drugo.MustNewApp(
		drugo.WithContext(ctx),
		drugo.WithRoot(root),
		drugo.WithService(ginsrv.New()),
	)

	// 获取 Gin 服务并添加路由
	ginSvc := drugo.MustGetService[*ginsrv.GinService](app, "gin")
	engine := ginSvc.Engine()

	// 手动添加路由
	router.Default().Register(func(r *gin.Engine) {
		r.GET("/hello", func(c *gin.Context) {
			app.Logger().MustGet("gin").Info("hello world",
				zap.String("url", c.Request.URL.String()),
			)
			c.JSON(200, gin.H{"message": "hello world"})
		})
	})

	// 设置所有注册的路由
	router.Default().Setup(engine)

	// 启动应用
	if err := app.Serve(ctx); err != nil {
		panic(err)
	}
}
```

### 3. 配置文件

**conf/gin.yaml**:

```yaml
gin:
  mode: release
  host: "0.0.0.0"
  shutdown_timeout: 30s
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s
  http:
    enabled: true
    port: 8080
  https:
    enabled: false
    port: 443
    cert_file: "./cert/server.crt"
    key_file: "./cert/server.key"
```

**conf/log.yaml**:

```yaml
log:
  level: "info"
  format: "json"
  max_size: 100
  max_backups: 30
  max_age: 7
  compress: true
  console: true
```

## 核心概念

### Service 接口

`Service` 是 Drugo 中最基本的服务单元，定义了服务的基本生命周期：

```go
type Service interface {
    Name() string                    // 服务名称
    Boot(ctx context.Context) error  // 初始化
    Close(ctx context.Context) error // 关闭
}
```

### Runner 接口

`Runner` 扩展了 `Service`，用于需要长期运行的服务（如 HTTP Server）：

```go
type Runner interface {
    Service
    Run(ctx context.Context) error   // 运行（阻塞直到上下文取消）
}
```

### 服务容器

服务容器负责管理所有服务实例，支持按名称绑定和获取：

```go
// 绑定服务
app.Container().Bind("myservice", myService)

// 获取服务
svc, err := app.Container().Get("myservice")

// 类型安全的获取
ginSvc := drugo.MustGetService[*ginsrv.GinService](app, "gin")
```

### 生命周期

Drugo 应用的完整生命周期：

```
┌─────────────────────────────────────────────────────────────┐
│                         Serve()                              │
├─────────────────────────────────────────────────────────────┤
│  1. Boot()     → 按注册顺序初始化所有服务                      │
│  2. Run()      → 并发启动所有 Runner 服务                      │
│  3. 信号监听    → 等待 SIGINT/SIGTERM                         │
│  4. Shutdown() → 逆序关闭所有服务（带超时控制）                 │
└─────────────────────────────────────────────────────────────┘
```

## 架构设计

### 模块结构

```
drugo/
├── kernel/          # 核心接口定义
│   ├── kernel.go    # Kernel 接口
│   ├── service.go   # Service/Runner 接口
│   ├── container.go # Container 接口
│   ├── context.go   # 上下文工具
│   └── error.go     # 错误定义
│
├── drugo/           # 框架实现
│   ├── drugo.go     # Drugo 核心实现
│   ├── container.go # 容器实现
│   ├── options.go   # 选项模式
│   └── service.go   # 服务工具函数
│
├── config/          # 配置管理
│   ├── manager.go   # 配置管理器
│   └── error.go     # 错误定义
│
├── log/             # 日志管理
│   ├── manager.go   # 日志管理器
│   ├── config.go    # 日志配置
│   └── log.go       # Zap 日志创建
│
└── pkg/             # 工具包
    ├── router/      # 路由注册表
    └── gomod/       # Go Module 工具
```

### 核心流程图

```
                    ┌──────────────────┐
                    │   MustNewApp()   │
                    └────────┬─────────┘
                             │
                             ▼
            ┌────────────────────────────────┐
            │  初始化配置管理器 (Config)       │
            │  初始化日志管理器 (Logger)       │
            │  注册用户服务到容器             │
            └────────────────┬───────────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │    Serve()     │
                    └────────┬───────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
    ┌─────────┐        ┌─────────┐        ┌──────────────┐
    │  Boot() │   →    │  Run()  │   →    │  Shutdown()  │
    └─────────┘        └─────────┘        └──────────────┘
         │                   │                   │
    按顺序初始化        并发运行 Runner      逆序关闭服务
    所有服务            阻塞等待            带超时控制
```

## 配置管理

配置管理器基于 [Viper](https://github.com/spf13/viper) 构建，提供：

- ✅ 多配置文件合并
- ✅ 按业务名称获取配置
- ✅ 配置热加载
- ✅ 重载回调机制

### 使用示例

```go
// 获取配置管理器
cfg := app.Config()

// 获取指定业务配置
ginConfig, err := cfg.Get("gin")
if err != nil {
    // 处理错误
}

// 读取配置值
port := ginConfig.GetInt("http.port")
host := ginConfig.GetString("host")

// 解析到结构体
var config GinConfig
ginConfig.Unmarshal(&config)

// 监听配置变化（热加载）
cfg.OnReload(func(m *config.Manager) error {
    log.Println("配置已重载")
    return nil
})
cfg.Watch()
```

详细文档请参阅 [config/README.md](./config/README.md)

## 日志管理

日志管理器基于 [Zap](https://github.com/uber-go/zap) 和 [Lumberjack](https://github.com/natefinch/lumberjack) 构建，提供：

- ✅ 多业务日志实例
- ✅ 动态级别调整
- ✅ 日志自动切分与压缩
- ✅ JSON/Console/Text 多种格式

### 使用示例

```go
// 获取日志管理器
logger := app.Logger()

// 获取指定业务的日志实例
appLog := logger.MustGet("app")
apiLog := logger.MustGet("api")

// 记录日志
appLog.Info("应用启动",
    zap.String("version", "1.0.0"),
    zap.Int("port", 8080),
)

apiLog.Info("收到请求",
    zap.String("method", "GET"),
    zap.String("path", "/api/users"),
)

// 动态调整日志级别（用于线上调试）
logger.SetLevel("app", "debug")
```

详细文档请参阅 [log/README.md](./log/README.md)

## 内置服务

### Gin HTTP 服务

内置的 Gin HTTP 服务提供：

- HTTP/HTTPS 双协议支持
- 可配置的超时时间
- 优雅停机

```go
import "github.com/qq1060656096/drugo-provider/ginsrv"

// 创建并注册 Gin 服务
app := drugo.MustNewApp(
    drugo.WithService(ginsrv.New()),
)

// 获取 Gin Engine 并添加路由
ginSvc := drugo.MustGetService[*ginsrv.GinService](app, "gin")
engine := ginSvc.Engine()

engine.GET("/hello", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "hello"})
})
```

### 自定义服务

实现 `Service` 或 `Runner` 接口来创建自定义服务：

```go
package myservice

import (
    "context"
    "github.com/qq1060656096/drugo/kernel"
)

var _ kernel.Service = (*MyService)(nil)

type MyService struct {
    name string
}

func (s *MyService) Name() string {
    return s.name
}

func (s *MyService) Boot(ctx context.Context) error {
    k := kernel.MustFromContext(ctx)
    logger := k.Logger().MustGet(s.Name())
    logger.Info("MyService booting")
    
    // 初始化逻辑...
    
    return nil
}

func (s *MyService) Close(ctx context.Context) error {
    k := kernel.MustFromContext(ctx)
    logger := k.Logger().MustGet(s.Name())
    logger.Info("MyService closing")
    
    // 清理逻辑...
    
    return nil
}

func New() *MyService {
    return &MyService{name: "myservice"}
}
```

如果服务需要持续运行（如消费者、定时任务），实现 `Runner` 接口：

```go
var _ kernel.Runner = (*MyWorker)(nil)

func (w *MyWorker) Run(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            // 工作逻辑...
        }
    }
}
```

## 路由注册

Drugo 提供了一个路由注册表，支持模块化路由管理：

```go
import "github.com/qq1060656096/drugo/pkg/router"

// 在模块的 init() 中注册路由
func init() {
    router.Default().Register(func(r *gin.Engine) {
        r.GET("/users", listUsers)
        r.POST("/users", createUser)
    })
}

// 在 main.go 中统一设置
func main() {
    app := drugo.MustNewApp(...)
    ginSvc := drugo.MustGetService[*ginsrv.GinService](app, "gin")
    
    // 执行所有注册的路由函数
    router.Default().Setup(ginSvc.Engine())
    
    app.Serve(ctx)
}
```

## 上下文工具

Drugo 将 Kernel 实例注入到 Context 中，方便在任何地方访问：

```go
// 从 Context 获取 Kernel
k := kernel.MustFromContext(ctx)

// 访问配置
cfg := k.Config().MustGet("app")

// 访问日志
logger := k.Logger().MustGet("app")

// 获取服务
svc, err := kernel.ServiceFromContext[*MyService](ctx, "myservice")
```

## 示例项目

完整的示例项目请参阅 [drugo-app](https://github.com/qq1060656096/drugo-app)：

```go
package main

import (
    "context"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/qq1060656096/drugo/drugo"
    "github.com/qq1060656096/drugo/pkg/gomod"
    "github.com/qq1060656096/drugo/pkg/router"
    "github.com/qq1060656096/drugo-provider/ginsrv"
    "go.uber.org/zap"

    // 导入模块以触发 init() 自动注册路由
    _ "github.com/qq1060656096/drugo-app/internal/user/api"
)

func main() {
    wd, _ := os.Getwd()
    root := gomod.GmodRoot(wd)
    ctx := context.Background()
    
    // 创建应用
    app := drugo.MustNewApp(
        drugo.WithContext(ctx),
        drugo.WithRoot(root),
        drugo.WithService(ginsrv.New()),
    )
    
    // 获取 Gin 服务并添加路由
    ginSvc := drugo.MustGetService[*ginsrv.GinService](app, "gin")
    engine := ginSvc.Engine()

    // 手动添加路由
    router.Default().Register(func(r *gin.Engine) {
        r.GET("/hello", func(c *gin.Context) {
            app.Logger().MustGet("gin").Info("hello world",
                zap.String("url", c.Request.URL.String()),
            )
            c.JSON(200, gin.H{"message": "hello world"})
        })
    })
    
    // 设置所有注册的路由
    router.Default().Setup(engine)

    // 启动应用
    if err := app.Serve(ctx); err != nil {
        panic(err)
    }
}
```

## Options 模式

Drugo 使用 Options 模式进行灵活配置：

```go
app := drugo.MustNewApp(
    // 设置项目根目录
    drugo.WithRoot("/path/to/project"),
    
    // 设置上下文
    drugo.WithContext(ctx),
    
    // 注册服务（使用服务自身的 Name()）
    drugo.WithService(myService),
    
    // 注册服务（指定名称）
    drugo.WithNameService("custom-name", myService),
    
    // 设置优雅停机超时时间
    drugo.WithShutdownTimeout(30 * time.Second),
)
```

## API 参考

### Kernel 接口

| 方法 | 说明 |
|------|------|
| `Container()` | 返回服务容器 |
| `Boot(ctx)` | 引导所有服务 |
| `Run(ctx)` | 运行所有 Runner 服务 |
| `Shutdown(ctx)` | 关闭所有服务 |
| `Serve(ctx)` | 完整生命周期（Boot + Run + Shutdown） |
| `Root()` | 返回项目根目录 |
| `Config()` | 返回配置管理器 |
| `Logger()` | 返回日志管理器 |

### Container 接口

| 方法 | 说明 |
|------|------|
| `Bind(name, service)` | 绑定服务到容器 |
| `Get(name)` | 获取服务 |
| `MustGet(name)` | 获取服务（失败时 panic） |
| `Services()` | 返回所有服务 |
| `Names()` | 返回所有服务名称 |

### 工具函数

| 函数 | 说明 |
|------|------|
| `drugo.GetService[T](k, name)` | 类型安全地获取服务 |
| `drugo.MustGetService[T](k, name)` | 类型安全地获取服务（失败时 panic） |
| `kernel.FromContext(ctx)` | 从上下文获取 Kernel |
| `kernel.MustFromContext(ctx)` | 从上下文获取 Kernel（失败时 panic） |
| `kernel.ServiceFromContext[T](ctx, name)` | 从上下文获取服务 |

## 依赖

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [spf13/viper](https://github.com/spf13/viper) - 配置管理
- [uber-go/zap](https://github.com/uber-go/zap) - 高性能日志
- [natefinch/lumberjack](https://github.com/natefinch/lumberjack) - 日志切分
- [fsnotify/fsnotify](https://github.com/fsnotify/fsnotify) - 文件监听

## 开发

```bash
# 运行测试
make test

# 运行测试（包含竞态检测）
make testa

# 检查测试覆盖率
make cover
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

本项目采用 [LICENSE](./LICENSE) 许可证。

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/qq1060656096">qq1060656096</a>
</p>

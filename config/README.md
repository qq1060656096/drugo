# Config 包使用文档

## 概述

`config` 包提供了强大而灵活的配置管理功能，专为多业务配置场景设计。它基于 [viper](https://github.com/spf13/viper) 和 [fsnotify](https://github.com/fsnotify/fsnotify) 构建，支持从 YAML 文件加载配置、配置缓存、热加载以及自定义回调等功能。

## 主要特性

- ✅ **多业务配置管理**：支持在一个目录中管理多个业务的配置文件
- ✅ **自动合并**：自动加载并合并目录中所有 `.yml` 和 `.yaml` 文件
- ✅ **配置缓存**：使用双重检查锁定模式实现高效的配置缓存
- ✅ **线程安全**：所有操作都是并发安全的
- ✅ **热加载**：支持监听配置文件变化并自动重载
- ✅ **回调机制**：支持注册配置重载时的回调函数
- ✅ **全局实例**：提供便捷的全局默认 Manager
- ✅ **错误处理**：定义了清晰的错误类型，便于错误判断

## 快速开始

### 安装

```bash
go get github.com/qq1060656096/drugo/config
```

### 基本使用

假设你的配置文件目录结构如下：

```
conf/
├── app.yaml
├── database.yaml
└── redis.yaml
```

其中 `app.yaml` 内容为：

```yaml
app:
  name: my-app
  version: 1.0.0
  port: 8080
```

使用示例：

```go
package main

import (
    "fmt"
    "github.com/yourusername/drugo/config"
)

func main() {
    // 初始化全局 Manager
    config.Init("./conf")
    
    // 获取全局 Manager
    manager := config.Default()
    
    // 获取指定业务配置
    appConfig, err := manager.Get("app")
    if err != nil {
        panic(err)
    }
    
    // 读取配置值
    name := appConfig.GetString("name")
    port := appConfig.GetInt("port")
    
    fmt.Printf("App: %s, Port: %d\n", name, port)
}
```

## API 详解

### Manager 结构

`Manager` 是配置管理的核心结构，负责加载、缓存和管理所有业务配置。

#### 创建 Manager

##### NewManager

```go
func NewManager(configDir string) (*Manager, error)
```

创建一个新的 Manager 实例。如果配置目录不存在或读取失败，返回错误。

**参数：**
- `configDir`: 配置文件所在目录的路径

**返回：**
- `*Manager`: Manager 实例
- `error`: 错误信息

**示例：**

```go
manager, err := config.NewManager("./conf")
if err != nil {
    log.Fatalf("Failed to create manager: %v", err)
}
```

##### MustNewManager

```go
func MustNewManager(configDir string) *Manager
```

类似于 `NewManager`，但如果发生错误会 panic。适合在初始化阶段使用。

**示例：**

```go
manager := config.MustNewManager("./conf")
```

#### 全局 Manager

##### Init

```go
func Init(configDir string)
```

初始化全局默认 Manager。此函数是并发安全的，多次调用只会初始化一次。如果初始化失败会 panic。

**示例：**

```go
func main() {
    config.Init("./conf")
    // ... 其他代码
}
```

##### Default

```go
func Default() *Manager
```

返回全局默认 Manager 实例。如果 `Init` 未被调用，返回 `nil`。

**示例：**

```go
manager := config.Default()
if manager == nil {
    log.Fatal("Config manager not initialized")
}
```

### 配置获取

#### Get

```go
func (m *Manager) Get(name string) (*viper.Viper, error)
```

获取指定业务名称的配置。首次获取时会从根配置中提取并缓存，后续获取直接返回缓存的配置。此方法是线程安全的。

**参数：**
- `name`: 业务配置名称（对应配置文件中的顶级键）

**返回：**
- `*viper.Viper`: 配置实例
- `error`: 如果配置不存在，返回 `ErrNotFound`

**示例：**

```go
dbConfig, err := manager.Get("database")
if err != nil {
    if config.IsNotFound(err) {
        log.Println("Database config not found")
    }
    return err
}

host := dbConfig.GetString("host")
port := dbConfig.GetInt("port")
```

#### MustGet

```go
func (m *Manager) MustGet(name string) *viper.Viper
```

类似于 `Get`，但如果发生错误会 panic。适合在确定配置一定存在的情况下使用。

**示例：**

```go
appConfig := manager.MustGet("app")
appName := appConfig.GetString("name")
```

#### Root

```go
func (m *Manager) Root() *viper.Viper
```

返回包含所有业务配置的根配置实例。可以使用点号分隔符访问嵌套配置。

**示例：**

```go
root := manager.Root()
appName := root.GetString("app.name")
dbHost := root.GetString("database.host")
```

### 配置信息

#### List

```go
func (m *Manager) List() []string
```

返回所有的业务配置名称的有序列表（按字母顺序排序）。

**示例：**

```go
names := manager.List()
fmt.Println("Cached configs:", names) // 输出: [app, database]
```

#### AllNames

```go
func (m *Manager) AllNames() []string
```

返回根配置中所有可用业务配置名称的有序列表，无论它们是否已被加载到缓存中。

**示例：**

```go
allNames := manager.AllNames()
fmt.Println("All available configs:", allNames) // 输出: [app, cache, database, redis]
```

### 配置重载

#### Reset

```go
func (m *Manager) Reset() error
```

重新加载配置并清空所有缓存的业务配置。此方法是线程安全的。

**示例：**

```go
if err := manager.Reset(); err != nil {
    log.Printf("Failed to reset config: %v", err)
}
```

#### OnReload

```go
func (m *Manager) OnReload(callback ReloadCallback)
```

注册配置重载时的回调函数。回调函数会在配置文件变化并成功重载后被调用。可以注册多个回调函数。

**回调函数签名：**

```go
type ReloadCallback func(m *Manager) error
```

**示例：**

```go
manager.OnReload(func(m *Manager) error {
    log.Println("Config reloaded!")
    
    // 重新获取配置
    appConfig := m.MustGet("app")
    newPort := appConfig.GetInt("port")
    
    // 执行一些更新操作
    // ...
    
    return nil
})
```

### 热加载

#### Watch

```go
func (m *Manager) Watch() error
```

启动配置文件的热加载监听。当配置目录中的 `.yml` 或 `.yaml` 文件发生变化时，会自动重新加载配置并调用所有注册的回调函数。此方法是幂等的，多次调用只会启动一次监听。

**示例：**

```go
if err := manager.Watch(); err != nil {
    log.Fatalf("Failed to start watching: %v", err)
}
defer manager.StopWatch()
```

#### StopWatch

```go
func (m *Manager) StopWatch()
```

停止配置文件的热加载监听。此方法是幂等的，多次调用是安全的。

**示例：**

```go
manager.StopWatch()
```

### 错误处理

包定义了以下错误类型：

```go
var (
    ErrNotFound     = errors.New("config: not found")
    ErrDirRead      = errors.New("config: directory read failed")
    ErrFileRead     = errors.New("config: file read failed")
    ErrDuplicateKey = errors.New("config: duplicate key")
)
```

提供了相应的错误判断函数：

```go
func IsNotFound(err error) bool
func IsDirRead(err error) bool
func IsFileRead(err error) bool
func IsDuplicateKey(err error) bool
```

**示例：**

```go
config, err := manager.Get("nonexistent")
if err != nil {
    if config.IsNotFound(err) {
        log.Println("Config not found, using defaults")
        // 使用默认配置
    } else {
        return err
    }
}
```

## 完整示例

### 示例 1：基本配置管理

```go
package main

import (
    "fmt"
    "log"
    "github.com/yourusername/drugo/config"
)

func main() {
    // 初始化配置管理器
    config.Init("./conf")
    manager := config.Default()
    
    // 获取应用配置
    appConfig := manager.MustGet("app")
    fmt.Printf("App Name: %s\n", appConfig.GetString("name"))
    fmt.Printf("App Port: %d\n", appConfig.GetInt("port"))
    
    // 获取数据库配置
    dbConfig, err := manager.Get("database")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("DB Host: %s\n", dbConfig.GetString("host"))
    fmt.Printf("DB Port: %d\n", dbConfig.GetInt("port"))
    
    // 查看所有可用配置
    allNames := manager.AllNames()
    fmt.Printf("Available configs: %v\n", allNames)
}
```

### 示例 2：配置热加载

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "github.com/yourusername/drugo/config"
)

func main() {
    // 初始化配置
    config.Init("./conf")
    manager := config.Default()
    
    // 注册重载回调
    manager.OnReload(func(m *config.Manager) error {
        fmt.Println("=== Config Reloaded ===")
        
        // 重新读取配置
        appConfig := m.MustGet("app")
        fmt.Printf("New App Port: %d\n", appConfig.GetInt("port"))
        
        return nil
    })
    
    // 启动热加载
    if err := manager.Watch(); err != nil {
        log.Fatal(err)
    }
    defer manager.StopWatch()
    
    fmt.Println("Watching for config changes... Press Ctrl+C to exit")
    
    // 等待退出信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    
    fmt.Println("\nShutting down...")
}
```

### 示例 3：多个回调处理

```go
package main

import (
    "fmt"
    "log"
    "github.com/yourusername/drugo/config"
)

// 服务配置
type ServiceConfig struct {
    Port int
    Host string
}

var currentConfig ServiceConfig

func main() {
    config.Init("./conf")
    manager := config.Default()
    
    // 初始化配置
    loadServiceConfig(manager)
    
    // 注册多个回调
    manager.OnReload(func(m *config.Manager) error {
        fmt.Println("Callback 1: Reloading service config")
        loadServiceConfig(m)
        return nil
    })
    
    manager.OnReload(func(m *config.Manager) error {
        fmt.Println("Callback 2: Notifying components")
        // 通知其他组件配置已更新
        notifyComponents()
        return nil
    })
    
    manager.OnReload(func(m *config.Manager) error {
        fmt.Println("Callback 3: Logging config change")
        log.Printf("Config updated: %+v", currentConfig)
        return nil
    })
    
    // 启动热加载
    if err := manager.Watch(); err != nil {
        log.Fatal(err)
    }
    defer manager.StopWatch()
    
    // 程序继续运行...
    select {}
}

func loadServiceConfig(m *config.Manager) {
    appConfig := m.MustGet("app")
    currentConfig = ServiceConfig{
        Port: appConfig.GetInt("port"),
        Host: appConfig.GetString("host"),
    }
}

func notifyComponents() {
    // 通知其他组件
    fmt.Println("Notifying all components about config change...")
}
```

### 示例 4：错误处理

```go
package main

import (
    "fmt"
    "log"
    "github.com/yourusername/drugo/config"
)

func main() {
    // 尝试创建 Manager
    manager, err := config.NewManager("./conf")
    if err != nil {
        if config.IsDirRead(err) {
            log.Fatal("Config directory not found or cannot be read")
        } else if config.IsFileRead(err) {
            log.Fatal("Failed to read config file")
        } else if config.IsDuplicateKey(err) {
            log.Fatal("Duplicate keys found in config files")
        } else {
            log.Fatalf("Unknown error: %v", err)
        }
    }
    
    // 尝试获取配置
    dbConfig, err := manager.Get("database")
    if err != nil {
        if config.IsNotFound(err) {
            fmt.Println("Database config not found, using defaults")
            // 使用默认配置
            dbConfig = getDefaultDBConfig()
        } else {
            log.Fatal(err)
        }
    }
    
    // 使用配置
    fmt.Printf("DB Host: %s\n", dbConfig.GetString("host"))
}

func getDefaultDBConfig() *viper.Viper {
    // 返回默认配置
    // ...
    return nil
}
```

## 配置文件组织

### 推荐的目录结构

```
conf/
├── app.yaml          # 应用配置
├── database.yaml     # 数据库配置
├── redis.yaml        # Redis 配置
├── kafka.yaml        # Kafka 配置
└── log.yaml          # 日志配置
```

### 配置文件格式

每个配置文件的顶级键代表一个业务配置。例如：

**app.yaml:**

```yaml
app:
  name: my-application
  version: 1.0.0
  port: 8080
  debug: false
```

**database.yaml:**

```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  username: admin
  password: secret
  dbname: mydb
  max_connections: 100
```

**redis.yaml:**

```yaml
redis:
  host: localhost
  port: 6379
  password: ""
  database: 0
  pool_size: 10
```

### 避免重复键

在不同的配置文件中使用相同的顶级键会导致错误。例如：

❌ **错误示例：**

**app.yaml:**
```yaml
database:
  host: localhost
```

**db.yaml:**
```yaml
database:  # 错误：与 app.yaml 中的 database 键重复
  host: remote
```

✅ **正确示例：**

每个配置文件使用唯一的顶级键，或将相同业务的配置合并到一个文件中。

## 最佳实践

### 1. 使用全局 Manager

对于大多数应用，使用全局 Manager 是最简单的方式：

```go
func init() {
    config.Init("./conf")
}

func GetAppConfig() *viper.Viper {
    return config.Default().MustGet("app")
}
```

### 2. 封装配置结构

将配置解析为结构体，便于类型安全地使用：

```go
type AppConfig struct {
    Name    string `mapstructure:"name"`
    Version string `mapstructure:"version"`
    Port    int    `mapstructure:"port"`
    Debug   bool   `mapstructure:"debug"`
}

func LoadAppConfig(manager *config.Manager) (*AppConfig, error) {
    vCfg := manager.MustGet("app")
    
    var cfg AppConfig
    if err := vCfg.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

### 3. 使用回调更新运行时配置

在配置重载时更新应用的运行时配置：

```go
var (
    appConfig *AppConfig
    configMu  sync.RWMutex
)

func main() {
    config.Init("./conf")
    manager := config.Default()
    
    // 初始化配置
    updateAppConfig(manager)
    
    // 注册重载回调
    manager.OnReload(func(m *config.Manager) error {
        updateAppConfig(m)
        return nil
    })
    
    manager.Watch()
    defer manager.StopWatch()
    
    // ...
}

func updateAppConfig(manager *config.Manager) {
    cfg, err := LoadAppConfig(manager)
    if err != nil {
        log.Printf("Failed to load app config: %v", err)
        return
    }
    
    configMu.Lock()
    appConfig = cfg
    configMu.Unlock()
}

func GetCurrentAppConfig() *AppConfig {
    configMu.RLock()
    defer configMu.RUnlock()
    return appConfig
}
```

### 4. 处理配置重载错误

在回调中妥善处理错误，避免影响其他回调的执行：

```go
manager.OnReload(func(m *config.Manager) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Config reload panic: %v", r)
        }
    }()
    
    // 执行重载逻辑
    if err := reloadComponents(m); err != nil {
        log.Printf("Failed to reload components: %v", err)
        // 返回错误会被记录，但不会停止其他回调
        return err
    }
    
    return nil
})
```

### 5. 优雅关闭

在应用退出时停止配置监听：

```go
func main() {
    config.Init("./conf")
    manager := config.Default()
    
    if err := manager.Watch(); err != nil {
        log.Fatal(err)
    }
    
    // 使用 defer 确保监听被停止
    defer manager.StopWatch()
    
    // 或使用信号处理
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        manager.StopWatch()
        os.Exit(0)
    }()
    
    // 应用逻辑...
}
```

### 6. 环境特定配置

可以根据环境变量加载不同的配置目录：

```go
func main() {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }
    
    configDir := fmt.Sprintf("./conf/%s", env)
    config.Init(configDir)
    
    // ...
}
```

目录结构：
```
conf/
├── development/
│   ├── app.yaml
│   └── database.yaml
├── production/
│   ├── app.yaml
│   └── database.yaml
└── test/
    ├── app.yaml
    └── database.yaml
```

## 常见问题

### Q1: 如何判断配置是否已初始化？

```go
manager := config.Default()
if manager == nil {
    log.Fatal("Config not initialized. Call config.Init() first.")
}
```

### Q2: 如何手动触发配置重载？

```go
if err := manager.Reset(); err != nil {
    log.Printf("Failed to reload config: %v", err)
}
```

### Q3: 热加载是否会影响性能？

配置的热加载使用了 fsnotify 进行文件监听，性能开销很小。配置的读取使用了缓存机制，只有在配置重载时才会重新加载文件。

### Q4: 多个回调的执行顺序是什么？

回调按照注册的顺序依次执行。如果某个回调返回错误，错误会被记录，但不会阻止后续回调的执行。

### Q5: 是否支持配置的动态添加和删除？

目前不支持动态添加配置文件。如果需要添加新的配置文件，需要调用 `Reset()` 方法重新加载所有配置。

### Q6: 如何避免配置键冲突？

确保不同配置文件中的顶级键是唯一的。如果检测到重复键，`NewManager` 会返回 `ErrDuplicateKey` 错误。

### Q7: 是否支持配置文件的嵌套目录？

目前只支持指定目录下的直接配置文件，不会递归扫描子目录。

### Q8: 配置文件必须是 YAML 格式吗？

是的，目前只支持 `.yml` 和 `.yaml` 格式的配置文件。如果需要支持其他格式，可以在加载配置前手动转换。

## 性能优化

### 配置缓存

`Manager` 使用双重检查锁定模式实现了高效的配置缓存：

- 首次获取配置时，会从根配置中提取并缓存
- 后续获取直接返回缓存的配置，无需重复解析
- 使用读写锁保证并发安全，读操作可以并发执行

### 并发性能

基准测试显示：

- 缓存命中的单次 `Get` 操作耗时约 10-20 纳秒
- 并发访问性能优异，支持数千个并发 goroutine

```go
// 基准测试结果示例
BenchmarkManager_Get_Cached-8        100000000    10.5 ns/op
BenchmarkManager_Get_Concurrent-8     50000000    25.3 ns/op
```

## 线程安全

所有 `Manager` 的方法都是线程安全的：

- `Get` / `MustGet`: 使用双重检查锁定
- `Reset`: 使用写锁保护
- `OnReload`: 使用写锁保护
- `Watch` / `StopWatch`: 使用锁和 `sync.Once` 保证幂等性

可以安全地在多个 goroutine 中并发调用这些方法。

## 依赖

- [github.com/spf13/viper](https://github.com/spf13/viper) - 配置解析
- [github.com/fsnotify/fsnotify](https://github.com/fsnotify/fsnotify) - 文件监听

## 许可证

请查看项目根目录的 LICENSE 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

---

如有任何问题或建议，请联系项目维护者。


# log

## 简介

`log` 包是一个基于 [uber-go/zap](https://github.com/uber-go/zap) 与 [lumberjack](https://github.com/natefinch/lumberjack) 的日志管理模块。

它的核心目标是：

- **多业务日志隔离**：不同 `bizName` 对应不同 `*zap.Logger`（文件名为 `${bizName}.log`）。
- **多输出（Outputs）模型**：同一个 logger 可以同时输出到 `console` 与 `file`，且每个输出可独立指定格式（`json` / `text`）。
- **运行时动态调整级别**：通过 `SetLevel` 修改指定业务 logger 的 `zap.AtomicLevel`。
- **并发安全**：`Manager` 内部缓存与创建逻辑支持并发调用。

## 安装

```bash
go get github.com/qq1060656096/drugo/log
```

## 快速开始

### 1. 创建 Manager 并获取业务 logger

```go
package main

import (
	"github.com/qq1060656096/drugo/log"
	"go.uber.org/zap"
)

func main() {
	cfg := log.Config{
		Level: "info",
		Outputs: []log.OutputConfig{
			{
				Type:   log.OutputTypeConsole,
				Format: log.FormatText,
			},
			{
				Type:   log.OutputTypeFile,
				Format: log.FormatJSON,
				File: &log.FileOutputConfig{
					Dir:        "./runtime/logs",
					MaxSize:    100,
					MaxBackups: 10,
					MaxAge:     30,
					Compress:   true,
				},
			},
		},
	}

	m, err := log.NewManager(cfg)
	if err != nil {
		panic(err)
	}
	defer m.Close()

	logger, err := m.Get("app")
	if err != nil {
		panic(err)
	}

	logger.Info("启动成功", zap.String("version", "1.0.0"))
}
```

### 2. 使用全局默认 Manager

`Init` 会通过 `sync.Once` **只初始化一次**，`Default` 在未初始化时会返回 `nil`。

```go
package main

import "github.com/qq1060656096/drugo/log"

func main() {
	log.Init(log.Config{
		Level: "info",
		Outputs: []log.OutputConfig{
			{Type: log.OutputTypeConsole},
		},
	})

	logger := log.Default().MustGet("app")
	logger.Info("hello")
}
```

## 配置说明

### Config

```go
type Config struct {
	Level   string         `yaml:"level" mapstructure:"level"`
	Outputs []OutputConfig `yaml:"outputs" mapstructure:"outputs"`
}
```

- **Level**
  - 为空时默认 `info`
  - 由 `zap.ParseAtomicLevel` 解析
- **Outputs**
  - 必填，不能为空，否则返回 `ErrEmptyLogOutputs`
  - 每个输出独立配置 `type` / `format` / `file`

### OutputConfig

```go
type OutputConfig struct {
	Type   string            `yaml:"type" mapstructure:"type"`     // console, file
	Format string            `yaml:"format" mapstructure:"format"` // json, text
	File   *FileOutputConfig `yaml:"file,omitempty" mapstructure:"file,omitempty"`
}
```

- **Type**
  - `console`：输出到 `os.Stdout` / `os.Stderr`（error 及以上走 stderr）
  - `file`：输出到文件（使用 lumberjack 进行滚动）
- **Format**
  - `json` / `text`
  - 为空时默认 `text`
- **File**
  - 仅当 `Type=file` 时需要
  - `Type=console` 且 `File!=nil` 会被判定为 `ErrInvalidConfigValue`

### FileOutputConfig

```go
type FileOutputConfig struct {
	Dir        string `yaml:"dir" mapstructure:"dir"`
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"`
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`
	Compress   bool   `yaml:"compress" mapstructure:"compress"`
}
```

- **Dir**
  - 必填，否则返回 `ErrEmptyLogDir`
- **MaxSize / MaxBackups / MaxAge**
  - 不能为负数，否则返回 `ErrInvalidConfigValue`
  - 为 0 时会在 `Validate()` 中自动填默认值：
    - `MaxSize=100`
    - `MaxBackups=10`
    - `MaxAge=30`

### YAML 示例

```yaml
log: # 日志模块配置
  level: info # 全局日志级别，可选值：debug / info / warn / error / dpanic / panic / fatal
  outputs: # 输出目标列表，可配置多个输出，支持 outputs.console 和 outputs.file
    - type: console        # 控制台输出
      format: text         # 输出格式，可选值：json / text

    - type: file           # 文件输出，支持切分与保留策略
      format: json         # 输出格式，可选值：json / text
      file:                # 文件输出详细配置
        dir: logs          # 日志存放目录
        max_size: 100      # 单个日志文件最大尺寸（MB）
        max_backups: 10    # 最大保留的旧文件数量
        max_age: 30        # 最大保留天数
        compress: true     # 是否压缩旧日志（gzip）
```

## 核心概念

### bizName

`Manager.Get(bizName)` 会为每个 `bizName` 创建并缓存独立 logger：

- 日志文件名：`${dir}/${bizName}.log`
- logger 默认携带字段：`biz=<bizName>`（由 `NewZapLogger` 注入）

### 动态日志级别

`SetLevel` / `GetLevel` 依赖内部缓存的 `zap.AtomicLevel`，因此：

- 你必须先调用一次 `Get(bizName)`（或 `MustGet`）创建该业务 logger
- 否则会返回 `ErrLoggerNotFound`

## 错误处理

`log` 包导出了哨兵错误与判断函数，便于外部精确处理：

- `ErrInvalidLogLevel` / `IsInvalidLogLevel`
- `ErrInvalidLogFormat` / `IsInvalidLogFormat`
- `ErrInvalidOutputType` / `IsInvalidOutputType`
- `ErrEmptyBizName` / `IsEmptyBizName`
- `ErrEmptyLogDir` / `IsEmptyLogDir`
- `ErrEmptyLogOutputs` / `IsEmptyLogOutputs`
- `ErrInvalidConfigValue` / `IsInvalidConfigValue`
- `ErrLoggerNotFound` / `IsLoggerNotFound`

## API 参考

### 创建与获取

| API | 说明 |
| --- | --- |
| `NewManager(cfg)` | 创建 `Manager`（会执行 `cfg.Validate()`） |
| `MustNewManager(cfg)` | 创建失败时 `panic` |
| `Init(cfg)` | 初始化全局默认 `Manager`（只执行一次） |
| `Default()` | 获取全局默认 `Manager`（未初始化返回 `nil`） |
| `(*Manager).Get(bizName)` | 获取/创建业务 logger（缓存） |
| `(*Manager).MustGet(bizName)` | 获取失败时 `panic` |

### 生命周期与管理

| API | 说明 |
| --- | --- |
| `(*Manager).Sync()` | 调用所有 logger 的 `Sync()`（会忽略 stdout/stderr 的 sync 错误） |
| `(*Manager).Close()` | 同步并清空缓存（之后再次 `Get` 会创建新实例） |
| `(*Manager).List()` | 列出已创建的 `bizName` |
| `(*Manager).Remove(bizName)` | 移除指定业务 logger（会先 `Sync()`） |

### 级别控制

| API | 说明 |
| --- | --- |
| `(*Manager).SetLevel(bizName, level)` | 动态更新业务 logger 级别（logger 未创建时返回 `ErrLoggerNotFound`） |
| `(*Manager).GetLevel(bizName)` | 获取业务 logger 当前级别 |

### 辅助函数

| API | 说明 |
| --- | --- |
| `Data(x)` | `zap.Any("data", x)` 的便捷封装 |
| `NewZapLogger(cfg, bizName)` | 创建底层 `zap.Logger`（一般不需要直接调用） |

## 相关链接

- [Zap 官方文档](https://pkg.go.dev/go.uber.org/zap)
- [Lumberjack 文档](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2)
- [项目主页](https://github.com/qq1060656096/drugo)

## License

本项目使用 [LICENSE](../LICENSE) 许可证。





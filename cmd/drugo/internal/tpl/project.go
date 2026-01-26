// Package tpl contains template strings for code generation.
package tpl

// Project templates - these are embedded as string constants for simplicity.
// In larger projects, consider using embed.FS for external template files.

const MainGoTpl = `package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	//biapi "github.com/qq1060656096/drugo-provider/biapi/api"
	"github.com/qq1060656096/drugo-provider/dbsvc"
	"github.com/qq1060656096/drugo-provider/ginsrv"
	"github.com/qq1060656096/drugo-provider/redissvc"

	"github.com/qq1060656096/drugo/drugo"
	"github.com/qq1060656096/drugo/pkg/gomod"
	"github.com/qq1060656096/drugo/pkg/router"
	"go.uber.org/zap"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	root := gomod.ProjectRoot(wd)
	ctx := context.Background()
	app := drugo.MustNewApp(
		drugo.WithContext(ctx),
		drugo.WithRoot(root),
		drugo.WithService(ginsrv.New()),
		drugo.WithService(dbsvc.New()),
		drugo.WithService(redissvc.New()),
	)
	drugo.SetApp(app)
	//biapi.Init("public", "test_common")
	ginService := drugo.MustGetService[*ginsrv.GinService](app, "gin")
	engine := ginService.Engine()

	// 示例路由
	router.Default().Register(func(r *gin.Engine) {
		r.GET("/health", func(c *gin.Context) {
			app.Logger().MustGet("gin").Info("health", zap.String("url", c.Request.URL.String()))
			c.JSON(200, gin.H{"status": "ok"})
		})
	})
	engine.Use(func(c *gin.Context) {
		c.Set(drugo.Name, app)
		c.Next()
	})
	// 自动注册所有模块路由
	router.Default().Setup(engine)

	err = app.Serve(ctx)
	if err != nil {
		panic(err)
	}
}
`

const GinYamlTpl = `gin:
  mode: release           # debug, release, test
  host: "0.0.0.0"
  shutdown_timeout: 30s   # 优雅关闭超时
  read_timeout: 15s       # 请求读取超时
  write_timeout: 15s      # 响应写入超时
  idle_timeout: 60s       # Keep-Alive 空闲超时
  # HTTP 配置
  http:
    enabled: true
    port: 18001

  # HTTPS 配置
  https:
    enabled: false
    port: 18443
    cert_file: "./cert/server.crt"
    key_file: "./cert/server.key"
    force_ssl: false
`

const LogYamlTpl = `log:
  level: "debug"          # debug, info, warn, error
  format: "json"          # console, json, text
  max_size: 100           # MB
  max_backups: 30
  max_age: 7              # days
  compress: true
  console: true           # 同时输出到控制台
`

const GoModTpl = `module {{.ModPath}}

go 1.25.0

require (
	github.com/gin-gonic/gin v1.11.0
	github.com/qq1060656096/drugo {{.Version}}
	github.com/qq1060656096/drugo-provider v0.0.6
	go.uber.org/zap v1.27.1
)
`

const MakefileTpl = `.PHONY: run build clean mod test help fmt vet

# 应用名称
APP_NAME := {{.Name}}
# 编译输出目录
BUILD_DIR := bin
# 主入口
MAIN_FILE := cmd/app/main.go

# 默认目标
.DEFAULT_GOAL := help

AIR_PKG := github.com/air-verse/air@latest

## run: 运行应用
run:
	@command -v air >/dev/null 2>&1 || { \
    		echo "🔧 air 未安装，正在安装..."; \
    		go install $(AIR_PKG); \
    	}
	air

## build: 编译应用
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

## clean: 清理编译文件和日志
clean:
	@rm -rf $(BUILD_DIR)
	@rm -rf runtime/logs/*.log
	@echo "清理完成"

## mod: 下载依赖
mod:
	go mod download
	go mod tidy

## test: 运行测试
test:
	go test -v -count=1 ./...

## testa: 运行测试（带竞态检测）
testa:
	go test -v -count=1 -race ./...

## fmt: 格式化代码
fmt:
	go fmt ./...

## vet: 静态检查
vet:
	go vet ./...

## help: 显示帮助信息
help:
	@echo "使用方法:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
`

const GitignoreTpl = `# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Coverage
*.out
coverage.*
*.coverprofile
profile.cov

# Go workspace
go.work
go.work.sum

# Env
.env

# Editor/IDE
.idea/
.vscode/

# Logs
*.log

# Runtime
runtime/logs/*.log
`

const DbYamlTpl = `db:
  # =========================
  # 默认数据库组
  # 用途：
  # - 非 SaaS / 单库模式
  # - 简单项目或只有一个业务数据库的场景
  # =========================
  default:
    # 默认数据库实例
    default:
      # 数据库实例标识
      # 用于 DB 注册、日志、监控等（非 db_name）
      name: "default"
      # 支持：mysql、postgres、sqlite、sqlserver 等
      driver_type: "mysql"
      host: "172.16.123.1"
      port: 3306
      user: "root"
      password: "123456"
      # 实际使用的数据库名
      db_name: "sys"
      # 字符集（可选，默认 utf8mb4）
      charset: "utf8mb4"
      # 最大空闲连接数
      max_idle_conns: 10
      # 最大打开连接数
      max_open_conns: 100
      # 连接最大生命周期（秒）
      # 超过该时间的连接会被回收
      conn_max_lifetime: 3600

  # =========================
  # 公共数据库组
  # 用途：
  # - 公共表
  # - 基础数据
  # - 多业务共享数据
  # =========================
  public:
    # 默认公共库实例
    default:
      # 数据库实例标识
      name: "default"
      # 数据库类型
      # 支持：mysql、postgres、sqlite、sqlserver 等
      driver_type: "mysql"
      host: "172.16.123.1"
      port: 3306
      user: "root"
      password: "123456"
      # 实际业务数据库名
      db_name: "test_common"
      charset: "utf8mb4"
      # 最大空闲连接数
      max_idle_conns: 10
      # 最大打开连接数
      max_open_conns: 100
      # 连接最大生命周期（秒）
      conn_max_lifetime: 3600

  # =========================
  # 业务数据库组
  # 用途：
  # - 多业务 / 多租户场景
  # - 每个业务可映射到不同数据库
  # =========================
  business:
    # 默认业务库实例（可扩展为 data_1 / data_2 / data_n）
    default:
      # 数据库实例标识
      name: "default"
      # 支持：mysql、postgres、sqlite、sqlserver 等
      driver_type: "mysql"
      host: "172.16.123.1"
      port: 3306
      user: "root"
      password: "123456"
      # 实际业务数据库名
      db_name: "test_data_1"
      charset: "utf8mb4"
      max_idle_conns: 10
      max_open_conns: 100
      conn_max_lifetime: 3600

`

const RedisYamlTpl = `redis:
  # =========================
  # 默认缓存 Redis 实例
  # 用途：用户登录态、Session、Token 等短生命周期数据
  # =========================
  default:
    # 实例名称（用于注册表 / 日志 / 监控标识）
    name: "default"
    # Redis 部署模式
    # standalone | sentinel | cluster
    mode: "standalone"
    # Redis 地址
    # standalone: host:port
    # sentinel/cluster: 多地址用逗号分隔
    addr: "localhost:6379"
    # Redis 访问密码（无密码留空）
    password: ""
    # 使用的 Redis DB 编号
    # 建议不同业务使用不同 DB 隔离
    db: 0

  # =========================
  # 会话缓存 Redis 实例
  # 用途：用户登录态、Session、Token 等短生命周期数据
  # =========================
  session:
    # 实例名称（用于注册表 / 日志 / 监控标识）
    name: "session"
    # Redis 部署模式
    # standalone | sentinel | cluster
    mode: "standalone"
    # Redis 地址
    # standalone: host:port
    # sentinel/cluster: 多地址用逗号分隔
    addr: "localhost:6379"
    # Redis 访问密码（无密码留空）
    password: ""
    # 使用的 Redis DB 编号
    # 建议不同业务使用不同 DB 隔离
    db: 1


  # =========================
  # 购物车缓存 Redis 实例
  # 用途：购物车、临时订单、用户操作状态
  # 特点：读写频繁、并发高
  # =========================
  cart:
    # 实例名称
    name: "cart"
    # Redis 部署模式
    mode: "standalone"
    # Redis 地址
    addr: "localhost:6379"
    # Redis 访问密码
    password: ""
    # 使用独立 DB，避免与 session 数据混用
    db: 2

`

const ReadmeTpl = `# {{.Name}}

基于 [Drugo](https://github.com/qq1060656096/drugo) 框架的 Go 应用程序。

## 快速开始

### 安装依赖

` + "```bash" + `
go mod tidy
` + "```" + `

### 运行应用

` + "```bash" + `
make run
` + "```" + `

### 编译应用

` + "```bash" + `
make build
` + "```" + `

## 项目结构

` + "```" + `
{{.Name}}/
├── cmd/
│   └── app/
│       └── main.go       # 应用入口
├── conf/
│   ├── gin.yaml          # Gin 服务配置
│   └── log.yaml          # 日志配置
├── internal/             # 内部模块
├── runtime/
│   └── logs/             # 运行时日志
├── go.mod
├── Makefile
└── README.md
` + "```" + `

## 创建新模块

使用 drugo CLI 创建新模块：

` + "```bash" + `
drugo new module <module-name>
` + "```" + `

例如：

` + "```bash" + `
drugo new module user
` + "```" + `

这将在 ` + "`internal/`" + ` 目录下创建标准的 CRUD 模块结构。

## 配置

配置文件位于 ` + "`conf/`" + ` 目录：

- ` + "`gin.yaml`" + ` - HTTP 服务器配置
- ` + "`log.yaml`" + ` - 日志配置

## License

MIT
`
const AirTomlTpl = `# 项目根目录，"." 表示当前目录
root = "."

# Air 编译后的临时文件目录
tmp_dir = "bin"

[build]
  # 🔴 最关键的一行
  # 编译命令
  # -o ./bin/app   → 编译后的二进制文件路径
  # ./cmd/server   → main.go 所在目录（不是文件）
  cmd = "go build -o ./bin/app ./cmd/app"

  # 运行的二进制文件
  bin = "bin/app"

  # 文件变更后，延迟多少毫秒再重启（防止频繁抖动）
  delay = 1000

  # 监听的文件后缀
  # 只要这些文件变化就会触发重启
  include_ext = ["go", "tpl", "tmpl", "html", "yaml", "yml"]

  # 排除监听的目录
  # tmp：Air 输出目录，必须排除
  # vendor：依赖
  # node_modules：前端依赖
  exclude_dir = ["tmp", "vendor", "node_modules"]

  # 编译失败时是否停止运行
  # true = 有编译错误就不重启（推荐）
  stop_on_error = true

[log]
  # 日志是否显示时间
  time = true

[color]
  # Air 各阶段日志颜色（纯视觉效果）
  main = "cyan"      # Air 主进程
  watcher = "yellow" # 文件监听
  build = "green"    # 编译阶段
  runner = "magenta" # 程序运行
`

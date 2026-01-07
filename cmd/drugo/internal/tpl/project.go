// Package tpl contains template strings for code generation.
package tpl

// Project templates - these are embedded as string constants for simplicity.
// In larger projects, consider using embed.FS for external template files.

const MainGoTpl = `package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo/drugo"
	"github.com/qq1060656096/drugo/pkg/gomod"
	"github.com/qq1060656096/drugo/pkg/router"
	"github.com/qq1060656096/drugo/provider/ginsrv"
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
	)
	ginService := drugo.MustGetService[*ginsrv.GinService](app, "gin")
	engine := ginService.Engine()

	// 示例路由
	router.Default().Register(func(r *gin.Engine) {
		r.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		r.GET("/hello", func(c *gin.Context) {
			app.Logger().MustGet("gin").Info("hello world", zap.String("url", c.Request.URL.String()))
			c.JSON(200, gin.H{"message": "hello world"})
		})
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
	github.com/qq1060656096/drugo latest
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

## run: 运行应用
run:
	go run $(MAIN_FILE)

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

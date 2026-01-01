package kernel

import (
	"context"

	"github.com/qq1060656096/drugo/config"
	"github.com/qq1060656096/drugo/log"
)

// Kernel 定义了 Drugo 内核的核心契约
type Kernel interface {
	// Container 返回依赖注入容器
	Container() Container[Service]

	// Boot 引导所有服务完成初始化
	Boot(ctx context.Context) error

	// Run 运行所有实现了 Runner 接口的服务（通常是常驻进程）
	Run(ctx context.Context) error

	// Shutdown 优雅关闭内核及所有服务
	Shutdown(ctx context.Context) error

	// Root 返回应用根目录
	Root() string

	// Config 返回配置管理器
	Config() *config.Manager

	// Logger 返回日志管理器
	Logger() *log.Manager

	// Serve 运行完整的应用生命周期（Boot + Run + 信号监听 + Shutdown）
	// 注意：应用可能不存在任何 Runner 服务，此时 Serve 应当正常返回。
	Serve(ctx context.Context) error
}

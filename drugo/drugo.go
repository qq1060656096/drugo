package drugo

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/qq1060656096/drugo/config"
	"github.com/qq1060656096/drugo/kernel"
	"github.com/qq1060656096/drugo/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// 框架元数据
const (
	Version = "1.0.0"
	Name    = "Drugo"
	logName = "app"
)

var _ kernel.Kernel = (*Drugo)(nil)

// Drugo 是框架的核心引擎结构体
// 它负责管理服务容器、上下文、配置以及日志系统
type Drugo struct {
	container       kernel.Container[kernel.Service]
	root            string
	ctx             context.Context
	config          *config.Manager
	logger          *log.Manager
	shutdownTimeout time.Duration
}

// Container 返回绑定的服务容器
func (d *Drugo) Container() kernel.Container[kernel.Service] {
	return d.container
}

// Context 返回应用程序的主上下文
func (d *Drugo) Context() context.Context {
	return d.ctx
}

// Root 返回项目的根目录路径
func (d *Drugo) Root() string {
	return d.root
}

// Boot 初始化所有已注册的服务
// 按照服务注册的顺序调用它们的 Boot 方法
func (d *Drugo) Boot(ctx context.Context) error {
	services := d.Container().Services()
	l := d.Logger().MustGet(logName)

	l.Info("framework boot start", zap.String("app", Name))
	l.Info("framework boot start services names " + strings.Join(d.serviceNames(), ","))

	if len(services) == 0 {
		l.Warn("no services registered to boot")
		return nil
	}

	ctx = kernel.WithContext(ctx, d)
	for i := range services {
		service := services[i]
		// 动态变量作为 Field 传入，而非拼接字符串
		l.Info("service booting", zap.String("service", service.Name()))

		if err := service.Boot(ctx); err != nil {
			l.Error("service boot failed",
				zap.String("service", service.Name()),
				zap.Error(err),
			)
			return err
		}
	}
	l.Info("framework boot complete")
	return nil
}

// Run 启动所有实现了 kernel.Runner 接口的服务
// 这些服务通常是常驻进程，如 HTTP Server 或消息消费者
func (d *Drugo) Run(ctx context.Context) error {
	services := d.Container().Services()
	l := d.Logger().MustGet(logName)

	l.Info("framework run start")

	if len(services) == 0 {
		l.Warn("no services to run")
		return nil
	}

	runnerCount := 0
	ctx = kernel.WithContext(ctx, d)
	g, ctx := errgroup.WithContext(ctx)

	for i := range services {
		service := services[i]
		runner, ok := service.(kernel.Runner)
		if !ok {
			continue
		}
		runnerCount++

		// 闭包捕获
		r := runner
		s := service
		g.Go(func() error {
			if err := r.Run(ctx); err != nil {
				l.Error("service run failed",
					zap.String("service", s.Name()),
					zap.Error(err),
				)
				return err
			}
			return nil
		})
	}

	if runnerCount < 1 {
		l.Warn("no runner services identified")
	}

	if err := g.Wait(); err != nil {
		l.Error("framework run interrupted by error", zap.Error(err))
		return err
	}

	l.Info("framework run complete")
	return nil
}

// Shutdown 优雅地关闭所有服务
// 会在指定的上下文超时时间内尝试调用所有服务的 Close 方法
func (d *Drugo) Shutdown(ctx context.Context) error {
	services := d.Container().Services()
	l := d.Logger().MustGet(logName)

	l.Info("framework shutdown start")

	if len(services) == 0 {
		return nil
	}

	ctx = kernel.WithContext(ctx, d)
	// 逆序关闭服务
	for i := len(services) - 1; i >= 0; i-- {
		service := services[i]
		l.Info("service shutting down", zap.String("service", service.Name()))

		if err := service.Close(ctx); err != nil {
			l.Error("service shutdown failed",
				zap.String("service", service.Name()),
				zap.Error(err),
			)
			// 继续尝试关闭其他服务，不应立即退出
		}
	}
	l.Info("framework shutdown complete")
	return nil
}

// Serve 是框架的启动入口
// 它封装了 Boot、Run 以及信号监听逻辑，实现了优雅停机
//
// 执行流程：
//  1. Boot
//  2. Run（异步）
//  3. 监听系统信号
//  4. Shutdown（带超时）
func (d *Drugo) Serve(ctx context.Context) error {
	l := d.Logger().MustGet(logName)

	l.Info("app starting",
		zap.String("name", Name),
		zap.String("version", Version),
	)

	if err := d.Boot(ctx); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	errChan := make(chan error, 1)
	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()
	go func() {
		// 无论 Run 成功/失败都要通知主流程（特别是没有 Runner 服务时）
		errChan <- d.Run(runCtx)
	}()

	var runErr error
	select {
	case err := <-errChan:
		// Run 可能立即返回（例如没有 Runner 服务），此时应当进入 Shutdown 并正常退出
		runErr = err
		if runErr != nil {
			l.Error("app exit with error", zap.Error(runErr))
		} else {
			l.Info("app run complete, initiating shutdown")
		}
	case sig := <-quit:
		l.Info("receive signal, initiating graceful shutdown",
			zap.String("signal", sig.String()),
		)
		// 通知所有 Runner 尽快退出
		cancelRun()
	}

	// 优雅停机超时控制
	timeout := d.shutdownTimeout
	if timeout <= 0 {
		timeout = DefaultShutdownTimeout
	}
	l.Info("initiating shutdown with timeout", zap.Duration("timeout", timeout))
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := d.Shutdown(timeoutCtx); err != nil {
		l.Error("app shutdown failed", zap.Error(err))
		// 如果 Run 已经报错，优先返回 Run 的错误；否则返回 Shutdown 错误
		if runErr != nil {
			return runErr
		}
		return err
	}

	l.Info("app exit successfully")
	return runErr
}

// Config 获取配置管理器
func (d *Drugo) Config() *config.Manager {
	return d.config
}

// Logger 获取日志管理器
func (d *Drugo) Logger() *log.Manager {
	return d.logger
}

func (d *Drugo) serviceNames() []string {
	services := d.Container().Services()
	l := len(services)
	result := make([]string, l)
	for i := range services {
		result[i] = services[i].Name()
	}
	return result
}

// MustNewApp 快速创建一个预集成了默认服务（HTTP, Demo）的 Drugo 应用
// 如果初始化失败会 panic
//
// 会自动注册：
//   - Config
//   - Logger
func MustNewApp(opts ...Option) *Drugo {
	app := New(opts...)

	// 初始化配置系统 (默认路径: project_root/conf)
	configDir := filepath.Join(app.Root(), "conf")
	app.config = config.MustNewManager(configDir)

	// 初始化日志系统 (默认路径: project_root/runtime/logs)
	logConfigDir := filepath.Join(app.Root(), "runtime/logs")
	logCfg := log.Config{
		Dir: logConfigDir,
	}

	// 尝试从配置文件加载日志配置
	if logConfig, err := app.Config().Get("log"); err == nil {
		_ = logConfig.Unmarshal(&logCfg)
	}

	var err error
	app.logger, err = log.NewManager(logCfg)
	if err != nil {
		panic(err) // NewApp 不返回 error，配置错误时 panic
	}
	drugoLog := app.Logger().MustGet(logName)
	drugoLog.Info("framework init")
	drugoLog.Info("framework init has service names: " + strings.Join(app.serviceNames(), ", "))
	drugoLog.Info("framework init has config dir: " + configDir)
	drugoLog.Info("framework init has log dir: " + logConfigDir)
	drugoLog.Info("framework init has config biz names: " + strings.Join(app.Config().List(), ", "))

	return app
}

// New 创建一个新的 Drugo 实例
func New(opts ...Option) *Drugo {
	// 1. 初始化默认选项
	o := &options{
		services: make([]map[string]kernel.Service, 0),
		ctx:      context.Background(),
		root:     ".", // 默认根目录为当前目录
	}

	// 2. 应用所有自定义选项
	for _, opt := range opts {
		opt(o)
	}

	// 3. 实例化 Drugo
	app := &Drugo{
		root:            o.root,
		ctx:             o.ctx,
		container:       NewContainer[kernel.Service](),
		shutdownTimeout: o.shutdownTimeout,
	}

	// 4. 将选项中的服务注册到容器中
	for _, serviceMap := range o.services {
		for name, service := range serviceMap {
			app.Container().Bind(name, service)
		}
	}

	return app
}

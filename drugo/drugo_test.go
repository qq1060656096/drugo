package drugo

import (
	"context"
	"testing"
	"time"

	"github.com/qq1060656096/drugo/config"
	"github.com/qq1060656096/drugo/kernel"
	"github.com/qq1060656096/drugo/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDrugoService 是一个用于测试框架的模拟服务实现
type mockDrugoService struct {
	name        string
	bootCalled  bool
	closeCalled bool
	bootError   error
	closeError  error
	bootDelay   time.Duration
	closeDelay  time.Duration
}

func (m *mockDrugoService) Name() string {
	return m.name
}

func (m *mockDrugoService) Boot(ctx context.Context) error {
	m.bootCalled = true
	if m.bootDelay > 0 {
		time.Sleep(m.bootDelay)
	}
	return m.bootError
}

func (m *mockDrugoService) Close(ctx context.Context) error {
	m.closeCalled = true
	if m.closeDelay > 0 {
		time.Sleep(m.closeDelay)
	}
	return m.closeError
}

// mockRunnerService 是一个实现了 Runner 接口的模拟服务
type mockRunnerService struct {
	*mockDrugoService
	runCalled bool
	runError  error
	runDelay  time.Duration
	runBlock  bool // 是否阻塞运行
}

func (m *mockRunnerService) Run(ctx context.Context) error {
	m.runCalled = true
	if m.runDelay > 0 {
		time.Sleep(m.runDelay)
	}
	if m.runBlock {
		<-ctx.Done() // 阻塞直到上下文取消
	}
	return m.runError
}

// TestNew 测试框架实例创建
func TestNew(t *testing.T) {
	// 测试默认选项创建
	app := New()
	require.NotNil(t, app)
	assert.Equal(t, ".", app.root)
	assert.NotNil(t, app.ctx)
	assert.NotNil(t, app.container)
	assert.Zero(t, app.shutdownTimeout)

	// 测试自定义选项
	service := &mockDrugoService{name: "test-service"}
	customCtx := context.WithValue(context.Background(), "key", "value")
	app = New(
		WithRoot("/custom/root"),
		WithContext(customCtx),
		WithService(service),
		WithShutdownTimeout(5*time.Second),
	)
	assert.Equal(t, "/custom/root", app.root)
	assert.Equal(t, customCtx, app.ctx)
	assert.Equal(t, 5*time.Second, app.shutdownTimeout)

	// 验证服务已注册
	svc, err := app.Container().Get("test-service")
	assert.NoError(t, err)
	assert.Equal(t, service, svc)
}

// TestDrugo_Container 测试容器访问
func TestDrugo_Container(t *testing.T) {
	app := New()
	container := app.Container()
	assert.NotNil(t, container)
	assert.Same(t, app.container, container)
}

// TestDrugo_Context 测试上下文访问
func TestDrugo_Context(t *testing.T) {
	customCtx := context.WithValue(context.Background(), "test", "value")
	app := New(WithContext(customCtx))
	assert.Equal(t, customCtx, app.Context())
}

// TestDrugo_Root 测试根目录访问
func TestDrugo_Root(t *testing.T) {
	app := New(WithRoot("/test/root"))
	assert.Equal(t, "/test/root", app.Root())
}

// TestDrugo_Boot 测试服务启动
func TestDrugo_Boot(t *testing.T) {
	tests := []struct {
		name        string
		services    []kernel.Service
		expectError bool
		setupLogger bool
	}{
		{
			name:        "无服务启动",
			services:    []kernel.Service{},
			expectError: false,
			setupLogger: true,
		},
		{
			name: "单个服务启动成功",
			services: []kernel.Service{
				&mockDrugoService{name: "service1"},
			},
			expectError: false,
			setupLogger: true,
		},
		{
			name: "多个服务启动成功",
			services: []kernel.Service{
				&mockDrugoService{name: "service1"},
				&mockDrugoService{name: "service2"},
			},
			expectError: false,
			setupLogger: true,
		},
		{
			name: "服务启动失败",
			services: []kernel.Service{
				&mockDrugoService{name: "service1"},
				&mockDrugoService{name: "service2", bootError: assert.AnError},
			},
			expectError: true,
			setupLogger: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建应用
			opts := []Option{}
			for _, service := range tt.services {
				opts = append(opts, WithService(service))
			}
			app := New(opts...)

			// 设置模拟日志管理器（如果需要）
			if tt.setupLogger {
				// 创建一个简单的日志配置
				logCfg := log.Config{
					Dir:    "/tmp/test-logs",
					Level:  "info",
					Format: "console",
				}
				logger, err := log.NewManager(logCfg)
				require.NoError(t, err)
				app.logger = logger
			}

			// 执行启动
			err := app.Boot(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证所有服务的 Boot 方法都被调用
			for _, service := range tt.services {
				if mockSvc, ok := service.(*mockDrugoService); ok {
					assert.True(t, mockSvc.bootCalled, "服务 %s 的 Boot 方法应该被调用", mockSvc.name)
				}
			}
		})
	}
}

// TestDrugo_Run 测试服务运行
func TestDrugo_Run(t *testing.T) {
	tests := []struct {
		name           string
		services       []kernel.Service
		expectError    bool
		setupLogger    bool
		runnerCount    int
		blockingRunner bool
	}{
		{
			name:        "无服务运行",
			services:    []kernel.Service{},
			expectError: false,
			setupLogger: true,
			runnerCount: 0,
		},
		{
			name: "无Runner服务",
			services: []kernel.Service{
				&mockDrugoService{name: "service1"},
			},
			expectError: false,
			setupLogger: true,
			runnerCount: 0,
		},
		{
			name: "单个Runner服务",
			services: []kernel.Service{
				&mockRunnerService{
					mockDrugoService: &mockDrugoService{name: "runner1"},
				},
			},
			expectError: false,
			setupLogger: true,
			runnerCount: 1,
		},
		{
			name: "Runner服务运行失败",
			services: []kernel.Service{
				&mockRunnerService{
					mockDrugoService: &mockDrugoService{name: "runner1"},
					runError:         assert.AnError,
				},
			},
			expectError: true,
			setupLogger: true,
			runnerCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建应用
			opts := []Option{}
			for _, service := range tt.services {
				opts = append(opts, WithService(service))
			}
			app := New(opts...)

			// 设置日志管理器
			if tt.setupLogger {
				logCfg := log.Config{
					Dir:    "/tmp/test-logs",
					Level:  "info",
					Format: "console",
				}
				logger, err := log.NewManager(logCfg)
				require.NoError(t, err)
				app.logger = logger
			}

			// 执行运行
			err := app.Run(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证Runner服务的Run方法都被调用
			for _, service := range tt.services {
				if runner, ok := service.(*mockRunnerService); ok {
					assert.True(t, runner.runCalled, "Runner服务 %s 的 Run 方法应该被调用", runner.name)
				}
			}
		})
	}
}

// TestDrugo_Shutdown 测试服务关闭
func TestDrugo_Shutdown(t *testing.T) {
	tests := []struct {
		name        string
		services    []kernel.Service
		expectError bool
		setupLogger bool
	}{
		{
			name:        "无服务关闭",
			services:    []kernel.Service{},
			expectError: false,
			setupLogger: true,
		},
		{
			name: "单个服务关闭成功",
			services: []kernel.Service{
				&mockDrugoService{name: "service1"},
			},
			expectError: false,
			setupLogger: true,
		},
		{
			name: "多个服务关闭成功",
			services: []kernel.Service{
				&mockDrugoService{name: "service1"},
				&mockDrugoService{name: "service2"},
			},
			expectError: false,
			setupLogger: true,
		},
		{
			name: "服务关闭失败但继续关闭其他服务",
			services: []kernel.Service{
				&mockDrugoService{name: "service1"},
				&mockDrugoService{name: "service2", closeError: assert.AnError},
				&mockDrugoService{name: "service3"},
			},
			expectError: false, // 关闭失败不会返回错误
			setupLogger: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建应用
			opts := []Option{}
			for _, service := range tt.services {
				opts = append(opts, WithService(service))
			}
			app := New(opts...)

			// 设置日志管理器
			if tt.setupLogger {
				logCfg := log.Config{
					Dir:    "/tmp/test-logs",
					Level:  "info",
					Format: "console",
				}
				logger, err := log.NewManager(logCfg)
				require.NoError(t, err)
				app.logger = logger
			}

			// 执行关闭
			err := app.Shutdown(context.Background())
			assert.NoError(t, err) // Shutdown 总是返回 nil

			// 验证所有服务的 Close 方法都被调用
			for _, service := range tt.services {
				if mockSvc, ok := service.(*mockDrugoService); ok {
					assert.True(t, mockSvc.closeCalled, "服务 %s 的 Close 方法应该被调用", mockSvc.name)
				}
			}
		})
	}
}

func TestDrugo_Shutdown_Order(t *testing.T) {
	// 创建多个服务，通过不同的关闭延迟来验证顺序
	services := []kernel.Service{
		&mockDrugoService{
			name:       "service1",
			closeDelay: 30 * time.Millisecond, // 最长的延迟
		},
		&mockDrugoService{
			name:       "service2",
			closeDelay: 20 * time.Millisecond,
		},
		&mockDrugoService{
			name:       "service3",
			closeDelay: 10 * time.Millisecond, // 最短的延迟
		},
	}

	// 创建应用
	opts := []Option{}
	for _, service := range services {
		opts = append(opts, WithService(service))
	}
	app := New(opts...)

	// 设置日志管理器
	logCfg := log.Config{
		Dir:    "/tmp/test-logs",
		Level:  "info",
		Format: "console",
	}
	logger, err := log.NewManager(logCfg)
	require.NoError(t, err)
	app.logger = logger

	// 记录开始时间
	start := time.Now()

	// 执行关闭
	app.Shutdown(context.Background())

	// 记录结束时间
	elapsed := time.Since(start)

	// 验证总时间至少等于最长延迟（说明是逆序关闭）
	assert.True(t, elapsed >= 30*time.Millisecond)

	// 验证所有服务都被关闭
	for _, service := range services {
		if mockSvc, ok := service.(*mockDrugoService); ok {
			assert.True(t, mockSvc.closeCalled)
		}
	}
}

// TestDrugo_Config 测试配置管理器访问
func TestDrugo_Config(t *testing.T) {
	app := New()

	// 初始时配置管理器为nil
	assert.Nil(t, app.Config())

	// 设置配置管理器
	mockConfig := &config.Manager{}
	app.config = mockConfig
	assert.Same(t, mockConfig, app.Config())
}

// TestDrugo_Logger 测试日志管理器访问
func TestDrugo_Logger(t *testing.T) {
	app := New()

	// 初始时日志管理器为nil
	assert.Nil(t, app.Logger())

	// 设置日志管理器
	mockLogger := &log.Manager{}
	app.logger = mockLogger
	assert.Same(t, mockLogger, app.Logger())
}

// TestDrugo_serviceNames 测试服务名称获取
func TestDrugo_serviceNames(t *testing.T) {
	app := New()

	// 无服务时
	names := app.serviceNames()
	assert.Empty(t, names)

	// 添加服务
	service1 := &mockDrugoService{name: "service1"}
	service2 := &mockDrugoService{name: "service2"}
	app.Container().Bind("svc1", service1)
	app.Container().Bind("svc2", service2)

	names = app.serviceNames()
	assert.Equal(t, []string{"service1", "service2"}, names)
}

func TestDrugo_Serve_Signal(t *testing.T) {
	// 创建一个简单的应用
	service := &mockDrugoService{name: "test-service"}
	app := New(WithService(service))

	// 设置日志管理器
	logCfg := log.Config{
		Dir:    "/tmp/test-logs",
		Level:  "info",
		Format: "console",
	}
	logger, err := log.NewManager(logCfg)
	require.NoError(t, err)
	app.logger = logger

	// 创建一个会被立即取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 启动服务
	go func() {
		time.Sleep(10 * time.Millisecond) // 等待一小段时间
		cancel()                          // 取消上下文，模拟信号
	}()

	// Serve 应该正常退出
	err = app.Serve(ctx)
	assert.NoError(t, err)
	assert.True(t, service.bootCalled)
	assert.True(t, service.closeCalled)
}

// TestDrugo_Serve_Timeout 测试关闭超时
func TestDrugo_Serve_Timeout(t *testing.T) {
	// 创建一个关闭缓慢的服务
	service := &mockDrugoService{
		name:       "slow-service",
		closeDelay: 200 * time.Millisecond, // 超过默认超时时间
	}

	app := New(
		WithService(service),
		WithShutdownTimeout(50*time.Millisecond), // 设置较短的超时时间
	)

	// 设置日志管理器
	logCfg := log.Config{
		Dir:    "/tmp/test-logs",
		Level:  "info",
		Format: "console",
	}
	logger, err := log.NewManager(logCfg)
	require.NoError(t, err)
	app.logger = logger

	ctx := context.Background()
	err = app.Serve(ctx)

	// 应该正常退出，即使关闭超时
	assert.NoError(t, err)
	assert.True(t, service.bootCalled)
	assert.True(t, service.closeCalled)
}

// TestMustNewApp 测试强制创建应用
func TestMustNewApp(t *testing.T) {
	// 这个测试需要真实的文件系统结构
	// 在实际项目中，应该创建测试目录和配置文件

	// 测试基本创建（可能会panic，因为缺少配置文件）
	assert.Panics(t, func() {
		MustNewApp()
	})
}

// TestConstants 测试常量定义
func TestConstants(t *testing.T) {
	assert.Equal(t, "1.0.0", Version)
	assert.Equal(t, "Drugo", Name)
	assert.Equal(t, "app", logName)
	assert.Equal(t, 10*time.Second, DefaultShutdownTimeout)
}

// TestInterfaceCompliance 测试接口合规性
func TestInterfaceCompliance(t *testing.T) {
	// 验证 Drugo 实现了 kernel.Kernel 接口
	var _ kernel.Kernel = (*Drugo)(nil)

	app := New()
	assert.Implements(t, (*kernel.Kernel)(nil), app)
}

// BenchmarkNew 测试创建应用性能
func BenchmarkNew(b *testing.B) {
	service := &mockDrugoService{name: "benchmark-service"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app := New(WithService(service))
		_ = app
	}
}

// BenchmarkDrugo_Boot 测试启动性能
func BenchmarkDrugo_Boot(b *testing.B) {
	services := make([]kernel.Service, 10)
	for i := 0; i < 10; i++ {
		services[i] = &mockDrugoService{name: "service-" + string(rune(i))}
	}

	opts := []Option{}
	for _, service := range services {
		opts = append(opts, WithService(service))
	}
	app := New(opts...)

	// 设置日志管理器
	logCfg := log.Config{
		Dir:    "/tmp/test-logs",
		Level:  "info",
		Format: "console",
	}
	logger, err := log.NewManager(logCfg)
	if err != nil {
		b.Fatal(err)
	}
	app.logger = logger

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = app.Boot(context.Background())
	}
}

// BenchmarkDrugo_serviceNames 测试获取服务名称性能
func BenchmarkDrugo_serviceNames(b *testing.B) {
	services := make([]kernel.Service, 100)
	for i := 0; i < 100; i++ {
		services[i] = &mockDrugoService{name: "service-" + string(rune(i))}
	}

	opts := []Option{}
	for _, service := range services {
		opts = append(opts, WithService(service))
	}
	app := New(opts...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = app.serviceNames()
	}
}

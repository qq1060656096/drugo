package kernel

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/qq1060656096/drugo/config"
	"github.com/qq1060656096/drugo/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockService 是一个用于测试的模拟服务实现
type MockService struct {
	name     string
	booted   bool
	closed   bool
	bootErr  error
	closeErr error
}

// NewMockService 创建一个新的模拟服务
func NewMockService(name string) *MockService {
	return &MockService{name: name}
}

// Name 实现 Service 接口
func (m *MockService) Name() string {
	return m.name
}

// Boot 实现 Service 接口
func (m *MockService) Boot(ctx context.Context) error {
	if m.bootErr != nil {
		return m.bootErr
	}
	m.booted = true
	return nil
}

// Close 实现 Service 接口
func (m *MockService) Close(ctx context.Context) error {
	m.closed = true // 无论是否成功，都标记为已关闭
	if m.closeErr != nil {
		return m.closeErr
	}
	return nil
}

// IsBooted 检查服务是否已启动
func (m *MockService) IsBooted() bool {
	return m.booted
}

// IsClosed 检查服务是否已关闭
func (m *MockService) IsClosed() bool {
	return m.closed
}

// SetBootError 设置启动时的错误
func (m *MockService) SetBootError(err error) {
	m.bootErr = err
}

// SetCloseError 设置关闭时的错误
func (m *MockService) SetCloseError(err error) {
	m.closeErr = err
}

// MockRunner 是一个用于测试的模拟运行器实现
type MockRunner struct {
	*MockService
	runErr   error
	runCount int
}

// NewMockRunner 创建一个新的模拟运行器
func NewMockRunner(name string) *MockRunner {
	return &MockRunner{
		MockService: NewMockService(name),
	}
}

// Run 实现 Runner 接口
func (m *MockRunner) Run(ctx context.Context) error {
	m.runCount++
	if m.runErr != nil {
		return m.runErr
	}
	
	// 模拟长期运行的服务，直到上下文取消
	<-ctx.Done()
	return ctx.Err()
}

// RunCount 返回 Run 方法被调用的次数
func (m *MockRunner) RunCount() int {
	return m.runCount
}

// SetRunError 设置运行时的错误
func (m *MockRunner) SetRunError(err error) {
	m.runErr = err
}

// MockContainer 是一个用于测试的模拟容器实现
type MockContainer struct {
	services map[string]Service
	getErr   map[string]error
}

// NewMockContainer 创建一个新的模拟容器
func NewMockContainer() *MockContainer {
	return &MockContainer{
		services: make(map[string]Service),
		getErr:   make(map[string]error),
	}
}

// Bind 实现 Container 接口
func (m *MockContainer) Bind(name string, service Service) {
	m.services[name] = service
}

// Get 实现 Container 接口
func (m *MockContainer) Get(name string) (Service, error) {
	if err, exists := m.getErr[name]; exists {
		return nil, err
	}
	
	if svc, exists := m.services[name]; exists {
		return svc, nil
	}
	
	return nil, NewServiceNotFound(name)
}

// MustGet 实现 Container 接口
func (m *MockContainer) MustGet(name string) Service {
	svc, err := m.Get(name)
	if err != nil {
		panic(err)
	}
	return svc
}

// Services 实现 Container 接口
func (m *MockContainer) Services() []Service {
	services := make([]Service, 0, len(m.services))
	for _, svc := range m.services {
		services = append(services, svc)
	}
	return services
}

// Names 实现 Container 接口
func (m *MockContainer) Names() []string {
	names := make([]string, 0, len(m.services))
	for name := range m.services {
		names = append(names, name)
	}
	return names
}

// SetGetError 为指定服务设置获取错误
func (m *MockContainer) SetGetError(name string, err error) {
	m.getErr[name] = err
}

// MockKernel 是一个用于测试的模拟内核实现
type MockKernel struct {
	container *MockContainer
}

// NewMockKernel 创建一个新的模拟内核
func NewMockKernel() *MockKernel {
	return &MockKernel{
		container: NewMockContainer(),
	}
}

// Container 实现 Kernel 接口
func (m *MockKernel) Container() Container[Service] {
	return m.container
}

// Boot 实现 Kernel 接口
func (m *MockKernel) Boot(ctx context.Context) error {
	return nil
}

// Run 实现 Kernel 接口
func (m *MockKernel) Run(ctx context.Context) error {
	return nil
}

// Shutdown 实现 Kernel 接口
func (m *MockKernel) Shutdown(ctx context.Context) error {
	return nil
}

// Root 实现 Kernel 接口
func (m *MockKernel) Root() string {
	return "/mock/root"
}

// Config 实现 Kernel 接口
func (m *MockKernel) Config() *config.Manager {
	return nil
}

// Logger 实现 Kernel 接口
func (m *MockKernel) Logger() *log.Manager {
	return nil
}

// Serve 实现 Kernel 接口
func (m *MockKernel) Serve(ctx context.Context) error {
	return nil
}

// GetMockContainer 获取模拟容器的引用，用于测试设置
func (m *MockKernel) GetMockContainer() *MockContainer {
	return m.container
}

// TestInterfaceImplementation 测试接口实现的正确性
func TestInterfaceImplementation(t *testing.T) {
	t.Run("MockService 实现 Service 接口", func(t *testing.T) {
		var _ Service = (*MockService)(nil)
		svc := NewMockService("test")
		assert.Equal(t, "test", svc.Name())
	})

	t.Run("MockRunner 实现 Runner 接口", func(t *testing.T) {
		var _ Runner = (*MockRunner)(nil)
		runner := NewMockRunner("test-runner")
		assert.Equal(t, "test-runner", runner.Name())
	})

	t.Run("MockContainer 实现 Container 接口", func(t *testing.T) {
		var _ Container[Service] = (*MockContainer)(nil)
		container := NewMockContainer()
		assert.NotNil(t, container)
	})

	t.Run("MockKernel 实现 Kernel 接口", func(t *testing.T) {
		var _ Kernel = (*MockKernel)(nil)
		kernel := NewMockKernel()
		assert.NotNil(t, kernel)
	})
}

// TestGetService 测试 GetService 函数
func TestGetService(t *testing.T) {
	t.Run("成功获取正确类型的服务", func(t *testing.T) {
		kernel := NewMockKernel()
		container := kernel.GetMockContainer()
		
		// 注册一个 MockService
		mockSvc := NewMockService("test-service")
		container.Bind("test-service", mockSvc)
		
		// 获取服务
		svc, err := GetService[*MockService](kernel, "test-service")
		require.NoError(t, err)
		assert.Equal(t, mockSvc, svc)
		assert.Equal(t, "test-service", svc.Name())
	})

	t.Run("服务不存在", func(t *testing.T) {
		kernel := NewMockKernel()
		
		// 尝试获取不存在的服务
		svc, err := GetService[*MockService](kernel, "non-existent")
		assert.Error(t, err)
		assert.True(t, IsServiceNotFound(err))
		var zero *MockService
		assert.Equal(t, zero, svc)
	})

	t.Run("服务类型不匹配", func(t *testing.T) {
		kernel := NewMockKernel()
		container := kernel.GetMockContainer()
		
		// 注册一个 MockService
		mockSvc := NewMockService("test-service")
		container.Bind("test-service", mockSvc)
		
		// 尝试获取为不同的类型
		svc, err := GetService[*MockRunner](kernel, "test-service")
		assert.Error(t, err)
		assert.True(t, IsServiceType(err))
		assert.Contains(t, err.Error(), "service test-service is not of type")
		var zero *MockRunner
		assert.Equal(t, zero, svc)
	})

	t.Run("容器获取错误", func(t *testing.T) {
		kernel := NewMockKernel()
		container := kernel.GetMockContainer()
		
		// 设置获取错误
		testErr := errors.New("container error")
		container.SetGetError("error-service", testErr)
		
		// 尝试获取服务
		svc, err := GetService[*MockService](kernel, "error-service")
		assert.Error(t, err)
		assert.Equal(t, testErr, err)
		var zero *MockService
		assert.Equal(t, zero, svc)
	})

	t.Run("获取 Runner 类型服务", func(t *testing.T) {
		kernel := NewMockKernel()
		container := kernel.GetMockContainer()
		
		// 注册一个 MockRunner
		mockRunner := NewMockRunner("test-runner")
		container.Bind("test-runner", mockRunner)
		
		// 获取为 Service 类型
		svc, err := GetService[Service](kernel, "test-runner")
		require.NoError(t, err)
		assert.Equal(t, mockRunner, svc)
		assert.Equal(t, "test-runner", svc.Name())
		
		// 获取为 Runner 类型
		runner, err := GetService[Runner](kernel, "test-runner")
		require.NoError(t, err)
		assert.Equal(t, mockRunner, runner)
	})

	t.Run("获取为接口类型", func(t *testing.T) {
		kernel := NewMockKernel()
		container := kernel.GetMockContainer()
		
		// 注册一个 MockService
		mockSvc := NewMockService("interface-service")
		container.Bind("interface-service", mockSvc)
		
		// 获取为 Booter 接口
		booter, err := GetService[Booter](kernel, "interface-service")
		require.NoError(t, err)
		assert.Equal(t, mockSvc, booter)
		
		// 获取为 Closer 接口
		closer, err := GetService[Closer](kernel, "interface-service")
		require.NoError(t, err)
		assert.Equal(t, mockSvc, closer)
	})
}

// TestGetService_DifferentTypes 测试不同类型的获取
func TestGetService_DifferentTypes(t *testing.T) {
	kernel := NewMockKernel()
	container := kernel.GetMockContainer()
	
	// 注册不同类型的服务
	mockSvc := NewMockService("service")
	mockRunner := NewMockRunner("runner")
	
	container.Bind("service", mockSvc)
	container.Bind("runner", mockRunner)
	
	t.Run("获取具体类型", func(t *testing.T) {
		svc, err := GetService[*MockService](kernel, "service")
		require.NoError(t, err)
		assert.Equal(t, mockSvc, svc)
		
		runner, err := GetService[*MockRunner](kernel, "runner")
		require.NoError(t, err)
		assert.Equal(t, mockRunner, runner)
	})
	
	t.Run("获取接口类型", func(t *testing.T) {
		svc, err := GetService[Service](kernel, "service")
		require.NoError(t, err)
		assert.Equal(t, mockSvc, svc)
		
		runner, err := GetService[Runner](kernel, "runner")
		require.NoError(t, err)
		assert.Equal(t, mockRunner, runner)
	})
	
	t.Run("类型不匹配", func(t *testing.T) {
		// 尝试将 Service 获取为 Runner
		_, err := GetService[Runner](kernel, "service")
		assert.Error(t, err)
		assert.True(t, IsServiceType(err))
		
		// 尝试将 Runner 获取为不匹配的具体类型
		_, err = GetService[*MockService](kernel, "runner")
		assert.Error(t, err)
		assert.True(t, IsServiceType(err))
	})
}

// TestMustGetService 测试 MustGetService 函数
func TestMustGetService(t *testing.T) {
	t.Run("成功获取服务", func(t *testing.T) {
		kernel := NewMockKernel()
		container := kernel.GetMockContainer()
		
		// 注册一个 MockService
		mockSvc := NewMockService("must-service")
		container.Bind("must-service", mockSvc)
		
		// 获取服务
		svc := MustGetService[*MockService](kernel, "must-service")
		assert.Equal(t, mockSvc, svc)
		assert.Equal(t, "must-service", svc.Name())
	})

	t.Run("服务不存在时 panic", func(t *testing.T) {
		kernel := NewMockKernel()
		
		// 验证 panic
		assert.Panics(t, func() {
			MustGetService[*MockService](kernel, "non-existent")
		})
	})

	t.Run("类型不匹配时 panic", func(t *testing.T) {
		kernel := NewMockKernel()
		container := kernel.GetMockContainer()
		
		// 注册一个 MockService
		mockSvc := NewMockService("type-mismatch")
		container.Bind("type-mismatch", mockSvc)
		
		// 验证 panic
		assert.Panics(t, func() {
			MustGetService[*MockRunner](kernel, "type-mismatch")
		})
	})

	t.Run("容器错误时 panic", func(t *testing.T) {
		kernel := NewMockKernel()
		container := kernel.GetMockContainer()
		
		// 设置获取错误
		testErr := errors.New("container error")
		container.SetGetError("panic-service", testErr)
		
		// 验证 panic
		assert.Panics(t, func() {
			MustGetService[*MockService](kernel, "panic-service")
		})
	})
}

// TestServiceLifecycle 测试服务生命周期
func TestServiceLifecycle(t *testing.T) {
	t.Run("服务启动和关闭", func(t *testing.T) {
		svc := NewMockService("lifecycle-service")
		
		// 初始状态
		assert.False(t, svc.IsBooted())
		assert.False(t, svc.IsClosed())
		
		// 启动服务
		ctx := context.Background()
		err := svc.Boot(ctx)
		require.NoError(t, err)
		assert.True(t, svc.IsBooted())
		assert.False(t, svc.IsClosed())
		
		// 关闭服务
		err = svc.Close(ctx)
		require.NoError(t, err)
		assert.True(t, svc.IsBooted())
		assert.True(t, svc.IsClosed())
	})

	t.Run("服务启动失败", func(t *testing.T) {
		svc := NewMockService("fail-boot")
		bootErr := errors.New("boot failed")
		svc.SetBootError(bootErr)
		
		ctx := context.Background()
		err := svc.Boot(ctx)
		assert.Error(t, err)
		assert.Equal(t, bootErr, err)
		assert.False(t, svc.IsBooted())
	})

	t.Run("服务关闭失败", func(t *testing.T) {
		svc := NewMockService("fail-close")
		closeErr := errors.New("close failed")
		svc.SetCloseError(closeErr)
		
		ctx := context.Background()
		err := svc.Close(ctx)
		assert.Error(t, err)
		assert.Equal(t, closeErr, err)
		assert.True(t, svc.IsClosed()) // 即使失败，状态也应该更新
	})
}

// TestRunnerLifecycle 测试运行器生命周期
func TestRunnerLifecycle(t *testing.T) {
	t.Run("运行器正常执行", func(t *testing.T) {
		runner := NewMockRunner("test-runner")
		
		// 初始状态
		assert.Equal(t, 0, runner.RunCount())
		
		// 启动运行器
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		
		done := make(chan error, 1)
		go func() {
			done <- runner.Run(ctx)
		}()
		
		// 等待超时
		select {
		case err := <-done:
			assert.Error(t, err)
			assert.Equal(t, context.DeadlineExceeded, err)
		case <-time.After(200 * time.Millisecond):
			t.Fatal("运行器应该在超时后返回")
		}
		
		assert.Equal(t, 1, runner.RunCount())
	})

	t.Run("运行器执行失败", func(t *testing.T) {
		runner := NewMockRunner("fail-runner")
		runErr := errors.New("run failed")
		runner.SetRunError(runErr)
		
		ctx := context.Background()
		err := runner.Run(ctx)
		assert.Error(t, err)
		assert.Equal(t, runErr, err)
		assert.Equal(t, 1, runner.RunCount())
	})
}

// TestContainerOperations 测试容器操作
func TestContainerOperations(t *testing.T) {
	container := NewMockContainer()
	
	t.Run("绑定和获取服务", func(t *testing.T) {
		svc1 := NewMockService("service1")
		svc2 := NewMockService("service2")
		
		// 绑定服务
		container.Bind("service1", svc1)
		container.Bind("service2", svc2)
		
		// 获取服务
		retrieved1, err := container.Get("service1")
		require.NoError(t, err)
		assert.Equal(t, svc1, retrieved1)
		
		retrieved2, err := container.Get("service2")
		require.NoError(t, err)
		assert.Equal(t, svc2, retrieved2)
	})

	t.Run("覆盖已存在的服务", func(t *testing.T) {
		svc1 := NewMockService("original")
		svc2 := NewMockService("replacement")
		
		// 绑定第一个服务
		container.Bind("test", svc1)
		retrieved, err := container.Get("test")
		require.NoError(t, err)
		assert.Equal(t, svc1, retrieved)
		
		// 覆盖服务
		container.Bind("test", svc2)
		retrieved, err = container.Get("test")
		require.NoError(t, err)
		assert.Equal(t, svc2, retrieved)
	})

	t.Run("MustGet 成功", func(t *testing.T) {
		svc := NewMockService("must-get")
		container.Bind("must-get", svc)
		
		retrieved := container.MustGet("must-get")
		assert.Equal(t, svc, retrieved)
	})

	t.Run("MustGet 失败时 panic", func(t *testing.T) {
		assert.Panics(t, func() {
			container.MustGet("non-existent")
		})
	})

	t.Run("Services 和 Names 方法", func(t *testing.T) {
		// 清空容器
		container = NewMockContainer()
		
		svc1 := NewMockService("svc1")
		svc2 := NewMockService("svc2")
		svc3 := NewMockService("svc3")
		
		container.Bind("svc1", svc1)
		container.Bind("svc2", svc2)
		container.Bind("svc3", svc3)
		
		services := container.Services()
		names := container.Names()
		
		assert.Len(t, services, 3)
		assert.Len(t, names, 3)
		
		// 验证包含所有名称
		nameSet := make(map[string]bool)
		for _, name := range names {
			nameSet[name] = true
		}
		assert.True(t, nameSet["svc1"])
		assert.True(t, nameSet["svc2"])
		assert.True(t, nameSet["svc3"])
	})
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	t.Run("服务类型错误消息格式", func(t *testing.T) {
		kernel := NewMockKernel()
		container := kernel.GetMockContainer()
		
		svc := NewMockService("test-service")
		container.Bind("test-service", svc)
		
		_, err := GetService[*MockRunner](kernel, "test-service")
		assert.Error(t, err)
		assert.True(t, IsServiceType(err))
		assert.Contains(t, err.Error(), "service test-service is not of type")
	})

	t.Run("服务未找到错误", func(t *testing.T) {
		kernel := NewMockKernel()
		
		_, err := GetService[*MockService](kernel, "missing")
		assert.Error(t, err)
		assert.True(t, IsServiceNotFound(err))
	})
}

// BenchmarkGetService 性能测试
func BenchmarkGetService(b *testing.B) {
	kernel := NewMockKernel()
	container := kernel.GetMockContainer()
	
	// 预注册一些服务
	for i := 0; i < 100; i++ {
		svc := NewMockService(fmt.Sprintf("service-%d", i))
		container.Bind(fmt.Sprintf("service-%d", i), svc)
	}
	
	b.ResetTimer()
	
	b.Run("正常获取", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = GetService[*MockService](kernel, "service-50")
		}
	})
	
	b.Run("类型不匹配", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = GetService[*MockRunner](kernel, "service-50")
		}
	})
}

// BenchmarkMustGetService 性能测试
func BenchmarkMustGetService(b *testing.B) {
	kernel := NewMockKernel()
	container := kernel.GetMockContainer()
	
	svc := NewMockService("benchmark-service")
	container.Bind("benchmark-service", svc)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = MustGetService[*MockService](kernel, "benchmark-service")
	}
}
